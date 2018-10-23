package movers

import (
	"math"
	"github.com/wieku/danser/bmath/curves"
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/settings"
	"github.com/wieku/danser/bmath"
)

type HalfCircleMover struct {
	ca                 curves.Curve
	startTime, endTime int64
	invert             float64
}

func NewHalfCircleMover() MultiPointMover {
	return &HalfCircleMover{invert: -1}
}

func (bm *HalfCircleMover) Reset() {
	bm.invert = -1
}

func (bm *HalfCircleMover) SetObjects(objs []objects.BaseObject) {
	end := objs[0]
	start := objs[1]

	endPos := end.GetBasicData().EndPos
	startPos := start.GetBasicData().StartPos
	bm.endTime = end.GetBasicData().EndTime
	bm.startTime = start.GetBasicData().StartTime

	if settings.Dance.HalfCircle.StreamTrigger < 0 || (bm.startTime-bm.endTime) < settings.Dance.HalfCircle.StreamTrigger {
		bm.invert = -1 * bm.invert
	}

	if endPos == startPos {
		bm.ca = curves.NewLinear(endPos, startPos)
		return
	}

	point := endPos.Mid(startPos)
	p := point.Sub(endPos).Rotate(bm.invert * math.Pi / 2).Scl(settings.Dance.HalfCircle.RadiusMultiplier).Add(point)
	bm.ca = curves.NewCirArc(endPos, p, startPos)
}

func (bm *HalfCircleMover) Update(time int64) bmath.Vector2d {
	return bm.ca.PointAt(float64(time-bm.endTime) / float64(bm.startTime-bm.endTime))
}

func (bm *HalfCircleMover) GetEndTime() int64 {
	return bm.startTime
}
