package curves

import (
	"github.com/wieku/danser-go/bmath"
	"math"
)

type Linear struct {
	Point1, Point2 bmath.Vector2d
}

func NewLinear(pt1, pt2 bmath.Vector2d) Linear {
	return Linear{pt1, pt2}
}

func (ln Linear) PointAt(t float64) bmath.Vector2d {
	return ln.Point2.Sub(ln.Point1).Scl(t).Add(ln.Point1)
}

func (ln Linear) GetStartAngle() float64 {
	return ln.Point1.AngleRV(ln.Point2)
}

func (ln Linear) GetEndAngle() float64 {
	return ln.Point2.AngleRV(ln.Point1)
}

func (ln Linear) GetLength() float64 {
	return ln.Point1.Dst(ln.Point2)
}

func (ln Linear) GetLength32() float32 {
	s1 := ln.Point2.X32() - ln.Point1.X32()
	s2 := ln.Point2.Y32() - ln.Point1.Y32()
	s := s1*s1 + s2*s2
	return float32(math.Sqrt(float64(s)))
}
