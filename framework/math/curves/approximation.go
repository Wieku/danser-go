package curves

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

func ApproximateCircularArc(pt1, pt2, pt3 vector.Vector2f, detail float32) []Linear {
	arc := NewCirArc(pt1, pt2, pt3)

	if arc.Unstable {
		return []Linear{NewLinear(pt1, pt2), NewLinear(pt2, pt3)}
	}

	segments := int(float64(arc.r) * arc.totalAngle * float64(detail))

	lines := make([]Linear, 0)

	p := pt1

	for i := 1; i < segments; i++ {
		vector3 := arc.PointAt(float32(i) / float32(segments))

		lines = append(lines, NewLinear(p, vector3))

		p = vector3
	}

	lines = append(lines, NewLinear(p, pt3))

	return lines
}

func ApproximateCatmullRom(points []vector.Vector2f, detail int) []Linear {
	catmull := NewCatmull(points)

	lines := make([]Linear, detail)

	for i := 0; i < detail; i++ {
		lines[i] = NewLinear(catmull.PointAt(float32(i)/float32(detail)), catmull.PointAt(float32(i+1)/float32(detail)))
	}

	return lines
}

func ApproximateBezier(points []vector.Vector2f) []Linear {
	extracted := NewBezierApproximator(points).CreateBezier()

	lines := make([]Linear, len(extracted)-1)

	for i := 0; i < len(lines); i++ {
		lines[i] = NewLinear(extracted[i], extracted[i+1])
	}

	return lines
}
