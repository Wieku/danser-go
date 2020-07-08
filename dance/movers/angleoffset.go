package movers

import (
	"math"
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
)

type AngleOffsetMover struct {
	lastAngle          float64
	lastPoint          bmath.Vector2d
	bz                 *curves.Bezier
	startTime, endTime int64
	invert             float64
}

func NewAngleOffsetMover() MultiPointMover {
	return &AngleOffsetMover{lastAngle: 0, invert: 1}
}

func (bm *AngleOffsetMover) Reset() {
	bm.lastAngle = 0
	bm.invert = 1
	bm.lastPoint = bmath.NewVec2d(0, 0)
}

func (bm *AngleOffsetMover) SetObjects(objs []objects.BaseObject) {
	end := objs[0]
	start := objs[1]

	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	distance := endPos.Dst(startPos)

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []bmath.Vector2d

	scaledDistance := distance * settings.Dance.Flower.DistanceMult
	newAngle := settings.Dance.Flower.AngleOffset * math.Pi / 180.0

	if end.GetBasicData().StartTime > 0 && settings.Dance.Flower.LongJump >= 0 && (startTime-endTime) > settings.Dance.Flower.LongJump {
		scaledDistance = float64(startTime-endTime) * settings.Dance.Flower.LongJumpMult
	}

	if endPos == startPos {
		if settings.Dance.Flower.LongJumpOnEqualPos {
			scaledDistance = float64(startTime-endTime) * settings.Dance.Flower.LongJumpMult
			if math.Abs(float64(startTime-endTime)) > 1 {
				bm.lastAngle += math.Pi
			}

			pt1 := bmath.NewVec2dRad(bm.lastAngle, scaledDistance).Add(endPos)

			if ok1 {
				pt1 = bmath.NewVec2dRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
			}

			if !ok2 {
				angle := bm.lastAngle - newAngle*bm.invert
				pt2 := bmath.NewVec2dRad(angle, scaledDistance).Add(startPos)
				if math.Abs(float64(startTime-endTime)) > 1 {
					bm.lastAngle = angle
				}
				points = []bmath.Vector2d{endPos, pt1, pt2, startPos}
			} else {
				pt2 := bmath.NewVec2dRad(s2.GetStartAngle(), scaledDistance).Add(startPos)
				points = []bmath.Vector2d{endPos, pt1, pt2, startPos}
			}

		} else {
			points = []bmath.Vector2d{endPos, startPos}
		}
	} else if ok1 && ok2 {
		bm.invert = -1 * bm.invert

		pt1 := bmath.NewVec2dRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
		pt2 := bmath.NewVec2dRad(s2.GetStartAngle(), scaledDistance).Add(startPos)

		points = []bmath.Vector2d{endPos, pt1, pt2, startPos}
	} else if ok1 {
		bm.invert = -1 * bm.invert
		if math.Abs(float64(startTime-endTime)) > 1 {
			bm.lastAngle = endPos.AngleRV(startPos) - newAngle*bm.invert
		} else {
			bm.lastAngle = s1.GetEndAngle()+math.Pi
		}

		pt1 := bmath.NewVec2dRad(s1.GetEndAngle(), scaledDistance).Add(endPos)
		pt2 := bmath.NewVec2dRad(bm.lastAngle, scaledDistance).Add(startPos)

		points = []bmath.Vector2d{endPos, pt1, pt2, startPos}
	} else if ok2 {
		if math.Abs(float64(startTime-endTime)) > 1 {
			bm.lastAngle += math.Pi
		}

		pt1 := bmath.NewVec2dRad(bm.lastAngle, scaledDistance).Add(endPos)
		pt2 := bmath.NewVec2dRad(s2.GetStartAngle(), scaledDistance).Add(startPos)

		points = []bmath.Vector2d{endPos, pt1, pt2, startPos}
	} else {
		if settings.Dance.Flower.UseNewStyle {
			if math.Abs(float64(startTime-endTime)) > 1 && bmath.AngleBetween(endPos, bm.lastPoint, startPos) >= settings.Dance.Flower.AngleOffset*math.Pi/180.0 {
				bm.invert = -1 * bm.invert
				newAngle = settings.Dance.Flower.StreamAngleOffset * math.Pi / 180.0
			}
		} else if startTime-endTime < settings.Dance.Flower.StreamTrigger {
			newAngle = settings.Dance.Flower.StreamAngleOffset * math.Pi / 180.0
		}

		angle := endPos.AngleRV(startPos) - newAngle*bm.invert
		if math.Abs(float64(startTime-endTime)) <= 1 {
			angle = bm.lastAngle
		}


		pt1 := bmath.NewVec2dRad(bm.lastAngle+math.Pi, scaledDistance).Add(endPos)
		pt2 := bmath.NewVec2dRad(angle, scaledDistance).Add(startPos)

		if scaledDistance > 2 {
			bm.lastAngle = angle
		}

		if !settings.Dance.Flower.UseNewStyle && startTime-endTime < settings.Dance.Flower.StreamTrigger && !(start.GetBasicData().SliderPoint && end.GetBasicData().SliderPoint) {
			bm.invert = -1 * bm.invert
		}

		points = []bmath.Vector2d{endPos, pt1, pt2, startPos}
	}

	bm.bz = curves.NewBezierNA(points)
	bm.endTime = endTime
	bm.startTime = startTime
	bm.lastPoint = endPos
}

func (bm *AngleOffsetMover) Update(time int64) bmath.Vector2d {
	t := float64(time-bm.endTime) / float64(bm.startTime-bm.endTime)
	t = math.Max(0.0, math.Min(1.0, t))
	return bm.bz.NPointAt(t)
}

func (bm *AngleOffsetMover) GetEndTime() int64 {
	return bm.startTime
}
