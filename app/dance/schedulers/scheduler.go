package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/graphics"
)

type Scheduler interface {
	Init(objects []objects.BaseObject, cursor *graphics.Cursor)
	Update(time int64)
}
