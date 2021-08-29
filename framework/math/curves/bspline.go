package curves

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

// NewBSpline creates a spline that goes through all given control points.
// points[1] and points[len(points)-2] are terminal tangents.
func NewBSpline(points []vector.Vector2f) *Spline {
	beziers := SolveBSpline(points)
	beziersC := make([]Curve, len(beziers))

	for i, b := range beziers {
		b.CalculateLength()
		beziersC[i] = b
	}

	return NewSpline(beziersC)
}

// NewBSplineW creates a spline that goes through all given control points with forced weights(lengths), useful when control points have to be passed at certain times.
// points[1] and points[len(points)-2] are terminal tangents.
func NewBSplineW(points []vector.Vector2f, weights []float32) *Spline {
	beziers := SolveBSpline(points)
	beziersC := make([]Curve, len(beziers))

	for i, b := range beziers {
		beziersC[i] = b
	}

	return NewSplineW(beziersC, weights)
}

// SolveBSpline calculates the spline that goes through all given control points.
// points[1] and points[len(points)-2] are terminal tangents
// Returns an array of bezier curves in NA (non-approximated) version for performance considerations.
func SolveBSpline(points1 []vector.Vector2f) []*Bezier {
	pointsLen := len(points1)

	points := make([]vector.Vector2f, 0, pointsLen)

	points = append(points, points1[0])
	points = append(points, points1[2:pointsLen-2]...)
	points = append(points, points1[pointsLen-1], points1[1], points1[pointsLen-2])

	n := len(points) - 2

	d := make([]vector.Vector2f, n)
	d[0] = points[n].Sub(points[0])
	d[n-1] = points[n+1].Sub(points[n-1]).Scl(-1)

	A := make([]vector.Vector2f, len(points))
	Bi := make([]float32, len(points))

	Bi[1] = -0.25
	A[1] = points[2].Sub(points[0]).Sub(d[0]).Scl(1.0 / 4)
	for i := 2; i < n-1; i++ {
		Bi[i] = -1 / (4 + Bi[i-1])
		A[i] = points[i+1].Sub(points[i-1]).Sub(A[i-1]).Scl(-1 * Bi[i])
	}

	for i := n - 2; i > 0; i-- {
		d[i] = A[i].Add(d[i+1].Scl(Bi[i]))
	}

	bezierPoints := []vector.Vector2f{points[0], points[0].Add(d[0])}

	for i := 1; i < n-1; i++ {
		bezierPoints = append(bezierPoints, points[i].Sub(d[i]), points[i], points[i].Add(d[i]))
	}

	bezierPoints = append(bezierPoints, points[n-1].Sub(d[n-1]), points[n-1])

	var beziers []*Bezier
	for i := 0; i < len(bezierPoints)-3; i += 3 {
		beziers = append(beziers, NewBezierNA(bezierPoints[i:i+4]))
	}

	return beziers
}
