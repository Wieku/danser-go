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
	lastAngle          float32
	lastPoint          vector.Vector2f
	bz                 *curves.Bezier
	startTime, endTime float64
	invert             float32
	diff               *difficulty.Difficulty
	id                 int
}

func NewAngleOffsetMover() MultiPointMover {
	return &AngleOffsetMover{lastAngle: 0, invert: 1}
}

func (bm *AngleOffsetMover) Reset(diff *difficulty.Difficulty, id int) {
	bm.diff = diff
	bm.lastAngle = 0
	bm.invert = 1
	bm.lastPoint = vector.NewVec2f(0, 0)
	bm.id = id
}

func (bm *AngleOffsetMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.Flower[bm.id%len(settings.CursorDance.MoverSettings.Flower)]

	end := objs[0]
	start := objs[1]

	endPos := end.GetStackedEndPositionMod(bm.diff.Mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(bm.diff.Mods)
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

			bm.lastAngle += math.Pi

			pt1 := vector.NewVec2fRad(bm.lastAngle, scaledDistance).Add(endPos)

			if ok1 {
				pt1 = vector.NewVec2fRad(s1.GetEndAngleMod(bm.diff.Mods), scaledDistance).Add(endPos)
			}

			if !ok2 {
				angle := bm.lastAngle - newAngle*bm.invert
				pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(startPos)

				bm.lastAngle = angle

				points = []vector.Vector2f{endPos, pt1, pt2, startPos}
			} else {
				pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(bm.diff.Mods), scaledDistance).Add(startPos)
				points = []vector.Vector2f{endPos, pt1, pt2, startPos}
			}
		} else {
			points = []vector.Vector2f{endPos, startPos}
		}
	} else if ok1 && ok2 {
		bm.invert = -1 * bm.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngleMod(bm.diff.Mods), scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(bm.diff.Mods), scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else if ok1 {
		bm.invert = -1 * bm.invert
		bm.lastAngle = endPos.AngleRV(startPos) - newAngle*bm.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngleMod(bm.diff.Mods), scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(bm.lastAngle, scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else if ok2 {
		bm.lastAngle += math.Pi

		pt1 := vector.NewVec2fRad(bm.lastAngle, scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngleMod(bm.diff.Mods), scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else {
		if bmath.AngleBetween32(endPos, bm.lastPoint, startPos) >= float32(config.AngleOffset)*math32.Pi/180.0 {
			bm.invert = -1 * bm.invert
			newAngle = float32(config.StreamAngleOffset) * math32.Pi / 180.0
		}

		angle := endPos.AngleRV(startPos) - newAngle*bm.invert

		pt1 := vector.NewVec2fRad(bm.lastAngle+math.Pi, scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(startPos)

		bm.lastAngle = angle

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	}

	bm.bz = curves.NewBezierNA(points)
	bm.endTime = endTime
	bm.startTime = startTime
	bm.lastPoint = endPos

	return 2
}

func (bm *AngleOffsetMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *AngleOffsetMover) GetEndTime() float64 {
	return bm.startTime
}
