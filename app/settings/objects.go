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
			BorderWidth:            1.0,
			Distortions: &distortions{
				Enabled:             true,
				ViewportSize:        0,
				UseCustomResolution: false,
				CustomResolutionX:   1920,
				CustomResolutionY:   1080,
			},
			Snaking: &snaking{
				In:                 true,
				Out:                true,
				OutFadeInstant:     true,
				DurationMultiplier: 0,
				FadeMultiplier:     0,
			},
		},
		Colors: &objectColors{
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
	LoadSpinners        bool `liveedit:"false"`
	ScaleToTheBeat      bool //true, objects size is changing with music peak amplitude
	StackEnabled        bool `label:"Enable stack leniency" liveedit:"false"` //true, stack leniency
	Sliders             *sliders
	Colors              *objectColors
}

type sliders struct {
	ForceSliderBallTexture bool `label:"Force slider ball texture on mandalas"`
	DrawEndCircles         bool
	DrawSliderFollowCircle bool
	DrawScorePoints        bool //true
	SliderMerge            bool
	BorderWidth            float64      `max:"9"`
	Distortions            *distortions `liveedit:"false"`
	Snaking                *snaking
}

type distortions struct {
	Enabled             bool
	ViewportSize        int    `showif:"Enabled=true" combo:"0|Use GPU's Viewport Size,8192|8192 (very low end GPUs),16384|16384 (lower/mid end GPUs),32768|32768 (higher end GPUs)" tooltip:"Some maps are made on lower end GPUs so sliders may not look correct on better graphic cards. Set it to 16384 in that case."`
	UseCustomResolution bool   `showif:"Enabled=true" tooltip:"Some sliders are made to specific screen size (most probably 1920x1080). Use this if you plan to render/watch at a different resolution."`
	customResolution    string `showif:"UseCustomResolution=true" vector:"true" left:"CustomResolutionX" right:"CustomResolutionY"`
	CustomResolutionX   int    `min:"1" max:"30720"`
	CustomResolutionY   int    `min:"1" max:"17280"`
}

type snaking struct {
	In                 bool
	Out                bool
	OutFadeInstant     bool    `label:"Fade out the slider instantly" showif:"Out=true"`
	DurationMultiplier float64 `scale:"100" format:"%.0f%%" label:"In duration multiplier" showif:"In=true" tooltip:"How much of slider's duration should be added to snake in time"`
	FadeMultiplier     float64 `scale:"100" format:"%.0f%%" label:"In fade multiplier" showif:"In=true" tooltip:"How close to slider's start time snake in should end"`
}

type objectColors struct {
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
