package ffmpeg

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/app/settings"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const MaxBuffers = 10

var cmd *exec.Cmd
var pipe io.WriteCloser

var queue chan func()
var endSync *sync.WaitGroup

var w, h int

type PBO struct {
	handle     uint32
	memPointer unsafe.Pointer
	data       []byte
}

func createPBO() *PBO {
	pbo := new(PBO)

	gl.CreateBuffers(1, &pbo.handle)
	gl.NamedBufferStorage(pbo.handle, w*h*3, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_READ_BIT)

	pbo.memPointer = gl.MapNamedBufferRange(pbo.handle, 0, w*h*3, gl.MAP_PERSISTENT_BIT|gl.MAP_READ_BIT)

	pbo.data = (*[1 << 30]byte)(pbo.memPointer)[: w*h*3 : w*h*3]

	return pbo
}

type syncObj struct {
	sync uintptr
	pbo  *PBO
}

var pboSync *sync.Mutex

var syncPool = make([]syncObj, 0)

var pboPool = make([]*PBO, 0)

func StartFFmpeg(fps, _w, _h int) {
	w = _w
	h = _h

	err := os.MkdirAll(settings.Recording.OutputDir, 0655)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	split := strings.Split(settings.Recording.EncoderOptions, " ")

	filters := settings.Recording.Filters
	if len(filters) > 0 {
		filters = "," + filters
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
	}

	options = append(options, append(split, "-pix_fmt", settings.Recording.PixelFormat, filepath.Join(settings.Recording.OutputDir, "video."+settings.Recording.Container))...)

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
	})

	pboSync = &sync.Mutex{}

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
	for len(syncPool) > 0 {
		CheckData()
	}

	close(queue)
	endSync.Wait()
	pipe.Close()
	cmd.Wait()
}

func Combine() {
	log.Println("Starting composing audio and video into one file...")
	cmd2 := exec.Command("ffmpeg",
		"-y",
		"-i", filepath.Join(settings.Recording.OutputDir, "video."+settings.Recording.Container),
		"-i", filepath.Join(settings.Recording.OutputDir, "audio.wav"),
		"-c:v", "copy",
		"-c:a", "aac",
		"-ab", "320k",
		filepath.Join(settings.Recording.OutputDir, "danser_"+time.Now().Format("2006-01-02_15-04-05")+"."+settings.Recording.Container),
	)
	cmd2.Start()
	cmd2.Wait()

	log.Println("Finished! Cleaning up...")

	os.Remove(filepath.Join(settings.Recording.OutputDir, "video."+settings.Recording.Container))
	//os.Remove(filepath.Join(settings.Recording.OutputDir,"audio.wav"))

	log.Println("Finished.")
}

func MakeFrame() {
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

	gl.Flush()

	fSync := gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)

	syncPool = append(syncPool, syncObj{
		sync: fSync,
		pbo:  pbo,
	})
}

func CheckData() {
	for {
		if len(syncPool) == 0 {
			return
		}

		peekVal := syncPool[0]

		var status int32
		gl.GetSynciv(peekVal.sync, gl.SYNC_STATUS, 1, nil, &status)

		if status == gl.SIGNALED {
			gl.DeleteSync(peekVal.sync)

			syncPool = syncPool[1:]

			queue <- func() {
				_, err := pipe.Write(peekVal.pbo.data)
				if err != nil {
					panic(err)
				}

				pboSync.Lock()
				pboPool = append(pboPool, peekVal.pbo)
				pboSync.Unlock()
			}

			continue
		}

		return
	}
}
