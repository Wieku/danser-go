package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type AggressiveMover struct {
	*basicMover

	lastAngle float32
	bz        *curves.Bezier
	endTime   float64
}

func NewAggressiveMover() MultiPointMover {
	return &AggressiveMover{basicMover: &basicMover{}}
}

func (mover *AggressiveMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.lastAngle = 0
}

func (mover *AggressiveMover) SetObjects(objs []objects.IHitObject) int {
	end := objs[0]
	start := objs[1]

	endPos := end.GetStackedEndPositionMod(mover.diff.Mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(mover.diff.Mods)
	startTime := start.GetStartTime()

	scaledDistance := float32(startTime - endTime)

	newAngle := mover.lastAngle + math.Pi
	if s, ok := end.(objects.ILongObject); ok {
		newAngle = s.GetEndAngleMod(mover.diff.Mods)
	}

	points := []vector.Vector2f{endPos, vector.NewVec2fRad(newAngle, scaledDistance).Add(endPos)}

	if scaledDistance > 1 {
		mover.lastAngle = points[1].AngleRV(startPos)
	}

	if s, ok := start.(objects.ILongObject); ok {
		points = append(points, vector.NewVec2fRad(s.GetStartAngleMod(mover.diff.Mods), scaledDistance).Add(startPos))
	}

	points = append(points, startPos)

	mover.bz = curves.NewBezierNA(points)
	mover.endTime = endTime
	mover.startTime = startTime

	return 2
}

func (mover *AggressiveMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.bz.PointAt(t)
}
