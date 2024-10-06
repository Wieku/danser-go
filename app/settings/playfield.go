package settings

var Playfield = initPlayfield()

func initPlayfield() *playfield {
	return &playfield{
		DrawObjects:                  true,
		DrawCursors:                  true,
		Scale:                        1,
		OsuShift:                     false,
		ShiftY:                       0,
		ShiftX:                       0,
		ScaleStoryboardWithPlayfield: false,
		MoveStoryboardWithPlayfield:  false,
		LeadInTime:                   5,
		LeadInHold:                   2,
		FadeOutTime:                  5,
		SeizureWarning: &seizure{
			Enabled:  true,
			Duration: 5,
		},
		Background: &background{
			LoadStoryboards: true,
			LoadVideos:      false,
			FlashToTheBeat:  false,
			Dim: &dim{
				Intro:  0,
				Normal: 0.95,
				Breaks: 0.5,
			},
			Parallax: &parallax{
				Enabled: true,
				Amount:  0.1,
				Speed:   0.5,
			},
			Blur: &blur{
				Enabled: false,
				Values: &dim2{
					Intro:  0,
					Normal: 0.6,
					Breaks: 0.3,
				},
			},
			Triangles: &triangles{
				Enabled:            false,
				Shadowed:           true,
				DrawOverBlur:       true,
				ParallaxMultiplier: 0.5,
				Density:            1,
				Scale:              1,
				Speed:              1,
			},
		},
		Logo: &logo{
			Enabled:      true,
			DrawSpectrum: false,
			Dim: &dim{
				Intro:  0,
				Normal: 1,
				Breaks: 1,
			},
		},
		Bloom: &bloom{
			Enabled:           false,
			BloomToTheBeat:    true,
			BloomBeatAddition: 0.3,
			Threshold:         0.0,
			Blur:              0.6,
			Power:             0.7,
		},
	}
}

type playfield struct {
	DrawObjects                  bool
	DrawCursors                  bool
	Scale                        float64  `label:"Playfield scale" min:"0.1" max:"2" liveedit:"false"`   //1, scale the playfield (1 means that 384 will be rescaled to 900 on FullHD monitor)
	OsuShift                     bool     `label:"Position the playfield like in osu!" liveedit:"false"` //false, offset the playfield like in osu! | Overrides ShiftY
	playfieldShift               string   `vector:"true" label:"Playfield shift" left:"ShiftX" right:"ShiftY" showif:"OsuShift=false" liveedit:"false"`
	ShiftX                       float64  `min:"-512" max:"512"` //offset the playfield by X osu!pixels
	ShiftY                       float64  `min:"-512" max:"512"` //offset the playfield by Y osu!pixels
	ScaleStoryboardWithPlayfield bool     `liveedit:"false"`
	MoveStoryboardWithPlayfield  bool     `tooltip:"Even if selected, \"Position the playfield like in osu!\" option won't affect the storyboard" liveedit:"false"`
	LeadInTime                   float64  `max:"10" format:"%.1fs" liveedit:"false"` //5
	LeadInHold                   float64  `max:"10" format:"%.1fs" liveedit:"false"` //2
	FadeOutTime                  float64  `max:"10" format:"%.1fs" liveedit:"false"` //5
	SeizureWarning               *seizure `liveedit:"false"`
	Background                   *background
	Logo                         *logo
	Bloom                        *bloom
}

type seizure struct {
	// Whether seizure warning should be displayed before intro
	Enabled bool

	Duration float64 `max:"10" format:"%.1fs"`
}

// Background controls
type background struct {
	// Whether storyboards should be loaded
	LoadStoryboards bool `liveedit:"false"`

	// Whether videos should be loaded
	LoadVideos bool `liveedit:"false"`

	FlashToTheBeat bool

	// Dim controls
	Dim *dim

	Parallax *parallax

	// Blur controls
	Blur *blur

	//Triangle controls
	Triangles *triangles
}

type parallax struct {
	Enabled bool

	// Amount of parallax, also scales bg by (1+Amount), set to 0 to disable it
	Amount float64 `min:"-1"`

	// Speed of parallax
	Speed float64
}

type blur struct {
	Enabled bool

	Values *dim2
}

type triangles struct {
	Enabled            bool
	Shadowed           bool
	DrawOverBlur       bool    `label:"Don't blur with background"`
	ParallaxMultiplier float32 `max:"2"`
	Density            float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
	Scale              float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
	Speed              float64 `min:"0.1" max:"5" scale:"100" format:"%.0f%%"`
}

type logo struct {
	Enabled      bool
	DrawSpectrum bool `label:"Draw spectrum analyzer"`
	Dim          *dim
}

type dim struct {
	// Value before drain time start
	Intro float64 `label:"During intro" scale:"100" format:"%.0f%%"`

	// Value during drain time
	Normal float64 `label:"During drain time" scale:"100" format:"%.0f%%"`

	// Value during breaks
	Breaks float64 `label:"During breaks" scale:"100" format:"%.0f%%"`
}

type dim2 struct {
	// Value before drain time start
	Intro float64 `label:"During intro" max:"2"`

	// Value during drain time
	Normal float64 `label:"During drain time" max:"2"`

	// Value during breaks
	Breaks float64 `label:"During breaks" max:"2"`
}

type bloom struct {
	Enabled           bool
	BloomToTheBeat    bool
	BloomBeatAddition float64 `max:"2"`
	Threshold         float64
	Blur              float64 `max:"2"`
	Power             float64 `max:"2"`
}
