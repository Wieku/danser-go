package movers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
)

type MultiPointMover interface {
	Reset()
	SetObjects(objs []objects.BaseObject)
	Update(time int64) bmath.Vector2f
	GetEndTime() int64
}
