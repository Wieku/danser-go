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
	endTime float64
}

func NewAxisMover() MultiPointMover {
	return &AxisMover{basicMover: &basicMover{}}
}

func (mover *AxisMover) SetObjects(objs []objects.IHitObject) int {
	end, start := objs[0], objs[1]
	endPos := end.GetStackedEndPositionMod(mover.diff.Mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(mover.diff.Mods)
	startTime := start.GetStartTime()

	var midP vector.Vector2f

	if math32.Abs(startPos.Sub(endPos).X) < math32.Abs(startPos.Sub(startPos).X) {
		midP = vector.NewVec2f(endPos.X, startPos.Y)
	} else {
		midP = vector.NewVec2f(startPos.X, endPos.Y)
	}

	mover.bz = curves.NewMultiCurve("L", []vector.Vector2f{endPos, midP, startPos})
	mover.endTime = endTime
	mover.startTime = startTime

	return 2
}

func (mover AxisMover) Update(time float64) vector.Vector2f {
	t := float32(time-mover.endTime) / float32(mover.startTime-mover.endTime)
	tr := bmath.ClampF32(math32.Sin(t*math32.Pi/2), 0, 1)
	return mover.bz.PointAt(tr)
}
