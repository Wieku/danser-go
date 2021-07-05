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
	endTime float64
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

	end := objs[0]
	start := objs[1]

	endPos := end.GetStackedEndPositionMod(mover.diff.Mods)
	startPos := start.GetStackedStartPositionMod(mover.diff.Mods)
	mover.endTime = end.GetEndTime()
	mover.startTime = start.GetStartTime()

	if config.StreamTrigger < 0 || (mover.startTime-mover.endTime) < float64(config.StreamTrigger) {
		mover.invert = -1 * mover.invert
	}

	if endPos == startPos {
		mover.ca = curves.NewLinear(endPos, startPos)
		return 2
	}

	point := endPos.Mid(startPos)
	p := point.Sub(endPos).Rotate(mover.invert * math.Pi / 2).Scl(float32(config.RadiusMultiplier)).Add(point)
	mover.ca = curves.NewCirArc(endPos, p, startPos)

	return 2
}

func (mover *HalfCircleMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.ca.PointAt(t)
}
