package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/input"
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
	"math/rand"
)

type GenericScheduler struct {
	cursor       *graphics.Cursor
	queue        []objects.IHitObject
	mover        movers.MultiPointMover
	lastTime     float64
	input        *input.NaturalInputProcessor
	mods         difficulty.Modifier
}

func NewGenericScheduler(mover func() movers.MultiPointMover) Scheduler {
	return &GenericScheduler{mover: mover()}
}

func (scheduler *GenericScheduler) Init(objs []objects.IHitObject, mods difficulty.Modifier, cursor *graphics.Cursor, spinnerMoverCtor func() spinners.SpinnerMover, initKeys bool) {
	scheduler.mods = mods
	scheduler.cursor = cursor
	scheduler.queue = objs

	if initKeys {
		scheduler.input = input.NewNaturalInputProcessor(objs, cursor)
	}

	scheduler.mover.Reset(mods)

	for i := 0; i < len(scheduler.queue); i++ {
		scheduler.queue = PreprocessQueue(i, scheduler.queue, (settings.Dance.SliderDance && !settings.Dance.RandomSliderDance) || (settings.Dance.RandomSliderDance && rand.Intn(2) == 0))
	}

	for i := 0; i < len(scheduler.queue); i++ {
		if s, ok := scheduler.queue[i].(*objects.Spinner); ok {
			scheduler.queue[i] = spinners.NewSpinner(s, spinnerMoverCtor)
		}
	}

	scheduler.queue = append([]objects.IHitObject{objects.DummyCircle(vector.NewVec2f(100, 100), 0)}, scheduler.queue...)

	toRemove := scheduler.mover.SetObjects(scheduler.queue) - 1
	scheduler.queue = scheduler.queue[toRemove:]
}

func (scheduler *GenericScheduler) Update(time float64) {
	if len(scheduler.queue) > 0 {
		move := true

		for i := 0; i < len(scheduler.queue); i++ {
			g := scheduler.queue[i]

			if g.GetStartTime() > time {
				break
			}

			move = false

			if (scheduler.lastTime <= g.GetStartTime() && time >= g.GetStartTime()) || (time >= g.GetStartTime() && time <= g.GetEndTime()) {
				scheduler.cursor.SetPos(g.GetStackedPositionAtMod(time, scheduler.mods))
			} else if time > g.GetEndTime() {
				toRemove := 1
				if i+1 < len(scheduler.queue) {
					toRemove = scheduler.mover.SetObjects(scheduler.queue[i:]) - 1
				}

				scheduler.queue = append(scheduler.queue[:i], scheduler.queue[i+toRemove:]...)
				i--

				move = true
			}
		}

		if move && scheduler.mover.GetEndTime() >= time {
			scheduler.cursor.SetPos(scheduler.mover.Update(time))
		}
	}

	if scheduler.input != nil {
		scheduler.input.Update(time)
	}

	scheduler.lastTime = time
}
