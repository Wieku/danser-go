package movers

import (
	math2 "github.com/wieku/danser/bmath"
	"github.com/wieku/danser/bmath/curves"
	//"osubot/io"
	"github.com/wieku/danser/beatmap/objects"
	"math"
	"github.com/wieku/danser/render"
)

const (
	ANGLE = math.Pi/2
	STRENGTH = 2.0/3
	STREAM = 130
)

type FlowerBezierMover struct {
	lastAngle float64
	bz curves.Bezier
	beginTime, endTime int64
	invert float64
}

func NewFlowerBezierMover() *FlowerBezierMover {
	return &FlowerBezierMover{lastAngle: 0, invert: 1}
}

func (bm *FlowerBezierMover) Reset() {
	bm.lastAngle = 0
	bm.invert = 1
}

func (bm *FlowerBezierMover) SetObjects(end, start objects.BaseObject) {
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	distance := endPos.Dst(startPos)

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []math2.Vector2d

	scaledDistance := distance * STRENGTH
	newAngle := ANGLE

	if endPos == startPos {
		if ANGLE == 0.0 {
			pt1 := math2.NewVec2dRad(bm.lastAngle + math.Pi, float64(startTime-endTime)/2).Add(endPos)
			points = []math2.Vector2d{endPos, pt1, startPos}
		} else {
			points = []math2.Vector2d{endPos, startPos}
		}
	} else if ok1 && ok2 {
		bm.invert = -1 * bm.invert

		pt1 := math2.NewVec2dRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
		pt2 := math2.NewVec2dRad(s2.GetStartAngle(), scaledDistance).Add(startPos)

		points = []math2.Vector2d{endPos, pt1, pt2, startPos}
	} else if ok1 {
		bm.invert = -1 * bm.invert
		bm.lastAngle = endPos.AngleRV(startPos) - newAngle * bm.invert

		pt1 := math2.NewVec2dRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
		pt2 := math2.NewVec2dRad(bm.lastAngle, scaledDistance).Add(startPos)

		points = []math2.Vector2d{endPos, pt1, pt2, startPos}
	} else if ok2 {
		bm.lastAngle += math.Pi

		pt1 := math2.NewVec2dRad(bm.lastAngle, scaledDistance).Add(endPos)
		pt2 := math2.NewVec2dRad(s2.GetStartAngle(), scaledDistance).Add(startPos)

		points = []math2.Vector2d{endPos, pt1, pt2, startPos}
	} else {
		if startTime - endTime < STREAM {
			newAngle = math.Pi/2
		}
		angle := endPos.AngleRV(startPos) - newAngle * bm.invert

		pt1 := math2.NewVec2dRad(bm.lastAngle + math.Pi, scaledDistance).Add(endPos)
		pt2 := math2.NewVec2dRad(angle, scaledDistance).Add(startPos)

		bm.lastAngle = angle

		if startTime - endTime < STREAM {
			bm.invert = -1 * bm.invert
		}

		points = []math2.Vector2d{endPos, pt1, pt2, startPos}
	}

	bm.bz = curves.NewBezier(points)

	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm FlowerBezierMover) Update(time int64, cursor *render.Cursor) {
	cursor.SetPos(bm.bz.NPointAt(float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)))
}