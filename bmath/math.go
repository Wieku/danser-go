package bmath

import "math"

func AngleBetween(centre, p1, p2 Vector2d) float64 {
	a := centre.Dst(p1)
	b := centre.Dst(p2)
	c := p1.Dst(p2)
	return math.Acos((a*a + b*b - c*c) / (2 * a * b))
}
