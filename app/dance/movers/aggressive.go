package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type AggressiveMover struct {
	*basicMover

	curve *curves.Bezier

	lastAngle float32
}

func NewAggressiveMover() MultiPointMover {
	return &AggressiveMover{basicMover: &basicMover{}}
}

func (mover *AggressiveMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.lastAngle = 0
}

func (mover *AggressiveMover) SetObjects(objs []objects.IHitObject) int {
	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	scaledDistance := float32(mover.endTime - mover.startTime)

	newAngle := mover.lastAngle + math.Pi
	if s, ok := start.(objects.ILongObject); ok {
		newAngle = s.GetEndAngleMod(mover.diff)
	}

	points := []vector.Vector2f{startPos, vector.NewVec2fRad(newAngle, scaledDistance).Add(startPos)}

	if scaledDistance > 1 {
		mover.lastAngle = points[1].AngleRV(endPos)
	}

	if s, ok := end.(objects.ILongObject); ok {
		points = append(points, vector.NewVec2fRad(s.GetStartAngleMod(mover.diff), scaledDistance).Add(endPos))
	}

	points = append(points, endPos)

	mover.curve = curves.NewBezierNA(points)

	return 2
}

func (mover *AggressiveMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
