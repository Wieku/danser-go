package movers

import (
	"github.com/wieku/danser-go/animation/easing"
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
	"math"
)

type LinearMover struct {
	bz                 *curves.Bezier
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

	waitTime := start.GetBasicData().StartTime - 380 //math.Max(0.0, 479.9999999999999 - 100.0)

	if waitTime > endTime {
		endTime = waitTime
	}

	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm LinearMover) Update(time int64) bmath.Vector2d {
	t := float64(time-bm.endTime) / float64(bm.beginTime-bm.endTime)
	t = math.Max(0.0, math.Min(1.0, t))
	return bm.bz.PointAt(easing.OutQuad(t))
}

func (bm *LinearMover) GetEndTime() int64 {
	return bm.beginTime
}
