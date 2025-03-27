package putils

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
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
