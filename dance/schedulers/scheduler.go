package schedulers

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/render"
)

type Scheduler interface {
	Init(objects []objects.BaseObject, cursor *render.Cursor)
	Update(time int64)
}
