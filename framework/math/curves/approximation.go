package curves

import (
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

func ApproximateCircularArc(pt1, pt2, pt3 vector.Vector2f, detail float32) []vector.Vector2f {
	arc := NewCirArc(pt1, pt2, pt3)

	if arc.Unstable {
		return []vector.Vector2f{pt1, pt2, pt3}
	}

	segments := int(float64(arc.r) * arc.totalAngle * float64(detail))

	points := []vector.Vector2f{pt1}

	for i := 1; i < segments; i++ {
		vector3 := arc.PointAt(float32(i) / float32(segments))

		points = append(points, vector3)
	}

	points = append(points, pt3)

	return points
}

func ApproximateCircularArcLazer(pt1, pt2, pt3 vector.Vector2f) []vector.Vector2f {
	arc := NewCirArc(pt1, pt2, pt3)

	if arc.Unstable {
		return []vector.Vector2f{pt1, pt2, pt3}
	}

	segments := 2
	if 2*arc.r > 0.1 {
		segments = max(2, int(math.Ceil(arc.totalAngle/(2*math.Acos(1-0.1/float64(arc.r))))))
	}

	pts := make([]vector.Vector2f, 0)

	for i := 0; i < segments; i++ {
		fract := float64(i) / float64(segments-1)
		pts = append(pts, arc.PointAt64(fract))
	}

	return pts
}

func ApproximateCatmullRom(points []vector.Vector2f, detail int) []vector.Vector2f {
	catmull := NewCatmull(points)

	outPoints := make([]vector.Vector2f, detail+1)

	for i := 0; i <= detail; i++ {
		outPoints[i] = catmull.PointAt(float32(i) / float32(detail))
	}

	return outPoints
}

func ApproximateBezier(points []vector.Vector2f) []vector.Vector2f {
	extracted := NewBezierApproximator(points).CreateBezier()
	return extracted
}
