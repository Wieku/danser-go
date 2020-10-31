package curves

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/vector"
	"sort"
)

const minPartWidth = 0.0001

type MultiCurve struct {
	sections   []float32
	lines      []Linear
	length     float32
	firstPoint vector.Vector2f
}

func NewMultiCurve(typ string, points []vector.Vector2f, desiredLength float64) *MultiCurve {
	lines := make([]Linear, 0)

	if len(points) < 3 {
		typ = "L"
	}

	switch typ {
	case "P":
		lines = append(lines, ApproximateCircularArc(points[0], points[1], points[2], 0.125)...)
	case "L":
		for i := 0; i < len(points)-1; i++ {
			lines = append(lines, NewLinear(points[i], points[i+1]))
		}
	case "B":
		lastIndex := 0

		for i := 0; i < len(points); i++ {
			multi := i < len(points)-2 && points[i] == points[i+1]

			if multi || i == len(points)-1 {
				subPoints := points[lastIndex : i+1]

				if len(subPoints) > 2 {
					lines = append(lines, ApproximateBezier(subPoints)...)
				} else if len(subPoints) == 2 {
					lines = append(lines, NewLinear(subPoints[0], subPoints[1]))
				}

				if multi {
					i++
				}

				lastIndex = i
			}
		}
	case "C":
		if points[0] != points[1] {
			points = append([]vector.Vector2f{points[0]}, points...)
		}

		if points[len(points)-1] != points[len(points)-2] {
			points = append(points, points[len(points)-1])
		}

		for i := 0; i < len(points)-3; i++ {
			lines = append(lines, ApproximateCatmullRom(points[i:i+4], 50)...)
		}
	}

	length := float32(0.0)

	for _, l := range lines {
		length += l.GetLength()
	}

	firstPoint := points[0]

	diff := float64(length) - desiredLength

	for len(lines) > 0 {
		line := lines[len(lines)-1]

		if float64(line.GetLength()) > diff+minPartWidth {
			if line.Point1 != line.Point2 {
				pt := line.PointAt((line.GetLength() - float32(diff)) / line.GetLength())
				lines[len(lines)-1] = NewLinear(line.Point1, pt)
			}

			break
		}

		diff -= float64(line.GetLength())
		lines = lines[:len(lines)-1]
	}

	length = 0.0

	for _, l := range lines {
		length += l.GetLength()
	}

	sections := make([]float32, len(lines)+1)
	sections[0] = 0.0
	prev := float32(0.0)

	for i := 0; i < len(lines); i++ {
		prev += lines[i].GetLength()
		sections[i+1] = prev
	}

	return &MultiCurve{sections, lines, length, firstPoint}
}

func (mCurve *MultiCurve) PointAt(t float32) vector.Vector2f {
	if len(mCurve.lines) == 0 {
		return mCurve.firstPoint
	}

	desiredWidth := mCurve.length * bmath.ClampF32(t, 0.0, 1.0)

	withoutFirst := mCurve.sections[1:]
	index := sort.Search(len(withoutFirst), func(i int) bool {
		return withoutFirst[i] >= desiredWidth
	})

	index = bmath.MinI(index, len(mCurve.lines)-1)

	return mCurve.lines[index].PointAt((desiredWidth - mCurve.sections[index]) / (mCurve.sections[index+1] - mCurve.sections[index]))
}

func (mCurve *MultiCurve) GetLength() float32 {
	return mCurve.length
}

func (mCurve *MultiCurve) GetStartAngle() float32 {
	if len(mCurve.lines) > 0 {
		return mCurve.lines[0].GetStartAngle()
	}

	return 0.0
}

func (mCurve *MultiCurve) getLineAt(t float32) Linear {
	if len(mCurve.lines) == 0 {
		return Linear{}
	}

	desiredWidth := mCurve.length * bmath.ClampF32(t, 0.0, 1.0)

	withoutFirst := mCurve.sections[1:]
	index := sort.Search(len(withoutFirst), func(i int) bool {
		return withoutFirst[i] >= desiredWidth
	})

	return mCurve.lines[index]
}

func (mCurve *MultiCurve) GetStartAngleAt(t float32) float32 {
	if len(mCurve.lines) == 0 {
		return 0
	}

	return mCurve.getLineAt(t).GetStartAngle()
}

func (mCurve *MultiCurve) GetEndAngle() float32 {
	if len(mCurve.lines) > 0 {
		return mCurve.lines[len(mCurve.lines)-1].GetEndAngle()
	}

	return 0.0
}

func (mCurve *MultiCurve) GetEndAngleAt(t float32) float32 {
	if len(mCurve.lines) == 0 {
		return 0
	}

	return mCurve.getLineAt(t).GetEndAngle()
}

func (mCurve *MultiCurve) GetLines() []Linear {
	return mCurve.lines
}
