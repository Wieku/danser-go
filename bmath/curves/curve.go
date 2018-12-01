package curves

import "danser/bmath"

type Curve interface {
	PointAt(t float64) bmath.Vector2d
	GetStartAngle() float64
	GetEndAngle() float64
	GetLength() float64
}
