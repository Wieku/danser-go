package movers

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/bmath"
)

type MultiPointMover interface {
	Reset()
	SetObjects(objs []objects.BaseObject)
	Update(time int64) bmath.Vector2d
	GetEndTime() int64
}
