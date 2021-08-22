package ffmpeg

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/util"
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
const MaxAudioBuffers = 2000

var filename string

var cmd *exec.Cmd

var videoPipe *files.NamedPipe

var videoQueue chan func()

var pboSync *sync.Mutex
var pboPool = make([]*PBO, 0)
var syncPool = make([]*PBO, 0)

var endSyncVideo *sync.WaitGroup

var audioPipe *files.NamedPipe

var audioQueue chan []byte

var audioSync *sync.Mutex
var audioPool = make([][]byte, 0)

var endSyncAudio *sync.WaitGroup

var blend *effects.Blend

var w, h int

type PBO struct {
	handle     uint32
	memPointer unsafe.Pointer
	data       []byte

	sync uintptr
}

func createPBO() *PBO {
	pbo := new(PBO)

	gl.CreateBuffers(1, &pbo.handle)
	gl.NamedBufferStorage(pbo.handle, w*h*3, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_READ_BIT)

	pbo.memPointer = gl.MapNamedBufferRange(pbo.handle, 0, w*h*3, gl.MAP_PERSISTENT_BIT|gl.MAP_READ_BIT)

	pbo.data = (*[1 << 30]byte)(pbo.memPointer)[: w*h*3 : w*h*3]

	return pbo
}

// check used encoders exist
func precheck() {
	out, err := exec.Command("ffmpeg", "-encoders").Output()
	if err != nil {
		panic(err)
	}

	encoders := strings.Split(string(out[:]), "\n")
	for i, v := range encoders {
		if strings.TrimSpace(v) == "------" {
			encoders = encoders[i+1:len(encoders)-1]
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

func StartFFmpeg(fps, _w, _h int) {
	precheck()

	log.Println("Starting encoding!")

	w, h = _w, _h

	err := os.MkdirAll(settings.Recording.OutputDir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	filename = util.RandomHexString(32)

	split := strings.Split(settings.Recording.EncoderOptions, " ")

	videoFilters := strings.TrimSpace(settings.Recording.Filters)
	if len(videoFilters) > 0 {
		videoFilters = "," + videoFilters
	}

	if settings.Recording.MotionBlur.Enabled {
		fps /= settings.Recording.MotionBlur.OversampleMultiplier
	}

	videoPipe, err = files.NewNamedPipe("")
	if err != nil {
		panic(err)
	}

	audioPipe, err = files.NewNamedPipe("")
	if err != nil {
		panic(err)
	}

	options := []string{
		"-y", //(optional) overwrite output file if it exists
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", w, h), //size of one frame
		"-pix_fmt", "rgb24",
		"-r", strconv.Itoa(fps), //frames per second
		"-i", videoPipe.Name(), //The input comes from a videoPipe


		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", "48000",
		"-ac", "2",
		"-probesize", "32",
		"-i", audioPipe.Name(),


		"-vf", "vflip" + videoFilters,
		"-profile:v", settings.Recording.Profile,
		"-preset", settings.Recording.Preset,
		"-vcodec", settings.Recording.Encoder,
		"-color_range", "1",
		"-colorspace", "1",
		"-color_trc", "1",
		"-color_primaries", "1",
		//"-movflags", "+write_colr",
		"-pix_fmt", settings.Recording.PixelFormat,
	}

	options = append(options, split...)

	audioFilters := strings.TrimSpace(settings.Recording.AudioFilters)
	if len(audioFilters) > 0 {
		options = append(options, "-af", audioFilters)
	}

	options = append(options,
		"-c:a", settings.Recording.AudioCodec,
		"-ab", settings.Recording.AudioBitrate,
	)

	movFlags := "+write_colr"

	if settings.Recording.Container == "mp4" {
		//options = append(options, "-movflags", "+faststart")
		movFlags += "+faststart"
	}

	options = append(options, "-movflags", movFlags)

	options = append(options, filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container))

	log.Println("Running ffmpeg with options:", options)

	cmd = exec.Command("ffmpeg", options...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	mainthread.Call(func() {
		for i := 0; i < MaxVideoBuffers; i++ {
			pboPool = append(pboPool, createPBO())
		}

		if settings.Recording.MotionBlur.Enabled {
			bFrames := settings.Recording.MotionBlur.BlendFrames
			blend = effects.NewBlend(w, h, bFrames, calculateWeights(bFrames))
		}
	})

	audioBufSize := 48000*4*2/1000

	for i := 0; i < MaxAudioBuffers; i++ {
		audioPool = append(audioPool, make([]byte, audioBufSize))
	}

	pboSync = &sync.Mutex{}
	audioSync = &sync.Mutex{}

	videoQueue = make(chan func(), MaxVideoBuffers)
	audioQueue = make(chan []byte, MaxAudioBuffers)

	startThreads()
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
	videoPipe.Close()

	log.Println("Finished! Stopping audio videoPipe...")

	close(audioQueue)
	endSyncAudio.Wait()
	audioPipe.Close()

	log.Println("Pipes closed.")

	cmd.Wait()

	log.Println("Ffmpeg finished.")
}

func PushAudio() {
	audioSync.Lock()

	data := audioPool[0]
	audioPool = audioPool[1:]

	audioSync.Unlock()

	bass.EncodePartD(data)

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
	gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(nil))

	pbo.sync = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)

	gl.Flush()

	syncPool = append(syncPool, pbo)
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
				_, err := videoPipe.Write(pbo.data)
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

func GetFileName() string {
	return filename
}