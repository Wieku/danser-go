package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type LinearMover struct {
	line               curves.Linear
	beginTime, endTime float64
	mods               difficulty.Modifier
}

func NewLinearMover() MultiPointMover {
	return &LinearMover{}
}

func (bm *LinearMover) Reset(mods difficulty.Modifier) {
	bm.mods = mods
}

func (bm *LinearMover) SetObjects(objs []objects.IHitObject) int {
	end, start := objs[0], objs[1]
	endPos := end.GetStackedEndPositionMod(bm.mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(bm.mods)
	startTime := start.GetStartTime()

	bm.line = curves.NewLinear(endPos, startPos)

	bm.endTime = math.Max(endTime, start.GetStartTime()-380)
	bm.beginTime = startTime

	return 2
}

func (bm LinearMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF64(float64(time-bm.endTime)/float64(bm.beginTime-bm.endTime), 0, 1)
	return bm.line.PointAt(float32(easing.OutQuad(t)))
}

func (bm *LinearMover) GetEndTime() float64 {
	return bm.beginTime
}
