package ffmpeg

import (
	"encoding/hex"
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

const MaxBuffers = 10

var filename string

var cmd *exec.Cmd
var pipe io.WriteCloser

var queue chan func()
var endSync *sync.WaitGroup

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

var pboSync *sync.RWMutex
var pboPool = make([]*PBO, 0)

var syncPool = make([]*PBO, 0)

var blend *effects.Blend

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
		codec_type := string(encoder[0][0])

		if string(encoder[0][3]) == "X" {
			continue // experimental codec
		}

		if !vfound && codec_type == "V" {
			vfound = encoder[1] == vcodec
		} else if !afound && codec_type == "A" {
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

	w, h = _w, _h

	err := os.MkdirAll(settings.Recording.OutputDir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	b := make([]byte, 16)
	rand.Read(b)

	filename = hex.EncodeToString(b)

	split := strings.Split(settings.Recording.EncoderOptions, " ")

	filters := strings.TrimSpace(settings.Recording.Filters)
	if len(filters) > 0 {
		filters = "," + filters
	}

	if settings.Recording.MotionBlur.Enabled {
		fps /= settings.Recording.MotionBlur.OversampleMultiplier
	}

	options := []string{
		"-y", //(optional) overwrite output file if it exists
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", w, h), //size of one frame
		"-pix_fmt", "rgb24",
		"-r", strconv.Itoa(fps), //frames per second
		"-i", "-", //The input comes from a pipe
		"-vf", "vflip" + filters,
		"-profile:v", settings.Recording.Profile,
		"-preset", settings.Recording.Preset,
		"-an", //Tells FFMPEG not to expect any audio
		"-vcodec", settings.Recording.Encoder,
		"-color_range", "1",
		"-colorspace", "1",
		"-color_trc", "1",
		"-color_primaries", "1",
		"-movflags", "+write_colr",
		"-pix_fmt", settings.Recording.PixelFormat,
	}

	options = append(options, split...)
	options = append(options, filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container))

	log.Println("Running ffmpeg with options:", options)

	cmd = exec.Command("ffmpeg", options...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	pipe, err = cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	mainthread.Call(func() {
		for i := 0; i < MaxBuffers; i++ {
			pboPool = append(pboPool, createPBO())
		}

		if settings.Recording.MotionBlur.Enabled {
			bFrames := settings.Recording.MotionBlur.BlendFrames
			blend = effects.NewBlend(w, h, bFrames, calculateWeights(bFrames))
		}
	})

	pboSync = &sync.RWMutex{}

	queue = make(chan func(), MaxBuffers)

	endSync = &sync.WaitGroup{}

	endSync.Add(1)

	go func() {
		for {
			f, keepOpen := <-queue

			if f != nil {
				f()
			}

			if !keepOpen {
				endSync.Done()
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

	close(queue)
	endSync.Wait()

	log.Println("Finished! Stopping ffmpeg...")

	pipe.Close()

	log.Println("Pipe closed.")

	cmd.Wait()

	log.Println("Ffmpeg finished.")
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

	pboSync.RLock()

	pbo := pboPool[0]
	pboPool = pboPool[1:]

	pboSync.RUnlock()

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

			queue <- func() {
				_, err := pipe.Write(pbo.data)
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