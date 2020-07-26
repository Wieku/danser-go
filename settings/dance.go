package settings

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

type dance struct {
	SliderDance       bool
	RandomSliderDance bool
	TAGSliderDance    bool
	SliderDance2B     bool
	Bezier            *bezier
	Flower            *flower
	HalfCircle        *circular
}
