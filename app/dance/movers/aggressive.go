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
	startTime   float64
}

func NewAggressiveMover() MultiPointMover {
	return &AggressiveMover{basicMover: &basicMover{}}
}

func (mover *AggressiveMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.lastAngle = 0
}

func (mover *AggressiveMover) SetObjects(objs []objects.IHitObject) int {
	start := objs[0]
	end := objs[1]

	startPos := start.GetStackedEndPositionMod(mover.diff.Mods)
	startTime := start.GetEndTime()
	endPos := end.GetStackedStartPositionMod(mover.diff.Mods)
	endTime := end.GetStartTime()

	scaledDistance := float32(endTime - startTime)

	newAngle := mover.lastAngle + math.Pi
	if s, ok := start.(objects.ILongObject); ok {
		newAngle = s.GetEndAngleMod(mover.diff.Mods)
	}

	points := []vector.Vector2f{startPos, vector.NewVec2fRad(newAngle, scaledDistance).Add(startPos)}

	if scaledDistance > 1 {
		mover.lastAngle = points[1].AngleRV(endPos)
	}

	if s, ok := end.(objects.ILongObject); ok {
		points = append(points, vector.NewVec2fRad(s.GetStartAngleMod(mover.diff.Mods), scaledDistance).Add(endPos))
	}

	points = append(points, endPos)

	mover.bz = curves.NewBezierNA(points)
	mover.startTime = startTime
	mover.endTime = endTime

	return 2
}

func (mover *AggressiveMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.startTime)/float32(mover.endTime-mover.startTime), 0, 1)
	return mover.bz.PointAt(t)
}
