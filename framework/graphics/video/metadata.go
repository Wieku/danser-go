package video

import (
	"encoding/json"
	"fmt"
	"github.com/wieku/danser-go/framework/files"
	"log"
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
	IsMOV    bool // We need that info to determine if we can apply "ignore_editlist" parameter
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
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
}

func LoadMetadata(path string) *Metadata {
	_, err := os.Open(path)
	if err != nil {
		return nil
	}

	probeExec, err := files.GetCommandExec("ffmpeg", "ffprobe")
	if err != nil {
		log.Println("ffprobe not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg")
		return nil
	}

	mData := getProbeOutput(probeExec, path, false)

	if mData == nil {
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
		FPS:      min(aFPS, rFPS),
		Duration: parseRate(mData.Streams[0].Duration),
		PixFmt:   mData.Streams[0].PixFmt,
		IsMOV:    strings.Contains(mData.Format.FormatName, "mov"),
	}
}

func getProbeOutput(probeExec, path string, mov bool) *probeOutput {
	var args []string

	if mov {
		args = append(args, "-ignore_editlist", "1")
	}

	args = append(args, "-i", path,
		"-select_streams", "v:0",
		"-show_entries", "stream",
		"-show_entries", "format",
		"-of", "json",
		"-loglevel", "quiet",
	)

	output, err := exec.Command(probeExec, args...).Output()

	if err != nil {
		if strings.Contains(err.Error(), "127") || strings.Contains(strings.ToLower(err.Error()), "0xc0000135") {
			log.Println(fmt.Sprintf("ffmpeg was installed incorrectly! Please make sure needed libraries (libs/*.so or bin/*.dll) are installed as well. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg. Error: %s", err))
		} else {
			log.Println(fmt.Sprintf("ffprobe: Failed to get media info. Error: %s", err))
		}

		return nil
	}

	mData := new(probeOutput)

	err = json.Unmarshal(output, mData)
	if err != nil {
		log.Println("Failed to parse video metadata:", err)
		return nil
	}

	// Run a second pass if mov/mp4 is detected
	if !mov && strings.Contains(mData.Format.FormatName, "mov") {
		return getProbeOutput(probeExec, path, true)
	}

	return mData
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
