package bmath

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

func AngleBetween(centre, p1, p2 vector.Vector2d) float64 {
	a := centre.Dst(p1)
	b := centre.Dst(p2)
	c := p1.Dst(p2)
	return math.Acos((a*a + b*b - c*c) / (2 * a * b))
}

func AngleBetween32(centre, p1, p2 vector.Vector2f) float32 {
	a := centre.Dst(p1)
	b := centre.Dst(p2)
	c := p1.Dst(p2)
	return math32.Acos((a*a + b*b - c*c) / (2 * a * b))
}

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
