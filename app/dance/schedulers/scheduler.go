package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
)

type Scheduler interface {
	Init(objects []objects.BaseObject, cursor *graphics.Cursor, spinnerMover spinners.SpinnerMover)
	Update(time int64)
}
