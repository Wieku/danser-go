package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/curves"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math/rand"
)

type SmoothScheduler struct {
	cursor             *graphics.Cursor
	queue              []objects.BaseObject
	curve              *curves.BSpline
	endTime, startTime int64
	lastLeft           bool
	moving             bool
	lastEnd            int64
	lastTime           int64
	spinnerMover       spinners.SpinnerMover
	lastLeftClick      int64
	lastRightClick     int64
	leftToRelease      bool
	rightToRelease     bool
}

func NewSmoothScheduler() Scheduler {
	return &SmoothScheduler{}
}

func (sched *SmoothScheduler) Init(objs []objects.BaseObject, cursor *graphics.Cursor, spinnerMover spinners.SpinnerMover) {
	sched.spinnerMover = spinnerMover
	sched.cursor = cursor

	sched.queue = append([]objects.BaseObject{objects.DummyCircle(vector.NewVec2f(100, 100), 0)}, objs...)
	/*sched.queue = PreprocessQueue(0, sched.queue, settings.Dance.SliderDance)*/
	for i := 0; i < len(sched.queue); i++ {
		sched.queue = PreprocessQueue(i, sched.queue, (settings.Dance.SliderDance && !settings.Dance.RandomSliderDance) || (settings.Dance.RandomSliderDance && rand.Intn(2) == 0))
	}

	if settings.Dance.SliderDance2B {
		for i := 0; i < len(sched.queue); i++ {
			if s, ok := sched.queue[i].(*objects.Slider); ok {
				sd := s.GetBasicData()

				for j := i + 1; j < len(sched.queue); j++ {
					od := sched.queue[j].GetBasicData()
					if (od.StartTime > sd.StartTime && od.StartTime < sd.EndTime) || (od.EndTime > sd.StartTime && od.EndTime < sd.EndTime) {
						sched.queue = PreprocessQueue(i, sched.queue, true)
						break
					}
				}
			}
		}
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
				if _, ok := g.(*objects.Spinner); ok {
					sched.cursor.SetPos(sched.spinnerMover.GetPositionAt(time))
				}

				if s, ok := sched.queue[i].(*objects.Slider); ok {
					sched.cursor.SetPos(s.GetPosition())
				}

				if !sched.moving {
					if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointStart {
						if !sched.lastLeft && g.GetBasicData().StartTime-sched.lastEnd < 130 {
							sched.cursor.LeftButton = true
							sched.lastLeft = true
							sched.leftToRelease = false
							sched.lastLeftClick = time
						} else {
							sched.cursor.RightButton = true
							sched.lastLeft = false
							sched.rightToRelease = false
							sched.lastRightClick = time
						}
					}

				}

				sched.moving = true
			} else if time > g.GetBasicData().StartTime && time > g.GetBasicData().EndTime {

				sched.moving = false
				if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointEnd {
					sched.leftToRelease = true
					sched.rightToRelease = true
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
			t := bmath.ClampF32(float32(time-sched.endTime)/float32(sched.startTime-sched.endTime), 0, 1)
			sched.cursor.SetPos(sched.curve.PointAt(t))
		}
	}

	if sched.leftToRelease && time-sched.lastLeftClick > 50 {
		sched.leftToRelease = false
		sched.cursor.LeftButton = false
	}

	if sched.rightToRelease && time-sched.lastRightClick > 50 {
		sched.rightToRelease = false
		sched.cursor.RightButton = false
	}

	sched.lastTime = time
}

