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
	cursor   *graphics.Cursor
	queue    []objects.IHitObject
	mover    movers.MultiPointMover
	lastTime float64
	input    *input.NaturalInputProcessor
	diff     *difficulty.Difficulty
	index    int
	id       int
}

func NewGenericScheduler(mover func() movers.MultiPointMover, index, id int) Scheduler {
	return &GenericScheduler{mover: mover(), index: index, id: id}
}

func (scheduler *GenericScheduler) Init(objs []objects.IHitObject, diff *difficulty.Difficulty, cursor *graphics.Cursor, spinnerMoverCtor func() spinners.SpinnerMover, initKeys bool) {
	scheduler.diff = diff
	scheduler.cursor = cursor
	scheduler.queue = objs

	scheduler.mover.Reset(diff, scheduler.id)

	config := settings.CursorDance.Movers[scheduler.index%len(settings.CursorDance.Movers)]

	// Slider dance / random slider dance resolving
	for i := 0; i < len(scheduler.queue); i++ {
		scheduler.queue = PreprocessQueue(i, scheduler.queue, (config.SliderDance && !config.RandomSliderDance) || (config.RandomSliderDance && rand.Intn(2) == 0))
	}

	// Convert spinners to pseudo spinners that have beginning and ending angles, simplifies mover codes as well
	for i := 0; i < len(scheduler.queue); i++ {
		if s, ok := scheduler.queue[i].(*objects.Spinner); ok {
			scheduler.queue[i] = spinners.NewSpinner(s, spinnerMoverCtor, scheduler.index)
		}
	}

	// Convert two overlapping circles (slider starts too if slider danced) to one double-tap circle
	for i := 0; i < len(scheduler.queue)-1; i++ {
		current, pOk := scheduler.queue[i].(*objects.Circle)
		next, cOk := scheduler.queue[i+1].(*objects.Circle)

		if pOk && cOk && (!current.SliderPoint || current.SliderPointStart || (current.SliderPointEnd && diff.CheckModActive(difficulty.Lazer))) && (!next.SliderPoint || next.SliderPointStart || (next.SliderPointEnd && diff.CheckModActive(difficulty.Lazer))) {
			dst := current.GetStackedEndPositionMod(diff).Dst(next.GetStackedStartPositionMod(diff))

			if dst <= float32(diff.CircleRadius*1.995) && next.GetStartTime()-current.GetEndTime() <= 3 { // Sacrificing a bit of UR for better looks
				sTime := (next.GetStartTime() + current.GetEndTime()) / 2

				if current.SliderPointEnd && diff.CheckModActive(difficulty.Lazer) { // Prioritize slider end timing
					sTime = current.GetEndTime()
				}

				dC := objects.DummyCircle(current.GetStackedEndPositionMod(diff).Add(next.GetStackedStartPositionMod(diff)).Scl(0.5), sTime)

				if !diff.CheckModActive(difficulty.Lazer) || (!current.SliderPointEnd && !next.SliderPointEnd) { // Don't double-click if any of them is a slider end
					dC.DoubleClick = true
				}

				scheduler.queue[i] = dC

				scheduler.queue = append(scheduler.queue[:i+1], scheduler.queue[i+2:]...)
			}
		}
	}

	// Spread overlapping circles timing-wise
	for i := 0; i < len(scheduler.queue)-1; i++ {
		current := scheduler.queue[i]

		for j := i + 1; j < len(scheduler.queue); j++ {
			o := scheduler.queue[j]

			if current.GetEndTime() < o.GetStartTime() {
				break
			}

			if c, cOk := o.(*objects.Circle); cOk && (!c.SliderPoint || c.SliderPointStart) {
				scheduler.queue[j] = objects.DummyCircle(c.GetStackedStartPositionMod(diff), c.GetStartTime()+1)
			}
		}
	}

	if initKeys {
		scheduler.input = input.NewNaturalInputProcessor(scheduler.queue, cursor, scheduler.mover)
	}

	scheduler.queue = append([]objects.IHitObject{objects.DummyCircle(vector.NewVec2f(100, 100), -500)}, scheduler.queue...)

	scheduler.cursor.SetPos(vector.NewVec2f(100, 100))
	scheduler.cursor.Update(0)

	toRemove := scheduler.mover.SetObjects(scheduler.queue) - 1
	scheduler.queue = scheduler.queue[toRemove:]
}

func (scheduler *GenericScheduler) Update(time float64) {
	if len(scheduler.queue) > 0 {
		useMover := true
		lastEndTime := 0.0

		for i := 0; i < len(scheduler.queue); i++ {
			g := scheduler.queue[i]

			gStartTime := scheduler.mover.GetObjectsStartTime(g)
			gEndTime := scheduler.mover.GetObjectsEndTime(g)

			if gStartTime > time {
				break
			}

			lastEndTime = max(lastEndTime, gEndTime)

			if scheduler.lastTime <= gStartTime || time <= gEndTime {
				if scheduler.lastTime <= gStartTime { // brief movement lock for ExGon mover
					useMover = false
				}

				scheduler.cursor.SetPos(scheduler.mover.GetObjectsPosition(time, g))
			}

			if time > gEndTime {
				upperLimit := len(scheduler.queue)

				for j := i; j < len(scheduler.queue); j++ {
					if scheduler.mover.GetObjectsEndTime(scheduler.queue[j]) >= lastEndTime {
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
