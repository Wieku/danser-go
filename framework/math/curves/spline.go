package curves

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/vector"
	"sort"
)

type Spline struct {
	sections   []float32
	path       []Curve
	length     float32
}

func NewSpline(curves []Curve) *Spline {
	sections := make([]float32, len(curves)+1)
	length := float32(0.0)

	for i, c := range curves {
		length += c.GetLength()
		sections[i+1] = length
	}

	return &Spline{sections, curves, length}
}

func NewSplineW(curves []Curve, widths []float32) *Spline {
	sections := make([]float32, len(curves)+1)
	length := float32(0.0)

	for i, c := range widths {
		length += c
		sections[i+1] = length
	}

	return &Spline{sections, curves, length}
}

func (spline *Spline) PointAt(t float32) vector.Vector2f {
	desiredWidth := spline.length * bmath.ClampF32(t, 0.0, 1.0)

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

func (spline *Spline) GetLength() float32 {
	return spline.length
}

func (spline *Spline) GetStartAngle() float32 {
	if len(spline.path) > 0 {
		return spline.path[0].GetStartAngle()
	}

	return 0.0
}

func (spline *Spline) getCurveAt(t float32) Curve {
	if len(spline.path) == 0 {
		return Linear{}
	}

	desiredWidth := spline.length * bmath.ClampF32(t, 0.0, 1.0)

	withoutFirst := spline.sections[1:]
	index := sort.Search(len(withoutFirst), func(i int) bool {
		return withoutFirst[i] >= desiredWidth
	})

	return spline.path[index]
}

func (spline *Spline) GetStartAngleAt(t float32) float32 {
	if len(spline.path) == 0 {
		return 0
	}

	return spline.getCurveAt(t).GetStartAngle()
}

func (spline *Spline) GetEndAngle() float32 {
	if len(spline.path) > 0 {
		return spline.path[len(spline.path)-1].GetEndAngle()
	}

	return 0.0
}

func (spline *Spline) GetEndAngleAt(t float32) float32 {
	if len(spline.path) == 0 {
		return 0
	}

	return spline.getCurveAt(t).GetEndAngle()
}

func (spline *Spline) GetCurves() []Curve {
	return spline.path
}
