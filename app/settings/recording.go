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
	resolution     string `vector:"true" left:"FrameWidth" right:"FrameHeight"`
	FrameWidth     int    `min:"1" max:"30720"`
	FrameHeight    int    `min:"1" max:"17280"`
	FPS            int    `string:"true" min:"1" max:"10727"`
	EncodingFPSCap int    `string:"true" min:"0" label:"Encoding FPS (Speed)"`
	Encoder        string `combo:"libx264|x264 (AVC),libx265|x265 (HEVC),h264_nvenc|NVIDIA NVENC H.264 (AVC),hevc_nvenc|NVIDIA NVENC H.265 (HEVC),h264_amf|AMD H.264 (AVC),hevc_amf|AMD H.265 (HEVC),h264_qsv|Intel QuickSync H.264 (AVC),hevc_qsv|Intel QuickSync H.265 (HEVC),libvpx-vp9|VP9"`
	EncoderOptions string `label:"Video Encoder Options"`
	Profile        string `label:"Video Encoder Profile"`
	Preset         string `label:"Video Encoder Preset"`
	PixelFormat    string `combo:"yuv420p|I420,yuv444p|I444,nv12|NV12,nv21|NV21"`
	Filters        string `label:"FFmpeg Video Filters"`
	AudioCodec     string `combo:"aac|AAC,libmp3lame|MP3,libopus|OPUS,flac|FLAC"`
	AudioOptions   string `label:"Audio Encoder Options"`
	AudioFilters   string `label:"FFmpeg Audio Filters"`
	OutputDir      string `path:"Select video output directory"`
	Container      string `combo:"mp4,mkv,webm"`
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
	OversampleMultiplier int `string:"true" min:"1" max:"512"`
	BlendFrames          int `string:"true" min:"1" max:"512"`
	BlendWeights         *blendWeights
}

type blendWeights struct {
	UseManualWeights bool
	ManualWeights    string
	AutoWeightsID    int     `combo:"0|Flat,1|Linear,2|InQuad,3|OutQuad,4|InOutQuad,5|InCubic,6|OutCubic,7|InOutCubic,8|InQuart,9|OutQuart,10|InOutQuart,11|InQuint,12|OutQuint,13|InOutQuint,14|InSine,15|OutSine,16|InOutSine,17|InExpo,18|OutExpo,19|InOutExpo,20|InCirc,21|OutCirc,22|InOutCirc,23|InBack,24|OutBack,25|InOutBack,26|Gauss,27|GaussSymmetric,28|PyramidSymmetric,29|SemiCircle"`
	GaussWeightsMult float64 `string:"true" min:"0" max:"10"`
}
