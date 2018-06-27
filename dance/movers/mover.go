package movers

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/bmath"
)

type MultiPointMover interface {
	Reset()
	SetObjects(objects []objects.BaseObject) (int, int64)
	Update(time int64) bmath.Vector2d
}