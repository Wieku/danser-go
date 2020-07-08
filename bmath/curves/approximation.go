package curves

import "github.com/wieku/danser-go/bmath"

func ApproximateCircularArc(pt1, pt2, pt3 bmath.Vector2d, detail float64) []Linear {
	arc := NewCirArc(pt1, pt2, pt3)

	if arc.Unstable {
		return []Linear{NewLinear(pt1, pt2), NewLinear(pt2, pt3)}
	}

	segments := int(arc.r * arc.totalAngle * detail)

	lines := make([]Linear, segments)

	for i := 0; i < segments; i++ {
		lines[i] = NewLinear(arc.PointAt(float64(i)/float64(segments)), arc.PointAt(float64(i+1)/float64(segments)))
	}

	return lines
}

func ApproximateCatmullRom(points []bmath.Vector2d, detail int) []Linear {
	catmull := NewCatmull(points)

	lines := make([]Linear, detail)

	for i := 0; i < detail; i++ {
		lines[i] = NewLinear(catmull.PointAt(float64(i)/float64(detail)), catmull.PointAt(float64(i+1)/float64(detail)))
	}

	return lines
}

func ApproximateBezier(points []bmath.Vector2d) []Linear {
	extracted := NewBezierApproximator(points).CreateBezier()

	lines := make([]Linear, len(extracted))

	for i := 0; i < len(extracted); i++ {
		lines[i] = NewLinear(extracted[i], extracted[i+1])
	}

	return lines
}
