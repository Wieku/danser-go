package curves

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

func ApproximateCircularArc(pt1, pt2, pt3 vector.Vector2f, detail float32) []Linear {
	arc := NewCirArc(pt1, pt2, pt3)

	if arc.Unstable {
		return []Linear{NewLinear(pt1, pt2), NewLinear(pt2, pt3)}
	}

	segments := int(arc.r * arc.totalAngle * detail)

	lines := make([]Linear, segments)

	for i := 0; i < segments; i++ {
		lines[i] = NewLinear(arc.PointAt(float32(i)/float32(segments)), arc.PointAt(float32(i+1)/float32(segments)))
	}

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
