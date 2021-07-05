package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
)

type MultiPointMover interface {
	Reset(diff *difficulty.Difficulty, id int)
	SetObjects(objs []objects.IHitObject) int
	Update(time float64) vector.Vector2f
	GetObjectsPosition(time float64, object objects.IHitObject) vector.Vector2f
	GetEndTime() float64
}

type basicMover struct {
	endTime float64
	id      int
	diff    *difficulty.Difficulty
}

func (mover *basicMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.diff = diff
	mover.id = id
}

func (mover *basicMover) GetObjectsPosition(time float64, object objects.IHitObject) vector.Vector2f {
	return object.GetStackedPositionAtMod(time, mover.diff.Mods)
}

func (mover *basicMover) GetEndTime() float64 {
	return mover.endTime
}
