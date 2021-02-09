package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type AxisMover struct {
	bz                 *curves.MultiCurve
	beginTime, endTime float64
	mods               difficulty.Modifier
}

func NewAxisMover() MultiPointMover {
	return &AxisMover{}
}

func (bm *AxisMover) Reset(mods difficulty.Modifier) {
	bm.mods = mods
}

func (bm *AxisMover) SetObjects(objs []objects.IHitObject) int {
	end, start := objs[0], objs[1]
	endPos := end.GetStackedEndPositionMod(bm.mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(bm.mods)
	startTime := start.GetStartTime()

	var midP vector.Vector2f

	if math32.Abs(startPos.Sub(endPos).X) < math32.Abs(startPos.Sub(startPos).X) {
		midP = vector.NewVec2f(endPos.X, startPos.Y)
	} else {
		midP = vector.NewVec2f(startPos.X, endPos.Y)
	}

	bm.bz = curves.NewMultiCurve("L", []vector.Vector2f{endPos, midP, startPos})
	bm.endTime = endTime
	bm.beginTime = startTime

	return 2
}

func (bm AxisMover) Update(time float64) vector.Vector2f {
	t := float32(time-bm.endTime) / float32(bm.beginTime-bm.endTime)
	tr := bmath.ClampF32(math32.Sin(t*math32.Pi/2), 0, 1)
	return bm.bz.PointAt(tr)
}

func (bm *AxisMover) GetEndTime() float64 {
	return bm.beginTime
}
