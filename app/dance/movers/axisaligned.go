package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/curves"
	"github.com/wieku/danser-go/app/bmath/math32"
)

type AxisMover struct {
	bz                 *curves.MultiCurve
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

	var midP bmath.Vector2f

	if math32.Abs(startPos.Sub(endPos).X) < math32.Abs(startPos.Sub(endPos).X) {
		midP = bmath.NewVec2f(endPos.X, startPos.Y)
	} else {
		midP = bmath.NewVec2f(startPos.X, endPos.Y)
	}

	bm.bz = curves.NewMultiCurve("L", []bmath.Vector2f{endPos, midP, startPos}, float64(endPos.Dst(midP)+midP.Dst(startPos)))
	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm AxisMover) Update(time int64) bmath.Vector2f {
	t := float32(time-bm.endTime) / float32(bm.beginTime-bm.endTime)
	tr := bmath.ClampF32(math32.Sin(t*math32.Pi/2), 0, 1)
	return bm.bz.PointAt(tr)
}

func (bm *AxisMover) GetEndTime() int64 {
	return bm.beginTime
}
