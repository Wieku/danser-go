package dance

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/dance/schedulers"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"sort"
	"strings"
)

type Controller interface {
	SetBeatMap(beatMap *beatmap.BeatMap)
	InitCursors()
	Update(time float64, delta float64)
	GetCursors() []*graphics.Cursor
}

type GenericController struct {
	bMap       *beatmap.BeatMap
	cursors    []*graphics.Cursor
	schedulers []schedulers.Scheduler
}

func NewGenericController() Controller {
	return &GenericController{}
}

func (controller *GenericController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *GenericController) InitCursors() {
	controller.cursors = make([]*graphics.Cursor, settings.TAG)
	controller.schedulers = make([]schedulers.Scheduler, settings.TAG)

	counter := make(map[string]int)

	// Mover initialization
	for i := range controller.cursors {
		controller.cursors[i] = graphics.NewCursor()

		mover := "flower"
		if len(settings.CursorDance.Movers) > 0 {
			mover = strings.ToLower(settings.CursorDance.Movers[i%len(settings.CursorDance.Movers)].Mover)
		}

		moverCtor, mName := movers.GetMoverCtorByName(mover)

		controller.schedulers[i] = schedulers.NewGenericScheduler(moverCtor, i, counter[mName])

		counter[mName]++
	}

	type Queue struct {
		hitObjects []objects.IHitObject
	}

	queues := make([]Queue, settings.TAG)

	queue := controller.bMap.GetObjectsCopy()

	// Convert retarded (0 length / 0ms) sliders to pseudo-circles
	for i := 0; i < len(queue); i++ {
		if s, ok := queue[i].(*objects.Slider); ok && s.IsRetarded() {
			queue = schedulers.PreprocessQueue(i, queue, true)
		}
	}

	// Convert sliders to pseudo-circles for tag cursors
	if !settings.CursorDance.ComboTag && !settings.CursorDance.Battle &&
		settings.CursorDance.TAGSliderDance && settings.TAG > 1 {
		for i := 0; i < len(queue); i++ {
			queue = schedulers.PreprocessQueue(i, queue, true)
		}
	}

	// Resolving 2B conflicts
	for i := 0; i < len(queue); i++ {
		if s, ok := queue[i].(*objects.Slider); ok {
			found := false

			// We need to loop backwards to look for overlapping spinners (p) that are separated by circles:
			// --ppppppppppppppppp------
			// ----------c--c-----------
			// ---------------ssssssss--
			// Looking just by i-1 (like i+1 in forward detection) wouldn't detect that scenario because objects
			// are not sorted by end times
			for j := i - 1; j >= 0; j-- {
				if o := queue[i-1]; o.GetEndTime() >= s.GetStartTime() {
					queue = schedulers.PreprocessQueue(i, queue, true)
					found = true
					break
				}
			}

			// If no conflict was detected in the past then look one object ahead, no looping is needed in this scenario
			if !found && i+1 < len(queue) {
				if o := queue[i+1]; o.GetStartTime() <= s.GetEndTime() {
					queue = schedulers.PreprocessQueue(i, queue, true)
				}
			}
		}
	}

	// Second 2B pass for spinners
	for i := 0; i < len(queue); i++ {
		if s, ok := queue[i].(*objects.Spinner); ok {
			var subSpinners []objects.IHitObject

			startTime := s.GetStartTime()

			for j := i + 1; j < len(queue); j++ {
				o := queue[j]

				if o.GetStartTime() >= s.GetEndTime() {
					break
				}

				if endTime := o.GetStartTime() - 30; endTime > startTime {
					subSpinners = append(subSpinners, objects.NewDummySpinner(startTime, endTime))
				}

				startTime = o.GetEndTime() + 30
			}

			if subSpinners != nil && len(subSpinners) > 0 {
				if s.GetEndTime() > startTime {
					subSpinners = append(subSpinners, objects.NewDummySpinner(startTime, s.GetEndTime()))
				}

				queue = append(queue[:i], append(subSpinners, queue[i+1:]...)...)
				sort.SliceStable(queue, func(i, j int) bool { return queue[i].GetStartTime() < queue[j].GetStartTime() })
			}
		}
	}

	// If DoSpinnersTogether is true with tag mode, allow all tag cursors to spin the same spinner with different movers
	for j, o := range queue {
		_, isSpinner := o.(*objects.Spinner)

		if (isSpinner && settings.CursorDance.DoSpinnersTogether) || settings.CursorDance.Battle {
			for i := range queues {
				queues[i].hitObjects = append(queues[i].hitObjects, o)
			}
		} else if settings.CursorDance.ComboTag {
			i := int(o.GetComboSet()) % settings.TAG
			queues[i].hitObjects = append(queues[i].hitObjects, o)
		} else {
			i := j % settings.TAG
			queues[i].hitObjects = append(queues[i].hitObjects, o)
		}
	}

	//Initialize spinner movers
	for i := range controller.cursors {
		spinMover := "circle"
		if len(settings.CursorDance.Spinners) > 0 {
			spinMover = settings.CursorDance.Spinners[i%len(settings.CursorDance.Spinners)].Mover
		}

		controller.schedulers[i].Init(queues[i].hitObjects, controller.bMap.Diff, controller.cursors[i], spinners.GetMoverCtorByName(spinMover), true)
	}
}

func (controller *GenericController) Update(time float64, delta float64) {
	for i := range controller.cursors {
		controller.schedulers[i].Update(time)
		controller.cursors[i].Update(delta)

		controller.cursors[i].LeftButton = controller.cursors[i].LeftKey || controller.cursors[i].LeftMouse
		controller.cursors[i].RightButton = controller.cursors[i].RightKey || controller.cursors[i].RightMouse
	}
}

func (controller *GenericController) GetCursors() []*graphics.Cursor {
	return controller.cursors
}
