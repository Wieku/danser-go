package putils

import (
	"math"

	"github.com/wieku/danser-go/framework/math/mutils"
)

func BPMToMillisecondsD(bpm float64) float64 {
	return BPMToMilliseconds(bpm, 4)
}

func BPMToMilliseconds(bpm float64, delimiter int) float64 {
	return 60000.0 / float64(delimiter) / bpm
}

func MillisecondsToBPMD(ms float64) float64 {
	return MillisecondsToBPM(ms, 4)
}

func MillisecondsToBPM(ms float64, delimiter int) float64 {
	return 60000.0 / (ms * float64(delimiter))
}

func Logistic(x, midpointOffset, multiplier, maxValue float64) float64 {
	return maxValue / (1 + math.Exp(multiplier*(midpointOffset-x)))
}

func LogisticE(exponent, maxValue float64) float64 {
	return maxValue / (1.0 + math.Exp(exponent))
}

func Smoothstep(x, start, end float64) float64 {
	x = mutils.Clamp((x-start)/(end-start), 0, 1)

	return x * x * (3.0 - 2.0*x)
}

func SmoothstepBellCurve(x, mean, width float64) float64 {
	x -= mean

	if x > 0 {
		x = width - x
	} else {
		x = width + x
	}

	return Smoothstep(x, 0, width)
}

func SmoothstepBellCurvex(x float64) float64 {
	x = 0.5 - math.Abs(x-0.5)
	x = mutils.Clamp(x*2.0, 0.0, 1.0)
	return x * x * (3.0 - 2.0*x)
}

func Smootherstep(x, start, end float64) float64 {
	x = mutils.Clamp((x-start)/(end-start), 0, 1)

	return x * x * x * (x*(6.0*x-15.0) + 10.0)
}

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func ReverseLerp(x, start, end float64) float64 {
	return mutils.Clamp((x-start)/(end-start), 0, 1)
}

func Norm(p float64, values ...float64) float64 {
	var sum float64
	for _, x := range values {
		sum += math.Pow(x, p)
	}

	return math.Pow(sum, 1/p)
}

func Powi(x float64, exp int) float64 {
	switch exp {
	case 0:
		return 1
	case 1:
		return x
	case 2:
		return x * x
	case 3:
		return x * x * x
	case 4:
		return x * x * x * x
	case 5:
		return x * x * x * x * x
	default:
		return math.Pow(x, float64(exp))
	}
}

// Erf approximates the error function using osu's DiffUtils implementation.
// https://github.com/ppy/osu/blob/master/osu.Game/Rulesets/Difficulty/Utils/DiffUtils.cs
func Erf(x float64) float64 {
	if x == 0 {
		return 0
	}

	if math.IsInf(x, 1) {
		return 1
	}

	if math.IsInf(x, -1) {
		return -1
	}

	if math.IsNaN(x) {
		return math.NaN()
	}

	// Constants for approximation (Abramowitz and Stegun formula 7.1.26).
	t := 1.0 / (1.0 + 0.3275911*math.Abs(x))
	tau := t * (0.254829592 + t*(-0.284496736+t*(1.421413741+t*(-1.453152027+t*1.061405429))))

	value := 1.0 - tau*math.Exp(-x*x)
	if x >= 0 {
		return value
	}

	return -value
}

// ErfInv approximates the inverse error function using osu's DiffUtils implementation.
func ErfInv(x float64) float64 {
	if x <= -1 {
		return math.Inf(-1)
	}

	if x >= 1 {
		return math.Inf(1)
	}

	if x == 0 {
		return 0
	}

	const a = 0.147
	sign := math.Copysign(1, x)
	x = math.Abs(x)

	ln := math.Log(1 - x*x)
	t1 := 2/(math.Pi*a) + ln/2
	t2 := ln / a
	baseApprox := math.Sqrt(t1*t1-t2) - t1

	// Correction reduces max error from -0.005 to -0.00045.
	correction := 0.0
	if x >= 0.85 {
		correction = math.Pow((x-0.85)/0.293, 8)
	}

	return sign * (math.Sqrt(baseApprox) + correction)
}
