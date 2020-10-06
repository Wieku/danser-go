package settings

var Dance *dance = initDance()

func initDance() *dance {
	return &dance{
		Movers:            []string{"flower"},
		SliderDance:       false,
		RandomSliderDance: false,
		TAGSliderDance:    false,
		SliderDance2B:     true,
		Bezier: &bezier{
			Aggressiveness:       60,
			SliderAggressiveness: 3,
		},
		Flower: &flower{
			UseNewStyle:        true,
			AngleOffset:        90,
			DistanceMult:       0.666,
			StreamTrigger:      130,
			StreamAngleOffset:  90,
			LongJump:           -1,
			LongJumpMult:       0.7,
			LongJumpOnEqualPos: false,
		},
		HalfCircle: &circular{
			RadiusMultiplier: 1,
			StreamTrigger:    130,
		},
	}
}

type dance struct {
	Movers            []string
	SliderDance       bool
	RandomSliderDance bool
	TAGSliderDance    bool
	SliderDance2B     bool
	Bezier            *bezier
	Flower            *flower
	HalfCircle        *circular
}

type bezier struct {
	Aggressiveness, SliderAggressiveness float64
}

type flower struct {
	UseNewStyle        bool
	AngleOffset        float64
	DistanceMult       float64
	StreamTrigger      int64
	StreamAngleOffset  float64
	LongJump           int64
	LongJumpMult       float64
	LongJumpOnEqualPos bool
}

type circular struct {
	RadiusMultiplier float64
	StreamTrigger    int64
}
