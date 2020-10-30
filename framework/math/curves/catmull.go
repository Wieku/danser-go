package curves

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type Catmull struct {
	points       []vector.Vector2f
	ApproxLength float32
}

func NewCatmull(points []vector.Vector2f) Catmull {

	if len(points) != 4 {
		panic("4 points are needed to create centripetal catmull rom")
	}

	cm := &Catmull{points: points}

	pointLength := points[1].Dst(points[2])

	pointLength = math32.Ceil(pointLength)

	for i := 1; i <= int(pointLength); i++ {
		cm.ApproxLength += cm.PointAt(float32(i) / pointLength).Dst(cm.PointAt(float32(i-1) / pointLength))
	}

	return *cm
}

func (cm Catmull) PointAt(t float32) vector.Vector2f {
	return findPoint(cm.points[0], cm.points[1], cm.points[2], cm.points[3], t)
}

func findPoint(vec1, vec2, vec3, vec4 vector.Vector2f, t float32) vector.Vector2f {
	t2 := t * t
	t3 := t * t2

	return vector.NewVec2f(0.5*(2*vec2.X+(-vec1.X+vec3.X)*t+(2*vec1.X-5*vec2.X+4*vec3.X-vec4.X)*t2+(-vec1.X+3*vec2.X-3*vec3.X+vec4.X)*t3),
		0.5*(2*vec2.Y+(-vec1.Y+vec3.Y)*t+(2*vec1.Y-5*vec2.Y+4*vec3.Y-vec4.Y)*t2+(-vec1.Y+3*vec2.Y-3*vec3.Y+vec4.Y)*t3))
}

func (cm Catmull) GetLength() float32 {
	return cm.ApproxLength
}

func (cm Catmull) GetStartAngle() float32 {
	return cm.points[0].AngleRV(cm.PointAt(1.0 / cm.ApproxLength))
}

func (cm Catmull) GetEndAngle() float32 {
	return cm.points[len(cm.points)-1].AngleRV(cm.PointAt((cm.ApproxLength - 1) / cm.ApproxLength))
}
