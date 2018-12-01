package movers

import (
	"math"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/bmath/sliders"
)

type AxisMover struct {
	bz                 sliders.SliderAlgo
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

	bm.bz = sliders.NewSliderAlgo("L", []bmath.Vector2d{endPos, midP, startPos}, endPos.Dst(midP)+midP.Dst(startPos))
	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm AxisMover) Update(time int64) bmath.Vector2d {
	t := float64(time-bm.endTime) / float64(bm.beginTime-bm.endTime)
	tr := math.Max(0.0, math.Min(1.0, math.Sin(t * math.Pi / 2)))
	return bm.bz.PointAt(tr)
}

func (bm *AxisMover) GetEndTime() int64 {
	return bm.beginTime
}