package schedulers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/dance/movers"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/settings"
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

	if settings.Dance.SliderDance2B {
		for i := 0; i < len(sched.queue); i++ {
			if s, ok := sched.queue[i].(*objects.Slider); ok {
				sd := s.GetBasicData()
				for j := i+1; j < len(sched.queue); j++ {
					od := sched.queue[j].GetBasicData()
					if (od.StartTime > sd.StartTime && od.StartTime < sd.EndTime) || (od.EndTime > sd.StartTime && od.EndTime < sd.EndTime) {
						sched.queue = PreprocessQueue(i, sched.queue, true)
						break
					}
				}
			}
		}
	}

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
					//if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointStart {
						if !sched.lastLeft && g.GetBasicData().StartTime-sched.lastEnd < 130 {
							sched.cursor.LeftButton = true
							sched.lastLeft = true
						} else {
							sched.cursor.RightButton = true
							sched.lastLeft = false
						}
					//}

				}
				sched.moving = true

			} else if time > g.GetBasicData().EndTime {

				sched.moving = false
				//if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointEnd || g.GetBasicData().SliderPointStart {
					sched.cursor.LeftButton = false
					sched.cursor.RightButton = false
				//}

				if i < len(sched.queue)-1 {
					sched.queue = append(sched.queue[:i], sched.queue[i+1:]...)
				} else if i < len(sched.queue) {
					sched.queue = sched.queue[:i]
				}
				i--

				if i+1 < len(sched.queue) {
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
