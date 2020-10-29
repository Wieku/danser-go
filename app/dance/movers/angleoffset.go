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

type AngleOffsetMover struct {
	lastAngle          float32
	lastPoint          vector.Vector2f
	bz                 *curves.Bezier
	startTime, endTime int64
	invert             float32
}

func NewAngleOffsetMover() MultiPointMover {
	return &AngleOffsetMover{lastAngle: 0, invert: 1}
}

func (bm *AngleOffsetMover) Reset() {
	bm.lastAngle = 0
	bm.invert = 1
	bm.lastPoint = vector.NewVec2f(0, 0)
}

func (bm *AngleOffsetMover) SetObjects(objs []objects.BaseObject) int {
	end := objs[0]
	start := objs[1]

	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	distance := endPos.Dst(startPos)

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []vector.Vector2f

	scaledDistance := distance * float32(settings.Dance.Flower.DistanceMult)
	newAngle := float32(settings.Dance.Flower.AngleOffset) * math32.Pi / 180.0

	if end.GetBasicData().StartTime > 0 && settings.Dance.Flower.LongJump >= 0 && (startTime-endTime) > settings.Dance.Flower.LongJump {
		scaledDistance = float32(startTime-endTime) * float32(settings.Dance.Flower.LongJumpMult)
	}

	if endPos == startPos {
		if settings.Dance.Flower.LongJumpOnEqualPos {
			scaledDistance = float32(startTime-endTime) * float32(settings.Dance.Flower.LongJumpMult)

			if math.Abs(float64(startTime-endTime)) > 1 {
				bm.lastAngle += math.Pi
			}

			pt1 := vector.NewVec2fRad(bm.lastAngle, scaledDistance).Add(endPos)

			if ok1 {
				pt1 = vector.NewVec2fRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
			}

			if !ok2 {
				angle := bm.lastAngle - newAngle*bm.invert
				pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(startPos)

				if math.Abs(float64(startTime-endTime)) > 1 {
					bm.lastAngle = angle
				}

				points = []vector.Vector2f{endPos, pt1, pt2, startPos}
			} else {
				pt2 := vector.NewVec2fRad(s2.GetStartAngle(), scaledDistance).Add(startPos)
				points = []vector.Vector2f{endPos, pt1, pt2, startPos}
			}
		} else {
			points = []vector.Vector2f{endPos, startPos}
		}
	} else if ok1 && ok2 {
		bm.invert = -1 * bm.invert

		pt1 := vector.NewVec2fRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngle(), scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else if ok1 {
		bm.invert = -1 * bm.invert
		if math.Abs(float64(startTime-endTime)) > 1 {
			bm.lastAngle = endPos.AngleRV(startPos) - newAngle*bm.invert
		} else {
			bm.lastAngle = s1.GetEndAngle() + math.Pi
		}

		pt1 := vector.NewVec2fRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(bm.lastAngle, scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else if ok2 {
		if math.Abs(float64(startTime-endTime)) > 1 {
			bm.lastAngle += math.Pi
		}

		pt1 := vector.NewVec2fRad(bm.lastAngle, scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(s2.GetStartAngle(), scaledDistance).Add(startPos)

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	} else {
		if math.Abs(float64(startTime-endTime)) > 1 && bmath.AngleBetween32(endPos, bm.lastPoint, startPos) >= float32(settings.Dance.Flower.AngleOffset)*math32.Pi/180.0 {
			bm.invert = -1 * bm.invert
			newAngle = float32(settings.Dance.Flower.StreamAngleOffset) * math32.Pi / 180.0
		}

		angle := endPos.AngleRV(startPos) - newAngle*bm.invert
		if math.Abs(float64(startTime-endTime)) <= 1 {
			angle = bm.lastAngle
		}

		pt1 := vector.NewVec2fRad(bm.lastAngle+math.Pi, scaledDistance).Add(endPos)
		pt2 := vector.NewVec2fRad(angle, scaledDistance).Add(startPos)

		if scaledDistance > 2 {
			bm.lastAngle = angle
		}

		points = []vector.Vector2f{endPos, pt1, pt2, startPos}
	}

	bm.bz = curves.NewBezierNA(points)
	bm.endTime = endTime
	bm.startTime = startTime
	bm.lastPoint = endPos

	return 2
}

func (bm *AngleOffsetMover) Update(time int64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *AngleOffsetMover) GetEndTime() int64 {
	return bm.startTime
}
