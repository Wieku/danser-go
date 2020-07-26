package movers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
	"math"
)

type AggressiveMover struct {
	lastAngle          float32
	bz                 *curves.Bezier
	startTime, endTime int64
}

func NewAggressiveMover() MultiPointMover {
	return &AggressiveMover{lastAngle: 0}
}

func (bm *AggressiveMover) Reset() {
	bm.lastAngle = 0
}

func (bm *AggressiveMover) SetObjects(objs []objects.BaseObject) {
	end := objs[0]
	start := objs[1]

	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	scaledDistance := float32(startTime - endTime)

	newAngle := bm.lastAngle + math.Pi
	if s, ok := end.(*objects.Slider); ok {
		newAngle = s.GetEndAngle()
	}

	points := []bmath.Vector2f{endPos, bmath.NewVec2fRad(newAngle, scaledDistance).Add(endPos)}

	if scaledDistance > 1 {
		bm.lastAngle = points[1].AngleRV(startPos)
	}

	if s, ok := start.(*objects.Slider); ok {
		points = append(points, bmath.NewVec2fRad(s.GetStartAngle(), scaledDistance).Add(startPos))
	}

	points = append(points, startPos)

	bm.bz = curves.NewBezierNA(points)
	bm.endTime = endTime
	bm.startTime = startTime
}

func (bm *AggressiveMover) Update(time int64) bmath.Vector2f {
	t := float32(time-bm.endTime) / float32(bm.startTime-bm.endTime)
	t = bmath.ClampF32(t, 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *AggressiveMover) GetEndTime() int64 {
	return bm.startTime
}
