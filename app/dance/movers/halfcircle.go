package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type HalfCircleMover struct {
	ca                 curves.Curve
	startTime, endTime float64
	invert             float32
	diff               *difficulty.Difficulty
	id                 int
}

func NewHalfCircleMover() MultiPointMover {
	return &HalfCircleMover{invert: -1}
}

func (bm *HalfCircleMover) Reset(diff *difficulty.Difficulty, id int) {
	bm.diff = diff
	bm.invert = -1
	bm.id = id
}

func (bm *HalfCircleMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.HalfCircle[bm.id%len(settings.CursorDance.MoverSettings.HalfCircle)]

	end := objs[0]
	start := objs[1]

	endPos := end.GetStackedEndPositionMod(bm.diff.Mods)
	startPos := start.GetStackedStartPositionMod(bm.diff.Mods)
	bm.endTime = end.GetEndTime()
	bm.startTime = start.GetStartTime()

	if config.StreamTrigger < 0 || (bm.startTime-bm.endTime) < float64(config.StreamTrigger) {
		bm.invert = -1 * bm.invert
	}

	if endPos == startPos {
		bm.ca = curves.NewLinear(endPos, startPos)
		return 2
	}

	point := endPos.Mid(startPos)
	p := point.Sub(endPos).Rotate(bm.invert * math.Pi / 2).Scl(float32(config.RadiusMultiplier)).Add(point)
	bm.ca = curves.NewCirArc(endPos, p, startPos)

	return 2
}

func (bm *HalfCircleMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.ca.PointAt(t)
}

func (bm *HalfCircleMover) GetEndTime() float64 {
	return bm.startTime
}
