package settings

var Dance *dance = initDance()

func initDance() *dance {
	return &dance{
		Movers:             []string{"spline"},
		Spinners:           []string{"circle"},
		DoSpinnersTogether: true,
		SpinnerRadius:      100,
		SliderDance:        false,
		RandomSliderDance:  false,
		TAGSliderDance:     false,
		SliderDance2B:      true,
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
		Spline: &spline{
			RotationalForce: false,
			StreamWobble:    true,
			WobbleScale:     0.67,
		},
	}
}

type dance struct {
	Movers             []string
	Spinners           []string
	DoSpinnersTogether bool
	SpinnerRadius      float64
	SliderDance        bool
	RandomSliderDance  bool
	TAGSliderDance     bool
	SliderDance2B      bool
	Bezier             *bezier
	Flower             *flower
	HalfCircle         *circular
	Spline             *spline
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

type spline struct {
	RotationalForce bool
	StreamWobble    bool
	WobbleScale     float64
}
