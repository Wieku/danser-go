package movers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
	"math"
)

type AxisMover struct {
	bz                 curves.MultiCurve
	beginTime, endTime int64
}

func NewAxisMover() MultiPointMover {
	return &AxisMover{}
}

func (bm *AxisMover) Reset() {

}

func (bm *AxisMover) SetObjects(objs []objects.BaseObject) {
	end, start := objs[0], objs[1]
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	var midP bmath.Vector2d

	if math.Abs(startPos.Sub(endPos).X) < math.Abs(startPos.Sub(endPos).X) {
		midP = bmath.NewVec2d(endPos.X, startPos.Y)
	} else {
		midP = bmath.NewVec2d(startPos.X, endPos.Y)
	}

	bm.bz = curves.NewMultiCurve("L", []bmath.Vector2d{endPos, midP, startPos}, endPos.Dst(midP)+midP.Dst(startPos), nil)
	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm AxisMover) Update(time int64) bmath.Vector2d {
	t := float64(time-bm.endTime) / float64(bm.beginTime-bm.endTime)
	tr := math.Max(0.0, math.Min(1.0, math.Sin(t*math.Pi/2)))
	return bm.bz.PointAt(tr)
}

func (bm *AxisMover) GetEndTime() int64 {
	return bm.beginTime
}
