package settings

var Objects = initObjects()

func initObjects() *objects {
	return &objects{
		DrawApproachCircles: true,
		DrawComboNumbers:    true,
		DrawFollowPoints:    true,
		LoadSpinners:        true,
		ScaleToTheBeat:      false,
		StackEnabled:        true,
		Sliders: &sliders{
			ForceSliderBallTexture: true,
			DrawEndCircles:         true,
			DrawSliderFollowCircle: true,
			DrawScorePoints:        true,
			SliderMerge:            false,
			SliderDistortions:      true,
			BorderWidth:            1.0,
			Quality: &quality{
				CircleLevelOfDetail: 50,
				PathLevelOfDetail:   50,
			},
			Snaking: &snaking{
				In:                 true,
				Out:                true,
				OutFadeInstant:     true,
				DurationMultiplier: 0,
				FadeMultiplier:     0,
			},
		},
		Colors: &objectcolors{
			MandalaTexturesTrigger: 5,
			MandalaTexturesAlpha:   0.3,
			Color: &color{
				EnableRainbow:         true,
				RainbowSpeed:          8,
				BaseColor:             DefaultsFactory.InitHSV(),
				EnableCustomHueOffset: false,
				HueOffset:             0,
				FlashToTheBeat:        false,
				FlashAmplitude:        100,
				currentHue:            0,
			},
			UseComboColors: false,
			ComboColors: []*HSV{
				DefaultsFactory.InitHSV(),
			},
			UseSkinComboColors:    false,
			UseBeatmapComboColors: false,
			Sliders: &sliderColors{
				WhiteScorePoints:      true,
				ScorePointColorOffset: 0,
				SliderBallTint:        false,
				Border: &borderColors{
					UseHitCircleColor: false,
					Color: &color{
						EnableRainbow: false,
						RainbowSpeed:  8,
						BaseColor: &HSV{
							0,
							0.0,
							1.0},
						EnableCustomHueOffset: false,
						HueOffset:             0,
						FlashToTheBeat:        false,
						FlashAmplitude:        100,
						currentHue:            0,
					},
					EnableCustomGradientOffset: true,
					CustomGradientOffset:       0,
				},
				Body: &bodyColors{
					UseHitCircleColor: true,
					Color: &color{
						EnableRainbow: false,
						RainbowSpeed:  8,
						BaseColor: &HSV{
							0,
							1.0,
							0.0},
						EnableCustomHueOffset: false,
						HueOffset:             0,
						FlashToTheBeat:        true,
						FlashAmplitude:        100,
						currentHue:            0,
					},
					InnerOffset: -0.5,
					OuterOffset: -0.05,
					InnerAlpha:  0.8,
					OuterAlpha:  0.8,
				},
			},
		},
	}
}

type objects struct {
	DrawApproachCircles bool //true
	DrawComboNumbers    bool
	DrawFollowPoints    bool
	LoadSpinners        bool
	ScaleToTheBeat      bool //true, objects size is changing with music peak amplitude
	StackEnabled        bool `label:"Enable stack leniency"` //true, stack leniency
	Sliders             *sliders
	Colors              *objectcolors
}

type sliders struct {
	ForceSliderBallTexture bool `label:"Force slider ball texture on mandalas"`
	DrawEndCircles         bool
	DrawSliderFollowCircle bool
	DrawScorePoints        bool //true
	SliderMerge            bool
	SliderDistortions      bool    //true, osu!stable slider distortions on aspire maps
	BorderWidth            float64 `max:"9"`
	Quality                *quality
	Snaking                *snaking
}

type quality struct {
	// Quality of slider unit circle, 50 means that circle will have 50 sides
	CircleLevelOfDetail int64 //30, number of triangles in a circle

	//Quality of slider path, 50 means that unit circle will be placed every 2 osu!pixels (100/PathLevelOfDetail)
	PathLevelOfDetail int64 //50, int(pixelLength*(PathLOD/100)) => number of slider path points
}

type snaking struct {
	In                 bool
	Out                bool
	OutFadeInstant     bool    `label:"Fade out the slider instantly" showif:"Out=true"`
	DurationMultiplier float64 `scale:"100" format:"%.0f%%" label:"In duration multiplier" showif:"In=true" tooltip:"How much of slider's duration should be added to snake in time"`
	FadeMultiplier     float64 `scale:"100" format:"%.0f%%" label:"In fade multiplier" showif:"In=true" tooltip:"How close to slider's start time snake in should end"`
}

type objectcolors struct {
	MandalaTexturesTrigger int     `label:"Use Mandala textures at x mirrors" string:"true"`      //5, minimum value of cursors needed to use more translucent texture
	MandalaTexturesAlpha   float64 `label:"Mandala textures opacity" scale:"100" format:"%.0f%%"` //0.3
	Color                  *color
	UseComboColors         bool   `label:"Use custom combo colors"`
	ComboColors            []*HSV `new:"InitHSV" label:"Custom combo colors" showif:"UseComboColors=true"`
	UseSkinComboColors     bool
	UseBeatmapComboColors  bool
	Sliders                *sliderColors
}

type sliderColors struct {
	WhiteScorePoints      bool    //true
	ScorePointColorOffset float64 `min:"-180" max:"180" format:"%.0f°" showif:"WhiteScorePoints=false"` //0.0, hue offset of the followpoint
	SliderBallTint        bool
	Border                *borderColors
	Body                  *bodyColors
}

type borderColors struct {
	UseHitCircleColor          bool
	Color                      *color
	EnableCustomGradientOffset bool
	CustomGradientOffset       float64 `min:"-180" max:"180" format:"%.0f°" showif:"EnableCustomGradientOffset=true"` //18, hue offset of slider outer border
}

type bodyColors struct {
	UseHitCircleColor bool
	Color             *color
	InnerOffset       float64 `min:"-2" max:"2"`
	OuterOffset       float64 `min:"-2" max:"2"`
	InnerAlpha        float64 `label:"Inner body opacity" scale:"100.0" format:"%.0f%%"`
	OuterAlpha        float64 `label:"Outer body opacity" scale:"100.0" format:"%.0f%%"`
}
