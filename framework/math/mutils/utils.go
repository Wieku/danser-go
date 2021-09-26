package mutils

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"math"
)

func ClampF32(x, min, max float32) float32 {
	return math32.Min(max, math32.Max(min, x))
}

func ClampF64(x, min, max float64) float64 {
	return math.Min(max, math.Max(min, x))
}

func MinI(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func MaxI(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func ClampI(x, min, max int) int {
	return MinI(max, MaxI(min, x))
}

func MinI64(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}

func MaxI64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func ClampI64(x, min, max int64) int64 {
	return MinI64(max, MaxI64(min, x))
}

func LerpF32(min, max, t float32) float32 {
	return min + (max-min) * t
}

func LerpF64(min, max, t float64) float64 {
	return min + (max-min) * t
}