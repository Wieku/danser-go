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
			PointFadeOutTime:     10,
			ShowPositionalMisses: true,
			PositionalMissScale:  1.5,
			ShowUnstableRate:     true,
			UnstableRateDecimals: 0,
			UnstableRateScale:    1.0,
		},
		AimErrorMeter: &aimError{
			hudElementPosition: &hudElementPosition{
				hudElement: &hudElement{
					Show:    false,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XPosition: 1350,
				YPosition: 650,
			},
			PointFadeOutTime:     10,
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
			hudElementPosition: &hudElementPosition{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XPosition: 5,
				YPosition: 150,
			},
			Color: &hsv{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			Decimals:         0,
			Align:            "CentreLeft",
			ShowInResults:    true,
			ShowPPComponents: false,
		},
		HitCounter: &hitCounter{
			hudElementPosition: &hudElementPosition{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XPosition: 5,
				YPosition: 190,
			},
			Color: []*hsv{
				{
					Hue:        0,
					Saturation: 0,
					Value:      1,
				},
			},

			Spacing:    48,
			FontScale:  1,
			Align:      "Left",
			ValueAlign: "Left",
			Vertical:   false,
			Show300:    false,
		},
		StrainGraph: &strainGraph{
			Show:      true,
			Opacity:   1,
			XPosition: 5,
			YPosition: 310,
			Align:     "BottomLeft",
			Width:     130,
			Height:    70,
			BgColor: &hsv{
				Hue:        0,
				Saturation: 0,
				Value:      0.2,
			},
			FgColor: &hsv{
				Hue:        297,
				Saturation: 0.4,
				Value:      0.92,
			},
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
			AlignRight:     false,
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
			HideInReplays:     false,
			FoldInReplays:     false,
			AdditionalSpacing: 0,
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
		ShowHitLighting:         false,
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
	StrainGraph             *strainGraph
	KeyOverlay              *hudElementOffset
	ScoreBoard              *scoreBoard
	Mods                    *mods
	Boundaries              *boundaries
	ShowResultsScreen       bool
	ResultsScreenTime       float64
	ResultsUseLocalTimeZone bool
	ShowWarningArrows       bool
	ShowHitLighting         bool
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

type hudElementPosition struct {
	*hudElement
	XPosition float64
	YPosition float64
}

type hitError struct {
	*hudElementOffset
	PointFadeOutTime     float64
	ShowPositionalMisses bool
	PositionalMissScale  float64
	ShowUnstableRate     bool
	UnstableRateDecimals int
	UnstableRateScale    float64
}

type aimError struct {
	*hudElementPosition
	PointFadeOutTime     float64
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
	*hudElementPosition
	Color            *hsv
	Decimals         int
	Align            string
	ShowInResults    bool
	ShowPPComponents bool
}

type hitCounter struct {
	*hudElementPosition
	Color      []*hsv
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
	HideInReplays     bool
	FoldInReplays     bool
	AdditionalSpacing float64
}

type strainGraph struct {
	Show      bool
	Opacity   float64
	XPosition float64
	YPosition float64
	Align     string
	Width     float64
	Height    float64
	BgColor   *hsv
	FgColor   *hsv
}
