package settings

var Playfield = initPlayfield()

func initPlayfield() *playfield {
	return &playfield{
		ShowWarning:          false,
		LeadInTime:           5,
		LeadInHold:           2,
		FadeOutTime:          5,
		BackgroundInDim:      0,
		BackgroundDim:        0.95,
		BackgroundDimBreaks:  0.95,
		BlurEnable:           true,
		BackgroundInBlur:     0,
		BackgroundBlur:       0.6,
		BackgroundBlurBreaks: 0.6,
		SpectrumInDim:        0,
		SpectrumDim:          1,
		SpectrumDimBreaks:    0,
		DrawObjects:          true,
		StoryboardEnabled:    true,
		Scale:                1,
		OsuShift:             false,
		FlashToTheBeat:       true,
		UnblurToTheBeat:      true,
		UnblurFill:           0.8,
		KiaiFactor:           1.1,
		BloomEnabled:         true,
		BloomToTheBeat:       true,
		BloomBeatAddition:    0.3,
		Bloom: &bloom{
			Threshold: 0.0,
			Blur:      0.6,
			Power:     0.7,
		},
	}
}

type playfield struct {
	ShowWarning          bool
	LeadInTime           float64 //5
	LeadInHold           float64 //2
	FadeOutTime          float64 //5
	BackgroundInDim      float64 //0, background dim at the start of app
	BackgroundDim        float64 // 0.95, background dim at the beatmap start
	BackgroundDimBreaks  float64 // 0.95, background dim at the breaks
	BlurEnable           bool    //true
	BackgroundInBlur     float64 //0, background blur at the start of app
	BackgroundBlur       float64 // 0.6, background blur at the beatmap start
	BackgroundBlurBreaks float64 // 0.6, background blur at the breaks
	SpectrumInDim        float64 //0, background dim at the start of app
	SpectrumDim          float64 // 0.95, background dim at the beatmap start
	SpectrumDimBreaks    float64 // 0.95, background dim at the breaks
	DrawObjects          bool
	StoryboardEnabled    bool
	Scale                float64 //1, scale the playfield (1 means that 384 will be rescaled to 900 on FullHD monitor)
	OsuShift             bool    //false, offset the playfield like in osu!
	FlashToTheBeat       bool    //true, background dim varies accoriding to music power
	UnblurToTheBeat      bool    //true, background blur varies accoriding to music power
	UnblurFill           float64 //0.8, if blur is set to 0.6, then on full beat blur will be equal to 0.12
	KiaiFactor           float64 //1.2, scale and flash factor during Kiai
	BloomEnabled         bool
	BloomToTheBeat       bool
	BloomBeatAddition    float64
	Bloom                *bloom
}

type bloom struct {
	Threshold, Blur, Power float64
}
