package curves

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type Bezier struct {
	points       []vector.Vector2f
	ApproxLength float32
}

func NewBezier(points []vector.Vector2f) *Bezier {
	bz := &Bezier{points: points}

	pointLength := float32(0.0)
	for i := 1; i < len(points); i++ {
		pointLength += points[i].Dst(points[i-1])
	}

	pointLength = math32.Ceil(pointLength * 30)

	for i := 1; i <= int(pointLength); i++ {
		bz.ApproxLength += bz.PointAt(float32(i) / pointLength).Dst(bz.PointAt(float32(i-1) / pointLength))
	}

	return bz
}

func NewBezierNA(points []vector.Vector2f) *Bezier {
	bz := &Bezier{points: points}
	bz.ApproxLength = 0.0
	return bz
}

func (bz *Bezier) PointAt(t float32) vector.Vector2f {
	x := float32(0.0)
	y := float32(0.0)
	n := len(bz.points) - 1
	for i := 0; i <= n; i++ {
		b := bernstein(int64(i), int64(n), t)
		x += bz.points[i].X * b
		y += bz.points[i].Y * b
	}
	return vector.NewVec2f(x, y)
}

func (bz Bezier) GetLength() float32 {
	return bz.ApproxLength
}

func (bz Bezier) GetStartAngle() float32 {
	return bz.points[0].AngleRV(bz.PointAt(1.0 / bz.ApproxLength))
}

func (bz Bezier) GetEndAngle() float32 {
	return bz.points[len(bz.points)-1].AngleRV(bz.PointAt((bz.ApproxLength - 1) / bz.ApproxLength))
}

func BinomialCoefficient(n, k int64) int64 {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}
	k = bmath.MinI64(k, n-k)
	var c int64 = 1
	var i int64 = 0
	for ; i < k; i++ {
		c = c * (n - i) / (i + 1)
	}

	return c
}

func bernstein(i, n int64, t float32) float32 {
	return float32(BinomialCoefficient(n, i)) * math32.Pow(t, float32(i)) * math32.Pow(1.0-t, float32(n-i))
}
