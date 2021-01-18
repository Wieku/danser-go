package video

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type VideoMetadata struct {
	Width  int
	Height int
	FPS    float64
}

func LoadMetadata(path string) *VideoMetadata {
	_, err := os.Open(path)
	if err != nil {
		return nil
	}

	cmd2 := exec.Command(
		"ffmpeg",
		"-i", path,
		"-f", "null",
	)

	pipe, err := cmd2.StderrPipe()

	if err != nil {
		log.Println("Failed to open pipe:", err)
		return nil
	}

	scanner := bufio.NewScanner(pipe)

	err = cmd2.Start()
	if err != nil {
		log.Println("Failed to start ffmpeg process:", err)
		return nil
	}

	metadata := new(VideoMetadata)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for:
		// "	Stream #0:0: Video: vp6f, yuv420p, 320x240, 314 kb/s, 30 tbr, 1k tbn, 1k tbc"
		// ----------------------------------------------------------^
		if strings.HasPrefix(line, "Stream #") && strings.Contains(line, "Video:") {
			regex1 := regexp.MustCompile("\\s(\\d+(\\.\\d+)?)(k?)\\stbr,")
			regex2 := regexp.MustCompile("\\s(\\d+)x(\\d+)[\\s,]")

			fpsRaw := strings.Split(regex1.FindString(line), " ")[1]

			multiplier := 1.0

			if strings.HasSuffix(fpsRaw, "k") {
				multiplier = 1000
				fpsRaw = fpsRaw[:len(fpsRaw)-1]
			}

			metadata.FPS, _ = strconv.ParseFloat(fpsRaw, 64)
			metadata.FPS *= multiplier

			resRaw := strings.Split(strings.TrimSpace(strings.ReplaceAll(regex2.FindString(line), ",", "")), "x")

			metadata.Width, _ = strconv.Atoi(resRaw[0])
			metadata.Height, _ = strconv.Atoi(resRaw[1])
		}
	}

	return metadata
}
