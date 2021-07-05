package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type AxisMover struct {
	*basicMover

	bz      *curves.MultiCurve
	startTime float64
}

func NewAxisMover() MultiPointMover {
	return &AxisMover{basicMover: &basicMover{}}
}

func (mover *AxisMover) SetObjects(objs []objects.IHitObject) int {
	start, end := objs[0], objs[1]
	startPos := start.GetStackedEndPositionMod(mover.diff.Mods)
	startTime := start.GetEndTime()
	endPos := end.GetStackedStartPositionMod(mover.diff.Mods)
	endTime := end.GetStartTime()

	var midP vector.Vector2f

	if math32.Abs(endPos.Sub(startPos).X) < math32.Abs(endPos.Sub(endPos).X) {
		midP = vector.NewVec2f(startPos.X, endPos.Y)
	} else {
		midP = vector.NewVec2f(endPos.X, startPos.Y)
	}

	mover.bz = curves.NewMultiCurve("L", []vector.Vector2f{startPos, midP, endPos})
	mover.startTime = startTime
	mover.endTime = endTime

	return 2
}

func (mover AxisMover) Update(time float64) vector.Vector2f {
	t := float32(time-mover.startTime) / float32(mover.endTime-mover.startTime)
	tr := bmath.ClampF32(math32.Sin(t*math32.Pi/2), 0, 1)
	return mover.bz.PointAt(tr)
}
