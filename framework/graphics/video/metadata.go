package video

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Metadata struct {
	Width    int
	Height   int
	FPS      float64
	Duration float64
	PixFmt   string
}

type probeOutput struct {
	Streams []stream `json:"streams"`
	Format  format   `json:"format"`
}

type stream struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	PixFmt       string `json:"pix_fmt"`
	AvgFramerate string `json:"avg_frame_rate"`
	Framerate    string `json:"r_frame_rate"`
	Duration     string `json:"duration"`
}

type format struct {
	Duration string `json:"duration"`
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
		"-show_entries", "format",
		"-of", "json",
		"-loglevel", "quiet",
	).Output()

	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			log.Println("ffprobe not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg")
		}

		return nil
	}

	mData := new(probeOutput)

	err = json.Unmarshal(output, mData)
	if err != nil {
		log.Println("Failed to parse video metadata:", err)
		return nil
	}

	aFPS := 1000.0
	rFPS := 1000.0

	if mData.Streams[0].AvgFramerate != "" {
		aFPS = parseRate(mData.Streams[0].AvgFramerate)
	}

	if mData.Streams[0].Framerate != "" {
		rFPS = parseRate(mData.Streams[0].Framerate)
	}

	if mData.Streams[0].Duration == "" {
		mData.Streams[0].Duration = mData.Format.Duration
	}

	return &Metadata{
		Width:    mData.Streams[0].Width,
		Height:   mData.Streams[0].Height,
		FPS:      math.Min(aFPS, rFPS),
		Duration: parseRate(mData.Streams[0].Duration),
		PixFmt:   mData.Streams[0].PixFmt,
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
