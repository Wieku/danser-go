package movers

import (
	"danser/beatmap/objects"
	"danser/bmath"
)

type MultiPointMover interface {
	Reset()
	SetObjects(objs []objects.BaseObject)
	Update(time int64) bmath.Vector2d
	GetEndTime() int64
}