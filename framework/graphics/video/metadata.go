package video

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Metadata struct {
	Width  int
	Height int
	FPS    float64
	PixFmt string
}

type probeOutput struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	PixFmt       string `json:"pix_fmt"`
	AvgFramerate string `json:"avg_frame_rate"`
	Framerate    string `json:"r_frame_rate"`
}

func LoadMetadata(path string) *Metadata {
	_, err := os.Open(path)
	if err != nil {
		return nil
	}

	output, err := exec.Command(
		"ffprobe",
		"-i", path,
		"-select_streams", "v:0",
		"-show_entries", "stream",
		"-of", "json",
		"-loglevel", "quiet",
	).Output()

	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			log.Println("ffprobe not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://ffmpeg.org/")
		}

		return nil
	}

	mData := new(probeOutput)

	err = json.Unmarshal(output, mData)
	if err != nil {
		log.Println("Failed to parse video metadata:", err)
		return nil
	}

	if mData.Streams[0].AvgFramerate == "" {
		mData.Streams[0].AvgFramerate = mData.Streams[0].Framerate
	}

	return &Metadata{
		Width:  mData.Streams[0].Width,
		Height: mData.Streams[0].Height,
		FPS:    parseRate(mData.Streams[0].AvgFramerate),
		PixFmt: mData.Streams[0].PixFmt,
	}
}

func parseRate(rate string) float64 {
	split := strings.Split(rate, "/")

	fps, _ := strconv.ParseFloat(split[0], 64)

	if len(split) > 1 {
		div, _ := strconv.ParseFloat(split[1], 64)
		fps /= div
	}

	return fps
}
