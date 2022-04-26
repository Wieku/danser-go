package mutils

import (
	"golang.org/x/exp/constraints"
	"math"
)

func ClampF[T constraints.Float](x, min, max T) T {
	return T(math.Min(float64(max), math.Max(float64(min), float64(x))))
}

func Min[T constraints.Integer](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func Max[T constraints.Integer](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func Clamp[T constraints.Integer](x, min, max T) T {
	return Min(max, Max(min, x))
}

func Lerp[T constraints.Float](min, max, t T) T {
	return min + (max-min)*t
}
