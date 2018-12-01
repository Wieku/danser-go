package curves

import math2 "danser/bmath"

type Linear struct {
	point1, point2 math2.Vector2d
}

func NewLinear(pt1, pt2 math2.Vector2d) Linear {
	return Linear{pt1, pt2}
}

func (ln Linear) PointAt(t float64) math2.Vector2d {
	return ln.point2.Sub(ln.point1).Scl(t).Add(ln.point1)
}

func (ln Linear) GetStartAngle() float64 {
	return ln.point1.AngleRV(ln.point2)
}

func (ln Linear) GetEndAngle() float64 {
	return ln.point2.AngleRV(ln.point1)
}

func (ln Linear) GetLength() float64 {
	return ln.point1.Dst(ln.point2)
}
