package curves

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

type BSpline struct {
	points       []vector.Vector2f
	timing       []float32
	subPoints    []vector.Vector2f
	path         []*Bezier
	ApproxLength float32
}

func NewBSpline(points1 []vector.Vector2f, timing []int64) *BSpline {

	pointsLen := len(points1)

	points := make([]vector.Vector2f, 0)

	points = append(points, points1[0])
	points = append(points, points1[2:pointsLen-2]...)
	points = append(points, points1[pointsLen-1], points1[1], points1[pointsLen-2])

	newTiming := make([]float32, len(timing))
	for i := range newTiming {
		newTiming[i] = float32(timing[i] - timing[0])
	}

	spline := &BSpline{points: points, timing: newTiming}

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

	converted := make([]float32, len(timing))

	for i, time := range timing {
		if i > 0 {
			converted[i-1] = float32(time - timing[i-1])
		}
	}

	firstMul := float32(1.0)
	if converted[0] > 600 {
		firstMul = converted[0] / 2
	}

	secondMul := float32(1.0)

	spline.subPoints = append(spline.subPoints, points[0], points[0].Add(d[0].SclOrDenorm(firstMul)))

	for i := 1; i < n-1; i++ {
		if converted[i] > 600 {
			secondMul = converted[i] / 2
		} else {
			secondMul = 1.0
		}

		spline.subPoints = append(spline.subPoints, points[i].Sub(d[i].SclOrDenorm(firstMul)), points[i], points[i].Add(d[i].SclOrDenorm(secondMul)))
		firstMul = secondMul
	}

	spline.subPoints = append(spline.subPoints, points[len(points)-3].Sub(d[n-1].SclOrDenorm(firstMul)), points[len(points)-3])

	spline.ApproxLength = spline.timing[len(spline.timing)-1]

	for i := 0; i < len(spline.subPoints)-3; i += 3 {
		c := NewBezierNA(spline.subPoints[i : i+4])
		spline.path = append(spline.path, c)
	}

	return spline
}

func (spline *BSpline) PointAt(t float32) vector.Vector2f {
	desiredWidth := spline.ApproxLength * t

	lineI := len(spline.timing) - 2

	for i, k := range spline.timing[:len(spline.timing)-1] {
		if k <= desiredWidth {
			lineI = i
		}
	}

	line := spline.path[lineI]

	return line.PointAt((desiredWidth - spline.timing[lineI]) / (spline.timing[lineI+1] - spline.timing[lineI]))
}

func (spline *BSpline) GetLength() float32 {
	return spline.ApproxLength
}

func (spline *BSpline) GetStartAngle() float32 {
	return spline.points[0].AngleRV(spline.PointAt(1.0 / spline.ApproxLength))
}

func (spline *BSpline) GetEndAngle() float32 {
	return spline.points[len(spline.points)-1].AngleRV(spline.PointAt((spline.ApproxLength - 1) / spline.ApproxLength))
}
