package easing

import "math"

var Easings = []func(float64) float64{
	Linear,
	QuadEaseOut,
	QuadEaseIn,
	QuadEaseIn,
	QuadEaseOut,
	QuadEaseInOut,
	CubicEaseIn,
	CubicEaseOut,
	CubicEaseInOut,
	QuartEaseIn,
	QuartEaseOut,
	QuartEaseInOut,
	QuintEaseIn,
	QuintEaseOut,
	QuintEaseInOut,
	SineEaseIn,
	SineEaseOut,
	SineEaseInOut,
	ExpoEaseIn,
	ExpoEaseOut,
	ExpoEaseInOut,
	CircularEaseIn,
	CircularEaseOut,
	CircularEaseInOut,
	ElasticEaseIn,
	ElasticEaseOut,
	ElasticEaseHalfOut,
	ElasticEaseQuartOut,
	ElasticEaseInOut,
	BackEaseIn,
	BackEaseOut,
	BackEaseInOut,
	BounceEaseIn,
	BounceEaseOut,
	BounceEaseInOut,
}

/*  Linear
-----------------------------------------------*/
func Linear(t float64) float64 {
	return t
}

/*  Quad
-----------------------------------------------*/
func QuadEaseIn(t float64) float64 {
	return t * t
}

func QuadEaseOut(t float64) float64 {
	return -(t * (t - 2))
}

func QuadEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	} else {
		return (-2 * t * t) + (4 * t) - 1
	}
}

/*  Cubic
-----------------------------------------------*/
func CubicEaseIn(t float64) float64 {
	return t * t * t
}

func CubicEaseOut(t float64) float64 {
	f := (t - 1)
	return f*f*f + 1
}

func CubicEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	} else {
		f := ((2 * t) - 2)
		return 0.5*f*f*f + 1
	}
}

/*  Quart
-----------------------------------------------*/
func QuartEaseIn(t float64) float64 {
	return t * t * t * t
}

func QuartEaseOut(t float64) float64 {
	f := (t - 1)
	return f*f*f*(1-t) + 1
}

func QuartEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	} else {
		f := (t - 1)
		return -8*f*f*f*f + 1
	}
}

/*  Quint
-----------------------------------------------*/
func QuintEaseIn(t float64) float64 {
	return t * t * t * t * t
}

func QuintEaseOut(t float64) float64 {
	f := (t - 1)
	return f*f*f*f*f + 1
}

func QuintEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	} else {
		f := ((2 * t) - 2)
		return 0.5*f*f*f*f*f + 1
	}
}

/*  Sine
-----------------------------------------------*/
func SineEaseIn(t float64) float64 {
	return math.Sin((t-1)*math.Pi/2) + 1
}

func SineEaseOut(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

func SineEaseInOut(t float64) float64 {
	return 0.5 * (1 - math.Cos(t*math.Pi))
}

/*  Circle
-----------------------------------------------*/
func CircularEaseIn(t float64) float64 {
	return 1 - math.Sqrt(1-(t*t))
}

func CircularEaseOut(t float64) float64 {
	return math.Sqrt((2 - t) * t)
}

func CircularEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 0.5 * (1 - math.Sqrt(1-4*(t*t)))
	} else {
		return 0.5 * (math.Sqrt(-((2 * t) - 3)*((2*t)-1)) + 1)
	}
}

/*  Expo
-----------------------------------------------*/
func ExpoEaseIn(t float64) float64 {
	if t == 0.0 {
		return t
	} else {
		return math.Pow(2, 10*(t-1))
	}
}

func ExpoEaseOut(t float64) float64 {
	if t == 1.0 {
		return t
	} else {
		return 1 - math.Pow(2, -10*t)
	}
}

func ExpoEaseInOut(t float64) float64 {
	if t == 0.0 || t == 1.0 {
		return t
	}

	if t < 0.5 {
		return 0.5 * math.Pow(2, (20*t)-10)
	} else {
		return -0.5*math.Pow(2, (-20*t)+10) + 1
	}
}

/*  Elastic
-----------------------------------------------*/
func ElasticEaseIn(t float64) float64 {
	return math.Sin(13*math.Pi*2*t) * math.Pow(2, 10*(t-1))
}

func ElasticEaseOut(t float64) float64 {
	return math.Sin(-13*math.Pi*2*(t+1))*math.Pow(2, -10*t) + 1
}

func ElasticEaseHalfOut(t float64) float64 {
	return math.Sin(-13*math.Pi*2*(t*0.5+1))*math.Pow(2, -10*t) + 1
}

func ElasticEaseQuartOut(t float64) float64 {
	return math.Sin(-13*math.Pi*2*(t*0.25+1))*math.Pow(2, -10*t) + 1
}

func ElasticEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 0.5 * math.Sin(13*math.Pi*2*(2*t)) * math.Pow(2, 10*((2*t)-1))
	} else {
		return 0.5 * (math.Sin(-13*math.Pi*2*((2*t-1)+1))*math.Pow(2, -10*(2*t-1)) + 2)
	}
}

/*  Back
-----------------------------------------------*/
func BackEaseIn(t float64) float64 {
	return t*t*t - t*math.Sin(t*math.Pi)
}

func BackEaseOut(t float64) float64 {
	f := (1 - t)
	return 1 - (f*f*f - f*math.Sin(f*math.Pi))
}

func BackEaseInOut(t float64) float64 {
	if t < 0.5 {
		f := 2 * t
		return 0.5 * (f*f*f - f*math.Sin(f*math.Pi))
	} else {
		f := (1 - (2*t - 1))
		return 0.5*(1-(f*f*f-f*math.Sin(f*math.Pi))) + 0.5
	}
}

/*  Bounce
-----------------------------------------------*/
func BounceEaseIn(t float64) float64 {
	return 1 - BounceEaseOut(1-t)
}

func BounceEaseOut(t float64) float64 {
	if t < 4/11.0 {
		return (121 * t * t) / 16.0
	} else if t < 8/11.0 {
		return (363 / 40.0 * t * t) - (99 / 10.0 * t) + 17/5.0
	} else if t < 9/10.0 {
		return (4356 / 361.0 * t * t) - (35442 / 1805.0 * t) + 16061/1805.0
	} else {
		return (54 / 5.0 * t * t) - (513 / 25.0 * t) + 268/25.0
	}
}

func BounceEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 0.5 * BounceEaseIn(t*2)
	} else {
		return 0.5*BounceEaseOut(t*2-1) + 0.5
	}
}
