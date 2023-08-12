package curves

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type Bezier struct {
	Points        []vector.Vector2f
	controlLength float32
	ApproxLength  float32
}

// Creates a bezier curve with approximated length
func NewBezier(points []vector.Vector2f) *Bezier {
	bz := NewBezierNA(points)
	bz.CalculateLength()

	return bz
}

// Creates a bezier curve with non-approximated length.
// To calculate that length, call (*Bezier).CalculateLength()
func NewBezierNA(points []vector.Vector2f) *Bezier {
	bz := &Bezier{Points: points}

	for i := 1; i < len(bz.Points); i++ {
		bz.controlLength += bz.Points[i].Dst(bz.Points[i-1])
	}

	bz.ApproxLength = bz.controlLength

	return bz
}

// Calculates the approximate length of the curve to 2 decimal points of accuracy in most cases
func (bz *Bezier) CalculateLength() {
	length := float32(0.0)

	sections := math32.Ceil(bz.controlLength)

	previous := bz.Points[0]
	for i := 1; i <= int(sections); i++ {
		current := bz.PointAt(float32(i) / sections)

		length += current.Dst(previous)

		previous = current
	}

	bz.ApproxLength = length
}

// https://en.wikipedia.org/wiki/B%C3%A9zier_curve#Terminology
func (bz *Bezier) PointAt(t float32) (p vector.Vector2f) {
	n := len(bz.Points) - 1
	for i := 0; i <= n; i++ {
		b := bernstein(int64(i), int64(n), t)
		p.X += bz.Points[i].X * b
		p.Y += bz.Points[i].Y * b
	}

	return
}

func (bz *Bezier) GetLength() float32 {
	return bz.ApproxLength
}

func (bz *Bezier) GetStartAngle() float32 {
	return bz.Points[0].AngleRV(bz.PointAt(1.0 / bz.controlLength))
}

func (bz *Bezier) GetEndAngle() float32 {
	return bz.Points[len(bz.Points)-1].AngleRV(bz.PointAt(1.0 - 1.0/bz.controlLength))
}

// https://en.wikipedia.org/wiki/Binomial_coefficient#Multiplicative_formula
func BinomialCoefficient(n, k int64) int64 {
	if k < 0 || k > n {
		return 0
	}

	if k == 0 || k == n {
		return 1
	}

	k = min(k, n-k)

	c := int64(1)
	for i := int64(1); i <= k; i++ {
		c *= (n + 1 - i) / (i)
	}

	return c
}

// https://en.wikipedia.org/wiki/Bernstein_polynomial
func bernstein(i, n int64, t float32) float32 {
	return float32(BinomialCoefficient(n, i)) * math32.Pow(t, float32(i)) * math32.Pow(1.0-t, float32(n-i))
}
