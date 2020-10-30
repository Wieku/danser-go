package dance

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/dance/schedulers"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"strings"
)

type Controller interface {
	SetBeatMap(beatMap *beatmap.BeatMap)
	InitCursors()
	Update(time int64, delta float64)
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

	for i := range controller.cursors {
		controller.cursors[i] = graphics.NewCursor()

		mover := "flower"
		if len(settings.Dance.Movers) > 0 {
			mover = strings.ToLower(settings.Dance.Movers[i%len(settings.Dance.Movers)])
		}

		var scheduler schedulers.Scheduler

		switch mover {
		case "spline":
			scheduler = schedulers.NewGenericScheduler(movers.NewSplineMover)
		case "bezier":
			scheduler = schedulers.NewGenericScheduler(movers.NewBezierMover)
		case "circular":
			scheduler = schedulers.NewGenericScheduler(movers.NewHalfCircleMover)
		case "linear":
			scheduler = schedulers.NewGenericScheduler(movers.NewLinearMover)
		case "axis":
			scheduler = schedulers.NewGenericScheduler(movers.NewAxisMover)
		case "aggressive":
			scheduler = schedulers.NewGenericScheduler(movers.NewAggressiveMover)
		case "momentum":
			scheduler = schedulers.NewGenericScheduler(movers.NewMomentumMover)
		default:
			scheduler = schedulers.NewGenericScheduler(movers.NewAngleOffsetMover)
		}

		controller.schedulers[i] = scheduler
	}

	type Queue struct {
		objs []objects.BaseObject
	}

	objs := make([]Queue, settings.TAG)

	queue := controller.bMap.GetObjectsCopy()

	if !settings.Dance.Battle && settings.Dance.TAGSliderDance && settings.TAG > 1 {
		for i := 0; i < len(queue); i++ {
			queue = schedulers.PreprocessQueue(i, queue, true)
		}
	}

	if settings.Dance.SliderDance2B {
		for i := 0; i < len(queue); i++ {
			if s, ok := queue[i].(*objects.Slider); ok {
				sd := s.GetBasicData()

				for j := i + 1; j < len(queue); j++ {
					od := queue[j].GetBasicData()
					if (od.StartTime > sd.StartTime && od.StartTime < sd.EndTime) || (od.EndTime > sd.StartTime && od.EndTime < sd.EndTime) {
						queue = schedulers.PreprocessQueue(i, queue, true)
						break
					}
				}
			}
		}
	}

	for j, o := range queue {
		if _, ok := o.(*objects.Spinner); (ok && settings.Dance.DoSpinnersTogether) || settings.Dance.Battle {
			for i := range objs {
				objs[i].objs = append(objs[i].objs, o)
			}
		} else {
			i := j % settings.TAG
			objs[i].objs = append(objs[i].objs, o)
		}
	}

	for i := range controller.cursors {
		spinMover := "circle"
		if len(settings.Dance.Spinners) > 0 {
			spinMover = settings.Dance.Spinners[i%len(settings.Dance.Spinners)]
		}

		controller.schedulers[i].Init(objs[i].objs, controller.cursors[i], spinners.GetMoverByName(spinMover))
	}
}

func (controller *GenericController) Update(time int64, delta float64) {
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
