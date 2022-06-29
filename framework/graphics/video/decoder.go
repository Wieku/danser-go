package video

import (
	"fmt"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/util/pixconv"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
)

const BufferSize = 3

type Frame []byte

type VideoDecoder struct {
	filePath string

	Metadata *Metadata

	decodingQueue chan Frame
	readyQueue    chan Frame

	ffmpegExec string
	command    *exec.Cmd
	pipe       io.ReadCloser
	running    bool
	wg         *sync.WaitGroup
	finished   bool
	format     pixconv.PixFmt
}

func NewVideoDecoder(filePath string) *VideoDecoder {
	metadata := LoadMetadata(filePath)
	if metadata == nil {
		return nil
	}

	ffmpegExec, err := files.GetCommandExec("ffmpeg", "ffmpeg")
	if err != nil {
		log.Println("ffprobe not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg")
	}

	decoder := &VideoDecoder{
		ffmpegExec: ffmpegExec,
		filePath:   filePath,
		Metadata:   metadata,
		wg:         &sync.WaitGroup{},
		format:     pixconv.RGB,
	}

	switch strings.ToLower(metadata.PixFmt) {
	case "yuv420p":
		decoder.format = pixconv.I420
	case "yuv422p":
		decoder.format = pixconv.I422
	case "yuv444p":
		decoder.format = pixconv.I444
	case "nv12":
		decoder.format = pixconv.NV12
	case "nv21":
		decoder.format = pixconv.NV21
	}

	return decoder
}

func (dec *VideoDecoder) StartFFmpeg(millis int64) {
	if dec.running {
		dec.running = false
		close(dec.decodingQueue)

		dec.wg.Wait()

		if dec.command != nil {
			err := dec.command.Process.Kill()
			if err != nil {
				panic(err)
			}
		}
	}

	dec.wg.Add(1)

	var args []string

	if millis != 0 {
		args = append(args,
			"-ss", fmt.Sprintf("%d.%d", millis/1000, millis%1000),
		)
	}

	args = append(args,
		"-i", dec.filePath,
		"-f", "rawvideo",
	)

	var readBuffer []byte

	if dec.format == pixconv.ARGB {
		args = append(args, "-pix_fmt", "rgb24")
	} else {
		args = append(args, "-pix_fmt", dec.Metadata.PixFmt)
		readBuffer = make([]byte, pixconv.GetRequiredBufferSize(dec.format, dec.Metadata.Width, dec.Metadata.Height))
	}

	args = append(args, "-")

	dec.command = exec.Command(
		dec.ffmpegExec,
		args...,
	)

	var err error
	dec.pipe, err = dec.command.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = dec.command.Start()
	if err != nil {
		if strings.Contains(err.Error(), "127") || strings.Contains(strings.ToLower(err.Error()), "0xc0000135") {
			panic(fmt.Sprintf("ffmpeg was installed incorrectly! Please make sure needed libraries (libs/*.so or bin/*.dll) are installed as well. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg. Error: %s", err))
		}

		panic(fmt.Sprintf("Failed to start ffmpeg decoder. Error: %s", err))
	}

	dec.running = true

	dec.decodingQueue = make(chan Frame, BufferSize)
	dec.readyQueue = make(chan Frame, BufferSize)

	for i := 0; i < BufferSize; i++ {
		dec.decodingQueue <- make([]byte, dec.Metadata.Width*dec.Metadata.Height*3)
	}

	goroutines.RunOS(func() {
		for dec.running {
			frame, opened := <-dec.decodingQueue
			if !opened {
				break
			}

			var err1 error

			if dec.format == pixconv.RGB {
				_, err1 = io.ReadFull(dec.pipe, frame)
			} else {
				_, err1 = io.ReadFull(dec.pipe, readBuffer)

				pixconv.Convert(readBuffer, dec.format, frame, pixconv.RGB, dec.Metadata.Width, dec.Metadata.Height)
			}

			if err1 != nil {
				dec.running = false
				dec.finished = true
			}

			dec.readyQueue <- frame
		}

		dec.wg.Done()
	})
}

func (dec *VideoDecoder) GetFrame() Frame {
	return <-dec.readyQueue
}

func (dec *VideoDecoder) Free(frame Frame) {
	dec.decodingQueue <- frame
}

func (dec *VideoDecoder) HasFinished() bool {
	return dec.finished
}
