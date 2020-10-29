package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/graphics"
)

type InputProcessor struct {
	queue  []objects.BaseObject
	cursor *graphics.Cursor

	lastLeft       bool
	moving         bool
	lastEnd        int64
	lastTime       int64
	lastLeftClick  int64
	lastRightClick int64
	leftToRelease  bool
	rightToRelease bool
}

func NewInputProcessor(objs []objects.BaseObject, cursor *graphics.Cursor) *InputProcessor {
	processor := new(InputProcessor)
	processor.cursor = cursor
	processor.queue = make([]objects.BaseObject, len(objs))

	copy(processor.queue, objs)

	return processor
}

func (processor *InputProcessor) Update(time int64) {
	if len(processor.queue) > 0 {
		for i := 0; i < len(processor.queue); i++ {
			g := processor.queue[i]
			if g.GetBasicData().StartTime > time {
				break
			}

			if time >= g.GetBasicData().StartTime && time <= g.GetBasicData().EndTime {
				if !processor.moving {
					if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointStart {
						if !processor.lastLeft && g.GetBasicData().StartTime-processor.lastEnd < 140 {
							processor.cursor.LeftKey = true
							processor.lastLeft = true
							processor.leftToRelease = false
							processor.lastLeftClick = time
						} else {
							processor.cursor.RightKey = true
							processor.lastLeft = false
							processor.rightToRelease = false
							processor.lastRightClick = time
						}
					}
				}

				processor.moving = true
			} else if time > g.GetBasicData().StartTime && time > g.GetBasicData().EndTime {

				processor.moving = false
				if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointEnd {
					processor.leftToRelease = true
					processor.rightToRelease = true
				}

				processor.lastEnd = g.GetBasicData().EndTime

				processor.queue = append(processor.queue[:i], processor.queue[i+1:]...)

				i--
			}
		}
	}

	if processor.leftToRelease && time-processor.lastLeftClick > 50 {
		processor.leftToRelease = false
		processor.cursor.LeftKey = false
	}

	if processor.rightToRelease && time-processor.lastRightClick > 50 {
		processor.rightToRelease = false
		processor.cursor.RightKey = false
	}

	processor.lastTime = time
}
