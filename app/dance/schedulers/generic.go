package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
	"math/rand"
)

type GenericScheduler struct {
	cursor       *graphics.Cursor
	queue        []objects.BaseObject
	mover        movers.MultiPointMover
	lastTime     int64
	spinnerMover spinners.SpinnerMover
	input        *InputProcessor
}

func NewGenericScheduler(mover func() movers.MultiPointMover) Scheduler {
	return &GenericScheduler{mover: mover()}
}

func (sched *GenericScheduler) Init(objs []objects.BaseObject, cursor *graphics.Cursor, spinnerMover spinners.SpinnerMover) {
	sched.spinnerMover = spinnerMover
	sched.cursor = cursor
	sched.queue = objs

	sched.input = NewInputProcessor(objs, cursor)

	sched.mover.Reset()

	for i := 0; i < len(sched.queue); i++ {
		sched.queue = PreprocessQueue(i, sched.queue, (settings.Dance.SliderDance && !settings.Dance.RandomSliderDance) || (settings.Dance.RandomSliderDance && rand.Intn(2) == 0))
	}

	sched.queue = append([]objects.BaseObject{objects.DummyCircle(vector.NewVec2f(100, 100), 0)}, sched.queue...)

	toRemove := sched.mover.SetObjects(sched.queue) - 1
	sched.queue = sched.queue[toRemove:]
}

func (sched *GenericScheduler) Update(time int64) {
	if len(sched.queue) > 0 {
		move := true

		for i := 0; i < len(sched.queue); i++ {
			g := sched.queue[i]

			if g.GetBasicData().StartTime > time {
				break
			}

			move = false

			if time >= g.GetBasicData().StartTime && time <= g.GetBasicData().EndTime {
				if _, ok := g.(*objects.Spinner); ok {
					if sched.lastTime < g.GetBasicData().StartTime {
						sched.spinnerMover.Init(g.GetBasicData().StartTime, g.GetBasicData().EndTime)
					}

					sched.cursor.SetPos(sched.spinnerMover.GetPositionAt(time))
				} else {
					sched.cursor.SetPos(g.GetPosition())
				}
			} else if time > g.GetBasicData().EndTime {
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

	sched.input.Update(time)

	sched.lastTime = time
}
