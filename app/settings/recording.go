package settings

import (
	"github.com/wieku/danser-go/framework/env"
	"path/filepath"
	"strings"
)

var Recording = initRecording()

func initRecording() *recording {
	return &recording{
		FrameWidth:     1920,
		FrameHeight:    1080,
		FPS:            60,
		EncodingFPSCap: 0,
		Encoder:        "libx264",
		X264Settings: &x264Settings{
			RateControl:       "crf",
			Bitrate:           "10M",
			CRF:               14,
			Profile:           "high",
			Preset:            "faster",
			AdditionalOptions: "",
		},
		X265Settings: &x265Settings{
			RateControl:       "crf",
			Bitrate:           "10M",
			CRF:               18,
			Preset:            "fast",
			AdditionalOptions: "",
		},
		H264NvencSettings: &h264NvencSettings{
			RateControl:       "cq",
			Bitrate:           "10M",
			CQ:                22,
			Profile:           "high",
			Preset:            "p7",
			AdditionalOptions: "",
		},
		HEVCNvencSettings: &hevcNvencSettings{
			RateControl:       "cq",
			Bitrate:           "10M",
			CQ:                24,
			Preset:            "p7",
			AdditionalOptions: "",
		},
		H264QSVSettings: &h264QSVSettings{
			RateControl:       "icq",
			Bitrate:           "10M",
			Quality:           15,
			Profile:           "high",
			Preset:            "slow",
			AdditionalOptions: "",
		},
		HEVCQSVSettings: &hevcQSVSettings{
			RateControl:       "icq",
			Bitrate:           "10M",
			Quality:           20,
			Preset:            "slow",
			AdditionalOptions: "",
		},
		CustomSettings: &custom{
			CustomOptions: "",
		},
		PixelFormat: "yuv420p",
		Filters:     "",
		AudioCodec:  "aac",
		AACSettings: &aacSettings{
			Bitrate:           "192k",
			AdditionalOptions: "",
		},
		MP3Settings: &mp3Settings{
			RateControl:       "abr",
			TargetBitrate:     "192k",
			AdditionalOptions: "",
		},
		OPUSSettings: &opusSettings{
			RateControl:       "vbr",
			TargetBitrate:     "192k",
			AdditionalOptions: "",
		},
		FLACSettings: &flacSettings{
			CompressionLevel:  12,
			AdditionalOptions: "",
		},
		CustomAudioSettings: &custom{
			CustomOptions: "",
		},
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
	resolution          string             `vector:"true" left:"FrameWidth" right:"FrameHeight"`
	FrameWidth          int                `min:"1" max:"30720"`
	FrameHeight         int                `min:"1" max:"17280"`
	FPS                 int                `string:"true" min:"1" max:"10727"`
	EncodingFPSCap      int                `string:"true" min:"0" max:"10727" label:"Max Encoding FPS (Speed)"`
	Encoder             string             `combo:"libx264|Software x264 (AVC),libx265|Software x265 (HEVC),h264_nvenc|NVIDIA NVENC H.264 (AVC),hevc_nvenc|NVIDIA NVENC H.265 (HEVC),h264_qsv|Intel QuickSync H.264 (AVC),hevc_qsv|Intel QuickSync H.265 (HEVC),libvpx-vp9|VP9"`
	X264Settings        *x264Settings      `json:"libx264" label:"Software x264 (AVC) Settings" showif:"Encoder=libx264"`
	X265Settings        *x265Settings      `json:"libx265" label:"Software x265 (HEVC) Settings" showif:"Encoder=libx265"`
	H264NvencSettings   *h264NvencSettings `json:"h264_nvenc" label:"NVIDIA NVENC H.264 (AVC) Settings" showif:"Encoder=h264_nvenc"`
	HEVCNvencSettings   *hevcNvencSettings `json:"hevc_nvenc" label:"NVIDIA NVENC H.265 (HEVC) Settings" showif:"Encoder=hevc_nvenc"`
	H264QSVSettings     *h264QSVSettings   `json:"h264_qsv" label:"Intel QuickSync H.264 (AVC) Settings" showif:"Encoder=h264_qsv"`
	HEVCQSVSettings     *hevcQSVSettings   `json:"hevc_qsv" label:"Intel QuickSync H.265 (HEVC) Settings" showif:"Encoder=hevc_qsv"`
	CustomSettings      *custom            `json:"custom" label:"Custom Encoder Settings" showif:"Encoder=!"`
	PixelFormat         string             `combo:"yuv420p|I420,yuv444p|I444,nv12|NV12,nv21|NV21" showif:"Encoder=!h264_qsv,!hevc_qsv"`
	Filters             string             `label:"FFmpeg Video Filters"`
	AudioCodec          string             `combo:"aac|AAC,libmp3lame|MP3,libopus|OPUS,flac|FLAC"`
	AACSettings         *aacSettings       `json:"aac" label:"AAC Settings" showif:"AudioCodec=aac"`
	MP3Settings         *mp3Settings       `json:"libmp3lame" label:"MP3 Settings" showif:"AudioCodec=libmp3lame"`
	OPUSSettings        *opusSettings      `json:"libopus" label:"OPUS Settings" showif:"AudioCodec=libopus"`
	FLACSettings        *flacSettings      `json:"flac" label:"FLAC Settings" showif:"AudioCodec=flac"`
	CustomAudioSettings *custom            `json:"customAudio" label:"Custom Audio Settings" showif:"AudioCodec=!"`
	//AudioOptions        string             `label:"Audio Encoder Options"`
	AudioFilters   string `label:"FFmpeg Audio Filters"`
	OutputDir      string `path:"Select video output directory"`
	Container      string `combo:"mp4,mkv,webm"`
	ShowFFmpegLogs bool
	MotionBlur     *motionblur

	outDir *string
}

func (g *recording) GetEncoderOptions() EncoderOptions {
	switch strings.ToLower(g.Encoder) {
	case "libx264":
		return g.X264Settings
	case "libx265":
		return g.X265Settings
	case "h264_nvenc":
		return g.H264NvencSettings
	case "hevc_nvenc":
		return g.HEVCNvencSettings
	case "h264_qsv":
		return g.H264QSVSettings
	case "hevc_qsv":
		return g.HEVCQSVSettings
	default:
		return g.CustomSettings
	}
}

func (g *recording) GetAudioOptions() EncoderOptions {
	switch strings.ToLower(g.AudioCodec) {
	case "aac":
		return g.AACSettings
	case "libmp3lame":
		return g.MP3Settings
	case "libopus":
		return g.OPUSSettings
	case "flac":
		return g.FLACSettings
	default:
		return g.CustomAudioSettings
	}
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

type EncoderOptions interface {
	GenerateFFmpegArgs() ([]string, error)
}

type custom struct {
	CustomOptions string
}

func (s *custom) GenerateFFmpegArgs() (ret []string, err error) {
	ret = parseCustomOptions(ret, s.CustomOptions)

	return
}

func parseCustomOptions(list []string, custom string) []string {
	if encOptions := strings.TrimSpace(custom); encOptions != "" {
		split := strings.Split(encOptions, " ")
		list = append(list, split...)
	}

	return list
}
