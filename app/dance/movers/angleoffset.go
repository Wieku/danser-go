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

type AngleOffsetMover struct {
	*basicMover

	lastAngle float32
	lastPoint vector.Vector2f
	bz        *curves.Bezier
	endTime   float64
	invert    float32
}

func NewAngleOffsetMover() MultiPointMover {
	return &AngleOffsetMover{basicMover: &basicMover{}}
}

func (mover *AngleOffsetMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.lastAngle = 0
	mover.invert = 1
	mover.lastPoint = vector.NewVec2f(0, 0)
}

func (mover *AngleOffsetMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.Flower[mover.id%len(settings.CursorDance.MoverSettings.Flower)]

	end := objs[0]
	start := objs[1]

	endPos := end.GetStackedEndPositionMod(mover.diff.Mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(mover.diff.Mods)
	startTime := start.GetStartTime()

	distance := endPos.Dst(startPos)

	s1, ok1 := end.(objects.ILongObject)
	s2, ok2 := start.(objects.ILongObject)

	var points []vector.Vector2f

	scaledDistance := distance * float32(config.DistanceMult)
	newAngle := float32(config.AngleOffset) * math32.Pi / 180.0

	if end.GetStartTime() > 0 && config.LongJump >= 0 && (startTime-endTime) > float64(config.LongJump) {
		scaledDistance = float32(startTime-endTime) * float32(config.LongJumpMult)
	}

	if endPos == startPos {
		if config.LongJumpOnEqualPos {
			scaledDistance = float32(startTime-endTime) * float32(config.LongJumpMult)

			mover.lastAngle += math.Pi

			pt1 := vector.NewVec2fRad(mover.lastAngle, scaledDistance).Add(endPos)

			if ok1 {
				pt1 = vector.NewVec2fRad(s1.GetEndAngleMod(mover.diff.Mods), scaledDistance).Add(endPos)
			}

			if !ok2 {
				angle := mover.lastAngle - newAngle*mover.invert
				pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(startPos)

				mover.lastAngle = angle

				points = []vector.Vector2f{endPos, pt1, pt2, startPos}
			} else {
				pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(mover.diff.Mods), scaledDistance).Add(startPos)
				points = []vector.Vector2f{endPos, pt1, pt2, startPos}
			}
		} else {
			points = []vector.Vector2f{endPos, startPos}
		}
	} else if ok1 && ok2 {
		mover.invert = -1 * mover.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngleMod(mover.diff.Mods), scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(mover.diff.Mods), scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else if ok1 {
		mover.invert = -1 * mover.invert
		mover.lastAngle = endPos.AngleRV(startPos) - newAngle*mover.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngleMod(mover.diff.Mods), scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(mover.lastAngle, scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else if ok2 {
		mover.lastAngle += math.Pi

		pt1 := vector.NewVec2fRad(mover.lastAngle, scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(mover.diff.Mods), scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else {
		if bmath.AngleBetween32(endPos, mover.lastPoint, startPos) >= float32(config.AngleOffset)*math32.Pi/180.0 {
			mover.invert = -1 * mover.invert
			newAngle = float32(config.StreamAngleOffset) * math32.Pi / 180.0
		}

		angle := endPos.AngleRV(startPos) - newAngle*mover.invert

		pt1 := vector.NewVec2fRad(mover.lastAngle+math.Pi, scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(startPos)

		mover.lastAngle = angle

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	}

	mover.bz = curves.NewBezierNA(points)
	mover.endTime = endTime
	mover.startTime = startTime
	mover.lastPoint = endPos

	return 2
}

func (mover *AngleOffsetMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.bz.PointAt(t)
}
