package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
)

type Scheduler interface {
	Init(objects []objects.IHitObject, mods difficulty.Modifier, cursor *graphics.Cursor, spinnerMover spinners.SpinnerMover, initKeys bool)
	Update(time float64)
}
