package schedulers

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/dance/movers"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/settings"
)

type GenericScheduler struct {
	cursor *render.Cursor
	queue  []objects.BaseObject
	mover  movers.MultiPointMover
	lastLeft bool
	moving bool
	lastEnd int64
}

func NewGenericScheduler(mover func() movers.MultiPointMover) Scheduler {
	return &GenericScheduler{mover: mover()}
}

func (sched *GenericScheduler) Init(objs []objects.BaseObject, cursor *render.Cursor) {
	sched.cursor = cursor
	sched.queue = objs
	sched.mover.Reset()
	sched.queue = PreprocessQueue(0, sched.queue, settings.Dance.SliderDance)
	sched.mover.SetObjects([]objects.BaseObject{objects.DummyCircle(bmath.NewVec2d(100, 100), 0), sched.queue[0]})
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
				sched.cursor.SetPos(g.GetPosition())

				if !sched.moving {
					if !sched.lastLeft && g.GetBasicData().StartTime-sched.lastEnd < 130 {
						sched.cursor.LeftButton = true
						sched.lastLeft = true
					} else {
						sched.cursor.RightButton = true
						sched.lastLeft = false
					}
				}
				sched.moving = true

			} else if time > g.GetBasicData().EndTime {

				sched.moving = false
				sched.cursor.LeftButton = false
				sched.cursor.RightButton = false

				if i < len(sched.queue)-1 {
					sched.queue = append(sched.queue[:i], sched.queue[i+1:]...)
				} else if i < len(sched.queue) {
					sched.queue = sched.queue[:i]
				}
				i--

				if len(sched.queue) > 0 {
					sched.queue = PreprocessQueue(i+1, sched.queue, settings.Dance.SliderDance)
					sched.mover.SetObjects([]objects.BaseObject{g, sched.queue[i+1]})
				}
				sched.lastEnd = g.GetBasicData().EndTime
				move = true
			}
		}

		if move && sched.mover.GetEndTime() >= time {
			sched.cursor.SetPos(sched.mover.Update(time))
		}

	}
}
