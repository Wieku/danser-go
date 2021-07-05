package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type BezierMover struct {
	*basicMover

	pt            vector.Vector2f
	bz            *curves.Bezier
	endTime       float64
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

	end := objs[0]
	start := objs[1]
	endPos := end.GetStackedEndPositionMod(mover.diff.Mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(mover.diff.Mods)
	startTime := start.GetStartTime()

	dst := endPos.Dst(startPos)

	if mover.previousSpeed < 0 {
		mover.previousSpeed = dst / float32(startTime-endTime)
	}

	s1, ok1 := end.(objects.ILongObject)
	s2, ok2 := start.(objects.ILongObject)

	var points []vector.Vector2f

	genScale := mover.previousSpeed

	aggressiveness := float32(config.Aggressiveness)
	sliderAggressiveness := float32(config.SliderAggressiveness)

	if endPos == startPos {
		points = []vector.Vector2f{endPos, startPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngleMod(mover.diff.Mods)
		startAngle := s2.GetStartAngleMod(mover.diff.Mods)
		mover.pt = vector.NewVec2fRad(endAngle, s1.GetStackedPositionAtMod(endTime-10, mover.diff.Mods).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		pt2 := vector.NewVec2fRad(startAngle, s2.GetStackedPositionAtMod(startTime+10, mover.diff.Mods).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []vector.Vector2f{endPos, mover.pt, pt2, startPos}
	} else if ok1 {
		endAngle := s1.GetEndAngleMod(mover.diff.Mods)
		pt1 := vector.NewVec2fRad(endAngle, s1.GetStackedPositionAtMod(endTime-10, mover.diff.Mods).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		mover.pt = vector.NewVec2fRad(startPos.AngleRV(mover.pt), genScale*aggressiveness).Add(startPos)
		points = []vector.Vector2f{endPos, pt1, mover.pt, startPos}
	} else if ok2 {
		startAngle := s2.GetStartAngleMod(mover.diff.Mods)
		mover.pt = vector.NewVec2fRad(endPos.AngleRV(mover.pt), genScale*aggressiveness).Add(endPos)
		pt1 := vector.NewVec2fRad(startAngle, s2.GetStackedPositionAtMod(startTime+10, mover.diff.Mods).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		points = []vector.Vector2f{endPos, mover.pt, pt1, startPos}
	} else {
		angle := endPos.AngleRV(mover.pt)
		if math32.IsNaN(angle) {
			angle = 0
		}
		mover.pt = vector.NewVec2fRad(angle, mover.previousSpeed*aggressiveness).Add(endPos)

		points = []vector.Vector2f{endPos, mover.pt, startPos}
	}

	mover.bz = curves.NewBezierNA(points)

	mover.endTime = endTime
	mover.startTime = startTime
	mover.previousSpeed = (dst + 1.0) / float32(startTime-endTime)

	return 2
}

func (mover *BezierMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.bz.PointAt(t)
}
