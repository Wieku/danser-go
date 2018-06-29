package schedulers

import (
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/dance/movers"
	"github.com/wieku/danser/settings"
	"github.com/wieku/danser/bmath"
)

type FlowerScheduler struct {
	cursor *render.Cursor
	queue []objects.BaseObject
	mover movers.MultiPointMover
}

func NewFlowerScheduler() Scheduler {
	return &FlowerScheduler{mover: movers.NewAngleOffsetMover()}
}

func (sched *FlowerScheduler) Init(objs []objects.BaseObject, cursor *render.Cursor) {
	sched.cursor = cursor
	sched.queue = objs
	sched.mover.Reset()
	sched.queue = preprocessQueue(0, sched.queue, settings.Dance.SliderDance)
	sched.mover.SetObjects([]objects.BaseObject{objects.DummyCircle(bmath.NewVec2d(100, 100), 0), sched.queue[0]})
}

func (sched *FlowerScheduler) Update(time int64) {
	if len(sched.queue) > 0 {
		move := true
		for i:=0; i < len(sched.queue); i++ {
			g := sched.queue[i]
			if g.GetBasicData().StartTime > time {
				break
			}

			move = false

			if time >= g.GetBasicData().StartTime && time <= g.GetBasicData().EndTime {
				sched.cursor.SetPos(g.GetPosition())
			} else if time > g.GetBasicData().EndTime {
				if i < len(sched.queue) -1 {
					sched.queue = append(sched.queue[:i], sched.queue[i+1:]...)
				} else if i < len(sched.queue) {
					sched.queue = sched.queue[:i]
				}
				i--

				if len(sched.queue) > 0 {
					sched.queue = preprocessQueue(i+1, sched.queue, settings.Dance.SliderDance)
					sched.mover.SetObjects([]objects.BaseObject{g, sched.queue[i+1]})
				}

				move = true
			}
		}

		if move && sched.mover.GetEndTime() >= time {
			sched.cursor.SetPos(sched.mover.Update(time))
		}

	}
}