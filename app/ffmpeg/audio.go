package ffmpeg

import (
	"fmt"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/goroutines"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const MaxAudioBuffers = 2000

var cmdAudio *exec.Cmd

var audioPipe io.WriteCloser

var audioQueue chan []byte

var audioSync *sync.Mutex
var audioPool = make([][]byte, 0)

var endSyncAudio *sync.WaitGroup

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
		"-i", inputName,

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

	options = append(options, filepath.Join(settings.Recording.GetOutputDir(), output+"_temp", "audio."+settings.Recording.Container))

	log.Println("Running ffmpeg with options:", options)

	cmdAudio = exec.Command(ffmpegExec, options...)

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
		panic(fmt.Sprintf("ffmpeg's audio process failed to start! Please check if audio parameters are entered correctly or audio codec is supported by provided container. Error: %s", err))
	}

	audioBufSize := bass.GetMixerRequiredBufferSize(1 / audioFPS)

	for i := 0; i < MaxAudioBuffers; i++ {
		audioPool = append(audioPool, make([]byte, audioBufSize))
	}

	audioSync = &sync.Mutex{}

	audioQueue = make(chan []byte, MaxAudioBuffers)

	endSyncAudio = &sync.WaitGroup{}

	endSyncAudio.Add(1)

	goroutines.RunOS(func() {
		for {
			data, keepOpen := <-audioQueue

			if data != nil {
				_, err := audioPipe.Write(data)
				if err != nil {
					panic(fmt.Sprintf("ffmpeg's audio process finished abruptly! Please check if you have enough storage or audio parameters are entered correctly. Error: %s", err))
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
	})
}

func stopAudio() {
	log.Println("Audio finished! Stopping audio pipe...")

	close(audioQueue)
	endSyncAudio.Wait()
	_ = audioPipe.Close()

	log.Println("Audio pipe closed. Waiting for audio ffmpeg process to finish...")

	_ = cmdAudio.Wait()

	log.Println("Audio process finished.")
}

func PushAudio() {
	audioSync.Lock()

	//spin until at least one audio buffer is free
	for len(audioPool) == 0 {
	}

	data := audioPool[0]
	audioPool = audioPool[1:]

	audioSync.Unlock()

	bass.ProcessMixer(data)

	audioQueue <- data
}
