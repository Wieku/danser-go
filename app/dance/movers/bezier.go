package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

type BezierMover struct {
	*basicMover

	curve *curves.Bezier

	pt            vector.Vector2f
	previousSpeed float32
	invert        float32
}

func NewBezierMover() MultiPointMover {
	return &BezierMover{basicMover: &basicMover{}}
}

func (mover *BezierMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.pt = vector.NewVec2f(512/2, 384/2)
	mover.invert = 1
	mover.previousSpeed = -1
}

func (mover *BezierMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.Bezier[mover.id%len(settings.CursorDance.MoverSettings.Bezier)]

	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	dst := startPos.Dst(endPos)

	if mover.previousSpeed < 0 {
		mover.previousSpeed = dst / float32(mover.endTime-mover.startTime)
	}

	s1, ok1 := start.(objects.ILongObject)
	s2, ok2 := end.(objects.ILongObject)

	var points []vector.Vector2f

	genScale := mover.previousSpeed

	aggressiveness := float32(config.Aggressiveness)
	sliderAggressiveness := float32(config.SliderAggressiveness)

	if startPos == endPos {
		points = []vector.Vector2f{startPos, endPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngleMod(mover.diff)
		startAngle := s2.GetStartAngleMod(mover.diff)
		mover.pt = vector.NewVec2fRad(endAngle, s1.GetStackedPositionAtMod(mover.startTime-10, mover.diff).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		pt2 := vector.NewVec2fRad(startAngle, s2.GetStackedPositionAtMod(mover.endTime+10, mover.diff).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		points = []vector.Vector2f{startPos, mover.pt, pt2, endPos}
	} else if ok1 {
		endAngle := s1.GetEndAngleMod(mover.diff)
		pt1 := vector.NewVec2fRad(endAngle, s1.GetStackedPositionAtMod(mover.startTime-10, mover.diff).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		mover.pt = vector.NewVec2fRad(endPos.AngleRV(mover.pt), genScale*aggressiveness).Add(endPos)
		points = []vector.Vector2f{startPos, pt1, mover.pt, endPos}
	} else if ok2 {
		startAngle := s2.GetStartAngleMod(mover.diff)
		mover.pt = vector.NewVec2fRad(startPos.AngleRV(mover.pt), genScale*aggressiveness).Add(startPos)
		pt1 := vector.NewVec2fRad(startAngle, s2.GetStackedPositionAtMod(mover.endTime+10, mover.diff).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		points = []vector.Vector2f{startPos, mover.pt, pt1, endPos}
	} else {
		angle := startPos.AngleRV(mover.pt)
		if math32.IsNaN(angle) {
			angle = 0
		}
		mover.pt = vector.NewVec2fRad(angle, mover.previousSpeed*aggressiveness).Add(startPos)

		points = []vector.Vector2f{startPos, mover.pt, endPos}
	}

	mover.curve = curves.NewBezierNA(points)

	mover.previousSpeed = (dst + 1.0) / float32(mover.endTime-mover.startTime)

	return 2
}

func (mover *BezierMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
