package settings

type danceOld struct {
	Movers             []string
	Spinners           []string
	DoSpinnersTogether bool
	SpinnerRadius      float64
	Battle             bool
	SliderDance        bool
	RandomSliderDance  bool
	TAGSliderDance     bool
	Bezier             *bezier
	Flower             *flower
	HalfCircle         *circular
	Spline             *spline
	Momentum           *momentum
	ExGon              *exgon
}

type bezier struct {
	Aggressiveness, SliderAggressiveness float64
}

type flower struct {
	AngleOffset        float64
	DistanceMult       float64
	StreamAngleOffset  float64
	LongJump           int64
	LongJumpMult       float64
	LongJumpOnEqualPos bool
}

type circular struct {
	RadiusMultiplier float64
	StreamTrigger    int64
}

type spline struct {
	RotationalForce  bool
	StreamHalfCircle bool
	StreamWobble     bool
	WobbleScale      float64
}

type momentum struct {
	SkipStackAngles bool
	StreamRestrict  bool
	DurationMult    float64
	DurationTrigger float64
	StreamMult      float64
	RestrictAngle   float64
	RestrictArea    float64
	RestrictInvert  bool
	DistanceMult    float64
	DistanceMultOut float64
}

type exgon struct {
	Delay int64
}

type linear struct {
	WaitForPreempt bool
	ReactionTime   float64
}
