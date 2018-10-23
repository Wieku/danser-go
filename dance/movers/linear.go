package movers

import (
	"math"
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/bmath/curves"
	"github.com/wieku/danser/bmath"
)

type LinearMover struct {
	bz                 curves.Bezier
	beginTime, endTime int64
}

func NewLinearMover() MultiPointMover {
	return &LinearMover{}
}

func (bm *LinearMover) Reset() {

}

func (bm *LinearMover) SetObjects(objs []objects.BaseObject) {
	end, start := objs[0], objs[1]
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	bm.bz = curves.NewBezier([]bmath.Vector2d{endPos, startPos})
	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm LinearMover) Update(time int64) bmath.Vector2d {
	t := float64(time-bm.endTime) / float64(bm.beginTime-bm.endTime)
	t = math.Max(0.0, math.Min(1.0, t))
	return bm.bz.NPointAt(math.Sin(t * math.Pi / 2))
}

func (bm *LinearMover) GetEndTime() int64 {
	return bm.beginTime
}
