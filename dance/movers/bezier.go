package movers

import (
	"math"
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
)

type BezierMover struct {
	pt                 bmath.Vector2d
	bz                 *curves.Bezier
	beginTime, endTime int64
	previousSpeed      float64
	invert             float64
}

func NewBezierMover() MultiPointMover {
	bm := &BezierMover{invert: 1}
	bm.pt = bmath.NewVec2d(512/2, 384/2)
	bm.previousSpeed = -1
	return bm
}

func (bm *BezierMover) Reset() {
	bm.pt = bmath.NewVec2d(512/2, 384/2)
	bm.invert = 1
	bm.previousSpeed = -1
}

func (bm *BezierMover) SetObjects(objs []objects.BaseObject) {
	end := objs[0]
	start := objs[1]
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	dst := endPos.Dst(startPos)

	if bm.previousSpeed < 0 {
		bm.previousSpeed = dst / float64(startTime-endTime)
	}

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []bmath.Vector2d

	genScale := bm.previousSpeed

	aggressiveness := settings.Dance.Bezier.Aggressiveness
	sliderAggressiveness := settings.Dance.Bezier.SliderAggressiveness

	if endPos == startPos {
		points = []bmath.Vector2d{endPos, startPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngle()
		startAngle := s2.GetStartAngle()
		bm.pt = bmath.NewVec2dRad(endAngle, s1.GetPointAt(endTime - 10).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		pt2 := bmath.NewVec2dRad(startAngle, s2.GetPointAt(startTime + 10).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []bmath.Vector2d{endPos, bm.pt, pt2, startPos}
	} else if ok1 {
		endAngle := s1.GetEndAngle()
		pt1 := bmath.NewVec2dRad(endAngle, s1.GetPointAt(endTime - 10).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		bm.pt = bmath.NewVec2dRad(startPos.AngleRV(bm.pt), genScale*aggressiveness).Add(startPos)
		points = []bmath.Vector2d{endPos, pt1, bm.pt, startPos}
	} else if ok2 {
		startAngle := s2.GetStartAngle()
		bm.pt = bmath.NewVec2dRad(endPos.AngleRV(bm.pt), genScale*aggressiveness).Add(endPos)
		pt1 := bmath.NewVec2dRad(startAngle, s2.GetPointAt(startTime + 10).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []bmath.Vector2d{endPos, bm.pt, pt1, startPos}
	} else {
		angle := endPos.AngleRV(bm.pt)
		if math.IsNaN(angle) {
			angle = 0
		}
		bm.pt = bmath.NewVec2dRad(angle, bm.previousSpeed*aggressiveness).Add(endPos)

		points = []bmath.Vector2d{endPos, bm.pt, startPos}
	}

	bm.bz = curves.NewBezier(points)

	bm.endTime = endTime
	bm.beginTime = startTime
	bm.previousSpeed = (dst + 1.0) / float64(startTime-endTime)
}

func (bm *BezierMover) Update(time int64) bmath.Vector2d {
	return bm.bz.NPointAt(float64(time-bm.endTime) / float64(bm.beginTime-bm.endTime))
}

func (bm *BezierMover) GetEndTime() int64 {
	return bm.beginTime
}
