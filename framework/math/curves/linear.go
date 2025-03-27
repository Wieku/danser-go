package curves

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

type Linear struct {
	Point1, Point2 vector.Vector2f
	customLength   float64
}

func NewLinear(pt1, pt2 vector.Vector2f) Linear {
	return Linear{pt1, pt2, float64(pt1.Dst(pt2))}
}

func NewLinearC(pt1, pt2 vector.Vector2f, customLength float64) Linear {
	return Linear{pt1, pt2, customLength}
}

func (ln Linear) PointAt(t float32) vector.Vector2f {
	return ln.Point1.Lerp(ln.Point2, t)
}

func (ln Linear) GetStartAngle() float32 {
	return ln.Point1.AngleRV(ln.Point2)
}

func (ln Linear) GetEndAngle() float32 {
	return ln.Point2.AngleRV(ln.Point1)
}

func (ln Linear) GetLength() float32 {
	return ln.Point1.Dst(ln.Point2)
}

func (ln Linear) GetLength87() float32 {
	return ln.Point1.Dst87(ln.Point2)
}

func (ln Linear) GetCustomLength() float64 {
	return ln.customLength
}
