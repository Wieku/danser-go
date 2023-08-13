package mutils

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
	"strings"
)

func Abs[T constraints.Integer | constraints.Float](a T) T {
	if a < 0 {
		return -a
	}

	return a
}

func Clamp[T constraints.Integer | constraints.Float](v, minV, maxV T) T {
	return min(maxV, max(minV, v))
}

func Lerp[T constraints.Integer | constraints.Float, V constraints.Float](min, max T, t V) T {
	return min + T(V(max-min)*t)
}

func Signum[T constraints.Float](a T) T {
	if a == 0 {
		return 0
	}

	if math.Signbit(float64(a)) {
		return -1
	}

	return 1
}

func Sanitize[T constraints.Float](v, maxV T) T {
	v = T(math.Mod(float64(v), float64(maxV)))
	if v < 0 {
		v += maxV
	}

	return v
}

func SanitizeAngle[T constraints.Float](v T) T {
	return Sanitize(v, T(2*math.Pi))
}

func SanitizeAngleArc[T constraints.Float](a T) T {
	sPi := T(math.Pi)

	if a < -sPi {
		a += 2 * sPi
	} else if a >= sPi {
		a -= 2 * sPi
	}

	return a
}

// FormatWOZeros formats the float with specified precision but removes trailing zeros
func FormatWOZeros[T constraints.Float](val T, precision int) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.*f", precision, val), "0"), ".")
}
