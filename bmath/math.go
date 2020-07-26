package bmath

import (
	"github.com/wieku/danser-go/bmath/math32"
	"math"
)

func AngleBetween(centre, p1, p2 Vector2d) float64 {
	a := centre.Dst(p1)
	b := centre.Dst(p2)
	c := p1.Dst(p2)
	return math.Acos((a*a + b*b - c*c) / (2 * a * b))
}

func AngleBetween32(centre, p1, p2 Vector2f) float32 {
	a := centre.Dst(p1)
	b := centre.Dst(p2)
	c := p1.Dst(p2)
	return math32.Acos((a*a + b*b - c*c) / (2 * a * b))
}
