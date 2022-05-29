package ffmpeg

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/effects"
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

var videoQueue chan func()

var pboSync *sync.Mutex
var pboPool = make([]*PBO, 0)
var syncPool = make([]*PBO, 0)

var endSyncVideo *sync.WaitGroup

var blend *effects.Blend

var w, h int

var limiter *frame.Limiter

type PBO struct {
	handle     uint32
	memPointer unsafe.Pointer
	data       []byte

	convFormat pixconv.PixFmt
	convData   []byte

	sync uintptr
}

func createPBO(format pixconv.PixFmt) *PBO {
	pbo := new(PBO)
	pbo.convFormat = format

	channels := 3
	if pbo.convFormat != pixconv.ARGB {
		channels = 4

		convSize := pixconv.GetRequiredBufferSize(pbo.convFormat, w, h)
		pbo.convData = make([]byte, convSize)
	}

	glSize := w * h * channels

	gl.CreateBuffers(1, &pbo.handle)
	gl.NamedBufferStorage(pbo.handle, glSize, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_READ_BIT)

	pbo.memPointer = gl.MapNamedBufferRange(pbo.handle, 0, glSize, gl.MAP_PERSISTENT_BIT|gl.MAP_READ_BIT)

	pbo.data = (*[1 << 30]byte)(pbo.memPointer)[:glSize:glSize]

	return pbo
}

func startVideo(fps, _w, _h int) {
	w, h = _w, _h

	if settings.Recording.MotionBlur.Enabled {
		fps /= settings.Recording.MotionBlur.OversampleMultiplier
	}

	parsedFormat := pixconv.ARGB

	switch strings.ToLower(settings.Recording.PixelFormat) {
	case "yuv420p":
		parsedFormat = pixconv.I420
	case "yuv422p":
		parsedFormat = pixconv.I422
	case "yuv444p":
		parsedFormat = pixconv.I444
	case "nv12":
		parsedFormat = pixconv.NV12
	case "nv21":
		parsedFormat = pixconv.NV21
	}

	inputPixFmt := "rgb24"
	if parsedFormat != pixconv.ARGB {
		inputPixFmt = strings.ToLower(settings.Recording.PixelFormat)
	}

	videoFilters := strings.TrimSpace(settings.Recording.Filters)
	if len(videoFilters) > 0 {
		videoFilters = "," + videoFilters
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
		"-y", //(optional) overwrite output file if it exists

		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", w, h), //size of one frame
		"-pix_fmt", inputPixFmt,
		"-r", strconv.Itoa(fps), //frames per second
		"-i", inputName, //The input comes from a videoPipe

		"-an",

		"-vf", "vflip" + videoFilters,
		"-c:v", settings.Recording.Encoder,
		"-color_range", "1",
		"-colorspace", "1",
		"-color_trc", "1",
		"-color_primaries", "1",
		"-movflags", "+write_colr",
	}

	if parsedFormat == pixconv.ARGB {
		options = append(options, "-pix_fmt", strings.ToLower(settings.Recording.PixelFormat))
	}

	encOptions, err := settings.Recording.GetEncoderOptions().GenerateFFmpegArgs()
	if err != nil {
		panic(fmt.Sprintf("encoder \"%s\": %s", settings.Recording.Encoder, err))
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

	if settings.Recording.ShowFFmpegLogs {
		cmdVideo.Stdout = os.Stdout
		cmdVideo.Stderr = os.Stderr
	}

	err = cmdVideo.Start()
	if err != nil {
		panic(fmt.Sprintf("ffmpeg's video process failed to start! Please check if video parameters are entered correctly or video codec is supported by provided container. Error: %s", err))
	}

	mainthread.Call(func() {
		for i := 0; i < MaxVideoBuffers; i++ {
			pboPool = append(pboPool, createPBO(parsedFormat))
		}

		if settings.Recording.MotionBlur.Enabled {
			bFrames := settings.Recording.MotionBlur.BlendFrames
			blend = effects.NewBlend(w, h, bFrames, calculateWeights(bFrames))
		}
	})

	pboSync = &sync.Mutex{}

	videoQueue = make(chan func(), MaxVideoBuffers)

	limiter = frame.NewLimiter(settings.Recording.EncodingFPSCap)

	endSyncVideo = &sync.WaitGroup{}

	endSyncVideo.Add(1)

	goroutines.RunOS(func() {
		for {
			f, keepOpen := <-videoQueue

			if f != nil {
				f()
			}

			if !keepOpen {
				endSyncVideo.Done()
				break
			}
		}
	})
}

func stopVideo() {
	log.Println("Waiting for video to finish writing...")

	for len(syncPool) > 0 {
		CheckData()
	}

	log.Println("Finished! Stopping video pipe...")

	close(videoQueue)
	endSyncVideo.Wait()
	_ = videoPipe.Close()

	log.Println("Video pipe closed. Waiting for video ffmpeg process to finish...")

	_ = cmdVideo.Wait()

	log.Println("Video process finished.")
}

func PreFrame() {
	if settings.Recording.MotionBlur.Enabled {
		blend.Begin()
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

		blend.Blend()
	}

	//spin until at least one pbo is free
	for len(pboPool) == 0 {
		CheckData()
	}

	pboSync.Lock()

	pbo := pboPool[0]
	pboPool = pboPool[1:]

	pboSync.Unlock()

	gl.MemoryBarrier(gl.PIXEL_BUFFER_BARRIER_BIT)

	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, pbo.handle)

	gl.PixelStorei(gl.PACK_ALIGNMENT, 1)

	glFmt := gl.RGB
	if pbo.convFormat != pixconv.ARGB {
		glFmt = gl.BGRA
	}

	gl.ReadPixels(0, 0, int32(w), int32(h), uint32(glFmt), gl.UNSIGNED_BYTE, gl.Ptr(nil))

	pbo.sync = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)

	gl.Flush()

	syncPool = append(syncPool, pbo)

	limiter.Sync()
}

func CheckData() {
	for {
		if len(syncPool) == 0 {
			return
		}

		pbo := syncPool[0]

		var status int32

		gl.GetSynciv(pbo.sync, gl.SYNC_STATUS, 1, nil, &status)

		if status == gl.SIGNALED {
			gl.DeleteSync(pbo.sync)

			syncPool = syncPool[1:]

			videoQueue <- func() {
				var err error

				if pbo.convFormat != pixconv.ARGB {
					pixconv.Convert(pbo.data, pixconv.ARGB, pbo.convData, pbo.convFormat, w, h)
					_, err = videoPipe.Write(pbo.convData)
				} else {
					_, err = videoPipe.Write(pbo.data)
				}

				if err != nil {
					panic(fmt.Sprintf("ffmpeg's video process finished abruptly! Please check if you have enough storage or video parameters are entered correctly. Error: %s", err))
				}

				pboSync.Lock()
				pboPool = append(pboPool, pbo)
				pboSync.Unlock()
			}

			continue
		}

		return
	}
}
