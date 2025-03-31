package curves

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"slices"
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
	sections []float32
	lines    []Linear
	length   float32

	points    []vector.Vector2f
	cumLength []float64

	firstPoint vector.Vector2f
}

func NewMultiCurve(curveDefs []CurveDef) *MultiCurve {
	lines := make([]Linear, 0)
	points := make([]vector.Vector2f, 0)

	for _, def := range curveDefs {
		var cPoints1 []vector.Vector2f

		switch def.CurveType {
		case CCirArc:
			cPoints1 = processPerfect(def.Points, false)
		case CLine:
			cPoints1 = processLinear(def.Points)
		case CBezier:
			cPoints1 = processBezier(def.Points)
		case CCatmull:
			cPoints1 = processCatmull(def.Points)
		}

		cPoints2 := cPoints1
		if def.CurveType == CCirArc {
			cPoints2 = processPerfect(def.Points, true)
		}

		nLines := make([]Linear, max(0, len(lines)+len(cPoints1)-1))
		copy(nLines, lines)
		for i := 0; i < len(cPoints1)-1; i++ {
			nLines[len(lines)+i] = NewLinear(cPoints1[i], cPoints1[i+1])
		}
		lines = nLines

		skip := 0
		if len(points) > 0 && points[len(points)-1] == cPoints2[0] {
			skip = 1
		}

		nPoints := make([]vector.Vector2f, len(points)+len(cPoints2)-skip)
		copy(nPoints, points)
		copy(nPoints[len(points):], cPoints2[skip:])
		points = nPoints
	}

	length := float32(0.0)

	for _, l := range lines {
		length += l.GetLength()
	}

	cumLength := make([]float64, max(1, len(points)))
	for i := 0; i < len(points)-1; i++ {
		cumLength[i+1] = cumLength[i] + float64(points[i+1].Dst(points[i]))
	}

	firstPoint := curveDefs[0].Points[0]

	sections := make([]float32, len(lines)+1)
	sections[0] = 0.0
	prev := float32(0.0)

	for i := 0; i < len(lines); i++ {
		prev += lines[i].GetLength()
		sections[i+1] = prev
	}

	return &MultiCurve{sections, lines, length, points, cumLength, firstPoint}
}

