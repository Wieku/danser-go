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

func same(o1 objects.BaseObject, o2 objects.BaseObject) bool {
	d1 := o1.GetBasicData()
	d2 := o2.GetBasicData()
	return d1.StartPos == d2.StartPos || (settings.Dance.Momentum.SkipStackAngles && d1.StartPos.Sub(d2.StackOffset) == d2.StartPos.Sub(d2.StackOffset))
}

func anorm(a float32) float32 {
	pi2 := 2 * math32.Pi
	a = math32.Mod(a, pi2)
	if a < 0 { a += pi2 }
	return a
}

func anorm2(a float32) float32 {
	a = anorm(a)
	if a > math32.Pi { a = -(2 * math32.Pi - a) }
	return a
}

func (bm *MomentumMover) SetObjects(objs []objects.BaseObject) int {
	i := 0
	if bm.first { i = 1 }

	end := objs[i+0]
	start := objs[i+1]
	hasNext := false
	var next *objects.Circle
	if len(objs) > 2 {
		if v, ok := objs[i+2].(*objects.Circle); ok {
			hasNext = true
			next = v
		}   
	}

	endPos := end.GetBasicData().EndPos
	startPos := start.GetBasicData().StartPos

	dst := endPos.Dst(startPos)

	var a2 float32
	fromSlider := false
	for i++; i < len(objs); i++ {
		o := objs[i]
		if s, ok := o.(*objects.Slider); ok && !s.IsRetarded() {
			a2 = s.GetStartAngle()
			fromSlider = true
			break
		}
		if i == len(objs) - 1 {
			a2 = bm.last.AngleRV(endPos)
			break
		}
		if !same(o, objs[i+1]) {
			a2 = o.GetBasicData().StartPos.AngleRV(objs[i+1].GetBasicData().StartPos)
			break
		}
	}

	s, ok1 := end.(*objects.Slider)
	if ok1 { ok1 = !s.IsRetarded() }

	// stream detection logic stolen from spline mover
	stream := false
	ms := settings.Dance.Momentum
	if hasNext && !fromSlider && ms.StreamRestrict {
		min := float32(25.0)
		max := float32(6000.0)
		nextPos := next.GetBasicData().StartPos
		sq1 := endPos.DstSq(startPos)
		sq2 := startPos.DstSq(nextPos)

		if sq1 >= min && sq2 >= min && sq1 <= max && sq2 <= max {
			stream = true
		}
	}

	var a1 float32
	if ok1 {
		a1 = s.GetEndAngle()
	} else if bm.first {
		a1 = a2 + math.Pi
	} else {
		a1 = endPos.AngleRV(bm.last)
	}

	a := endPos.AngleRV(startPos)
	offset := float32(ms.RestrictAngle * math.Pi / 180.0)
	sangle := float32(ms.StreamAngle * math.Pi / 180.0)

	multend := ms.DistanceMult
	multstart := ms.DistanceMultEnd

	if stream && math32.Abs(anorm(a2 - startPos.AngleRV(endPos))) < anorm((2 * math32.Pi) - offset) {
		if anorm(a1 - a) > math32.Pi {
			a2 = a - sangle
		} else {
			a2 = a + sangle
		}
		multend = ms.StreamMult
		multstart = ms.StreamMult
	} else if !fromSlider && math32.Abs(anorm2(a2 - startPos.AngleRV(endPos))) < offset {
		a = startPos.AngleRV(endPos)
		if anorm(a2 - a) < offset {
			a2 = a - offset
		} else {
			a2 = a + offset
		}
	}

	p1 := vector.NewVec2fRad(a1, dst * float32(multend)).Add(endPos)
	p2 := vector.NewVec2fRad(a2, dst * float32(multstart)).Add(startPos)

	if !same(end, start) {
		bm.last = p2
	}

	bm.bz = curves.NewBezierNA([]vector.Vector2f{endPos, p1, p2, startPos})
	bm.endTime = end.GetBasicData().EndTime
	bm.startTime = start.GetBasicData().StartTime
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
