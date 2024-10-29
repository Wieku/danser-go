package settings

import (
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/util"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
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
		AV1NvencSettings: &av1NvencSettings{
			RateControl:       "cbr",
			Bitrate:           "5M",
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
			OversampleMultiplier: 16,
			BlendFrames:          24,
			BlendFunctionID:      27,
			GaussWeightsMult:     1.5,
		},
	}
}

type recording struct {
	resolution          string             `vector:"true" combo:"640x360|360p,854x480|480p,1280x720|720p (HD),1920x1080|1080p (FullHD),2560x1440|1440p (WQHD),3840x2160|2160p (4K),custom" left:"FrameWidth" right:"FrameHeight"`
	FrameWidth          int                `min:"1" max:"30720"`
	FrameHeight         int                `min:"1" max:"17280"`
	FPS                 int                `label:"FPS (PLEASE READ TOOLTIP)" string:"true" min:"1" max:"10727" tooltip:"IMPORTANT: If you plan to have a \"high fps\" video, use Motion Blur below instead of setting FPS to absurd numbers. Setting the value too high will result in a broken video!"`
	EncodingFPSCap      int                `string:"true" min:"0" max:"10727" label:"Max Encoding FPS (Speed)" tooltip:"Limits the speed at which danser renders the video. If FPS is set to 60 and this option to 30, then it means 2 minute map will take at least 4 minutes to render"`
	Encoder             string             `combo:"libx264|Software x264 (AVC),libx265|Software x265 (HEVC),h264_nvenc|NVIDIA NVENC H.264 (AVC),hevc_nvenc|NVIDIA NVENC H.265 (HEVC),av1_nvenc|NVIDIA NVENC AV1,h264_qsv|Intel QuickSync H.264 (AVC),hevc_qsv|Intel QuickSync H.265 (HEVC)" comboSrc:"EncoderOptions" tooltip:"Hardware encoding with AMD GPUs is not supported because software encoding provides better performance and results"`
	X264Settings        *x264Settings      `json:"libx264" label:"Software x264 (AVC) Settings" showif:"Encoder=libx264"`
	X265Settings        *x265Settings      `json:"libx265" label:"Software x265 (HEVC) Settings" showif:"Encoder=libx265"`
	H264NvencSettings   *h264NvencSettings `json:"h264_nvenc" label:"NVIDIA NVENC H.264 (AVC) Settings" showif:"Encoder=h264_nvenc"`
	HEVCNvencSettings   *hevcNvencSettings `json:"hevc_nvenc" label:"NVIDIA NVENC H.265 (HEVC) Settings" showif:"Encoder=hevc_nvenc"`
	AV1NvencSettings    *av1NvencSettings  `json:"av1_nvenc" label:"NVIDIA NVENC AV1 Settings" showif:"Encoder=av1_nvenc"`
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
	Container      string `combo:"mp4,mkv"`
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
	case "av1_nvenc":
		return g.AV1NvencSettings
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
	OversampleMultiplier int           `string:"true" min:"1" max:"512" tooltip:"Multiplier for FPS. FPS=60 and Oversample=16 means original footage has 960fps before blending to 60fps"`
	BlendFrames          int           `string:"true" min:"1" max:"512" tooltip:"How many frames should be blended together.\nValue 1.5x bigger than Oversample multiplier is recommended"`
	BlendFunctionID      int           `label:"Blending function" combo:"0|Flat,1|Linear,2|InQuad,3|OutQuad,4|InOutQuad,5|InCubic,6|OutCubic,7|InOutCubic,8|InQuart,9|OutQuart,10|InOutQuart,11|InQuint,12|OutQuint,13|InOutQuint,14|InSine,15|OutSine,16|InOutSine,17|InExpo,18|OutExpo,19|InOutExpo,20|InCirc,21|OutCirc,22|InOutCirc,23|InBack,24|OutBack,25|InOutBack,26|Gauss,27|GaussSymmetric,28|PyramidSymmetric,29|SemiCircle"`
	GaussWeightsMult     float64       `string:"true" min:"0" max:"10" label:"Gauss weights multiplier"`
	BlendWeights         *blendWeights `json:",omitempty"` // Deprecated
}

type blendWeights struct {
	UseManualWeights bool
	ManualWeights    string  `showif:"UseManualWeights=true"`
	AutoWeightsID    int     `showif:"UseManualWeights=false" label:"Blending function" combo:"0|Flat,1|Linear,2|InQuad,3|OutQuad,4|InOutQuad,5|InCubic,6|OutCubic,7|InOutCubic,8|InQuart,9|OutQuart,10|InOutQuart,11|InQuint,12|OutQuint,13|InOutQuint,14|InSine,15|OutSine,16|InOutSine,17|InExpo,18|OutExpo,19|InOutExpo,20|InCirc,21|OutCirc,22|InOutCirc,23|InBack,24|OutBack,25|InOutBack,26|Gauss,27|GaussSymmetric,28|PyramidSymmetric,29|SemiCircle"`
	GaussWeightsMult float64 `showif:"UseManualWeights=false" string:"true" min:"0" max:"10" label:"Gauss weights multiplier"`
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

var encoderCacheCreated bool
var encoderCache []string

func (d *defaultsFactory) EncoderOptions() []string {
	if !encoderCacheCreated {
		encoderCacheCreated = true

		eField, _ := reflect.ValueOf(initRecording()).Type().Elem().FieldByName("Encoder")

		possibleEncoders := strings.Split(eField.Tag.Get("combo"), ",")
		encoderCache = slices.Clone(possibleEncoders)

		ffmpegExec, err := files.GetCommandExec("ffmpeg", "ffmpeg")
		if err != nil { // Fail silently and show all encoders
			return encoderCache
		}

		// control group, if libx264 fails it means ffmpeg was not installed correctly
		ctrl := exec.Command(ffmpegExec, "-f", "lavfi", "-i", "color=black:s=240x144", "-vframes", "1", "-an", "-c:v", "libx264", "-f", "null", "-")
		if ctrl.Run() != nil {
			return encoderCache
		}

		toRemove := util.Balance(8, encoderCache, func(encoder string) (string, bool) {
			eName := strings.Split(encoder, "|")[0]

			cmd := exec.Command(ffmpegExec, "-f", "lavfi", "-i", "color=black:s=240x144", "-vframes", "1", "-an", "-c:v", eName, "-f", "null", "-")
			if cmd.Run() != nil {
				return encoder, true
			}

			return "", false
		})

		encoderCache = slices.DeleteFunc(encoderCache, func(s string) bool { return slices.Contains(toRemove, s) })
	}

	return encoderCache
}
