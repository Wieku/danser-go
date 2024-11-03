package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type LinearMover struct {
	*basicMover

	line   curves.Linear
	simple bool
}

func NewLinearMover() MultiPointMover {
	return &LinearMover{basicMover: &basicMover{}}
}

func NewLinearMoverSimple() MultiPointMover {
	return &LinearMover{
		basicMover: &basicMover{},
		simple:     true,
	}
}

func (mover *LinearMover) SetObjects(objs []objects.IHitObject) int {
	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	mover.line = curves.NewLinear(startPos, endPos)

	if mover.simple {
		mover.startTime = max(mover.startTime, mover.endTime-(mover.diff.Preempt-100*mover.diff.Speed))
	} else {
		config := settings.CursorDance.MoverSettings.Linear[mover.id%len(settings.CursorDance.MoverSettings.Linear)]

		if config.WaitForPreempt {
			mover.startTime = max(mover.startTime, mover.endTime-(mover.diff.Preempt-config.ReactionTime*mover.diff.Speed))
		}
	}

	return 2
}

func (mover *LinearMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.line.PointAt(float32(easing.OutQuad(t)))
}

func (mover *LinearMover) GetObjectsPosition(time float64, object objects.IHitObject) vector.Vector2f {
	config := settings.CursorDance.MoverSettings.Linear[mover.id%len(settings.CursorDance.MoverSettings.Linear)]

	if !config.ChoppyLongObjects || mover.simple || object.GetType() == objects.CIRCLE {
		return mover.basicMover.GetObjectsPosition(time, object)
	}

	timeDiff := math.Mod(time-object.GetStartTime(), sixtyTime)

	time1 := time - timeDiff
	time2 := time1 + sixtyTime

	pos1 := object.GetStackedPositionAtMod(time1, mover.diff)
	pos2 := object.GetStackedPositionAtMod(time2, mover.diff)

	return pos1.Lerp(pos2, float32((time-time1)/sixtyTime))
}
