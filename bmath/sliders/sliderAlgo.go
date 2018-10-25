package sliders

import (
	m2 "github.com/wieku/danser/bmath"
	"github.com/wieku/danser/bmath/curves"
	"github.com/wieku/danser/bmath"
)

type SliderAlgo struct {
	curves   []curves.Curve
	sections []float64
	length   float64
	scale    float64
}

func NewSliderAlgo(typ string, points []m2.Vector2d, desiredLength float64) SliderAlgo {
	var curveList []curves.Curve

	length := 0.0
	if len(points) < 3 {
		typ = "L"
	}
	switch typ {
	case "P":
		c := curves.NewCirArc(points[0], points[1], points[2])
		if !c.Unstable {
			curveList = append(curveList, c)
			length += c.GetLength()
			break
		}
		fallthrough
	case "L":
		for i := 1; i < len(points); i++ {
			c := curves.NewLinear(points[i-1], points[i])
			curveList = append(curveList, c)
			length += c.GetLength()
		}
		break
	case "B":
		lastIndex := 0
		for i, p := range points {
			if (i == len(points)-1 && p != points[i-1]) || (i < len(points)-1 && points[i+1] == p) {
				pts := points[lastIndex : i+1]
				var c curves.Curve
				if len(pts) > 2 {
					c = curves.NewBezier(pts)
				} else if len(pts) == 1 {
					c = curves.NewLinear(pts[0], pts[0])
				} else {
					c = curves.NewLinear(pts[0], pts[1])
				}

				curveList = append(curveList, c)
				length += c.GetLength()
				lastIndex = i + 1
			}
		}
		break
	case "C":

		if points[0] != points[1] {
			points = append([]bmath.Vector2d{points[0]}, points...)
		}

		if points[len(points)-1] != points[len(points)-2] {
			points = append(points, points[len(points)-1])
		}

		for i := 0; i < len(points)-3; i++ {
			c := curves.NewCatmull(points[i : i+4])
			curveList = append(curveList, c)
			length += c.GetLength()
		}
		break
	}

	scale := -1.0

	if length > desiredLength {
		scale = desiredLength / length
	} else if desiredLength > length {
		last := curveList[len(curveList)-1]
		p2 := last.PointAt(1)
		p3 := bmath.NewVec2dRad(last.GetEndAngle(), desiredLength-length).Add(p2)
		c := curves.NewLinear(p2, p3)
		curveList = append(curveList, c)
		length += c.GetLength()
	}

	sections := make([]float64, len(curveList)+1)
	sections[0] = 0.0
	prev := 0.0
	if len(curveList) > 1 {
		for i := 0; i < len(curveList); i++ {
			prev += curveList[i].GetLength() / length
			sections[i+1] = prev
		}
	}

	return SliderAlgo{curveList, sections, length, scale}
}

func (sa *SliderAlgo) PointAt(t float64) m2.Vector2d {
	if sa.scale > -0.5 {
		t *= sa.scale
	}
	if len(sa.curves) == 1 {
		return sa.curves[0].PointAt(t)
	} else {
		t = sa.sections[len(sa.sections)-1] * t
		for i := 1; i < len(sa.sections); i++ {
			if t <= sa.sections[i] || i == len(sa.sections)-1 {
				prc := (t - sa.sections[i-1]) / (sa.sections[i] - sa.sections[i-1])
				return sa.curves[i-1].PointAt(prc)
			}
		}
	}

	return m2.NewVec2d(512/2, 384/2)
}

func (sa *SliderAlgo) GetLength() float64 {
	if sa.scale > -0.5 {
		return sa.length * sa.scale
	}
	return sa.length
}
