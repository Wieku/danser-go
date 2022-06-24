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
			StaticUnstableRate:   false,
			ScaleWithSpeed:       false,
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
			StaticUnstableRate:   false,
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
			StaticScore:     false,
			StaticAccuracy:  false,
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
		ComboCounter: &comboCounter{
			hudElementOffset: &hudElementOffset{
				hudElement: &hudElement{
					Show:    true,
					Scale:   1.0,
					Opacity: 1.0,
				},
				XOffset: 0,
				YOffset: 0,
			},
			Static: false,
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
			Color: &HSV{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			Decimals:         0,
			Align:            "CentreLeft",
			ShowInResults:    true,
			ShowPPComponents: false,
			Static:           false,
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
			Color: []*HSV{
				{
					Hue:        0,
					Saturation: 0,
					Value:      1,
				},
			},

			Spacing:          48,
			FontScale:        1,
			Align:            "Left",
			ValueAlign:       "Left",
			Vertical:         false,
			Show300:          false,
			ShowSliderBreaks: false,
		},
		StrainGraph: &strainGraph{
			Show:      true,
			Opacity:   1,
			XPosition: 5,
			YPosition: 310,
			Align:     "BottomLeft",
			Width:     130,
			Height:    70,
			BgColor: &HSV{
				Hue:        0,
				Saturation: 0,
				Value:      0.2,
			},
			FgColor: &HSV{
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
			ModsOnly:       false,
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
			BorderColor: &HSV{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			BorderOpacity: 1,
			BackgroundColor: &HSV{
				Hue:        0,
				Saturation: 1,
				Value:      0,
			},
			BackgroundOpacity: 0.5,
		},
		Underlay: &underlay{
			Path:       "",
			AboveHpBar: false,
		},
		HUDFont:                 "",
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
	ComboCounter            *comboCounter
	PPCounter               *ppCounter
	HitCounter              *hitCounter
	StrainGraph             *strainGraph
	KeyOverlay              *hudElementOffset
	ScoreBoard              *scoreBoard
	Mods                    *mods
	Boundaries              *boundaries
	Underlay                *underlay
	HUDFont                 string `file:"Select HUD font" filter:"TrueType/OpenType Font (*.ttf, *.otf)|ttf,otf"`
	ShowResultsScreen       bool
	ResultsScreenTime       float64 `label:"Results screen duration" min:"1" max:"20" format:"%.1fs"`
	ResultsUseLocalTimeZone bool    `label:"Show PC's time zone instead of UTC"`
	ShowWarningArrows       bool
	ShowHitLighting         bool
	FlashlightDim           float64
	PlayUsername            string
	UseLazerPP              bool
}

type boundaries struct {
	Enabled bool

	BorderThickness float64 `min:"0.5" max:"10" format:"%.1f o!px"`
	BorderFill      float64

	BorderColor   *HSV    `short:"true"`
	BorderOpacity float64 `scale:"100.0" format:"%.0f%%"`

	BackgroundColor   *HSV    `short:"true"`
	BackgroundOpacity float64 `scale:"100.0" format:"%.0f%%"`
}

type hudElement struct {
	Show    bool
	Scale   float64 `max:"3" scale:"100.0" format:"%.0f%%"`
	Opacity float64 `scale:"100.0" format:"%.0f%%"`
}

type hudElementOffset struct {
	*hudElement
	offset  string  `vector:"true" left:"XOffset" right:"YOffset"`
	XOffset float64 `min:"-10000" max:"10000"`
	YOffset float64 `min:"-10000" max:"10000"`
}

type hudElementPosition struct {
	*hudElement
	position  string  `vector:"true" left:"XPosition" right:"YPosition"`
	XPosition float64 `min:"-10000" max:"10000"`
	YPosition float64 `min:"-10000" max:"10000"`
}

type hitError struct {
	*hudElementOffset
	PointFadeOutTime     float64 `max:"10" format:"%.1fs"`
	ShowPositionalMisses bool
	PositionalMissScale  float64 `min:"1" max:"2" scale:"100" format:"%.0f%%"`
	ShowUnstableRate     bool
	UnstableRateDecimals int     `max:"5"`
	UnstableRateScale    float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
	StaticUnstableRate   bool
	ScaleWithSpeed       bool
}

type aimError struct {
	*hudElementPosition
	PointFadeOutTime     float64 `max:"10" format:"%.1fs"`
	DotScale             float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
	Align                string  `combo:"TopLeft,Top,TopRight,Left,Centre,Right,BottomLeft,Bottom,BottomRight"`
	ShowUnstableRate     bool
	UnstableRateScale    float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
	UnstableRateDecimals int     `max:"5"`
	StaticUnstableRate   bool
	CapPositionalMisses  bool
	AngleNormalized      bool
}

type score struct {
	*hudElementOffset
	ProgressBar     string `combo:"Pie,Bar,BottomRight,Bottom"`
	ShowGradeAlways bool   `label:"Always show grade"`
	StaticScore     bool
	StaticAccuracy  bool
}

type comboCounter struct {
	*hudElementOffset
	Static bool
}

type ppCounter struct {
	*hudElementPosition
	Color            *HSV   `short:"true"`
	Decimals         int    `max:"5"`
	Align            string `combo:"TopLeft,Top,TopRight,Left,Centre,Right,BottomLeft,Bottom,BottomRight"`
	ShowInResults    bool
	ShowPPComponents bool `label:"Show PP breakdown"`
	Static           bool
}

type hitCounter struct {
	*hudElementPosition
	Color            []*HSV  `new:"InitHSV" label:"Color list"`
	Spacing          float64 `string:"true" min:"0" max:"1366"`
	FontScale        float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
	Align            string  `combo:"TopLeft,Top,TopRight,Left,Centre,Right,BottomLeft,Bottom,BottomRight"`
	ValueAlign       string  `combo:"TopLeft,Top,TopRight,Left,Centre,Right,BottomLeft,Bottom,BottomRight"`
	Vertical         bool
	Show300          bool `label:"Show perfect hits"`
	ShowSliderBreaks bool
}

type scoreBoard struct {
	*hudElementOffset
	ModsOnly       bool
	AlignRight     bool
	HideOthers     bool
	ShowAvatars    bool
	ExplosionScale float64 `min:"0.1" max:"2" scale:"100" format:"%.0f%%"`
}

type mods struct {
	*hudElementOffset
	HideInReplays     bool
	FoldInReplays     bool
	AdditionalSpacing float64 `string:"true" min:"-1366" max:"1366"`
}

type strainGraph struct {
	Show    bool
	Opacity float64 `scale:"100.0" format:"%.0f%%"`

	position  string  `vector:"true" left:"XPosition" right:"YPosition"`
	XPosition float64 `min:"-10000" max:"10000"`
	YPosition float64 `min:"-10000" max:"10000"`

	Align string `combo:"TopLeft,Top,TopRight,Left,Centre,Right,BottomLeft,Bottom,BottomRight"`

	size   string  `vector:"true" left:"Width" right:"Height"`
	Width  float64 `string:"true" min:"1" max:"10000"`
	Height float64 `string:"true" min:"1" max:"768"`

	BgColor *HSV `label:"Background color" short:"true"`
	FgColor *HSV `label:"Foreground color" short:"true"`
}

type underlay struct {
	Path       string `file:"Select underlay image" filter:"PNG file (*.png)|png"`
	AboveHpBar bool
}
