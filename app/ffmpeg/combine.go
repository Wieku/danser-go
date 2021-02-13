package ffmpeg

import (
	"github.com/wieku/danser-go/app/settings"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Combine() {
	log.Println("Starting composing audio and video into one file...")
	cmd2 := exec.Command("ffmpeg",
		"-y",
		"-i", filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container),
		"-i", filepath.Join(settings.Recording.OutputDir, filename+".wav"),
		"-c:v", "copy",
		"-c:a", "aac",
		"-ab", "320k",
		filepath.Join(settings.Recording.OutputDir, "danser_"+time.Now().Format("2006-01-02_15-04-05")+"."+settings.Recording.Container),
	)
	cmd2.Start()
	cmd2.Wait()

	log.Println("Finished! Cleaning up...")

	os.Remove(filepath.Join(settings.Recording.OutputDir, filename+"."+settings.Recording.Container))
	os.Remove(filepath.Join(settings.Recording.OutputDir, filename+".wav"))

	log.Println("Finished.")
}
