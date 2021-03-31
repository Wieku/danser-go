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
		output = "danser_"+time.Now().Format("2006-01-02_15-04-05")
	}

	log.Println("Starting composing audio and video into one file...")
	cmd2 := exec.Command("ffmpeg",
		"-y",
		"-i", filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container),
		"-i", filepath.Join(settings.Recording.OutputDir, filename+".wav"),
		"-c:v", "copy",
		"-c:a", settings.Recording.AudioCodec,
		"-ab", settings.Recording.AudioBitrate,
		filepath.Join(settings.Recording.OutputDir, output+"."+settings.Recording.Container),
	)
	cmd2.Start()
	cmd2.Wait()

	log.Println("Finished! Cleaning up...")

	os.Remove(filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container))
	os.Remove(filepath.Join(settings.Recording.OutputDir, filename+".wav"))

	log.Println("Finished.")
}
