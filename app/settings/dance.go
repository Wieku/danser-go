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
			StreamRestrict:  true,
			DurationMult:    2,
			DurationTrigger: 500,
			StreamMult:      1,
			StreamAngle:     90,
			RestrictAngle:   89,
			DistanceMult:    0.7,
			DistanceMultEnd: 0.7,
			DistanceMultOut: 0.45,
		},
		ExGon: &exgon{
			Delay: 50,
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
	StreamAngle     float64
	RestrictAngle   float64
	DistanceMult    float64
	DistanceMultEnd float64
	DistanceMultOut float64
}

type exgon struct {
	Delay int64
}
