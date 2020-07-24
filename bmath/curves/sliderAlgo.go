package curves

import (
	"github.com/wieku/danser-go/bmath"
)

const minPartWidth = 0.0001

type MultiCurve struct {
	sections []float64
	length   float64
	lines    []Linear
}

func NewMultiCurve(typ string, points []bmath.Vector2d, desiredLength float64) *MultiCurve {
	lines := make([]Linear, 0)

	if len(points) < 3 {
		typ = "L"
	}

	switch typ {
	case "P":
		lines = append(lines, ApproximateCircularArc(points[0], points[1], points[2], 0.125)...)
		break
	case "L":
		for i := 0; i < len(points)-1; i++ {
			lines = append(lines, NewLinear(points[i], points[i+1]))
		}
		break
	case "B":
		lastIndex := 0
		for i, p := range points {
			if (i == len(points)-1 && p != points[i-1]) || (i < len(points)-1 && points[i+1] == p) {
				pts := points[lastIndex : i+1]

				if len(pts) > 2 {
					lines = append(lines, ApproximateBezier(pts)...)
				} else if len(pts) == 1 {
					lines = append(lines, NewLinear(pts[0], pts[0]))
				} else {
					lines = append(lines, NewLinear(pts[0], pts[1]))
				}

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
			lines = append(lines, ApproximateCatmullRom(points[i:i+4], 50)...)
		}
		break
	}

	length := 0.0

	for _, l := range lines {
		length += l.GetLength()
	}

	if desiredLength >= 0 {
		if length > desiredLength {
			diff := length - desiredLength
			length -= diff
			for i := len(lines) - 1; i >= 0 && diff > 0.0; i-- {
				line := lines[i]

				if line.GetLength() >= diff+minPartWidth {
					pt := line.PointAt((line.GetLength() - diff) / line.GetLength())
					lines[i] = NewLinear(line.Point1, pt)
					break
				}

				diff -= line.GetLength()
				lines = lines[:len(lines)-1]
			}

		} else if desiredLength > length {
			last := lines[len(lines)-1]

			p1 := last.PointAt(1)
			p2 := bmath.NewVec2dRad(last.GetEndAngle(), desiredLength-length).Add(p1)

			c := NewLinear(p1, p2)

			length += c.GetLength()
			lines = append(lines, c)
		}
	}

	sections := make([]float64, len(lines)+1)
	sections[0] = 0.0
	prev := 0.0

	for i := 0; i < len(lines); i++ {
		prev += lines[i].GetLength()
		sections[i+1] = prev
	}

	return &MultiCurve{sections, length, lines}
}

func (sa *MultiCurve) PointAt(t float64) bmath.Vector2d {

	desiredWidth := sa.length * t

	lineI := len(sa.sections) - 2

	for i, k := range sa.sections[:len(sa.sections)-1] {
		if k <= desiredWidth {
			lineI = i
		}
	}

	line := sa.lines[lineI]

	return line.PointAt((desiredWidth - sa.sections[lineI]) / (sa.sections[lineI+1] - sa.sections[lineI]))
}

func (sa *MultiCurve) GetLength() float64 {
	return sa.length
}

func (sa *MultiCurve) GetStartAngle() float64 {
	return sa.lines[0].GetStartAngle()
}

func (sa *MultiCurve) GetEndAngle() float64 {
	return sa.lines[len(sa.lines)-1].GetEndAngle()
}

func (ln *MultiCurve) GetLines() []Linear {
	return ln.lines
}
