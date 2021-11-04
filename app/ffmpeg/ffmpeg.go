package ffmpeg

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/frame"
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
	"time"
	"unsafe"
)

const MaxVideoBuffers = 10
const MaxAudioBuffers = 2000

var cmdVideo *exec.Cmd
var cmdAudio *exec.Cmd

var videoPipe io.WriteCloser

var videoQueue chan func()

var pboSync *sync.Mutex
var pboPool = make([]*PBO, 0)
var syncPool = make([]*PBO, 0)

var endSyncVideo *sync.WaitGroup

var audioPipe io.WriteCloser

var audioQueue chan []byte

var audioSync *sync.Mutex
var audioPool = make([][]byte, 0)

var endSyncAudio *sync.WaitGroup

var blend *effects.Blend

var w, h int

var limiter *frame.Limiter

var output string

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

// check used encoders exist
func preCheck() {
	out, err := exec.Command("ffmpeg", "-encoders").Output()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			panic("ffmpeg not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg")
		}

		panic(err)
	}

	encoders := strings.Split(string(out[:]), "\n")
	for i, v := range encoders {
		if strings.TrimSpace(v) == "------" {
			encoders = encoders[i+1 : len(encoders)-1]
			break
		}
	}

	vcodec := settings.Recording.Encoder
	acodec := settings.Recording.AudioCodec
	vfound := false
	afound := false

	for _, v := range encoders {
		encoder := strings.SplitN(strings.TrimSpace(v), " ", 3)
		codecType := string(encoder[0][0])

		if string(encoder[0][3]) == "X" {
			continue // experimental codec
		}

		if !vfound && codecType == "V" {
			vfound = encoder[1] == vcodec
		} else if !afound && codecType == "A" {
			afound = encoder[1] == acodec
		}
	}

	if !vfound {
		panic(fmt.Sprintf("Video codec %q does not exist", vcodec))
	}

	if !afound {
		panic(fmt.Sprintf("Audio codec %q does not exist", acodec))
	}
}

