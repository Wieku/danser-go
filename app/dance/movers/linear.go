package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
)

type LinearMover struct {
	line               curves.Linear
	beginTime, endTime int64
}

func NewLinearMover() MultiPointMover {
	return &LinearMover{}
}

func (bm *LinearMover) Reset() {

}

func (bm *LinearMover) SetObjects(objs []objects.BaseObject) int {
	end, start := objs[0], objs[1]
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	bm.line = curves.NewLinear(endPos, startPos)

	bm.endTime = bmath.MaxI64(endTime, start.GetBasicData().StartTime-380)
	bm.beginTime = startTime

	return 2
}

func (bm LinearMover) Update(time int64) vector.Vector2f {
	t := bmath.ClampF64(float64(time-bm.endTime)/float64(bm.beginTime-bm.endTime), 0, 1)
	return bm.line.PointAt(float32(easing.OutQuad(t)))
}

func (bm *LinearMover) GetEndTime() int64 {
	return bm.beginTime
}
