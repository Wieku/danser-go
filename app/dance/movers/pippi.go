package movers

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type PippiMover struct {
	line               curves.Linear
	beginTime, endTime float64
	mods               difficulty.Modifier
}

func NewPippiMover() MultiPointMover {
	return &PippiMover{}
}

func (bm *PippiMover) Reset(mods difficulty.Modifier, _ int) {
	bm.mods = mods
}

func (bm *PippiMover) SetObjects(objs []objects.IHitObject) int {
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

func (bm PippiMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF64((time-bm.endTime)/(bm.beginTime-bm.endTime), 0, 1)
	pos := bm.line.PointAt(float32(t))
	radius := float32(32 * 0.98)
	num := float32(float32(time) / 100)
	pos = vector.NewVec2f(pos.X+math32.Cos(num)*radius, pos.Y+math32.Sin(num)*radius)
	return pos
}

func (bm *PippiMover) GetEndTime() float64 {
	return bm.beginTime
}
