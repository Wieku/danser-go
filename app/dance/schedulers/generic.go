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
	spinnerMover spinners.SpinnerMover
	input        *input.NaturalInputProcessor
	mods         difficulty.Modifier
}

func NewGenericScheduler(mover func() movers.MultiPointMover) Scheduler {
	return &GenericScheduler{mover: mover()}
}

func (sched *GenericScheduler) Init(objs []objects.IHitObject, mods difficulty.Modifier, cursor *graphics.Cursor, spinnerMover spinners.SpinnerMover, initKeys bool) {
	sched.mods = mods
	sched.spinnerMover = spinnerMover
	sched.cursor = cursor
	sched.queue = objs

	if initKeys {
		sched.input = input.NewNaturalInputProcessor(objs, cursor)
	}

	sched.mover.Reset(mods)

	for i := 0; i < len(sched.queue); i++ {
		sched.queue = PreprocessQueue(i, sched.queue, (settings.Dance.SliderDance && !settings.Dance.RandomSliderDance) || (settings.Dance.RandomSliderDance && rand.Intn(2) == 0))
	}

	sched.queue = append([]objects.IHitObject{objects.DummyCircle(vector.NewVec2f(100, 100), 0)}, sched.queue...)

	toRemove := sched.mover.SetObjects(sched.queue) - 1
	sched.queue = sched.queue[toRemove:]
}

func (sched *GenericScheduler) Update(time float64) {
	if len(sched.queue) > 0 {
		move := true

		for i := 0; i < len(sched.queue); i++ {
			g := sched.queue[i]

			if g.GetStartTime() > time {
				break
			}

			move = false

			if time >= g.GetStartTime() && time <= g.GetEndTime() {
				if _, ok := g.(*objects.Spinner); ok {
					if sched.lastTime < g.GetStartTime() {
						sched.spinnerMover.Init(g.GetStartTime(), g.GetEndTime())
					}

					sched.cursor.SetPos(sched.spinnerMover.GetPositionAt(time))
				} else {
					sched.cursor.SetPos(g.GetStackedPositionAtMod(time, sched.mods))
				}
			} else if time > g.GetEndTime() {
				toRemove := 1
				if i+1 < len(sched.queue) {
					toRemove = sched.mover.SetObjects(sched.queue[i:]) - 1
				}

				sched.queue = append(sched.queue[:i], sched.queue[i+toRemove:]...)
				i--

				move = true
			}
		}

		if move && sched.mover.GetEndTime() >= time {
			sched.cursor.SetPos(sched.mover.Update(time))
		}
	}

	if sched.input != nil {
		sched.input.Update(time)
	}

	sched.lastTime = time
}
