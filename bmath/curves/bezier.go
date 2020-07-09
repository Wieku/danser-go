package curves

import (
	math2 "github.com/wieku/danser-go/bmath"
	"math"
)

type Bezier struct {
	points       []math2.Vector2d
	ApproxLength float64
}

func NewBezier(points []math2.Vector2d) *Bezier {
	bz := &Bezier{points: points}

	pointLength := 0.0
	for i := 1; i < len(points); i++ {
		pointLength += points[i].Dst(points[i-1])
	}

	pointLength = math.Ceil(pointLength * 30)

	for i := 1; i <= int(pointLength); i++ {
		bz.ApproxLength += bz.PointAt(float64(i) / pointLength).Dst(bz.PointAt(float64(i-1) / pointLength))
	}

	return bz
}

func NewBezierNA(points []math2.Vector2d) *Bezier {
	bz := &Bezier{points: points}
	bz.ApproxLength = 0.0
	return bz
}

func (bz *Bezier) PointAt(t float64) math2.Vector2d {
	x := 0.0
	y := 0.0
	n := len(bz.points) - 1
	for i := 0; i <= n; i++ {
		b := bernstein(int64(i), int64(n), t)
		x += bz.points[i].X * b
		y += bz.points[i].Y * b
	}
	return math2.NewVec2d(x, y)
}

func (bz Bezier) GetLength() float64 {
	return bz.ApproxLength
}

func (bz Bezier) GetStartAngle() float64 {
	return bz.points[0].AngleRV(bz.PointAt(1.0 / bz.ApproxLength))
}

func (bz Bezier) GetEndAngle() float64 {
	return bz.points[len(bz.points)-1].AngleRV(bz.PointAt((bz.ApproxLength - 1) / bz.ApproxLength))
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func BinomialCoefficient(n, k int64) int64 {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}
	k = min(k, n-k)
	var c int64 = 1
	var i int64 = 0
	for ; i < k; i++ {
		c = c * (n - i) / (i + 1)
	}

	return c
}

func bernstein(i, n int64, t float64) float64 {
	return float64(BinomialCoefficient(n, i)) * math.Pow(t, float64(i)) * math.Pow(1.0-t, float64(n-i))
}
