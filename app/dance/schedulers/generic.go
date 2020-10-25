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
	lastLeft     bool
	moving       bool
	lastEnd      int64
	lastTime     int64
	spinnerMover spinners.SpinnerMover
}

func NewGenericScheduler(mover func() movers.MultiPointMover) Scheduler {
	return &GenericScheduler{mover: mover()}
}

func (sched *GenericScheduler) Init(objs []objects.BaseObject, cursor *graphics.Cursor, spinnerMover spinners.SpinnerMover) {
	sched.spinnerMover = spinnerMover
	sched.cursor = cursor
	sched.queue = objs
	sched.mover.Reset()
	sched.queue = PreprocessQueue(0, sched.queue, settings.Dance.SliderDance)

	sched.mover.SetObjects([]objects.BaseObject{objects.DummyCircle(vector.NewVec2f(100, 100), 0), sched.queue[0]})
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
					sched.queue = PreprocessQueue(i+1, sched.queue, (settings.Dance.SliderDance && !settings.Dance.RandomSliderDance) || (settings.Dance.RandomSliderDance && rand.Intn(2) == 0))
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

	sched.lastTime = time
}
