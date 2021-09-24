package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		HitErrorMeter: &hitError{
			hudElementOffset: &hudElementOffset{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XOffset: 0,
				YOffset: 0,
			},
			ShowPositionalMisses: true,
			ShowUnstableRate:     true,
			UnstableRateDecimals: 0,
			UnstableRateScale:    1.0,
		},
		AimErrorMeter: &aimError{
			hudElement: &hudElement{
				Show:    false,
				Scale:   1.0,
				Opacity: 1.0,
			},
			XPosition:            1350,
			YPosition:            650,
			DotScale:             1,
			Align:                "Right",
			ShowUnstableRate:     false,
			UnstableRateScale:    1,
			UnstableRateDecimals: 0,
			CapPositionalMisses:  true,
			AngleNormalized:      false,
		},
		Score: &score{
			hudElementOffset: &hudElementOffset{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XOffset: 0,
				YOffset: 0,
			},
			ProgressBar:     "Pie",
			ShowGradeAlways: false,
		},
		HpBar: &hudElementOffset{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			XOffset: 0,
			YOffset: 0,
		},
		ComboCounter: &hudElementOffset{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			XOffset: 0,
			YOffset: 0,
		},
		PPCounter: &ppCounter{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			Color: &hsv{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			XPosition:        5,
			YPosition:        150,
			Decimals:         0,
			Align:            "CentreLeft",
			ShowInResults:    true,
			ShowPPComponents: false,
		},
		HitCounter: &hitCounter{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			Color: []*hsv{
				{
					Hue:        0,
					Saturation: 0,
					Value:      1,
				},
			},
			XPosition:  5,
			YPosition:  190,
			Spacing:    48,
			FontScale:  1,
			Align:      "Left",
			ValueAlign: "Left",
			Vertical:   false,
			Show300:    false,
		},
		KeyOverlay: &hudElementOffset{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			XOffset: 0,
			YOffset: 0,
		},
		ScoreBoard: &scoreBoard{
			hudElementOffset: &hudElementOffset{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XOffset: 0,
				YOffset: 0,
			},
			AlignRight:     true,
			HideOthers:     false,
			ShowAvatars:    false,
			ExplosionScale: 1.0,
		},
		Mods: &mods{
			hudElementOffset: &hudElementOffset{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XOffset: 0,
				YOffset: 0,
			},
			HideInReplays: false,
			FoldInReplays: false,
		},
		Boundaries: &boundaries{
			Enabled:         true,
			BorderThickness: 1,
			BorderFill:      1,
			BorderColor: &hsv{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			BorderOpacity: 1,
			BackgroundColor: &hsv{
				Hue:        0,
				Saturation: 1,
				Value:      0,
			},
			BackgroundOpacity: 0.5,
		},
		ShowResultsScreen:       true,
		ResultsScreenTime:       5,
		ResultsUseLocalTimeZone: false,
		ShowWarningArrows:       true,
		FlashlightDim:           1,
		PlayUsername:            "Guest",
		UseLazerPP:              false,
	}
}

type gameplay struct {
	HitErrorMeter           *hitError
	AimErrorMeter           *aimError
	Score                   *score
	HpBar                   *hudElementOffset
	ComboCounter            *hudElementOffset
	PPCounter               *ppCounter
	HitCounter              *hitCounter
	KeyOverlay              *hudElementOffset
	ScoreBoard              *scoreBoard
	Mods                    *mods
	Boundaries              *boundaries
	ShowResultsScreen       bool
	ResultsScreenTime       float64
	ResultsUseLocalTimeZone bool
	ShowWarningArrows       bool
	FlashlightDim           float64
	PlayUsername            string
	UseLazerPP              bool
}

type boundaries struct {
	Enabled bool

	BorderThickness float64
	BorderFill      float64

	BorderColor   *hsv
	BorderOpacity float64

	BackgroundColor   *hsv
	BackgroundOpacity float64
}

type hudElement struct {
	Show    bool
	Scale   float64
	Opacity float64
}

type hudElementOffset struct {
	*hudElement
	XOffset float64
	YOffset float64
}

type hitError struct {
	*hudElementOffset
	ShowPositionalMisses bool
	ShowUnstableRate     bool
	UnstableRateDecimals int
	UnstableRateScale    float64
}

type aimError struct {
	*hudElement
	XPosition            float64
	YPosition            float64
	DotScale             float64
	Align                string
	ShowUnstableRate     bool
	UnstableRateScale    float64
	UnstableRateDecimals int
	CapPositionalMisses  bool
	AngleNormalized      bool
}

type score struct {
	*hudElementOffset
	ProgressBar     string
	ShowGradeAlways bool
}

type ppCounter struct {
	*hudElement
	Color            *hsv
	XPosition        float64
	YPosition        float64
	Decimals         int
	Align            string
	ShowInResults    bool
	ShowPPComponents bool
}

type hitCounter struct {
	*hudElement
	Color      []*hsv
	XPosition  float64
	YPosition  float64
	Spacing    float64
	FontScale  float64
	Align      string
	ValueAlign string
	Vertical   bool
	Show300    bool
}

type scoreBoard struct {
	*hudElementOffset
	AlignRight     bool
	HideOthers     bool
	ShowAvatars    bool
	ExplosionScale float64
}

type mods struct {
	*hudElementOffset
	HideInReplays bool
	FoldInReplays bool
}
