package curves

import "github.com/wieku/danser-go/bmath"

type Curve interface {
	NPointAt(t float64) bmath.Vector2d
	PointAt(t float64) bmath.Vector2d
	GetStartAngle() float64
	GetEndAngle() float64
	GetLength() float64
	GetLines() []Linear
}
