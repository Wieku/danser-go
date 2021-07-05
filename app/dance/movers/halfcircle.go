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
	*basicMover

	ca      curves.Curve
	startTime float64
	invert  float32
}

func NewHalfCircleMover() MultiPointMover {
	return &HalfCircleMover{basicMover: &basicMover{}}
}

func (mover *HalfCircleMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.invert = -1
}

func (mover *HalfCircleMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.HalfCircle[mover.id%len(settings.CursorDance.MoverSettings.HalfCircle)]

	start := objs[0]
	end := objs[1]

	startPos := start.GetStackedEndPositionMod(mover.diff.Mods)
	endPos := end.GetStackedStartPositionMod(mover.diff.Mods)
	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	if config.StreamTrigger < 0 || (mover.endTime-mover.startTime) < float64(config.StreamTrigger) {
		mover.invert = -1 * mover.invert
	}

	if startPos == endPos {
		mover.ca = curves.NewLinear(startPos, endPos)
		return 2
	}

	point := startPos.Mid(endPos)
	p := point.Sub(startPos).Rotate(mover.invert * math.Pi / 2).Scl(float32(config.RadiusMultiplier)).Add(point)
	mover.ca = curves.NewCirArc(startPos, p, endPos)

	return 2
}

func (mover *HalfCircleMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.startTime)/float32(mover.endTime-mover.startTime), 0, 1)
	return mover.ca.PointAt(t)
}
