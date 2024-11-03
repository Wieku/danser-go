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

type PippiMover struct {
	*basicMover

	curve curves.Curve
}

func NewPippiMover() MultiPointMover {
	return &PippiMover{basicMover: &basicMover{}}
}

func (mover *PippiMover) SetObjects(objs []objects.IHitObject) int {
	start, end := objs[0], objs[1]

	startC, cOk := start.(*objects.Circle)
	endC, eOk := end.(*objects.Circle)

	mover.startTime = max(start.GetEndTime(), end.GetStartTime()-(mover.diff.Preempt-100*mover.diff.Speed))
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	timeDifference := mover.endTime - mover.startTime

	points := make([]vector.Vector2f, 0, int(math.Ceil(timeDifference/sixtyTime)))

	if cOk && startC.DoubleClick {
		points = append(points, startPos)
	} else {
		points = append(points, mover.modifyPos(start.GetEndTime(), start.GetType() == objects.SPINNER, startPos))
	}

	for t := sixtyTime; t < timeDifference; t += sixtyTime {
		f := t / timeDifference

		basePos := mover.modifyPos(mover.startTime+t, false, startPos.Lerp(endPos, float32(f)))

		if cOk && startC.DoubleClick {
			basePos = startPos.Lerp(basePos, float32(f))
		}

		if eOk && endC.DoubleClick {
			basePos = basePos.Lerp(endPos, float32(f))
		}

		points = append(points, basePos)
	}

	if eOk && endC.DoubleClick {
		points = append(points, endPos)
	} else {
		points = append(points, mover.modifyPos(mover.endTime, end.GetType() == objects.SPINNER, endPos))
	}

	mover.curve = curves.NewMultiCurve([]curves.CurveDef{{CurveType: curves.CLine, Points: points}})

	return 2
}

func (mover *PippiMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(easing.OutQuad(t)))
}

func (mover *PippiMover) GetObjectsPosition(time float64, object objects.IHitObject) vector.Vector2f {
	c, ok := object.(*objects.Circle)
	if ok && c.DoubleClick {
		return c.GetStackedStartPositionMod(mover.diff)
	}

	return mover.modifyPos(time, object.GetType() == objects.SPINNER, mover.basicMover.GetObjectsPosition(time, object))
}

func (mover *PippiMover) modifyPos(time float64, spinner bool, pos vector.Vector2f) vector.Vector2f {
	config := settings.CursorDance.MoverSettings.Pippi[mover.id%len(settings.CursorDance.MoverSettings.Pippi)]

	rad := math.Mod(time/1000*config.RotationSpeed, 1) * 2 * math.Pi

	radius := config.SpinnerRadius
	if !spinner {
		radius = mover.diff.CircleRadius * mutils.Clamp(config.RadiusMultiplier, 0, 1)
	}

	mVec := vector.NewVec2fRad(float32(rad), float32(radius))

	return pos.Add(mVec)
}
