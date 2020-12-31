package settings

var Recording = initRecording()

func initRecording() *recording {
	return &recording{
		FrameWidth:     1920,
		FrameHeight:    1080,
		FPS:            60,
		Encoder:        "libx264",
		EncoderOptions: "-b:v 10M",
		Profile:        "high",
		Preset:         "slow",
		PixelFormat:    "yuv420p",
		Filters:        "",
		OutputDir:      "videos",
		Container:      "mp4",
		MotionBlur: &motionblur{
			Enabled:              false,
			OversampleMultiplier: 3,
			BlendFrames:          5,
			BlendWeights: &blendWeights{
				UseManualWeights: false,
				ManualWeights:    "1 1.7 2.1 4.1 5",
				AutoWeightsID:    1,
				AutoWeightsScale: 1,
			},
		},
	}
}

type recording struct {
	FrameWidth     int
	FrameHeight    int
	FPS            int
	Encoder        string
	EncoderOptions string
	Profile        string
	Preset         string
	PixelFormat    string
	Filters        string
	OutputDir      string
	Container      string
	MotionBlur     *motionblur
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
	AutoWeightsScale float64
}
