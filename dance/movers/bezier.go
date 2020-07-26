package movers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
	"github.com/wieku/danser-go/bmath/math32"
	"github.com/wieku/danser-go/settings"
)

type BezierMover struct {
	pt                 bmath.Vector2f
	bz                 *curves.Bezier
	beginTime, endTime int64
	previousSpeed      float32
	invert             float32
}

func NewBezierMover() MultiPointMover {
	bm := &BezierMover{invert: 1}
	bm.pt = bmath.NewVec2f(512/2, 384/2)
	bm.previousSpeed = -1
	return bm
}

func (bm *BezierMover) Reset() {
	bm.pt = bmath.NewVec2f(512/2, 384/2)
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
		bm.previousSpeed = dst / float32(startTime-endTime)
	}

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []bmath.Vector2f

	genScale := bm.previousSpeed

	aggressiveness := float32(settings.Dance.Bezier.Aggressiveness)
	sliderAggressiveness := float32(settings.Dance.Bezier.SliderAggressiveness)

	if endPos == startPos {
		points = []bmath.Vector2f{endPos, startPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngle()
		startAngle := s2.GetStartAngle()
		bm.pt = bmath.NewVec2fRad(endAngle, s1.GetPointAt(endTime-10).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		pt2 := bmath.NewVec2fRad(startAngle, s2.GetPointAt(startTime+10).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []bmath.Vector2f{endPos, bm.pt, pt2, startPos}
	} else if ok1 {
		endAngle := s1.GetEndAngle()
		pt1 := bmath.NewVec2fRad(endAngle, s1.GetPointAt(endTime-10).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		bm.pt = bmath.NewVec2fRad(startPos.AngleRV(bm.pt), genScale*aggressiveness).Add(startPos)
		points = []bmath.Vector2f{endPos, pt1, bm.pt, startPos}
	} else if ok2 {
		startAngle := s2.GetStartAngle()
		bm.pt = bmath.NewVec2fRad(endPos.AngleRV(bm.pt), genScale*aggressiveness).Add(endPos)
		pt1 := bmath.NewVec2fRad(startAngle, s2.GetPointAt(startTime+10).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []bmath.Vector2f{endPos, bm.pt, pt1, startPos}
	} else {
		angle := endPos.AngleRV(bm.pt)
		if math32.IsNaN(angle) {
			angle = 0
		}
		bm.pt = bmath.NewVec2fRad(angle, bm.previousSpeed*aggressiveness).Add(endPos)

		points = []bmath.Vector2f{endPos, bm.pt, startPos}
	}

	bm.bz = curves.NewBezier(points)

	bm.endTime = endTime
	bm.beginTime = startTime
	bm.previousSpeed = (dst + 1.0) / float32(startTime-endTime)
}

func (bm *BezierMover) Update(time int64) bmath.Vector2f {
	return bm.bz.PointAt(float32(time-bm.endTime) / float32(bm.beginTime-bm.endTime))
}

func (bm *BezierMover) GetEndTime() int64 {
	return bm.beginTime
}
