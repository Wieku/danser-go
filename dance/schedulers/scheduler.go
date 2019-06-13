package schedulers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/render"
)

type Scheduler interface {
	Init(objects []objects.BaseObject, cursor *render.Cursor)
	Update(time int64)
}
