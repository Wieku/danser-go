package video

import (
	"fmt"
	"github.com/wieku/danser-go/framework/util/pixconv"
	"io"
	"os/exec"
	"runtime"
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

	command  *exec.Cmd
	pipe     io.ReadCloser
	running  bool
	wg       *sync.WaitGroup
	finished bool
	format   pixconv.PixFmt

	Channels int
}

func NewVideoDecoder(filePath string) *VideoDecoder {
	metadata := LoadMetadata(filePath)
	if metadata == nil {
		return nil
	}

	decoder := &VideoDecoder{
		filePath: filePath,
		Metadata: metadata,
		wg:       &sync.WaitGroup{},
		format:   pixconv.ARGB,
		Channels: 3,
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

	if decoder.format != pixconv.ARGB {
		decoder.Channels = 4
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
		"ffmpeg",
		args...,
	)

	var err error
	dec.pipe, err = dec.command.StdoutPipe()
	if err != nil {
		panic(err)
	}

	dec.command.Start()
	dec.running = true

	dec.decodingQueue = make(chan Frame, BufferSize)
	dec.readyQueue = make(chan Frame, BufferSize)

	for i := 0; i < BufferSize; i++ {
		dec.decodingQueue <- make([]byte, dec.Metadata.Width*dec.Metadata.Height*dec.Channels)
	}

	go func() {
		runtime.LockOSThread()

		for dec.running {
			frame, opened := <-dec.decodingQueue
			if !opened {
				break
			}

			var err1 error

			if dec.format == pixconv.ARGB {
				_, err1 = io.ReadFull(dec.pipe, frame)
			} else {
				_, err1 = io.ReadFull(dec.pipe, readBuffer)

				pixconv.Convert(readBuffer, dec.format, frame, pixconv.ARGB, dec.Metadata.Width, dec.Metadata.Height)
			}

			if err1 != nil {
				dec.running = false
				dec.finished = true
			}

			dec.readyQueue <- frame
		}

		dec.wg.Done()
	}()
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
