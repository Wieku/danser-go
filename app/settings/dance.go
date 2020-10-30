package settings

var Dance *dance = initDance()

func initDance() *dance {
	return &dance{
		Movers:             []string{"spline"},
		Spinners:           []string{"circle"},
		DoSpinnersTogether: true,
		SpinnerRadius:      100,
		Battle:             false,
		SliderDance:        false,
		RandomSliderDance:  false,
		TAGSliderDance:     false,
		SliderDance2B:      true,
		Bezier: &bezier{
			Aggressiveness:       60,
			SliderAggressiveness: 3,
		},
		Flower: &flower{
			AngleOffset:        90,
			DistanceMult:       0.666,
			StreamAngleOffset:  90,
			LongJump:           -1,
			LongJumpMult:       0.7,
			LongJumpOnEqualPos: false,
		},
		HalfCircle: &circular{
			RadiusMultiplier: 1,
			StreamTrigger:    130,
		},
		Spline: &spline{
			RotationalForce:  false,
			StreamHalfCircle: true,
			StreamWobble:     true,
			WobbleScale:      0.67,
		},
		Momentum: &momentum{
			SkipStackAngles: false,
			RestrictAngle:   80,
			DistanceMult:    0.666,
			DistanceMultEnd: 0.666,
		},
	}
}

type dance struct {
	Movers             []string
	Spinners           []string
	DoSpinnersTogether bool
	SpinnerRadius      float64
	Battle             bool
	SliderDance        bool
	RandomSliderDance  bool
	TAGSliderDance     bool
	SliderDance2B      bool
	Bezier             *bezier
	Flower             *flower
	HalfCircle         *circular
	Spline             *spline
	Momentum           *momentum
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
	RestrictAngle   float64
	DistanceMult    float64
	DistanceMultEnd float64
}
