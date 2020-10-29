package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type BezierMover struct {
	pt                 vector.Vector2f
	bz                 *curves.Bezier
	beginTime, endTime int64
	previousSpeed      float32
	invert             float32
}

func NewBezierMover() MultiPointMover {
	bm := &BezierMover{invert: 1}
	bm.pt = vector.NewVec2f(512/2, 384/2)
	bm.previousSpeed = -1
	return bm
}

func (bm *BezierMover) Reset() {
	bm.pt = vector.NewVec2f(512/2, 384/2)
	bm.invert = 1
	bm.previousSpeed = -1
}

func (bm *BezierMover) SetObjects(objs []objects.BaseObject) int {
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

	var points []vector.Vector2f

	genScale := bm.previousSpeed

	aggressiveness := float32(settings.Dance.Bezier.Aggressiveness)
	sliderAggressiveness := float32(settings.Dance.Bezier.SliderAggressiveness)

	if endPos == startPos {
		points = []vector.Vector2f{endPos, startPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngle()
		startAngle := s2.GetStartAngle()
		bm.pt = vector.NewVec2fRad(endAngle, s1.GetPointAt(endTime-10).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		pt2 := vector.NewVec2fRad(startAngle, s2.GetPointAt(startTime+10).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []vector.Vector2f{endPos, bm.pt, pt2, startPos}
	} else if ok1 {
		endAngle := s1.GetEndAngle()
		pt1 := vector.NewVec2fRad(endAngle, s1.GetPointAt(endTime-10).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		bm.pt = vector.NewVec2fRad(startPos.AngleRV(bm.pt), genScale*aggressiveness).Add(startPos)
		points = []vector.Vector2f{endPos, pt1, bm.pt, startPos}
	} else if ok2 {
		startAngle := s2.GetStartAngle()
		bm.pt = vector.NewVec2fRad(endPos.AngleRV(bm.pt), genScale*aggressiveness).Add(endPos)
		pt1 := vector.NewVec2fRad(startAngle, s2.GetPointAt(startTime+10).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []vector.Vector2f{endPos, bm.pt, pt1, startPos}
	} else {
		angle := endPos.AngleRV(bm.pt)
		if math32.IsNaN(angle) {
			angle = 0
		}
		bm.pt = vector.NewVec2fRad(angle, bm.previousSpeed*aggressiveness).Add(endPos)

		points = []vector.Vector2f{endPos, bm.pt, startPos}
	}

	bm.bz = curves.NewBezier(points)

	bm.endTime = endTime
	bm.beginTime = startTime
	bm.previousSpeed = (dst + 1.0) / float32(startTime-endTime)

	return 2
}

func (bm *BezierMover) Update(time int64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.beginTime-bm.endTime), 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *BezierMover) GetEndTime() int64 {
	return bm.beginTime
}
