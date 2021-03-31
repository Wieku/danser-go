package input

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/graphics"
)

type NaturalInputProcessor struct {
	queue  []objects.IHitObject
	cursor *graphics.Cursor

	lastLeft       bool
	moving         bool
	lastEnd        float64
	lastTime       float64
	lastLeftClick  float64
	lastRightClick float64
	leftToRelease  bool
	rightToRelease bool
}

func NewNaturalInputProcessor(objs []objects.IHitObject, cursor *graphics.Cursor) *NaturalInputProcessor {
	processor := new(NaturalInputProcessor)
	processor.cursor = cursor
	processor.queue = make([]objects.IHitObject, len(objs))

	copy(processor.queue, objs)

	return processor
}

func (processor *NaturalInputProcessor) Update(time float64) {
	if len(processor.queue) > 0 {
		for i := 0; i < len(processor.queue); i++ {
			g := processor.queue[i]
			if g.GetStartTime() > time {
				break
			}

			if (processor.lastTime <= g.GetStartTime() && time >= g.GetStartTime()) || (time >= g.GetStartTime() && time <= g.GetEndTime()) {
				if !processor.moving {
					//if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointStart {
						if !processor.lastLeft && g.GetStartTime()-processor.lastEnd < 140 {
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
					//}
				}

				processor.moving = true
			} else if time > g.GetStartTime() && time > g.GetEndTime() {

				processor.moving = false
				//if !g.GetBasicData().SliderPoint || g.GetBasicData().SliderPointEnd {
					processor.leftToRelease = true
					processor.rightToRelease = true
				//}

				processor.lastEnd = g.GetEndTime()

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
