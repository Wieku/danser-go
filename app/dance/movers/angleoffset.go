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

type AngleOffsetMover struct {
	*basicMover

	curve *curves.Bezier

	lastAngle float32
	lastPoint vector.Vector2f
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

	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	timeDelta := mover.endTime - mover.startTime

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	distance := startPos.Dst(endPos)

	s1, ok1 := start.(objects.ILongObject)
	s2, ok2 := end.(objects.ILongObject)

	var points []vector.Vector2f

	scaledDistance := distance * float32(config.DistanceMult)
	newAngle := float32(config.AngleOffset) * math32.Pi / 180.0

	if start.GetStartTime() > 0 && config.LongJump >= 0 && timeDelta > float64(config.LongJump) {
		scaledDistance = float32(timeDelta) * float32(config.LongJumpMult)
	}

	if startPos == endPos {
		if config.LongJumpOnEqualPos {
			scaledDistance = float32(timeDelta) * float32(config.LongJumpMult)

			mover.lastAngle += math.Pi

			pt1 := vector.NewVec2fRad(mover.lastAngle, scaledDistance).Add(startPos)

			if ok1 {
				pt1 = vector.NewVec2fRad(s1.GetEndAngleMod(mover.diff), scaledDistance).Add(startPos)
			}

			if !ok2 {
				angle := mover.lastAngle - newAngle*mover.invert
				pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(endPos)

				mover.lastAngle = angle

				points = []vector.Vector2f{startPos, pt1, pt2, endPos}
			} else {
				pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(mover.diff), scaledDistance).Add(endPos)
				points = []vector.Vector2f{startPos, pt1, pt2, endPos}
			}
		} else {
			points = []vector.Vector2f{startPos, endPos}
		}
	} else if ok1 && ok2 {
		mover.invert = -1 * mover.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngleMod(mover.diff), scaledDistance).Add(startPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(mover.diff), scaledDistance).Add(endPos)

		points = []vector.Vector2f{startPos, pt1, pt2, endPos}
	} else if ok1 {
		mover.invert = -1 * mover.invert
		mover.lastAngle = startPos.AngleRV(endPos) - newAngle*mover.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngleMod(mover.diff), scaledDistance).Add(startPos)
		pt2 := vector.NewVec2fRad(mover.lastAngle, scaledDistance).Add(endPos)

		points = []vector.Vector2f{startPos, pt1, pt2, endPos}
	} else if ok2 {
		mover.lastAngle += math.Pi

		pt1 := vector.NewVec2fRad(mover.lastAngle, scaledDistance).Add(startPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(mover.diff), scaledDistance).Add(endPos)

		points = []vector.Vector2f{startPos, pt1, pt2, endPos}
	} else {
		if vector.AngleBetween32(startPos, mover.lastPoint, endPos) >= float32(config.AngleOffset)*math32.Pi/180.0 {
			mover.invert = -1 * mover.invert
			newAngle = float32(config.StreamAngleOffset) * math32.Pi / 180.0
		}

		angle := startPos.AngleRV(endPos) - newAngle*mover.invert

		pt1 := vector.NewVec2fRad(mover.lastAngle+math.Pi, scaledDistance).Add(startPos)
		pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(endPos)

		mover.lastAngle = angle

		points = []vector.Vector2f{startPos, pt1, pt2, endPos}
	}

	mover.curve = curves.NewBezierNA(points)
	mover.lastPoint = startPos

	return 2
}

func (mover *AngleOffsetMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
