package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type LinearMover struct {
	line               curves.Linear
	startTime, endTime float64
	diff               *difficulty.Difficulty
	simple             bool
	id                 int
}

func NewLinearMover() MultiPointMover {
	return &LinearMover{}
}

func NewLinearMoverSimple() MultiPointMover {
	return &LinearMover{simple: true}
}

func (bm *LinearMover) Reset(diff *difficulty.Difficulty, id int) {
	bm.diff = diff
	bm.id = id
}

func (bm *LinearMover) SetObjects(objs []objects.IHitObject) int {
	end, start := objs[0], objs[1]
	endPos := end.GetStackedEndPositionMod(bm.diff.Mods)
	endTime := end.GetEndTime()
	startPos := start.GetStackedStartPositionMod(bm.diff.Mods)
	startTime := start.GetStartTime()

	bm.line = curves.NewLinear(endPos, startPos)

	bm.endTime = endTime
	bm.startTime = startTime

	if bm.simple {
		bm.endTime = math.Max(endTime, start.GetStartTime()-(bm.diff.Preempt-100*bm.diff.Speed))
	} else {
		config := settings.CursorDance.MoverSettings.Linear[bm.id%len(settings.CursorDance.MoverSettings.Linear)]

		if config.WaitForPreempt {
			bm.endTime = math.Max(endTime, start.GetStartTime()-(bm.diff.Preempt-config.ReactionTime*bm.diff.Speed))
		}
	}

	return 2
}

func (bm LinearMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF64((time-bm.endTime)/(bm.startTime-bm.endTime), 0, 1)
	return bm.line.PointAt(float32(easing.OutQuad(t)))
}

func (bm *LinearMover) GetEndTime() float64 {
	return bm.startTime
}
