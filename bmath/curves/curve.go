package curves

import "danser/bmath"

type Curve interface {
	PointAt(t float64) bmath.Vector2d
	GetLength() float64
}
