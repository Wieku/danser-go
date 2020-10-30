package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type HalfCircleMover struct {
	ca                 curves.Curve
	startTime, endTime int64
	invert             float32
}

func NewHalfCircleMover() MultiPointMover {
	return &HalfCircleMover{invert: -1}
}

func (bm *HalfCircleMover) Reset() {
	bm.invert = -1
}

func (bm *HalfCircleMover) SetObjects(objs []objects.BaseObject) int {
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
		return 2
	}

	point := endPos.Mid(startPos)
	p := point.Sub(endPos).Rotate(bm.invert * math.Pi / 2).Scl(float32(settings.Dance.HalfCircle.RadiusMultiplier)).Add(point)
	bm.ca = curves.NewCirArc(endPos, p, startPos)

	return 2
}

func (bm *HalfCircleMover) Update(time int64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.ca.PointAt(t)
}

func (bm *HalfCircleMover) GetEndTime() int64 {
	return bm.startTime
}
