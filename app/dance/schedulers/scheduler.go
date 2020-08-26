package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/render"
)

type Scheduler interface {
	Init(objects []objects.BaseObject, cursor *render.Cursor)
	Update(time int64)
}
