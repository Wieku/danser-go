package movers

import (
	math2 "danser/bmath"
	"danser/bmath/curves"
	//"osubot/io"
	"danser/beatmap/objects"
	"math"
	/*"danser/render"*/
)

const (
	BEZIER_AGGRESSIVENESS = 60.0        // TODO:
	BEZIER_SLIDER_AGGRESSIVENESS = 3  // make these parameters changeable at runtime
)

type BezierMover struct {
	pt math2.Vector2d
	bz curves.Bezier
	beginTime, endTime int64
	previousSpeed float64
	invert float64
}

func NewBezierMover() *BezierMover {
	bm := &BezierMover{invert:1}
	bm.pt = math2.NewVec2d(512/2, 384/2)
	bm.previousSpeed = -1
	return bm
}

func (bm *BezierMover) Reset() {
	bm.pt = math2.NewVec2d(512/2, 384/2)
	bm.invert = 1
	bm.previousSpeed = -1
}

func (bm *BezierMover) SetObjects(end, start objects.BaseObject) {
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	dst := endPos.Dst(startPos)

	if bm.previousSpeed < 0 {
		bm.previousSpeed = dst / float64(startTime - endTime) * BEZIER_AGGRESSIVENESS * sliderMult(end, start)
	}

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []math2.Vector2d

	genScale := /*bmath.Sqrt(dst / 150.0)*/1.0

	if endPos == startPos {
		points = []math2.Vector2d{endPos, startPos}
	} else if ((!ok1 && !ok2) || (ok1 && !ok2) || (!ok1 && ok2)) && startTime - endTime > 500 && dst < 20 {
		mid := endPos.Mid(startPos)
		m2 := mid.Sub(math2.NewVec2d(512/2, 384/2))

		bm.pt = math2.NewVec2d(m2.X/math.Abs(m2.X) * math.Sqrt(float64(startTime - endTime))*12, 0).Add(mid)

		points = []math2.Vector2d{endPos, bm.pt, startPos}
	} else if startTime - end.GetBasicData().StartTime < 100 || dst < 50 {
		mid := endPos.Mid(startPos)
		bm.pt = mid.Sub(endPos).Rotate(bm.invert * math.Pi / 2).Add(mid)
		bm.invert = -bm.invert
		points = []math2.Vector2d{endPos, bm.pt, startPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngle()
		startAngle := s2.GetStartAngle()
		dst1 := s1.GetPointAt(endTime-10).Dst(endPos)
		dst2 := s2.GetPointAt(startTime+10).Dst(startPos)
		pt1 := math2.NewVec2dRad(endAngle,  genScale * dst1* BEZIER_AGGRESSIVENESS*BEZIER_SLIDER_AGGRESSIVENESS/15.0).Add(endPos)
		bm.pt = pt1
		pt2 := math2.NewVec2dRad(startAngle, genScale * dst2*BEZIER_AGGRESSIVENESS*BEZIER_SLIDER_AGGRESSIVENESS/15.0).Add(startPos)
		points = []math2.Vector2d{endPos, pt1, pt2, startPos}
	} else if ok1 {
		endAngle := s1.GetEndAngle()
		dst1 := s1.GetPointAt(endTime-10).Dst(endPos)
		bm.pt = math2.NewVec2dRad(endAngle, genScale * dst1*BEZIER_AGGRESSIVENESS*BEZIER_SLIDER_AGGRESSIVENESS/10.0).Add(endPos)
		points = []math2.Vector2d{endPos, bm.pt, startPos}
	} else if ok2 {
		startAngle := s2.GetStartAngle()
		dst2 := s2.GetPointAt(startTime+10).Dst(startPos)
		bm.pt = math2.NewVec2dRad(endPos.AngleRV(bm.pt), genScale * dst2*BEZIER_AGGRESSIVENESS*BEZIER_SLIDER_AGGRESSIVENESS/10.0).Add(endPos)
		pt1 := math2.NewVec2dRad(startAngle, genScale * dst2*BEZIER_AGGRESSIVENESS*BEZIER_SLIDER_AGGRESSIVENESS/10.0).Add(startPos)

		point := getMPoint(bm.pt, startPos, endPos)
		pt2 := bm.pt
		scale := math.Min(4, bm.pt.Dst(startPos)/startPos.Dst(endPos))
		bm.pt = point.Add(point.Sub(endPos).Scl(scale))

		points = []math2.Vector2d{endPos, pt2, bm.pt, pt1, startPos}
	} else {
		angle := endPos.AngleRV(bm.pt)
		if math.IsNaN(angle) {
			angle = 0
		}
		bm.pt = math2.NewVec2dRad(angle, bm.previousSpeed).Add(endPos)

		point := getMPoint(bm.pt, startPos, endPos)
		pt1 := bm.pt
		scale := math.Min(bm.pt.Dst(startPos)/startPos.Dst(endPos), 4)

 		bm.pt = point.Add(point.Sub(endPos).Scl(scale))

		points = []math2.Vector2d{endPos, pt1, bm.pt, startPos}
	}

	bm.bz = curves.NewBezier(points)

	bm.endTime = endTime
	bm.beginTime = startTime
	bm.previousSpeed = (dst+1.0) / float64(startTime-endTime) * BEZIER_AGGRESSIVENESS * sliderMult(end, start)
}

func (bm BezierMover) Update(time int64/*, cursor *render.Cursor*/) {
	//cursor.SetPos(bm.bz.NPointAt(float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)))
	//io.MouseMoveVec(bm.bz.NPointAt(float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)))
}