package ffmpeg

import (
	"fmt"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/files"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var ffmpegExec string

var output string

// check used encoders exist
func preCheck() {
	var err error

	ffmpegExec, err = files.GetCommandExec("ffmpeg", "ffmpeg")
	if err != nil {
		panic("ffmpeg not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg")
	}

	log.Println("FFmpeg exec location:", ffmpegExec)

	out, err := exec.Command(ffmpegExec, "-encoders").Output()
	if err != nil {
		if strings.Contains(err.Error(), "127") || strings.Contains(strings.ToLower(err.Error()), "0xc0000135") {
			panic(fmt.Sprintf("ffmpeg was installed incorrectly! Please make sure needed libraries (libs/*.so or bin/*.dll) are installed as well. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg. Error: %s", err))
		}

		panic(fmt.Sprintf("Failed to get encoder info. Error: %s", err))
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
	preCheck()

	if strings.TrimSpace(_output) == "" {
		_output = "danser_" + time.Now().Format("2006-01-02_15-04-05")
	}

	output = _output

	log.Println("Starting encoding!")

	_ = os.RemoveAll(filepath.Join(settings.Recording.GetOutputDir(), output+"_temp"))

	err := os.MkdirAll(filepath.Join(settings.Recording.GetOutputDir(), output+"_temp"), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	startVideo(fps, _w, _h)
	startAudio(audioFPS)
}

func StopFFmpeg() {
	log.Println("Finishing rendering...")

	stopVideo()
	stopAudio()

	log.Println("Ffmpeg finished.")

	combine()
}

func combine() {
	options := []string{
		"-y",
		"-i", filepath.Join(settings.Recording.GetOutputDir(), output+"_temp", "video."+settings.Recording.Container),
		"-i", filepath.Join(settings.Recording.GetOutputDir(), output+"_temp", "audio."+settings.Recording.Container),
		"-c:v", "copy",
		"-c:a", "copy",
	}

	if settings.Recording.Container == "mp4" {
		options = append(options, "-movflags", "+faststart")
	}

	finalOutputPath := filepath.Join(settings.Recording.GetOutputDir(), output+"."+settings.Recording.Container)

	options = append(options, finalOutputPath)

	log.Println("Starting composing audio and video into one file...")
	log.Println("Running ffmpeg with options:", options)
	cmd2 := exec.Command(ffmpegExec, options...)

	if settings.Recording.ShowFFmpegLogs {
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
	}

	if err := cmd2.Start(); err != nil {
		log.Println("Failed to start ffmpeg:", err)
	} else {
		if err = cmd2.Wait(); err != nil {
			log.Println("ffmpeg finished abruptly! Please check if you have enough storage. Error:", err)
		} else {
			log.Println("Finished!")
			log.Println("Video is available at:", finalOutputPath)
		}
	}

	cleanup()
}

func cleanup() {
	log.Println("Cleaning up intermediate files...")

	_ = os.RemoveAll(filepath.Join(settings.Recording.GetOutputDir(), output+"_temp"))

	log.Println("Finished.")
}
