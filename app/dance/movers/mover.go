package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
)

type MultiPointMover interface {
	Reset()
	SetObjects(objs []objects.IHitObject) int
	Update(time int64) vector.Vector2f
	GetEndTime() int64
}
