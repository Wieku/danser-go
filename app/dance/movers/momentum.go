package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

// https://github.com/TechnoJo4/osu/blob/master/osu.Game.Rulesets.Osu/Replays/Movers/MomentumMover.cs

type MomentumMover struct {
	bz        *curves.Bezier
	last      vector.Vector2f
	startTime int64
	endTime   int64
	first     bool
}

func NewMomentumMover() MultiPointMover {
	return &MomentumMover{last: vector.NewVec2f(0, 0), first: true}
}

func (bm *MomentumMover) Reset() {
	bm.first = true
	bm.last = vector.NewVec2f(0, 0)
}

func same(o1 objects.IHitObject, o2 objects.IHitObject) bool {
	return o1.GetStackedStartPosition() == o2.GetStackedStartPosition() || (settings.Dance.Momentum.SkipStackAngles && o1.GetStartPosition() == o2.GetStartPosition())
}

func (bm *MomentumMover) SetObjects(objs []objects.IHitObject) int {
	i := 0
	if bm.first { i = 1 }

	end := objs[i+0]
	start := objs[i+1]

	endPos := end.GetStackedEndPosition()
	startPos := start.GetStackedStartPosition()

	dst := endPos.Dst(startPos)

	var a2 float32
	fromSlider := false
	for i++; i < len(objs); i++ {
		o := objs[i]
		if s, ok := o.(*objects.Slider); ok {
			a2 = s.GetStartAngle()
			fromSlider = true
			break
		}
		if i == len(objs) - 1 {
			a2 = bm.last.AngleRV(endPos)
			break
		}
		if !same(o, objs[i+1]) {
			a2 = o.GetStackedStartPosition().AngleRV(objs[i+1].GetStackedStartPosition())
			break
		}
	}

	var a1 float32
	if s, ok := end.(*objects.Slider); ok {
		a1 = s.GetEndAngle()
	} else if bm.first {
		a1 = a2 + math.Pi
	} else {
		a1 = endPos.AngleRV(bm.last)
	}

	a := startPos.AngleRV(endPos)
	offset := float32(settings.Dance.Momentum.RestrictAngle * math.Pi / 180.0)
	if !fromSlider && math32.Abs(a2 - a) < offset {
		if a2 - a < offset {
			a2 = a - offset
		} else {
			a2 = a + offset
		}
	}

	p1 := vector.NewVec2fRad(a1, dst * float32(settings.Dance.Momentum.DistanceMult)).Add(endPos)
	p2 := vector.NewVec2fRad(a2, dst * float32(settings.Dance.Momentum.DistanceMultEnd)).Add(startPos)

	if !same(end, start) {
		bm.last = p2
	}

	bm.bz = curves.NewBezierNA([]vector.Vector2f{endPos, p1, p2, startPos})
	bm.endTime = end.GetEndTime()
	bm.startTime = start.GetStartTime()
	bm.first = false

	return 2
}

func (bm *MomentumMover) Update(time int64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *MomentumMover) GetEndTime() int64 {
	return bm.startTime
}
