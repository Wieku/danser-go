package curves

import (
	math2 "github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath"
	"math"
)

type Catmull struct {
	points       []math2.Vector2d
	ApproxLength float64
}

func NewCatmull(points []math2.Vector2d) Catmull {

	if len(points) != 4 {
		panic("4 points are needed to create centripetal catmull rom")
	}

	cm := &Catmull{points: points}

	pointLength := points[1].Dst(points[2])

	pointLength = math.Ceil(pointLength)

	for i := 1; i <= int(pointLength); i++ {
		cm.ApproxLength += cm.NPointAt(float64(i) / pointLength).Dst(cm.NPointAt(float64(i-1) / pointLength))
	}

	return *cm
}

func (cm Catmull) NPointAt(t float64) math2.Vector2d {
	return findPoint(cm.points[0], cm.points[1], cm.points[2], cm.points[3], t)
}

func findPoint(vec1, vec2, vec3, vec4 bmath.Vector2d, t float64) bmath.Vector2d {
	t2 := t * t
	t3 := t * t2

	return bmath.NewVec2d(0.5*(2*vec2.X+(-vec1.X+vec3.X)*t+(2*vec1.X-5*vec2.X+4*vec3.X-vec4.X)*t2+(-vec1.X+3*vec2.X-3*vec3.X+vec4.X)*t3),
		0.5*(2*vec2.Y+(-vec1.Y+vec3.Y)*t+(2*vec1.Y-5*vec2.Y+4*vec3.Y-vec4.Y)*t2+(-vec1.Y+3*vec2.Y-3*vec3.Y+vec4.Y)*t3))
}

//It's not a neat solution, but it works
//This calculates point on catmull curve with constant velocity
func (cm Catmull) PointAt(t float64) math2.Vector2d {
	desiredWidth := cm.ApproxLength * t
	width := 0.0
	pos := cm.points[1]
	c := 0.0
	for width < desiredWidth {
		pt := cm.NPointAt(c)
		width += pt.Dst(pos)
		if width > desiredWidth {
			return pos
		}
		pos = pt
		c += 1.0 / float64(cm.ApproxLength*2-1)
	}

	return pos
}

func (cm Catmull) GetLength() float64 {
	return cm.ApproxLength
}

func (cm Catmull) GetStartAngle() float64 {
	return cm.points[0].AngleRV(cm.NPointAt(1.0 / cm.ApproxLength))
}

func (cm Catmull) GetEndAngle() float64 {
	return cm.points[len(cm.points)-1].AngleRV(cm.NPointAt((cm.ApproxLength - 1) / cm.ApproxLength))
}

func (ln Catmull) GetPoints(num int) []math2.Vector2d {
	t0 := 1 / float64(num-1)

	points := make([]math2.Vector2d, num)
	t := 0.0
	for i := 0; i < num; i += 1 {
		points[i] = ln.PointAt(t)
		t += t0
	}

	return points
}
