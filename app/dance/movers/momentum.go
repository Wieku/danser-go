package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

// https://github.com/TechnoJo4/osu/blob/master/osu.Game.Rulesets.Osu/Replays/Movers/MomentumMover.cs

type MomentumMover struct {
	*basicMover

	curve *curves.Bezier

	last      vector.Vector2f
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

func same(diff *difficulty.Difficulty, o1 objects.IHitObject, o2 objects.IHitObject, skipStackAngles bool) bool {
	return o1.GetStackedStartPositionMod(diff) == o2.GetStackedStartPositionMod(diff) || (skipStackAngles && o1.GetStartPosition() == o2.GetStartPosition())
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

	start, end := objs[i], objs[i+1]

	hasNext := false
	var next objects.IHitObject
	if len(objs) > 2 {
		if _, ok := objs[i+2].(*objects.Circle); ok {
			hasNext = true
		}
		next = objs[i+2]
	}

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	dst := startPos.Dst(endPos)

	var a2 float32
	fromLong := false
	for i++; i < len(objs); i++ {
		o := objs[i]
		if s, ok := o.(objects.ILongObject); ok {
			a2 = s.GetStartAngleMod(mover.diff)
			fromLong = true
			break
		}
		if i == len(objs)-1 {
			a2 = mover.last.AngleRV(startPos)
			break
		}
		if !same(mover.diff, o, objs[i+1], ms.SkipStackAngles) {
			a2 = o.GetStackedStartPositionMod(mover.diff).AngleRV(objs[i+1].GetStackedStartPositionMod(mover.diff))
			break
		}
	}

	var sq1, sq2 float32
	if next != nil {
		nextPos := next.GetStackedStartPositionMod(mover.diff)
		sq1 = startPos.DstSq(endPos)
		sq2 = endPos.DstSq(nextPos)
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
	if s, ok := start.(objects.ILongObject); ok {
		a1 = s.GetEndAngleMod(mover.diff)
	} else if mover.first {
		a1 = a2 + math.Pi
	} else {
		a1 = startPos.AngleRV(mover.last)
	}

	mult := ms.DistanceMultOut

	ac := a2 - endPos.AngleRV(startPos)
	area := float32(ms.RestrictArea * math.Pi / 180.0)

	if area > 0 && stream && anorm(ac) < anorm((2*math32.Pi)-area) {
		a := startPos.AngleRV(endPos)

		sangle := float32(0.5 * math.Pi)
		if anorm(a1-a) > math32.Pi {
			a2 = a - sangle
		} else {
			a2 = a + sangle
		}

		mult = ms.StreamMult
	} else if !fromLong && area > 0 && math32.Abs(anorm2(ac)) < area {
		a := endPos.AngleRV(startPos)

		offset := float32(ms.RestrictAngle * math.Pi / 180.0)
		if (anorm(a2-a) < offset) != ms.RestrictInvert {
			a2 = a + offset
		} else {
			a2 = a - offset
		}

		mult = ms.DistanceMult
	} else if next != nil && !fromLong {
		r := sq1 / (sq1 + sq2)
		a := startPos.AngleRV(endPos)
		a2 = a + r*anorm2(a2-a)
	}

	startTime := start.GetEndTime()
	endTime := end.GetStartTime()
	duration := endTime - startTime

	if ms.DurationTrigger > 0 && duration >= ms.DurationTrigger {
		mult *= ms.DurationMult * (duration / ms.DurationTrigger)
	}

	p1 := vector.NewVec2fRad(a1, dst*float32(mult)).Add(startPos)
	p2 := vector.NewVec2fRad(a2, dst*float32(mult)).Add(endPos)

	if !same(mover.diff, start, end, ms.SkipStackAngles) {
		mover.last = p2
		mover.curve = curves.NewBezierNA([]vector.Vector2f{startPos, p1, p2, endPos})
	} else {
		mover.curve = curves.NewBezierNA([]vector.Vector2f{startPos, endPos})
	}

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()
	mover.first = false

	return 2
}

func (mover *MomentumMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
