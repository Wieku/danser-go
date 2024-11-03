package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

type AxisMover struct {
	*basicMover

	curve *curves.MultiCurve
}

func NewAxisMover() MultiPointMover {
	return &AxisMover{basicMover: &basicMover{}}
}

func (mover *AxisMover) SetObjects(objs []objects.IHitObject) int {
	start, end := objs[0], objs[1]

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()

	startPos := start.GetStackedEndPositionMod(mover.diff)
	endPos := end.GetStackedStartPositionMod(mover.diff)

	var midP vector.Vector2f

	if math32.Abs(endPos.Sub(startPos).X) < math32.Abs(endPos.Sub(endPos).X) {
		midP = vector.NewVec2f(startPos.X, endPos.Y)
	} else {
		midP = vector.NewVec2f(endPos.X, startPos.Y)
	}

	mover.curve = curves.NewMultiCurve([]curves.CurveDef{{CurveType: curves.CLine, Points: []vector.Vector2f{startPos, midP, endPos}}})

	return 2
}

func (mover AxisMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(easing.OutSine(t)))
}