func StartFFmpeg(fps, _w, _h int, audioFPS float64, _output string) {
	if strings.TrimSpace(_output) == "" {
		_output = "danser_" + time.Now().Format("2006-01-02_15-04-05")
	}

	output = _output

	w, h = _w, _h

	preCheck()

	log.Println("Starting encoding!")

	_ = os.RemoveAll(filepath.Join(settings.Recording.OutputDir, output+"_temp"))

	err := os.MkdirAll(filepath.Join(settings.Recording.OutputDir, output+"_temp"), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	startVideo(fps)
	startAudio(audioFPS)

	startThreads()
}

func startVideo(fps int) {
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
		"-profile:v", settings.Recording.Profile,
		"-preset", settings.Recording.Preset,
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

	encOptions := strings.TrimSpace(settings.Recording.EncoderOptions)
	if encOptions != "" {
		split := strings.Split(encOptions, " ")
		options = append(options, split...)
	}

	options = append(options, filepath.Join(settings.Recording.OutputDir, output+"_temp", "video."+settings.Recording.Container))

	log.Println("Running ffmpeg with options:", options)

	cmdVideo = exec.Command("ffmpeg", options...)

	var err error

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
		panic(err)
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
}

func startAudio(audioFPS float64) {
	inputName := "-"

	if runtime.GOOS != "windows" {
		pipe, err := files.NewNamedPipe("")
		if err != nil {
			panic(err)
		}

		inputName = pipe.Name()
		audioPipe = pipe
	}

	options := []string{
		"-y",

		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", "48000",
		"-ac", "2",
		"-i", inputName,//audioPipe.Name(),

		"-nostats", //hide audio encoding statistics because video ones are more important
		"-vn",
	}

	audioFilters := strings.TrimSpace(settings.Recording.AudioFilters)
	if len(audioFilters) > 0 {
		options = append(options, "-af", audioFilters)
	}

	options = append(options, "-c:a", settings.Recording.AudioCodec)

	encOptions := strings.TrimSpace(settings.Recording.AudioOptions)
	if encOptions != "" {
		split := strings.Split(encOptions, " ")
		options = append(options, split...)
	}

	options = append(options, filepath.Join(settings.Recording.OutputDir, output+"_temp", "audio."+settings.Recording.Container))

	log.Println("Running ffmpeg with options:", options)

	cmdAudio = exec.Command("ffmpeg", options...)

	var err error

	if runtime.GOOS == "windows" {
		audioPipe, err = cmdAudio.StdinPipe()
		if err != nil {
			panic(err)
		}
	}

	if settings.Recording.ShowFFmpegLogs {
		cmdAudio.Stdout = os.Stdout
		cmdAudio.Stderr = os.Stderr
	}

	err = cmdAudio.Start()
	if err != nil {
		panic(err)
	}

	audioBufSize := bass.GetMixerRequiredBufferSize(1 / audioFPS)

	for i := 0; i < MaxAudioBuffers; i++ {
		audioPool = append(audioPool, make([]byte, audioBufSize))
	}

	audioSync = &sync.Mutex{}

	audioQueue = make(chan []byte, MaxAudioBuffers)
}

func startThreads() {
	endSyncVideo = &sync.WaitGroup{}
	endSyncAudio = &sync.WaitGroup{}

	endSyncVideo.Add(1)
	endSyncAudio.Add(1)

	go func() {
		runtime.LockOSThread()

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
	}()

	go func() {
		runtime.LockOSThread()

		for {
			data, keepOpen := <-audioQueue

			if data != nil {
				_, err := audioPipe.Write(data)
				if err != nil {
					panic(err)
				}

				audioSync.Lock()

				audioPool = append(audioPool, data)

				audioSync.Unlock()
			}

			if !keepOpen {
				endSyncAudio.Done()
				break
			}
		}
	}()
}

func StopFFmpeg() {
	log.Println("Finishing rendering...")

	for len(syncPool) > 0 {
		CheckData()
	}

	log.Println("Finished! Stopping video videoPipe...")

	close(videoQueue)
	endSyncVideo.Wait()
	_ = videoPipe.Close()

	log.Println("Finished! Stopping audio videoPipe...")

	close(audioQueue)
	endSyncAudio.Wait()
	_ = audioPipe.Close()

	log.Println("Pipes closed.")

	_ = cmdVideo.Wait()
	_ = cmdAudio.Wait()

	log.Println("Ffmpeg finished.")

	combine()
}

func combine() {
	options := []string{
		"-y",
		"-i", filepath.Join(settings.Recording.OutputDir, output+"_temp", "video."+settings.Recording.Container),
		"-i", filepath.Join(settings.Recording.OutputDir, output+"_temp", "audio."+settings.Recording.Container),
		"-c:v", "copy",
		"-c:a", "copy",
	}

	if settings.Recording.Container == "mp4" {
		options = append(options, "-movflags", "+faststart")
	}

	options = append(options, filepath.Join(settings.Recording.OutputDir, output+"."+settings.Recording.Container))

	log.Println("Starting composing audio and video into one file...")
	log.Println("Running ffmpeg with options:", options)
	cmd2 := exec.Command("ffmpeg", options...)

	if settings.Recording.ShowFFmpegLogs {
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
	}

	if err := cmd2.Start(); err != nil {
		log.Println("Failed to start ffmpeg:", err)
	} else {
		if err = cmd2.Wait(); err != nil {
			log.Println("ffmpeg finished abruptly! Please check if you have enough storage or audio bitrate is entered correctly.")
		} else {
			log.Println("Finished!")
		}
	}

	cleanup()
}

func cleanup() {
	log.Println("Cleaning up intermediate files...")

	_ = os.RemoveAll(filepath.Join(settings.Recording.OutputDir, output+"_temp"))

	log.Println("Finished.")
}

func PushAudio() {
	audioSync.Lock()

	//spin until at least one audio buffer is free
	for len(audioPool) == 0 {}

	data := audioPool[0]
	audioPool = audioPool[1:]

	audioSync.Unlock()

	bass.ProcessMixer(data)

	audioQueue <- data
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
					panic(err)
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
