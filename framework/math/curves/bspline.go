package curves

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/vector"
	"sort"
)

type BSpline struct {
	points       []vector.Vector2f
	sections     []float32
	path         []*Bezier
	ApproxLength float32
}

func NewBSpline(points []vector.Vector2f, timing []int64) *BSpline {
	newTiming := make([]float32, len(timing))
	for i := range newTiming {
		newTiming[i] = float32(timing[i] - timing[0])
	}

	spline := &BSpline{
		points:   points,
		sections: newTiming,
		path:     SolveBSpline(points),
	}

	for i, b := range spline.path {
		d := spline.sections[i+1] - spline.sections[i]

		scl := float32(1.0)
		if d > 600 {
			scl = d / 2
		}

		b.Points[1] = b.Points[0].Add(b.Points[1].Sub(b.Points[0]).SclOrDenorm(scl))
		b.Points[2] = b.Points[3].Add(b.Points[2].Sub(b.Points[3]).SclOrDenorm(scl))
	}

	spline.ApproxLength = spline.sections[len(spline.sections)-1]

	return spline
}

func (spline *BSpline) PointAt(t float32) vector.Vector2f {
	desiredWidth := spline.ApproxLength * bmath.ClampF32(t, 0.0, 1.0)

	withoutFirst := spline.sections[1:]
	index := sort.Search(len(withoutFirst), func(i int) bool {
		return withoutFirst[i] >= desiredWidth
	})

	index = bmath.MinI(index, len(spline.path)-1)

	if spline.sections[index+1]-spline.sections[index] == 0 {
		return spline.path[index].PointAt(0)
	}

	return spline.path[index].PointAt((desiredWidth - spline.sections[index]) / (spline.sections[index+1] - spline.sections[index]))
}

func (spline *BSpline) GetLength() float32 {
	return spline.ApproxLength
}

func (spline *BSpline) GetStartAngle() float32 {
	return spline.path[0].GetStartAngle()
}

func (spline *BSpline) GetEndAngle() float32 {
	return spline.path[len(spline.path)-1].GetEndAngle()
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

	subPoints := []vector.Vector2f{points[0], points[0].Add(d[0])}

	for i := 1; i < n-1; i++ {
		subPoints = append(subPoints, points[i].Sub(d[i]), points[i], points[i].Add(d[i]))
	}

	subPoints = append(subPoints, points[n-1].Sub(d[n-1]), points[n-1])

	var beziers []*Bezier
	for i := 0; i < len(subPoints)-3; i += 3 {
		beziers = append(beziers, NewBezierNA(subPoints[i:i+4]))
	}

	return beziers
}
