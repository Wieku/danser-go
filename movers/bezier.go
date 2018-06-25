package movers

import (
	math2 "github.com/wieku/danser/bmath"
	"github.com/wieku/danser/bmath/curves"
	//"osubot/io"
	"github.com/wieku/danser/beatmap/objects"
	"math"
	/*"github.com/wieku/danser/render"*/
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/settings"
)

type BezierMover struct {
	pt math2.Vector2d
	bz curves.Bezier
	beginTime, endTime int64
	previousSpeed float64
	invert float64
}

func NewBezierMover() Mover {
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
		bm.previousSpeed = dst / float64(startTime - endTime)
	}

	s1, ok1 := end.(*objects.Slider)
	s2, ok2 := start.(*objects.Slider)

	var points []math2.Vector2d

	genScale := bm.previousSpeed

	aggressiveness := settings.Dance.Bezier.Aggressiveness
	sliderAggressiveness := settings.Dance.Bezier.SliderAggressiveness
	
	if endPos == startPos {
		points = []math2.Vector2d{endPos, startPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngle()
		startAngle := s2.GetStartAngle()
		bm.pt = math2.NewVec2dRad(endAngle,  genScale * aggressiveness * sliderAggressiveness).Add(endPos)
		pt2 := math2.NewVec2dRad(startAngle, genScale * aggressiveness * sliderAggressiveness).Add(startPos)
		points = []math2.Vector2d{endPos, bm.pt, pt2, startPos}
	} else if ok1 {
		endAngle := s1.GetEndAngle()
		pt1 := math2.NewVec2dRad(endAngle,  genScale * aggressiveness * sliderAggressiveness).Add(endPos)
		bm.pt = math2.NewVec2dRad(startPos.AngleRV(bm.pt), genScale * aggressiveness).Add(startPos)
		points = []math2.Vector2d{endPos, pt1, bm.pt, startPos}
	} else if ok2 {
		startAngle := s2.GetStartAngle()

		bm.pt = math2.NewVec2dRad(endPos.AngleRV(bm.pt), genScale * aggressiveness).Add(endPos)
		pt1 := math2.NewVec2dRad(startAngle, genScale * aggressiveness * sliderAggressiveness).Add(startPos)

		points = []math2.Vector2d{endPos, bm.pt, pt1, startPos}
	} else {
		angle := endPos.AngleRV(bm.pt)
		if math.IsNaN(angle) {
			angle = 0
		}
		bm.pt = math2.NewVec2dRad(angle, bm.previousSpeed * aggressiveness).Add(endPos)

		points = []math2.Vector2d{endPos, bm.pt, startPos}
	}

	bm.bz = curves.NewBezier(points)

	bm.endTime = endTime
	bm.beginTime = startTime
	bm.previousSpeed = (dst+1.0) / float64(startTime-endTime)
}

func (bm BezierMover) Update(time int64, cursor *render.Cursor) {
	cursor.SetPos(bm.bz.NPointAt(float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)))
}