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
	Aggressiveness       float64 `max:"200" format:"%.0f"`
	SliderAggressiveness float64 `max:"50" format:"%.0f"`
}

func (d *defaultsFactory) InitBezier() *bezier {
	return &bezier{
		Aggressiveness:       60,
		SliderAggressiveness: 3,
	}
}

type flower struct {
	AngleOffset        float64 `min:"0" max:"360" format:"%.0f°"`
	DistanceMult       float64 `max:"3"`
	StreamAngleOffset  float64 `min:"0" max:"360" format:"%.0f°"`
	LongJump           int64   `min:"-1" max:"1000"`
	LongJumpMult       float64 `max:"3"`
	LongJumpOnEqualPos bool
}

func (d *defaultsFactory) InitFlower() *flower {
	return &flower{
		AngleOffset:        90,
		DistanceMult:       0.666,
		StreamAngleOffset:  90,
		LongJump:           -1,
		LongJumpMult:       0.7,
		LongJumpOnEqualPos: false,
	}
}

type circular struct {
	RadiusMultiplier float64 `min:"0.1" max:"3"`
	StreamTrigger    int64   `max:"500" format:"%dms"`
}

func (d *defaultsFactory) InitCircular() *circular {
	return &circular{
		RadiusMultiplier: 1,
		StreamTrigger:    130,
	}
}

type spline struct {
	RotationalForce  bool
	StreamHalfCircle bool
	StreamWobble     bool
	WobbleScale      float64 `max:"2"`
}

func (d *defaultsFactory) InitSpline() *spline {
	return &spline{
		RotationalForce:  false,
		StreamHalfCircle: true,
		StreamWobble:     true,
		WobbleScale:      0.67,
	}
}

type momentum struct {
	SkipStackAngles bool
	StreamRestrict  bool
	DurationMult    float64 `max:"8"`
	DurationTrigger float64 `max:"4000" format:"%.0fms"`
	StreamMult      float64 `min:"-10" max:"10"`
	RestrictAngle   float64 `min:"0" max:"180" format:"%.0f°"`
	RestrictArea    float64 `min:"0" max:"180" format:"%.0f°"`
	RestrictInvert  bool
	DistanceMult    float64 `min:"-4" max:"4"`
	DistanceMultOut float64 `min:"-4" max:"4"`
}

func (d *defaultsFactory) InitMomentum() *momentum {
	return &momentum{
		SkipStackAngles: false,
		StreamRestrict:  true,
		StreamMult:      0.7,
		DurationMult:    2,
		DurationTrigger: 500,
		RestrictAngle:   90,
		RestrictArea:    40,
		RestrictInvert:  true,
		DistanceMult:    0.6,
		DistanceMultOut: 0.45,
	}
}

type exgon struct {
	Delay int64 `min:"10" max:"500" format:"%dms"`
}

func (d *defaultsFactory) InitExGon() *exgon {
	return &exgon{
		Delay: 50,
	}
}

type linear struct {
	WaitForPreempt    bool
	ReactionTime      float64 `min:"10" max:"500" format:"%.0fms"`
	ChoppyLongObjects bool
}

func (d *defaultsFactory) InitLinear() *linear {
	return &linear{
		WaitForPreempt:    true,
		ReactionTime:      100,
		ChoppyLongObjects: false,
	}
}

type pippi struct {
	RotationSpeed    float64 `min:"0.1" max:"6" format:"%.fx"`
	RadiusMultiplier float64
	SpinnerRadius    float64 `max:"200" format:"%.fo!px"`
}

func (d *defaultsFactory) InitPippi() *pippi {
	return &pippi{
		RotationSpeed:    1.6,
		RadiusMultiplier: 0.98,
		SpinnerRadius:    100,
	}
}
