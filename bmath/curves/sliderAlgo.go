package curves

import (
	"github.com/wieku/danser-go/bmath"
	m2 "github.com/wieku/danser-go/bmath"
	"log"
)

type MultiCurve struct {
	curves   []Curve
	sections []float64
	length   float64
	scale    float64
	lines 	[]Linear
}

func NewMultiCurve(typ string, points []m2.Vector2d, desiredLength float64, customTiming []float64) MultiCurve {
	var curveList []Curve
	var lines []Linear

	length := 0.0
	if len(points) < 3 {
		typ = "L"
	}

	index := 0

	switch typ {
	case "P":
		c := NewCirArc(points[0], points[1], points[2])
		if !c.Unstable {
			curveList = append(curveList, c)
			length += c.GetLength()
			lines = append(lines, c.GetLines()...)
			break
		}
		fallthrough
	case "L":
		for i := 1; i < len(points); i++ {
			c := NewLinear(points[i-1], points[i])
			curveList = append(curveList, c)
			if customTiming == nil {
				length += c.GetLength()
			} else {
				length += customTiming[index]
			}

			lines = append(lines, c.GetLines()...)

			index++
		}
		break
	case "B":
		lastIndex := 0
		for i, p := range points {
			if (i == len(points)-1 && p != points[i-1]) || (i < len(points)-1 && points[i+1] == p) {
				pts := points[lastIndex : i+1]
				var c Curve
				if len(pts) > 2 {
					c = NewBezier(pts)
				} else if len(pts) == 1 {
					c = NewLinear(pts[0], pts[0])
				} else {
					c = NewLinear(pts[0], pts[1])
				}

				curveList = append(curveList, c)

				lines = append(lines, c.GetLines()...)

				if customTiming == nil {
					length += c.GetLength()
				} else {
					length += customTiming[index]
				}

				index++
				lastIndex = i + 1
			}
		}
		break
	case "CB":
		for i := 0; i < len(points)-3; i+=3 {
			c := NewBezier(points[i : i+4])
			curveList = append(curveList, c)
			if customTiming == nil {
				length += c.GetLength()
			} else {
				length += customTiming[index]
			}

			lines = append(lines, c.GetLines()...)

			index++
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
			c := NewCatmull(points[i : i+4])
			curveList = append(curveList, c)
			lines = append(lines, c.GetLines()...)
			if customTiming == nil {
				length += c.GetLength()
			} else {
				length += customTiming[index]
			}

			index++
		}
		break
	}

	scale := -1.0

	if desiredLength >= 0 {
		if length > desiredLength {
			log.Println("shortening", desiredLength, length, len(lines))
			scale = desiredLength / length
			diff := length - desiredLength
			length -= diff
			for i := len(lines) - 1; i >= 0 && diff > 0.0 ; i-- {
				line := lines[i]
				log.Println(i, line, line.GetLength(), diff)
				if diff >= line.GetLength() {
					diff -= line.GetLength()
					lines = lines[:len(lines) - 1]
				} else {
					log.Println((line.GetLength()-diff) / line.GetLength())
					log.Println(line.PointAt((line.GetLength()-diff) / line.GetLength()))
					pt := line.PointAt((line.GetLength()-diff) / line.GetLength())
					lines[i] = Linear{line.Point1, pt}
					break
				}
			}

			log.Println(lines[len(lines) - 1])

		} else if desiredLength > length {
			log.Println("lenghtening", desiredLength)
			last := lines[len(lines)-1]
			p2 := last.PointAt(1)
			p3 := bmath.NewVec2dRad(last.GetEndAngle(), desiredLength-length).Add(p2)
			c := NewLinear(p2, p3)
			curveList = append(curveList, c)
			lines = append(lines, c)
			length += c.GetLength()
		}
	}

	sections := make([]float64, len(lines)+1)
	sections[0] = 0.0
	prev := 0.0
	//if len(lines) > 1 {
		for i := 0; i < len(lines); i++ {
			if customTiming == nil {
				prev += lines[i].GetLength()
			} else {
				prev += customTiming[i] / length
			}
			sections[i+1] = prev
		}
	//}

	return MultiCurve{curveList, sections, length, scale, lines}
}

func (sa *MultiCurve) PointAt(t float64) m2.Vector2d {
	/*if sa.scale > -0.5 {
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
	}*/

	desiredWidth := sa.length * t

	//log.Println(sa.sections, sa.length)

	lineI := len(sa.sections)-2
	//log.Println(lineI)
	for i, k := range sa.sections[:len(sa.sections)-1]  {
		if k <= desiredWidth {
			lineI = i
		}
	}

	//log.Println(lineI)


	//lineI := sort.SearchFloat64s(bz.sections[:len(bz.sections)-2], desiredWidth)

	//log.Println(lineI, len(bz.sections), len(bz.lines))
	line := sa.lines[lineI]

	/*log.Println(line.Point1, line.Point2)

	log.Println((desiredWidth-sa.sections[lineI])/(sa.sections[lineI+1]-sa.sections[lineI]))
	log.Println(sa.sections[lineI+1]-sa.sections[lineI])
	log.Println(sa.sections[lineI+1], sa.sections[lineI])
	log.Println(line.PointAt((desiredWidth-sa.sections[lineI])/(sa.sections[lineI+1]-sa.sections[lineI])))
	log.Println()*/

	return line.PointAt((desiredWidth-sa.sections[lineI])/(sa.sections[lineI+1]-sa.sections[lineI]))

	//return m2.NewVec2d(512/2, 384/2)
}

func (sa *MultiCurve) NPointAt(t float64) m2.Vector2d {
	if sa.scale > -0.5 {
		t *= sa.scale
	}
	if len(sa.curves) == 1 {
		return sa.curves[0].NPointAt(t)
	} else {
		t = sa.sections[len(sa.sections)-1] * t
		for i := 1; i < len(sa.sections); i++ {
			if t <= sa.sections[i] || i == len(sa.sections)-1 {
				prc := (t - sa.sections[i-1]) / (sa.sections[i] - sa.sections[i-1])
				return sa.curves[i-1].NPointAt(prc)
			}
		}
	}

	return m2.NewVec2d(512/2, 384/2)
}

func (sa *MultiCurve) GetLength() float64 {
	/*if sa.scale > -0.5 {
		return sa.length * sa.scale
	}*/
	return sa.length
}

func (sa *MultiCurve) GetStartAngle() float64 {
	return sa.curves[0].GetStartAngle()
}

func (sa *MultiCurve) GetEndAngle() float64 {
	return sa.curves[len(sa.curves)-1].GetEndAngle()
}

func (ln *MultiCurve) GetLines() []Linear {
	return ln.lines
}