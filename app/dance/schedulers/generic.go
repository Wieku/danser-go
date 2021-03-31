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
	"math"
	"math/rand"
)

type GenericScheduler struct {
	cursor   *graphics.Cursor
	queue    []objects.IHitObject
	mover    movers.MultiPointMover
	lastTime float64
	input    *input.NaturalInputProcessor
	mods     difficulty.Modifier
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

	for i := 0; i < len(scheduler.queue); i++ {
		if _, ok := scheduler.queue[i].(*objects.Circle); !ok {
			continue
		}

		remove := false

		if i > 0 {
			p := scheduler.queue[i-1]
			c := scheduler.queue[i]

			if p.GetStackedEndPositionMod(mods).Dst(c.GetStackedStartPositionMod(mods)) <= 3 && c.GetStartTime()-p.GetEndTime() <= 3 {
				remove = true
			}
		}

		if i < len(scheduler.queue)-1 {
			p := scheduler.queue[i]
			c := scheduler.queue[i+1]

			if p.GetStackedEndPositionMod(mods).Dst(c.GetStackedStartPositionMod(mods)) <= 3 && c.GetStartTime()-p.GetEndTime() <= 3 {
				remove = true
			}
		}

		if remove {
			scheduler.queue = append(scheduler.queue[:i], scheduler.queue[i+1:]...)
			//we don't do "i--" here because we don't want to remove too much
		}
	}

	scheduler.queue = append([]objects.IHitObject{objects.DummyCircle(vector.NewVec2f(100, 100), 0)}, scheduler.queue...)

	toRemove := scheduler.mover.SetObjects(scheduler.queue) - 1
	scheduler.queue = scheduler.queue[toRemove:]
}

func (scheduler *GenericScheduler) Update(time float64) {
	if len(scheduler.queue) > 0 {
		useMover := true
		lastEndTime := 0.0

		for i := 0; i < len(scheduler.queue); i++ {
			g := scheduler.queue[i]

			if g.GetStartTime() > time {
				break
			}

			lastEndTime = math.Max(lastEndTime, g.GetEndTime())

			if time <= g.GetEndTime() {
				if scheduler.lastTime <= g.GetStartTime() { // brief movement lock for ExGon mover
					useMover = false
				}

				scheduler.cursor.SetPos(g.GetStackedPositionAtMod(time, scheduler.mods))
			} else {
				upperLimit := len(scheduler.queue)

				for j := i; j < len(scheduler.queue); j++ {
					if scheduler.queue[j].GetEndTime() >= lastEndTime {
						break
					}

					upperLimit = j + 1
				}

				toRemove := 1

				if upperLimit-i > 1 {
					toRemove = scheduler.mover.SetObjects(scheduler.queue[i:upperLimit]) - 1
				}

				scheduler.queue = append(scheduler.queue[:i], scheduler.queue[i+toRemove:]...)
				i--
			}
		}

		if useMover && scheduler.mover.GetEndTime() >= time {
			scheduler.cursor.SetPos(scheduler.mover.Update(time))
		}
	}

	if scheduler.input != nil {
		scheduler.input.Update(time)
	}

	scheduler.lastTime = time
}
