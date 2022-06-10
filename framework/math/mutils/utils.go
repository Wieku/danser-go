package mutils

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
	"strings"
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

// FormatWOZeros formats the float with specified precision but removes trailing zeros
func FormatWOZeros[T constraints.Float](val T, precision int) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", precision, val), "0"), ".")
}