func NewMultiCurveT(curveDefs []CurveDef, desiredLength float64) *MultiCurve {
	mCurve := NewMultiCurve(curveDefs)

	length64 := 0.0

	for _, l := range mCurve.lines {
		length64 += float64(l.GetLength87())
	}

	if length64 > 0 && desiredLength != 0 {
		diff := length64 - desiredLength

		for len(mCurve.lines) > 0 {
			line := mCurve.lines[len(mCurve.lines)-1]

			if float64(line.GetLength87()) > diff+minPartWidth {
				if line.Point1 != line.Point2 {
					nor := line.Point2.Sub(line.Point1).Nor87()

					pt := line.Point1.Add(nor.Scl87(line.GetLength87() - float32(diff)))
					mCurve.lines[len(mCurve.lines)-1] = NewLinear(line.Point1, pt)
				}

				break
			}

			diff -= float64(line.GetLength87())
			mCurve.lines = mCurve.lines[:len(mCurve.lines)-1]
		}
	}

	if desiredLength != 0 && mCurve.GetLengthLazer() != desiredLength {
		if len(mCurve.points) >= 2 && mCurve.points[len(mCurve.points)-1] == mCurve.points[len(mCurve.points)-2] && desiredLength > mCurve.GetLengthLazer() {
			mCurve.cumLength = append(mCurve.cumLength, mCurve.GetLengthLazer())
		} else {
			mCurve.cumLength = mCurve.cumLength[:len(mCurve.cumLength)-1]

			pathEndIndex := len(mCurve.points) - 1

			if mCurve.GetLengthLazer() > desiredLength {
				for len(mCurve.cumLength) > 0 && mCurve.cumLength[len(mCurve.cumLength)-1] >= desiredLength {
					mCurve.cumLength = mCurve.cumLength[:len(mCurve.cumLength)-1]
					mCurve.points = mCurve.points[:len(mCurve.points)-1]
					pathEndIndex--
				}
			}

			if pathEndIndex <= 0 {
				mCurve.cumLength = append(mCurve.cumLength, 0)
			} else {
				dir := mCurve.points[pathEndIndex].Sub(mCurve.points[pathEndIndex-1]).Nor()

				mCurve.points[pathEndIndex] = mCurve.points[pathEndIndex-1].Add(dir.Scl(float32(desiredLength - mCurve.cumLength[len(mCurve.cumLength)-1])))
				mCurve.cumLength = append(mCurve.cumLength, desiredLength)
			}
		}
	}

	mCurve.length = 0.0

	for _, l := range mCurve.lines {
		mCurve.length += l.GetLength()
	}

	mCurve.sections = make([]float32, len(mCurve.lines)+1)
	mCurve.sections[0] = 0.0
	prev := float32(0.0)

	length64 = 0.0

	for i := 0; i < len(mCurve.lines); i++ {
		prevL := length64
		length64 += float64(mCurve.lines[i].GetLength87())
		mCurve.lines[i].customLength = length64 - prevL

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

func (mCurve *MultiCurve) PointAtLazer(t float64) vector.Vector2f {
	if len(mCurve.lines) == 0 || mCurve.GetLengthLazer() == 0 {
		return mCurve.firstPoint
	}

	d := mutils.Clamp(t, 0, 1) * mCurve.GetLengthLazer()

	i, _ := slices.BinarySearch(mCurve.cumLength, d)

	if i <= 0 {
		return mCurve.firstPoint
	} else if i >= len(mCurve.points) {
		return mCurve.points[len(mCurve.points)-1]
	}

	p0 := mCurve.points[i-1]
	p1 := mCurve.points[i]

	d0 := mCurve.cumLength[i-1]
	d1 := mCurve.cumLength[i]

	// Avoid division by and almost-zero number in case two points are extremely close to each other.
	if mutils.Abs(d0-d1) < 0.00000001 {
		return p0
	}

	w := (d - d0) / (d1 - d0)
	return p0.Add(p1.Sub(p0).Scl(float32(w)))
}

func (mCurve *MultiCurve) GetLength() float32 {
	return mCurve.length
}

func (mCurve *MultiCurve) GetLengthLazer() float64 {
	if len(mCurve.cumLength) == 0 {
		return 0
	}

	return mCurve.cumLength[len(mCurve.cumLength)-1]
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

func processPerfect(points []vector.Vector2f, lazer bool) (outPoints []vector.Vector2f) {
	if len(points) > 3 {
		outPoints = processBezier(points)
	} else if len(points) < 3 || vector.IsStraightLine32(points[0], points[1], points[2]) {
		outPoints = processLinear(points)
	} else if lazer {
		outPoints = ApproximateCircularArcLazer(points[0], points[1], points[2])
	} else {
		outPoints = ApproximateCircularArc(points[0], points[1], points[2], 0.125)
	}

	return
}

func processLinear(points []vector.Vector2f) (outPoints []vector.Vector2f) {
	for i := 0; i < len(points); i++ {
		if i < len(points)-1 && points[i] == points[i+1] { // skip red anchors, present in old maps like 243
			continue
		}

		outPoints = append(outPoints, points[i])
	}

	return
}

func processBezier(points []vector.Vector2f) (outPoints []vector.Vector2f) {
	lastIndex := 0

	for i := 0; i < len(points); i++ {
		multi := i < len(points)-2 && points[i] == points[i+1]

		if multi || i == len(points)-1 {
			subPoints := points[lastIndex : i+1]

			var inter []vector.Vector2f

			if len(subPoints) == 2 {
				inter = []vector.Vector2f{subPoints[0], subPoints[1]}
			} else {
				inter = ApproximateBezier(subPoints)
			}

			if len(outPoints) == 0 || outPoints[len(outPoints)-1] != inter[0] {
				outPoints = append(outPoints, inter...)
			} else {
				outPoints = append(outPoints, inter[1:]...)
			}

			if multi {
				i++
			}

			lastIndex = i
		}
	}

	return
}

func processCatmull(points []vector.Vector2f) (outPoints []vector.Vector2f) {
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

		outPoints = append(outPoints, ApproximateCatmullRom([]vector.Vector2f{p1, p2, p3, p4}, 50)...)
	}

	return
}
