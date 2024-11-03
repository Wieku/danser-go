package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type HalfCircleMover struct {
	*basicMover

	curve curves.Curve

	invert float32
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

	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	if config.StreamTrigger < 0 || (mover.endTime-mover.startTime) < float64(config.StreamTrigger) {
		mover.invert = -1 * mover.invert
	}

	if startPos == endPos {
		mover.curve = curves.NewLinear(startPos, endPos)
	} else {
		point := startPos.Mid(endPos)
		p := point.Sub(startPos).Rotate(mover.invert * math.Pi / 2).Scl(float32(config.RadiusMultiplier)).Add(point)
		mover.curve = curves.NewCirArc(startPos, p, endPos)
	}

	return 2
}

func (mover *HalfCircleMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
