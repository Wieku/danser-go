package settings

var CursorDance = initCursorDance()

func initCursorDance() *cursorDance {
	return &cursorDance{
		Movers: []*mover{
			{
				Mover:             "spline",
				SliderDance:       false,
				RandomSliderDance: false,
			},
		},
		Spinners: []*spinner{
			{
				Mover:  "circle",
				Radius: 100,
			},
		},
		Battle:             false,
		DoSpinnersTogether: true,
		TAGSliderDance:     false,
		MoverSettings: &moverSettings{
			Bezier: []*bezier{
				{
					Aggressiveness:       60,
					SliderAggressiveness: 3,
				},
			},
			Flower: []*flower{
				{
					AngleOffset:        90,
					DistanceMult:       0.666,
					StreamAngleOffset:  90,
					LongJump:           -1,
					LongJumpMult:       0.7,
					LongJumpOnEqualPos: false,
				},
			},
			HalfCircle: []*circular{
				{
					RadiusMultiplier: 1,
					StreamTrigger:    130,
				},
			},
			Spline: []*spline{
				{
					RotationalForce:  false,
					StreamHalfCircle: true,
					StreamWobble:     true,
					WobbleScale:      0.67,
				},
			},
			Momentum: []*momentum{
				{
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
				},
			},
			ExGon: []*exgon{
				{
					Delay: 50,
				},
			},
		},
	}
}

type mover struct {
	Mover             string
	SliderDance       bool
	RandomSliderDance bool
}

type spinner struct {
	Mover  string
	Radius float64
}

type cursorDance struct {
	Movers             []*mover
	Spinners           []*spinner
	Battle             bool
	DoSpinnersTogether bool
	TAGSliderDance     bool
	MoverSettings      *moverSettings
}

type moverSettings struct {
	Bezier     []*bezier
	Flower     []*flower
	HalfCircle []*circular
	Spline     []*spline
	Momentum   []*momentum
	ExGon      []*exgon
}
