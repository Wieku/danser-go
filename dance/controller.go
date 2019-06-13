package dance

import (
	"github.com/wieku/danser-go/beatmap"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/dance/schedulers"
	"github.com/wieku/danser-go/dance/movers"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/beatmap/objects"
	"strings"
)

type Controller interface {
	SetBeatMap(beatMap *beatmap.BeatMap)
	InitCursors()
	Update(time int64, delta float64)
	GetCursors() []*render.Cursor
}

var Mover = movers.NewAngleOffsetMover

func SetMover(name string) {
	name = strings.ToLower(name)

	if name == "bezier" {
		Mover = movers.NewBezierMover
	} else if name == "circular" {
		Mover = movers.NewHalfCircleMover
	} else if name == "linear" {
		Mover = movers.NewLinearMover
	} else if name == "axis" {
		Mover = movers.NewAxisMover
	} else if name == "aggressive" {
		Mover = movers.NewAggressiveMover
	} else {
		Mover = movers.NewAngleOffsetMover
	}
}

type GenericController struct {
	bMap       *beatmap.BeatMap
	cursors    []*render.Cursor
	schedulers []schedulers.Scheduler
}

func NewGenericController() Controller {
	return &GenericController{}
}

func (controller *GenericController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *GenericController) InitCursors() {
	controller.cursors = make([]*render.Cursor, settings.TAG)
	controller.schedulers = make([]schedulers.Scheduler, settings.TAG)

	for i := range controller.cursors {
		controller.cursors[i] = render.NewCursor()
		controller.schedulers[i] = schedulers.NewGenericScheduler(Mover)
	}

	type Queue struct {
		objs []objects.BaseObject
	}

	objs := make([]Queue, settings.TAG)

	queue := controller.bMap.GetObjectsCopy()

	if settings.Dance.TAGSliderDance && settings.TAG > 1 {
		for i := 0; i < len(queue); i++ {
			queue = schedulers.PreprocessQueue(i, queue, true)
		}
	}

	for j, o := range queue {
		i := j % settings.TAG
		objs[i].objs = append(objs[i].objs, o)
	}

	for i := range controller.cursors {
		controller.schedulers[i].Init(objs[i].objs, controller.cursors[i])
	}

}

func (controller *GenericController) Update(time int64, delta float64) {
	for i := range controller.cursors {
		controller.schedulers[i].Update(time)
		controller.cursors[i].Update(delta)
	}
}

func (controller *GenericController) GetCursors() []*render.Cursor {
	return controller.cursors
}
