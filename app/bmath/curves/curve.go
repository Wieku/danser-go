package curves

import "github.com/wieku/danser-go/app/bmath"

type Curve interface {
	PointAt(t float32) bmath.Vector2f
	GetStartAngle() float32
	GetEndAngle() float32
	GetLength() float32
}
