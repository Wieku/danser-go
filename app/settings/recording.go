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
}
