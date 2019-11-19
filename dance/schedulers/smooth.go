package schedulers

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/settings"
)

type SmoothScheduler struct {
	cursor *render.Cursor
	queue  []objects.BaseObject
	curve 	curves.BSpline
	endTime, startTime int64
	lastLeft bool
	moving bool
	lastEnd int64
}

func NewSmoothScheduler() Scheduler {
	return &SmoothScheduler{}
}

func (sched *SmoothScheduler) Init(objs []objects.BaseObject, cursor *render.Cursor) {
	sched.cursor = cursor
	sched.queue = append([]objects.BaseObject{objects.DummyCircle(bmath.NewVec2d(100, 100), 0)}, objs...)
	/*sched.queue = PreprocessQueue(0, sched.queue, settings.Dance.SliderDance)*/
	for i := 0; i < len(sched.queue); i++ {
		sched.queue = PreprocessQueue(i, sched.queue, settings.Dance.SliderDance)
	}
	sched.InitCurve(0)
}

func (sched *SmoothScheduler) Update(time int64) {
	if len(sched.queue) > 0 {
		move := true
		for i := 0; i < len(sched.queue); i++ {
			g := sched.queue[i]
			if g.GetBasicData().StartTime > time {
				break
			}

			move = false

			if time >= g.GetBasicData().StartTime && time <= g.GetBasicData().EndTime {
				if s, ok := sched.queue[i].(*objects.Slider); ok {
					sched.cursor.SetPos(s.GetPosition())
				}

				if s, ok := sched.queue[i].(*objects.Spinner); ok {
					sched.cursor.SetPos(s.GetPosition())
				}

				if !sched.moving {
					if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointStart {
						if !sched.lastLeft && g.GetBasicData().StartTime-sched.lastEnd < 130 {
							sched.cursor.LeftButton = true
							sched.lastLeft = true
						} else {
							sched.cursor.RightButton = true
							sched.lastLeft = false
						}
					}

				}
				sched.moving = true
			} else if time > g.GetBasicData().EndTime {

				sched.moving = false
				if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointEnd {
					sched.cursor.LeftButton = false
					sched.cursor.RightButton = false
				}
				sched.lastEnd = g.GetBasicData().EndTime

				if len(sched.queue) > 1 {
					if _, ok := sched.queue[i].(*objects.Slider); ok {
						sched.InitCurve(i)
					}

					if _, ok := sched.queue[i].(*objects.Spinner); ok {
						sched.InitCurve(i)
					}
				}

				if i < len(sched.queue)-1 {
					sched.queue = append(sched.queue[:i], sched.queue[i+1:]...)
				} else if i < len(sched.queue) {
					sched.queue = sched.queue[:i]
				}
				i--

				if len(sched.queue) > 0 {
					sched.queue = PreprocessQueue(i+1, sched.queue, settings.Dance.SliderDance)
				}

				move = true
			}
		}

		if move && sched.startTime >= time {
			sched.cursor.SetPos(sched.curve.NPointAt(float64(time-sched.endTime)/float64(sched.startTime-sched.endTime)))
		}

	}
}

func (sched *SmoothScheduler) InitCurve(index int) {
	points := make([]bmath.Vector2d, 0)
	timing := make([]int64, 0)
	var endTime, startTime int64
	for i := index; i < len(sched.queue); i++ {
		if i == index {
			if s, ok := sched.queue[i].(*objects.Slider); ok {
				points = append(points, s.GetBasicData().EndPos, bmath.NewVec2dRad(s.GetEndAngle(), s.GetBasicData().EndPos.Dst(sched.queue[i+1].GetBasicData().StartPos)*0.7).Add(s.GetBasicData().EndPos))
				timing = append(timing, s.GetBasicData().EndTime)
			}
			if s, ok := sched.queue[i].(*objects.Circle); ok {
				points = append(points, s.GetBasicData().EndPos, sched.queue[i+1].GetBasicData().StartPos.Sub(s.GetBasicData().EndPos).Scl(0.333).Add(s.GetBasicData().EndPos))
				timing = append(timing, s.GetBasicData().StartTime)
			}
			if s, ok := sched.queue[i].(*objects.Spinner); ok {
				points = append(points, s.GetBasicData().EndPos, sched.queue[i+1].GetBasicData().StartPos.Sub(s.GetBasicData().EndPos).Scl(0.333).Add(s.GetBasicData().EndPos))
				timing = append(timing, s.GetBasicData().EndTime)
			}
			endTime = sched.queue[i].GetBasicData().EndTime
			continue
		}

		if s, ok := sched.queue[i].(*objects.Circle); ok {
			points = append(points, s.GetBasicData().EndPos)
			timing = append(timing, s.GetBasicData().StartTime)
		}

		_, ok1 := sched.queue[i].(*objects.Slider)
		_, ok2 := sched.queue[i].(*objects.Spinner)

		ok := ok1 || ok2

		if ok || i == len(sched.queue) -1 {
			if s, ok := sched.queue[i].(*objects.Slider); ok {
				timing = append(timing, s.GetBasicData().StartTime)
				points = append(points, bmath.NewVec2dRad(s.GetStartAngle(), s.GetBasicData().StartPos.Dst(sched.queue[i-1].GetBasicData().EndPos)*0.7).Add(s.GetBasicData().StartPos), s.GetBasicData().StartPos)
			}
			if s, ok := sched.queue[i].(*objects.Circle); ok {
				timing = append(timing, s.GetBasicData().StartTime)
				points = append(points, sched.queue[i-1].GetBasicData().EndPos.Sub(s.GetBasicData().StartPos).Scl(0.333).Add(s.GetBasicData().StartPos), s.GetBasicData().StartPos)
			}

			if s, ok := sched.queue[i].(*objects.Spinner); ok {
				timing = append(timing, s.GetBasicData().StartTime)
				points = append(points, sched.queue[i-1].GetBasicData().EndPos.Sub(s.GetBasicData().StartPos).Scl(0.333).Add(s.GetBasicData().StartPos), s.GetBasicData().StartPos)
			}

			startTime = sched.queue[i].GetBasicData().StartTime
			break
		}
	}
	sched.startTime = startTime
	sched.endTime = endTime
	//log.Println(points)
	sched.curve = curves.NewBSpline(points, timing)
}