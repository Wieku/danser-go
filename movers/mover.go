package movers

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/render"
)

type Mover interface {
	Reset()
	SetObjects(end, start objects.BaseObject)
	Update(time int64, cursor *render.Cursor)
}