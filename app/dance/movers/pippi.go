package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type PippiMover struct {
	*basicMover

	line curves.Linear
}

func NewPippiMover() MultiPointMover {
	return &PippiMover{basicMover: &basicMover{}}
}

func (mover *PippiMover) SetObjects(objs []objects.IHitObject) int {
	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff.Mods)
	endPos := end.GetStackedStartPositionMod(mover.diff.Mods)

	mover.line = curves.NewLinear(startPos, endPos)

	return 2
}

func (mover *PippiMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF64((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	pos := mover.line.PointAt(float32(easing.OutQuad(t)))

	return mover.modifyPos(time, false, pos)
}

func (mover *PippiMover) GetObjectsPosition(time float64, object objects.IHitObject) vector.Vector2f {
	return mover.modifyPos(time, object.GetType() == objects.SPINNER, mover.basicMover.GetObjectsPosition(time, object))
}

func (mover *PippiMover) modifyPos(time float64, spinner bool, pos vector.Vector2f) vector.Vector2f {
	config := settings.CursorDance.MoverSettings.Pippi[mover.id%len(settings.CursorDance.MoverSettings.Pippi)]

	rad := math.Mod(time/1000*config.RotationSpeed, 1) * 2 * math.Pi

	radius := config.SpinnerRadius
	if !spinner {
		radius = mover.diff.CircleRadius * bmath.ClampF64(config.RadiusMultiplier, 0, 1)
	}

	mVec := vector.NewVec2fRad(float32(rad), float32(radius))

	return pos.Add(mVec)
}
