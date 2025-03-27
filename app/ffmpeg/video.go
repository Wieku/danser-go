package ffmpeg

import (
	"bufio"
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/util/pixconv"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

const MaxVideoBuffers = 10

var cmdVideo *exec.Cmd

var videoPipe io.WriteCloser

var videoWriteQueue chan *PBO
var endSyncVideo *sync.WaitGroup

var videoError string
var videoErrorWait *sync.WaitGroup

var freePBOPool chan *PBO

var frameReadQueue = make([]*PBO, 0)

var blend *effects.Blend

var w, h int

var limiter *frame.Limiter

var parsedFormat pixconv.PixFmt

type PBO struct {
	handle     uint32
	memPointer unsafe.Pointer
	data       []byte

	convFormat pixconv.PixFmt

	sync uintptr
}

func createPBO(format pixconv.PixFmt) *PBO {
	pbo := new(PBO)
	pbo.convFormat = format

	glSize := w * h * 3

	if pbo.convFormat == pixconv.I420 || pbo.convFormat == pixconv.NV12 {
		glSize = w * h * 3 / 2
	}

	gl.CreateBuffers(1, &pbo.handle)
	gl.NamedBufferStorage(pbo.handle, glSize, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_COHERENT_BIT|gl.MAP_READ_BIT)

	pbo.memPointer = gl.MapNamedBufferRange(pbo.handle, 0, glSize, gl.MAP_PERSISTENT_BIT|gl.MAP_COHERENT_BIT|gl.MAP_READ_BIT)

	pbo.data = (*[1 << 30]byte)(pbo.memPointer)[:glSize:glSize]

	return pbo
}

var rgbToYuvConverter *effects.RGBYUV

func startVideo(fps, _w, _h int) {
	w, h = _w, _h

	if settings.Recording.MotionBlur.Enabled {
		fps /= settings.Recording.MotionBlur.OversampleMultiplier
	}

	encoder := strings.ToLower(settings.Recording.Encoder)
	outputFormat := strings.ToLower(settings.Recording.PixelFormat)

	if strings.HasSuffix(encoder, "_qsv") { // qsv works best with nv12 format
		outputFormat = "nv12"
	} else if encoder == "libsvtav1" {
		outputFormat = "yuv420p"
	}

	parsedFormat = pixconv.ARGB

	switch outputFormat {
	case "yuv420p":
		parsedFormat = pixconv.I420
	case "yuv444p":
		parsedFormat = pixconv.I444
	case "nv12":
		parsedFormat = pixconv.NV12
	}

	var filters []string

	inputPixFmt := "rgb24"
	if parsedFormat != pixconv.ARGB {
		inputPixFmt = outputFormat
	} else {
		filters = append(filters, "vflip")
	}

	videoFilters := strings.TrimSpace(settings.Recording.Filters)
	if len(videoFilters) > 0 {
		filters = append(filters, videoFilters)
	}

	inputName := "-"

	if runtime.GOOS != "windows" {
		pipe, err := files.NewNamedPipe("")
		if err != nil {
			panic(err)
		}

		inputName = pipe.Name()
		videoPipe = pipe
	}

	options := []string{
		"-y",  //(optional) overwrite output file if it exists
		"-an", // no audio
		"-f", "rawvideo",
		"-c:v", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", w, h), //size of one frame
		"-pix_fmt", inputPixFmt,
		"-r", strconv.Itoa(fps), //frames per second
	}

	if inputPixFmt != "rgb24" {
		options = append(options,
			"-color_range", "1",
			"-colorspace", "1",
			"-color_trc", "1",
			"-color_primaries", "1",
		)
	}

	options = append(options,
		"-i", inputName, //The input comes from a videoPipe
	)

	if len(filters) > 0 {
		options = append(options, "-vf", strings.Join(filters, ","))
	}

	options = append(options,
		"-c:v", encoder,
		"-color_range", "1",
		"-colorspace", "1",
		"-color_trc", "1",
		"-color_primaries", "1",
		"-movflags", "+write_colr",
	)

	if parsedFormat == pixconv.ARGB {
		options = append(options, "-pix_fmt", outputFormat)
	}

	encOptions, err := settings.Recording.GetEncoderOptions().GenerateFFmpegArgs()
	if err != nil {
		panic(fmt.Sprintf("encoder \"%s\": %s", encoder, err))
	} else if encOptions != nil {
		options = append(options, encOptions...)
	}

	options = append(options, filepath.Join(settings.Recording.GetOutputDir(), output+"_temp", "video."+settings.Recording.Container))

	log.Println("Running ffmpeg with options:", options)

	cmdVideo = exec.Command(ffmpegExec, options...)

	if runtime.GOOS == "windows" {
		videoPipe, err = cmdVideo.StdinPipe()
		if err != nil {
			panic(err)
		}
	}

	rFile, oFile, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	outList := []io.Writer{oFile}
	errList := []io.Writer{oFile}

	if settings.Recording.ShowFFmpegLogs {
		outList = append(outList, os.Stdout)
		errList = append(errList, os.Stderr)
	}

	cmdVideo.Stdout = io.MultiWriter(outList...)
	cmdVideo.Stderr = io.MultiWriter(errList...)

	err = cmdVideo.Start()
	if err != nil {
		panic(fmt.Sprintf("ffmpeg's video process failed to start! Please check if video parameters are entered correctly or video codec is supported by provided container. Error: %s", err))
	}

	freePBOPool = make(chan *PBO, MaxVideoBuffers)

	goroutines.CallMain(func() {
		if parsedFormat != pixconv.ARGB {
			rgbToYuvConverter = effects.NewRGBYUV(w, h, parsedFormat != pixconv.I444 && parsedFormat != pixconv.I422)
		}

		for i := 0; i < MaxVideoBuffers; i++ {
			freePBOPool <- createPBO(parsedFormat)
		}

		if settings.Recording.MotionBlur.Enabled {
			bFrames := settings.Recording.MotionBlur.BlendFrames
			blend = effects.NewBlend(w, h, bFrames, calculateWeights(bFrames))
		}
	})

	videoWriteQueue = make(chan *PBO, MaxVideoBuffers)

	limiter = frame.NewLimiter(settings.Recording.EncodingFPSCap)

	videoErrorWait = &sync.WaitGroup{}
	videoErrorWait.Add(1)

	goroutines.Run(func() {
		sc := bufio.NewScanner(rFile)

		for sc.Scan() {
			line := sc.Text()

			cutIndex := strings.Index(line, "] ") //searching for encoder error

			if cutIndex > -1 {
				cutLine := line[cutIndex+2:]
				lineLower := strings.ToLower(cutLine)

				if strings.Contains(lineLower, "error setting") ||
					strings.Contains(lineLower, "error initializing") ||
					strings.Contains(lineLower, "error creating") ||
					strings.Contains(lineLower, "invalid") ||
					strings.Contains(lineLower, "incompatible") ||
					strings.Contains(lineLower, "not divisible") ||
					strings.Contains(lineLower, "exceeds") ||
					strings.Contains(lineLower, "failed") ||
					strings.Contains(lineLower, "no capable devices found") ||
					strings.Contains(lineLower, "does not support") {

					videoError = encoder + ": " + cutLine

					oFile.Close()
				}
			}
		}

		videoErrorWait.Done()
	})

	endSyncVideo = &sync.WaitGroup{}
	endSyncVideo.Add(1)

	goroutines.RunOS(func() {
		for pbo := range videoWriteQueue {
			if _, err2 := videoPipe.Write(pbo.data); err2 != nil {
				errorMsg := err2.Error()

				videoErrorWait.Wait()

				if videoError != "" {
					errorMsg = videoError
				}

				panic(fmt.Sprintf("ffmpeg's video process finished abruptly! Please check if you have enough storage or video parameters are entered correctly. Error: %s", errorMsg))
			}

			freePBOPool <- pbo
		}

		endSyncVideo.Done()
	})
}

func stopVideo() {
	log.Println("Waiting for video to finish writing...")

	checkData(true, true)

	close(videoWriteQueue)

	endSyncVideo.Wait()

	log.Println("Finished! Stopping video pipe...")

	_ = videoPipe.Close()

	log.Println("Video pipe closed. Waiting for video ffmpeg process to finish...")

	_ = cmdVideo.Wait()

	log.Println("Video process finished.")
}

func PreFrame() {
	if settings.Recording.MotionBlur.Enabled {
		blend.Begin()
	} else if rgbToYuvConverter != nil {
		rgbToYuvConverter.Begin()
	}
}

var frameNumber = int64(-1)

func MakeFrame() {
	frameNumber++

	if settings.Recording.MotionBlur.Enabled {
		blend.End()

		if frameNumber%int64(settings.Recording.MotionBlur.OversampleMultiplier) != 0 {
			return
		}

		if rgbToYuvConverter != nil {
			rgbToYuvConverter.Begin()
		}

		blend.Blend()
	}

	var yuvFull, yuvHalf []texture.Texture

	if rgbToYuvConverter != nil {
		rgbToYuvConverter.End()

		yuvFull, yuvHalf = rgbToYuvConverter.Draw()
	}

	checkData(len(freePBOPool) == 0, false) // Force wait for at least one frame to be retrieved if pbo pool is empty

	pbo := <-freePBOPool // Wait for free PBO

	//gl.MemoryBarrier(gl.PIXEL_BUFFER_BARRIER_BIT)

	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, pbo.handle)

	gl.PixelStorei(gl.PACK_ALIGNMENT, 1)

	if pbo.convFormat == pixconv.NV12 {
		gl.GetTextureSubImage(yuvFull[0].GetID(), 0, 0, 0, 0, int32(w), int32(h), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h), gl.Ptr(nil))

		gl.GetTextureSubImage(yuvHalf[0].GetID(), 0, 0, 0, 0, int32(w/2), int32(h/2), 1, gl.RG, gl.UNSIGNED_BYTE, int32(w*h/2), gl.PtrOffset(w*h))
	} else if pbo.convFormat == pixconv.I420 {
		gl.GetTextureSubImage(yuvFull[0].GetID(), 0, 0, 0, 0, int32(w), int32(h), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h), gl.Ptr(nil))

		gl.GetTextureSubImage(yuvHalf[0].GetID(), 0, 0, 0, 0, int32(w/2), int32(h/2), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h/4), gl.PtrOffset(w*h))
		gl.GetTextureSubImage(yuvHalf[1].GetID(), 0, 0, 0, 0, int32(w/2), int32(h/2), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h/4), gl.PtrOffset(w*h*5/4))
	} else if pbo.convFormat != pixconv.ARGB { //Read as yuv444p
		gl.GetTextureSubImage(yuvFull[0].GetID(), 0, 0, 0, 0, int32(w), int32(h), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h), gl.Ptr(nil))
		gl.GetTextureSubImage(yuvFull[1].GetID(), 0, 0, 0, 0, int32(w), int32(h), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h), gl.PtrOffset(w*h))
		gl.GetTextureSubImage(yuvFull[2].GetID(), 0, 0, 0, 0, int32(w), int32(h), 1, gl.RED, gl.UNSIGNED_BYTE, int32(w*h), gl.PtrOffset(w*h*2))
	} else {
		gl.ReadPixels(0, 0, int32(w), int32(h), uint32(gl.RGB), gl.UNSIGNED_BYTE, gl.Ptr(nil))
	}

	pbo.sync = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)

	gl.Flush()

	frameReadQueue = append(frameReadQueue, pbo)

	checkData(false, false)

	limiter.Sync()
}

func checkData(waitForFirst, waitForAll bool) { // I tried to do that on another thread, but it needs another opengl context and creates other funky problems
	for i := 0; len(frameReadQueue) > 0; i++ {
		pbo := frameReadQueue[0]

		status := int32(gl.SIGNALED)

		if (i == 0 && waitForFirst) || waitForAll {
			for {
				iStat := gl.ClientWaitSync(pbo.sync, 0, gl.TIMEOUT_IGNORED)

				if iStat == gl.ALREADY_SIGNALED || iStat == gl.CONDITION_SATISFIED {
					break
				}
			}
		} else {
			gl.GetSynciv(pbo.sync, gl.SYNC_STATUS, 1, nil, &status)
		}

		if status != gl.SIGNALED {
			return
		}

		gl.DeleteSync(pbo.sync)

		frameReadQueue = frameReadQueue[1:]

		videoWriteQueue <- pbo
	}
}
