package ffmpeg

import (
	"github.com/wieku/danser-go/app/settings"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func Combine(output string) {
	if strings.TrimSpace(output) == "" {
		output = "danser_" + time.Now().Format("2006-01-02_15-04-05")
	}

	options := []string{
		"-y",
		"-i", filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container),
		"-i", filepath.Join(settings.Recording.OutputDir, filename+".wav"),
		"-c:v", "copy",
	}

	filters := strings.TrimSpace(settings.Recording.AudioFilters)
	if len(filters) > 0 {
		options = append(options, "-af", filters)
	}

	options = append(options,
		"-c:a", settings.Recording.AudioCodec,
		"-ab", settings.Recording.AudioBitrate,
		)

	if settings.Recording.Container == "mp4" {
		options = append(options, "-movflags", "+faststart")
	}

	options = append(options, filepath.Join(settings.Recording.OutputDir, output+"."+settings.Recording.Container))

	log.Println("Starting composing audio and video into one file...")
	log.Println("Running ffmpeg with options:", options)
	cmd2 := exec.Command("ffmpeg", options...)
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	if err := cmd2.Start(); err != nil {
		log.Println("Failed to start ffmpeg:", err)
	} else {
		if err = cmd2.Wait(); err != nil {
			log.Println("ffmpeg finished abruptly! Please check if you have enough storage or audio bitrate is entered correctly.")
		} else {
			log.Println("Finished!")
		}
	}

	log.Println("Cleaning up intermediate files...")

	_ = os.Remove(filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container))
	_ = os.Remove(filepath.Join(settings.Recording.OutputDir, filename+".wav"))

	log.Println("Finished.")
}
