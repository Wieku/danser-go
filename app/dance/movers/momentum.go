package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
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
	*basicMover

	bz        *curves.Bezier
	last      vector.Vector2f
	endTime   float64
	first     bool
	wasStream bool
}

func NewMomentumMover() MultiPointMover {
	return &MomentumMover{basicMover: &basicMover{}}
}

func (mover *MomentumMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.first = true
	mover.last = vector.NewVec2f(0, 0)
}

func same(mods difficulty.Modifier, o1 objects.IHitObject, o2 objects.IHitObject, skipStackAngles bool) bool {
	return o1.GetStackedStartPositionMod(mods) == o2.GetStackedStartPositionMod(mods) || (skipStackAngles && o1.GetStartPosition() == o2.GetStartPosition())
}

func anorm(a float32) float32 {
	pi2 := 2 * math32.Pi
	a = math32.Mod(a, pi2)
	if a < 0 {
		a += pi2
	}

	return a
}

func anorm2(a float32) float32 {
	a = anorm(a)
	if a > math32.Pi {
		a = -(2*math32.Pi - a)
	}

	return a
}

func (mover *MomentumMover) SetObjects(objs []objects.IHitObject) int {
	ms := settings.CursorDance.MoverSettings.Momentum[mover.id%len(settings.CursorDance.MoverSettings.Momentum)]

	i := 0

	end := objs[i+0]
	start := objs[i+1]

	hasNext := false
	var next objects.IHitObject
	if len(objs) > 2 {
		if _, ok := objs[i+2].(*objects.Circle); ok {
			hasNext = true
		}
		next = objs[i+2]
	}

	endPos := end.GetStackedEndPositionMod(mover.diff.Mods)
	startPos := start.GetStackedStartPositionMod(mover.diff.Mods)

	dst := endPos.Dst(startPos)

	var a2 float32
	fromLong := false
	for i++; i < len(objs); i++ {
		o := objs[i]
		if s, ok := o.(objects.ILongObject); ok {
			a2 = s.GetStartAngleMod(mover.diff.Mods)
			fromLong = true
			break
		}
		if i == len(objs)-1 {
			a2 = mover.last.AngleRV(endPos)
			break
		}
		if !same(mover.diff.Mods, o, objs[i+1], ms.SkipStackAngles) {
			a2 = o.GetStackedStartPositionMod(mover.diff.Mods).AngleRV(objs[i+1].GetStackedStartPositionMod(mover.diff.Mods))
			break
		}
	}

	var sq1, sq2 float32
	if next != nil {
		nextPos := next.GetStackedStartPositionMod(mover.diff.Mods)
		sq1 = endPos.DstSq(startPos)
		sq2 = startPos.DstSq(nextPos)
	}

	// stream detection logic stolen from spline mover
	stream := false
	if hasNext && !fromLong && ms.StreamRestrict {
		min := float32(25.0)
		max := float32(10000.0)

		if sq1 >= min && sq1 <= max && mover.wasStream || (sq2 >= min && sq2 <= max) {
			stream = true
		}
	}

	mover.wasStream = stream

	var a1 float32
	if s, ok := end.(objects.ILongObject); ok {
		a1 = s.GetEndAngleMod(mover.diff.Mods)
	} else if mover.first {
		a1 = a2 + math.Pi
	} else {
		a1 = endPos.AngleRV(mover.last)
	}

	mult := ms.DistanceMultOut

	ac := a2 - startPos.AngleRV(endPos)
	area := float32(ms.RestrictArea * math.Pi / 180.0)

	if area > 0 && stream && anorm(ac) < anorm((2*math32.Pi)-area) {
		a := endPos.AngleRV(startPos)

		sangle := float32(0.5 * math.Pi)
		if anorm(a1-a) > math32.Pi {
			a2 = a - sangle
		} else {
			a2 = a + sangle
		}

		mult = ms.StreamMult
	} else if !fromLong && area > 0 && math32.Abs(anorm2(ac)) < area {
		a := startPos.AngleRV(endPos)

		offset := float32(ms.RestrictAngle * math.Pi / 180.0)
		if (anorm(a2-a) < offset) != ms.RestrictInvert {
			a2 = a + offset
		} else {
			a2 = a - offset
		}

		mult = ms.DistanceMult
	} else if next != nil && !fromLong {
		r := sq1 / (sq1 + sq2)
		a := endPos.AngleRV(startPos)
		a2 = a + r*anorm2(a2-a)
	}

	endTime := end.GetEndTime()
	startTime := start.GetStartTime()
	duration := startTime - endTime

	if ms.DurationTrigger > 0 && duration >= ms.DurationTrigger {
		mult *= ms.DurationMult * (duration / ms.DurationTrigger)
	}

	p1 := vector.NewVec2fRad(a1, dst*float32(mult)).Add(endPos)
	p2 := vector.NewVec2fRad(a2, dst*float32(mult)).Add(startPos)

	if !same(mover.diff.Mods, end, start, ms.SkipStackAngles) {
		mover.last = p2
		mover.bz = curves.NewBezierNA([]vector.Vector2f{endPos, p1, p2, startPos})
	} else {
		mover.bz = curves.NewBezierNA([]vector.Vector2f{endPos, startPos})
	}

	mover.endTime = end.GetEndTime()
	mover.startTime = start.GetStartTime()
	mover.first = false

	return 2
}

func (mover *MomentumMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.bz.PointAt(t)
}
