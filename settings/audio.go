package settings

var Audio = initAudio()

func initAudio() *audio {
	return &audio{
		GeneralVolume:             0.5,
		MusicVolume:               0.5,
		SampleVolume:              0.5,
		Offset:                    0,
		IgnoreBeatmapSamples:      false,
		IgnoreBeatmapSampleVolume: false,
	}
}

type audio struct {
	GeneralVolume             float64 //0.5
	MusicVolume               float64 //=0.5
	SampleVolume              float64 //=0.5
	Offset                    int64
	IgnoreBeatmapSamples      bool //= false
	IgnoreBeatmapSampleVolume bool //= false
}
