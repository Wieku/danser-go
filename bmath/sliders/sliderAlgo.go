package sliders

import (
	m2 "github.com/wieku/danser/bmath"
	"github.com/wieku/danser/bmath/curves"
)

type SliderAlgo struct {
	curves []curves.Curve
	sections []float64
	Length float64
}

func NewSliderAlgo(typ string, points []m2.Vector2d) SliderAlgo {
	var curveList []curves.Curve

	var length float64 = 0.0
	if len(points) < 3 {
		typ = "L"
	}
	switch typ {
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
			if i == len(points) - 1 || points[i+1] == p {
				c := curves.NewBezier(points[lastIndex:i+1])
				curveList = append(curveList, c)
				length += c.GetLength()
				lastIndex = i+1
			}
		}
		break
	case "P":
		c := curves.NewCirArc(points[0], points[1], points[2])
		curveList = append(curveList, c)
		length += c.GetLength()
		break
	}

	sections := make([]float64, len(curveList)+1)
	sections[0] = 0.0
	prev := 0.0
	if len(curveList) > 1 {
		for i := 0; i < len(curveList); i++ {
			prev += curveList[i].GetLength()/length
			sections[i+1] = prev
		}
	}

	return SliderAlgo{curveList, sections, length}
}

func (sa SliderAlgo) PointAt(t float64) m2.Vector2d {
	if len(sa.curves) == 1 {
		return sa.curves[0].PointAt(t)
	} else {
		t = sa.sections[len(sa.sections)-1] * t
		for i := 1; i < len(sa.sections); i++ {
			if t <= sa.sections[i] || i == len(sa.sections) - 1 {
				prc := (t - sa.sections[i-1]) / (sa.sections[i] - sa.sections[i-1])
				return sa.curves[i-1].PointAt(prc)
			}
		}
	}

	return m2.NewVec2d(512/2,384/2)
}

