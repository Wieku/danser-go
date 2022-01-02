package settings

import (
	"github.com/wieku/danser-go/framework/env"
	"path/filepath"
)

var Recording = initRecording()

func initRecording() *recording {
	return &recording{
		FrameWidth:     1920,
		FrameHeight:    1080,
		FPS:            60,
		EncodingFPSCap: 0,
		Encoder:        "libx264",
		EncoderOptions: "-crf 14",
		Profile:        "high",
		Preset:         "faster",
		PixelFormat:    "yuv420p",
		Filters:        "",
		AudioCodec:     "aac",
		AudioOptions:   "-b:a 320k",
		AudioFilters:   "",
		OutputDir:      "videos",
		Container:      "mp4",
		ShowFFmpegLogs: true,
		MotionBlur: &motionblur{
			Enabled:              false,
			OversampleMultiplier: 3,
			BlendFrames:          5,
			BlendWeights: &blendWeights{
				UseManualWeights: false,
				ManualWeights:    "1 1.7 2.1 4.1 5",
				AutoWeightsID:    1,
				GaussWeightsMult: 1.5,
			},
		},
	}
}

type recording struct {
	FrameWidth     int
	FrameHeight    int
	FPS            int
	EncodingFPSCap int
	Encoder        string
	EncoderOptions string
	Profile        string
	Preset         string
	PixelFormat    string
	Filters        string
	AudioCodec     string
	AudioOptions   string
	AudioFilters   string
	OutputDir      string
	Container      string
	ShowFFmpegLogs bool
	MotionBlur     *motionblur

	outDir *string
}

func (g *recording) GetOutputDir() string {
	if g.outDir == nil {
		dir := filepath.Join(env.DataDir(), g.OutputDir)

		if filepath.IsAbs(g.OutputDir) {
			dir = g.OutputDir
		}

		g.outDir = &dir
	}

	return *g.outDir
}

type motionblur struct {
	Enabled              bool
	OversampleMultiplier int
	BlendFrames          int
	BlendWeights         *blendWeights
}

type blendWeights struct {
	UseManualWeights bool
	ManualWeights    string
	AutoWeightsID    int
	GaussWeightsMult float64
}
