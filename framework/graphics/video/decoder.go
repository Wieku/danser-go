package video

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

const BufferSize = 3

type Frame []byte

type VideoDecoder struct {
	filePath string

	Metadata *VideoMetadata

	emptyQueue chan Frame

	readyQueue chan Frame

	command  *exec.Cmd
	pipe     io.ReadCloser
	running  bool
	wg       *sync.WaitGroup
	finished bool
}

func NewVideoDecoder(filePath string) *VideoDecoder {
	metadata := LoadMetadata(filePath)
	if metadata == nil {
		return nil
	}

	return &VideoDecoder{
		filePath: filePath,
		Metadata: metadata,
		wg:       &sync.WaitGroup{},
	}
}

func (dec *VideoDecoder) StartFFmpeg(millis int64) {
	if dec.running {
		dec.running = false
		close(dec.emptyQueue)

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
		"-pix_fmt", "rgb24",
		"-",
	)

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

	dec.emptyQueue = make(chan Frame, BufferSize)
	dec.readyQueue = make(chan Frame, BufferSize)

	for i := 0; i < BufferSize; i++ {
		dec.emptyQueue <- make([]byte, dec.Metadata.Width*dec.Metadata.Height*3)
	}

	go func() {
		for dec.running {
			frame, closed := <-dec.emptyQueue
			if !closed {
				break
			}

			_, err = io.ReadFull(dec.pipe, frame)
			if err != nil {
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
	dec.emptyQueue <- frame
}

func (dec *VideoDecoder) HasFinished() bool {
	return dec.finished
}
