package movers

import (
	math2 "github.com/wieku/danser/bmath"
	"github.com/wieku/danser/bmath/curves"
	"github.com/wieku/danser/beatmap/objects"
	"math"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/settings"
)

type FlowerBezierMover struct {
	lastAngle float64
	bz curves.Bezier
	beginTime, endTime int64
	invert float64
}

func NewFlowerBezierMover() Mover {
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

	scaledDistance := distance * settings.Dance.Flower.DistanceMult
	newAngle := settings.Dance.Flower.AngleOffset * math.Pi / 180.0

	if end.GetBasicData().StartTime > 0 && settings.Dance.Flower.LongJump >= 0 && (startTime-endTime) > settings.Dance.Flower.LongJump {
		scaledDistance = float64(startTime-endTime) * settings.Dance.Flower.LongJumpMult
	}

	if endPos == startPos {
		if settings.Dance.Flower.LongJumpOnEqualPos {
			scaledDistance = float64(startTime-endTime) * settings.Dance.Flower.LongJumpMult
			bm.lastAngle += math.Pi

			pt1 := math2.NewVec2dRad(bm.lastAngle, scaledDistance).Add(endPos)

			if ok1 {
				pt1 = math2.NewVec2dRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
			}

			if !ok2 {
				angle := bm.lastAngle - newAngle * bm.invert
				pt2 := math2.NewVec2dRad(angle, scaledDistance).Add(startPos)
				bm.lastAngle = angle
				points = []math2.Vector2d{endPos, pt1, pt2, startPos}
			} else {
				pt2 := math2.NewVec2dRad(s2.GetStartAngle(), scaledDistance).Add(startPos)
				points = []math2.Vector2d{endPos, pt1, pt2, startPos}
			}

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
		if startTime - endTime < settings.Dance.Flower.StreamTrigger {
			newAngle = settings.Dance.Flower.StreamAngleOffset * math.Pi / 180.0
		}
		angle := endPos.AngleRV(startPos) - newAngle * bm.invert

		pt1 := math2.NewVec2dRad(bm.lastAngle + math.Pi, scaledDistance).Add(endPos)
		pt2 := math2.NewVec2dRad(angle, scaledDistance).Add(startPos)

		bm.lastAngle = angle

		if startTime - endTime < settings.Dance.Flower.StreamTrigger {
			bm.invert = -1 * bm.invert
		}

		points = []math2.Vector2d{endPos, pt1, pt2, startPos}
	}

	bm.bz = curves.NewBezier(points)
	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm FlowerBezierMover) Update(time int64, cursor *render.Cursor) {
	t := float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)
	t = math.Max(0.0, math.Min(1.0, t))
	cursor.SetPos(bm.bz.NPointAt(t))
}