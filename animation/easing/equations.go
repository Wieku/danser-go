package easing

import "math"

var back_s = 1.70158
var elastic_a = 0.0
var elastic_p = 0.0
var elastic_setA = false
var elastic_setP = false

func Linear(t float64) float64 {
	return t
}

func BackIn(t float64) float64 {
	return t * t * ((back_s + 1) * t - back_s)
}

func BackOut(t float64) float64 {
	t -= 1
	return t * t * ((back_s + 1) * t + back_s) + 1
}

func BackInOut(t float64) float64 {
	s := back_s * 1.525
	t *= 2
	if t < 1 {
		return 0.5 * (t * t * ((s + 1) * t - s))
	} else {
		t -= 2
		return 0.5 * (t * t * ((s + 1) * t + s) + 2)
	}
}

func SineIn(t float64) float64 {
	return - math.Cos(t * math.Pi / 2) + 1
}

func SineOut(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

func SineInOut(t float64) float64 {
	return -0.5 * (math.Cos(t * math.Pi) - 1)
}