func (sched *SmoothScheduler) InitCurve(index int) {
	points := make([]vector.Vector2f, 0)
	timing := make([]int64, 0)

	var endTime, startTime int64
	var angle float32
	var stream bool

	for i := index; i < len(sched.queue); i++ {
		if i == index {
			if s, ok := sched.queue[i].(*objects.Slider); ok {
				points = append(points, s.GetBasicData().EndPos, vector.NewVec2fRad(s.GetEndAngle(), s.GetBasicData().EndPos.Dst(sched.queue[i+1].GetBasicData().StartPos)*0.7).Add(s.GetBasicData().EndPos))
				//timing = append(timing, s.GetBasicData().EndTime)
			}

			if s, ok := sched.queue[i].(*objects.Circle); ok {
				points = append(points, s.GetBasicData().EndPos, sched.queue[i+1].GetBasicData().StartPos.Sub(s.GetBasicData().EndPos).Scl(0.333).Add(s.GetBasicData().EndPos))
				//timing = append(timing, s.GetBasicData().StartTime)
			}

			if s, ok := sched.queue[i].(*objects.Spinner); ok {
				points = append(points, sched.spinnerMover.GetPositionAt(s.GetBasicData().EndTime), sched.queue[i+1].GetBasicData().StartPos.Sub(s.GetBasicData().EndPos).Scl(0.333).Add(s.GetBasicData().EndPos))
			}

			timing = append(timing, bmath.MaxI64(sched.queue[i].GetBasicData().StartTime, sched.queue[i].GetBasicData().EndTime))
			endTime = bmath.MaxI64(sched.queue[i].GetBasicData().StartTime, sched.queue[i].GetBasicData().EndTime)

			continue
		}

		_, ok1 := sched.queue[i].(*objects.Slider)
		_, ok2 := sched.queue[i].(*objects.Spinner)

		ok := ok1 || ok2

		if ok || i == len(sched.queue)-1 {
			if s, ok := sched.queue[i].(*objects.Slider); ok {
				timing = append(timing, s.GetBasicData().StartTime)
				points = append(points, vector.NewVec2fRad(s.GetStartAngle(), s.GetBasicData().StartPos.Dst(sched.queue[i-1].GetBasicData().EndPos)*0.7).Add(s.GetBasicData().StartPos), s.GetBasicData().StartPos)
			}

			if s, ok := sched.queue[i].(*objects.Circle); ok {
				timing = append(timing, s.GetBasicData().StartTime)
				points = append(points, sched.queue[i-1].GetBasicData().EndPos.Sub(s.GetBasicData().StartPos).Scl(0.333).Add(s.GetBasicData().StartPos), s.GetBasicData().StartPos)
			}

			if s, ok := sched.queue[i].(*objects.Spinner); ok {
				sched.spinnerMover.Init(s.GetBasicData().StartTime, s.GetBasicData().EndTime)
				timing = append(timing, s.GetBasicData().StartTime)
				points = append(points, sched.queue[i-1].GetBasicData().EndPos.Sub(sched.spinnerMover.GetPositionAt(s.GetBasicData().StartTime)).Scl(0.333).Add(s.GetBasicData().StartPos), s.GetBasicData().StartPos)
			}

			startTime = sched.queue[i].GetBasicData().StartTime

			break
		} else if i < len(sched.queue)-1 && i-1 > index {
			if _, ok := sched.queue[i].(*objects.Circle); ok {
				pos1 := sched.queue[i-1].GetBasicData().StartPos
				pos2 := sched.queue[i].GetBasicData().StartPos
				pos3 := sched.queue[i+1].GetBasicData().StartPos

				min := float32(25.0)
				max := float32(4000.0)
				if stream {
					max = 8000
				}

				sq1 := pos1.DstSq(pos2)
				sq2 := pos2.DstSq(pos3)

				if sq1 > max && sq2 > max && settings.Dance.Spline.RotationalForce {
					if stream {
						angle = 0
						stream = false
					} else {
						ang := int(math32.Abs(pos1.Sub(pos2).AngleR() - pos1.Sub(pos3).AngleR()))

						if ang == 0 {
							angle *= -1
						} else {
							angle = float32(ang) * 90 / 180 * math32.Pi
						}
					}
				} else if sq1 >= min && sq2 >= min && sq1 <= max && sq2 <= max && settings.Dance.Spline.StreamWobble {
					if stream {
						angle *= -1
					} else {
						pp1 := points[len(points)-1]

						shoeF := pp1.X*pos2.Y + pos2.X*pos3.Y + pos3.X*pp1.Y
						shoeS := pp1.Y*pos2.X + pos2.Y*pos3.X + pos3.Y*pp1.X

						sig := (shoeF - shoeS) > 0

						angle = math32.Pi / 2
						if !sig {
							angle *= -1
						}

						stream = true
					}
				} else {
					stream = false
					angle = 0
				}

				if math32.Abs(angle) > 0.01 {
					mid := pos1.Mid(pos2)

					scale := float32(1.0)
					if stream {
						scale = float32(settings.Dance.Spline.WobbleScale)
					}

					p4 := mid.Sub(pos1).Scl(scale).Rotate(angle).Add(mid)

					points = append(points, p4)
					timing = append(timing, (sched.queue[i].GetBasicData().StartTime-sched.queue[i-1].GetBasicData().StartTime)/2+sched.queue[i-1].GetBasicData().StartTime)
				}
			}
		}

		if s, ok := sched.queue[i].(*objects.Circle); ok {
			points = append(points, s.GetBasicData().EndPos)
			timing = append(timing, s.GetBasicData().StartTime)
		}
	}

	sched.startTime = startTime
	sched.endTime = endTime
	sched.curve = curves.NewBSpline(points, timing)
}
