package curves

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"sort"
)

type CType int

const (
	CLine = CType(iota)
	CBezier
	CCirArc
	CCatmull
)

type CurveDef struct {
	CurveType CType
	Points    []vector.Vector2f
}

const minPartWidth = 0.0001

type MultiCurve struct {
	sections   []float32
	lines      []Linear
	length     float32
	firstPoint vector.Vector2f
}

func NewMultiCurve(curveDefs []CurveDef) *MultiCurve {
	lines := make([]Linear, 0)

	for _, def := range curveDefs {
		var cLines []Linear

		switch def.CurveType {
		case CCirArc:
			cLines = processPerfect(def.Points)
		case CLine:
			cLines = processLinear(def.Points)
		case CBezier:
			cLines = processBezier(def.Points)
		case CCatmull:
			cLines = processCatmull(def.Points)
		}

		nLines := make([]Linear, len(lines)+len(cLines))
		copy(nLines, lines)
		copy(nLines[len(lines):], cLines)
		lines = nLines
	}

	length := float32(0.0)

	for _, l := range lines {
		length += l.GetLength()
	}

	firstPoint := curveDefs[0].Points[0]

	sections := make([]float32, len(lines)+1)
	sections[0] = 0.0
	prev := float32(0.0)

	for i := 0; i < len(lines); i++ {
		prev += lines[i].GetLength()
		sections[i+1] = prev
	}

	return &MultiCurve{sections, lines, length, firstPoint}
}

func NewMultiCurveT(curveDefs []CurveDef, desiredLength float64) *MultiCurve {
	mCurve := NewMultiCurve(curveDefs)

	if mCurve.length > 0 && desiredLength != 0 {
		diff := float64(mCurve.length) - desiredLength

		for len(mCurve.lines) > 0 {
			line := mCurve.lines[len(mCurve.lines)-1]

			if float64(line.GetLength()) > diff+minPartWidth {
				if line.Point1 != line.Point2 {
					pt := line.PointAt((line.GetLength() - float32(diff)) / line.GetLength())
					mCurve.lines[len(mCurve.lines)-1] = NewLinear(line.Point1, pt)
				}

				break
			}

			diff -= float64(line.GetLength())
			mCurve.lines = mCurve.lines[:len(mCurve.lines)-1]
		}
	}

	mCurve.length = 0.0

	for _, l := range mCurve.lines {
		mCurve.length += l.GetLength()
	}

	mCurve.sections = make([]float32, len(mCurve.lines)+1)
	mCurve.sections[0] = 0.0
	prev := float32(0.0)

	for i := 0; i < len(mCurve.lines); i++ {
		prev += mCurve.lines[i].GetLength()
		mCurve.sections[i+1] = prev
	}

	return mCurve
}

func (mCurve *MultiCurve) PointAt(t float32) vector.Vector2f {
	if len(mCurve.lines) == 0 || mCurve.length == 0 {
		return mCurve.firstPoint
	}

	desiredWidth := mCurve.length * mutils.Clamp(t, 0.0, 1.0)

	withoutFirst := mCurve.sections[1:]
	index := sort.Search(len(withoutFirst), func(i int) bool {
		return withoutFirst[i] >= desiredWidth
	})

	index = min(index, len(mCurve.lines)-1)

	if mCurve.sections[index+1]-mCurve.sections[index] == 0 {
		return mCurve.lines[index].Point1
	}

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

	desiredWidth := mCurve.length * mutils.Clamp(t, 0.0, 1.0)

	withoutFirst := mCurve.sections[1:]
	index := sort.Search(len(withoutFirst), func(i int) bool {
		return withoutFirst[i] >= desiredWidth
	})

	index = min(index, len(mCurve.lines)-1)

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

func processPerfect(points []vector.Vector2f) (lines []Linear) {
	if len(points) > 3 {
		lines = processBezier(points)
	} else if len(points) < 3 || vector.IsStraightLine32(points[0], points[1], points[2]) {
		lines = processLinear(points)
	} else {
		lines = append(lines, ApproximateCircularArc(points[0], points[1], points[2], 0.125)...)
	}

	return
}

func processLinear(points []vector.Vector2f) (lines []Linear) {
	for i := 0; i < len(points)-1; i++ {
		if points[i] == points[i+1] { // skip red anchors, present in old maps like 243
			continue
		}

		lines = append(lines, NewLinear(points[i], points[i+1]))
	}

	return
}

func processBezier(points []vector.Vector2f) (lines []Linear) {
	lastIndex := 0

	for i := 0; i < len(points); i++ {
		multi := i < len(points)-2 && points[i] == points[i+1]

		if multi || i == len(points)-1 {
			subPoints := points[lastIndex : i+1]

			if len(subPoints) == 2 {
				lines = append(lines, NewLinear(subPoints[0], subPoints[1]))
			} else {
				lines = append(lines, ApproximateBezier(subPoints)...)
			}

			if multi {
				i++
			}

			lastIndex = i
		}
	}

	return
}

func processCatmull(points []vector.Vector2f) (lines []Linear) {
	for i := 0; i < len(points)-1; i++ {
		var p1, p2, p3, p4 vector.Vector2f

		if i-1 >= 0 {
			p1 = points[i-1]
		} else {
			p1 = points[i]
		}

		p2 = points[i]

		if i+1 < len(points) {
			p3 = points[i+1]
		} else {
			p3 = p2.Add(p2.Sub(p1))
		}

		if i+2 < len(points) {
			p4 = points[i+2]
		} else {
			p4 = p3.Add(p3.Sub(p2))
		}

		lines = append(lines, ApproximateCatmullRom([]vector.Vector2f{p1, p2, p3, p4}, 50)...)
	}

	return
}
