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
	startTime       float64
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

	fff := objs[0]
	end := objs[1]

	startPos := fff.GetStackedEndPositionMod(mover.diff.Mods)
	startTime := fff.GetEndTime()
	endPos := end.GetStackedStartPositionMod(mover.diff.Mods)
	endTime := end.GetStartTime()

	dst := startPos.Dst(endPos)

	if mover.previousSpeed < 0 {
		mover.previousSpeed = dst / float32(endTime-startTime)
	}

	s1, ok1 := fff.(objects.ILongObject)
	s2, ok2 := end.(objects.ILongObject)

	var points []vector.Vector2f

	genScale := mover.previousSpeed

	aggressiveness := float32(config.Aggressiveness)
	sliderAggressiveness := float32(config.SliderAggressiveness)

	if startPos == endPos {
		points = []vector.Vector2f{startPos, endPos}
	} else if ok1 && ok2 {
		endAngle := s1.GetEndAngleMod(mover.diff.Mods)
		startAngle := s2.GetStartAngleMod(mover.diff.Mods)
		mover.pt = vector.NewVec2fRad(endAngle, s1.GetStackedPositionAtMod(startTime-10, mover.diff.Mods).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		pt2 := vector.NewVec2fRad(startAngle, s2.GetStackedPositionAtMod(endTime+10, mover.diff.Mods).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		points = []vector.Vector2f{startPos, mover.pt, pt2, endPos}
	} else if ok1 {
		endAngle := s1.GetEndAngleMod(mover.diff.Mods)
		pt1 := vector.NewVec2fRad(endAngle, s1.GetStackedPositionAtMod(startTime-10, mover.diff.Mods).Dst(startPos)*aggressiveness*sliderAggressiveness/10).Add(startPos)
		mover.pt = vector.NewVec2fRad(endPos.AngleRV(mover.pt), genScale*aggressiveness).Add(endPos)
		points = []vector.Vector2f{startPos, pt1, mover.pt, endPos}
	} else if ok2 {
		startAngle := s2.GetStartAngleMod(mover.diff.Mods)
		mover.pt = vector.NewVec2fRad(startPos.AngleRV(mover.pt), genScale*aggressiveness).Add(startPos)
		pt1 := vector.NewVec2fRad(startAngle, s2.GetStackedPositionAtMod(endTime+10, mover.diff.Mods).Dst(endPos)*aggressiveness*sliderAggressiveness/10).Add(endPos)
		points = []vector.Vector2f{startPos, mover.pt, pt1, endPos}
	} else {
		angle := startPos.AngleRV(mover.pt)
		if math32.IsNaN(angle) {
			angle = 0
		}
		mover.pt = vector.NewVec2fRad(angle, mover.previousSpeed*aggressiveness).Add(startPos)

		points = []vector.Vector2f{startPos, mover.pt, endPos}
	}

	mover.bz = curves.NewBezierNA(points)

	mover.startTime = startTime
	mover.endTime = endTime
	mover.previousSpeed = (dst + 1.0) / float32(endTime-startTime)

	return 2
}

func (mover *BezierMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.startTime)/float32(mover.endTime-mover.startTime), 0, 1)
	return mover.bz.PointAt(t)
}
