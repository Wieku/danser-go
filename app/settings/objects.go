package settings

var Objects = initObjects()

func initObjects() *objects {
	return &objects{
		MandalaTexturesTrigger: 5,
		MandalaTexturesAlpha:   0.3,
		ForceSliderBallTexture: true,
		DrawApproachCircles:    true,
		DrawComboNumbers:       true,
		DrawReverseArrows:      true,
		DrawSliderFollowCircle: true,
		LoadSpinners:           false,
		Colors: &color{
			EnableRainbow: true,
			RainbowSpeed:  8,
			BaseColor: &hsv{
				0,
				1.0,
				1.0},
			EnableCustomHueOffset: false,
			HueOffset:             0,
			FlashToTheBeat:        true,
			FlashAmplitude:        100,
			currentHue:            0,
		},
		ObjectsSize:                   -1,
		CSMult:                        1.0,
		ScaleToTheBeat:                false,
		SliderLOD:                     30,
		SliderPathLOD:                 50,
		SliderSnakeIn:                 true,
		SliderSnakeInMult:             0.0,
		SliderSnakeOut:                true,
		SliderMerge:                   true,
		SliderDistortions:             true,
		DrawFollowPoints:              true,
		WhiteFollowPoints:             true,
		FollowPointColorOffset:        0.0,
		EnableCustomSliderBorderColor: false,
		CustomSliderBorderColor: &color{
			EnableRainbow: false,
			RainbowSpeed:  8,
			BaseColor: &hsv{
				0,
				0.0,
				1.0},
			EnableCustomHueOffset: false,
			HueOffset:             0,
			FlashToTheBeat:        true,
			FlashAmplitude:        100,
			currentHue:            0,
		},
		EnableCustomSliderBorderGradientOffset: true,
		SliderBorderGradientOffset:             18,
		StackEnabled:                           true,
	}
}

type objects struct {
	MandalaTexturesTrigger                 int     //5, minimum value of cursors needed to use more translucent texture
	MandalaTexturesAlpha                   float64 //0.3
	ForceSliderBallTexture                 bool    //true, if this is disabled, mandala texture will be used for slider ball
	DrawApproachCircles                    bool    //true
	DrawComboNumbers                       bool
	DrawReverseArrows                      bool
	DrawSliderFollowCircle                 bool
	LoadSpinners                           bool
	Colors                                 *color
	ObjectsSize                            float64 //-1, objects radius in osu!pixels. If value is less than 0, beatmap's CS will be used
	CSMult                                 float64 //1.2, if ObjectsSize is -1, then CS value will be multiplied by this
	ScaleToTheBeat                         bool    //true, objects size is changing with music peak amplitude
	SliderLOD                              int64   //30, number of triangles in a circle
	SliderPathLOD                          int64   //50, int(pixelLength*(PathLOD/100)) => number of slider path points
	SliderSnakeIn                          bool
	SliderSnakeInMult                      float64
	SliderSnakeOut                         bool
	SliderMerge                            bool
	SliderDistortions                      bool    //true, osu!stable slider distortions on aspire maps
	DrawFollowPoints                       bool    //true
	WhiteFollowPoints                      bool    //true
	FollowPointColorOffset                 float64 //0.0, hue offset of the followpoint
	EnableCustomSliderBorderColor          bool
	CustomSliderBorderColor                *color
	EnableCustomSliderBorderGradientOffset bool
	SliderBorderGradientOffset             float64 //18, hue offset of slider outer border
	StackEnabled                           bool    //true, stack leniency
}
