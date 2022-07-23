package input

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const singleTapThreshold = 140

type NaturalInputProcessor struct {
	queue  []objects.IHitObject
	cursor *graphics.Cursor

	lastTime float64

	wasLeftBefore  bool
	previousEnd    float64
	releaseLeftAt  float64
	releaseRightAt float64
	mover          movers.MultiPointMover
}

func NewNaturalInputProcessor(objs []objects.IHitObject, cursor *graphics.Cursor, mover movers.MultiPointMover) *NaturalInputProcessor {
	processor := new(NaturalInputProcessor)
	processor.mover = mover
	processor.cursor = cursor
	processor.queue = make([]objects.IHitObject, len(objs))
	processor.releaseLeftAt = -10000000
	processor.releaseRightAt = -10000000

	copy(processor.queue, objs)

	return processor
}

func (processor *NaturalInputProcessor) Update(time float64) {
	if len(processor.queue) > 0 {
		for i := 0; i < len(processor.queue); i++ {
			g := processor.queue[i]

			gStartTime := processor.mover.GetObjectsStartTime(g)
			gEndTime := processor.mover.GetObjectsEndTime(g)

			if gStartTime > time {
				break
			}

			if processor.lastTime < gStartTime && time >= gStartTime {
				startTime := gStartTime
				endTime := gEndTime

				releaseAt := endTime + 50.0

				if i+1 < len(processor.queue) {
					nTime := processor.mover.GetObjectsStartTime(processor.queue[mutils.Min(i+2, len(processor.queue)-1)])

					releaseAt = mutils.ClampF(nTime-2, endTime+1, releaseAt)
				}

				shouldBeLeft := !processor.wasLeftBefore && startTime-processor.previousEnd < singleTapThreshold

				if shouldBeLeft {
					processor.releaseLeftAt = releaseAt
				} else {
					processor.releaseRightAt = releaseAt
				}

				processor.wasLeftBefore = shouldBeLeft

				processor.previousEnd = endTime

				processor.queue = append(processor.queue[:i], processor.queue[i+1:]...)

				i--
			}
		}
	}

	processor.cursor.LeftKey = time < processor.releaseLeftAt
	processor.cursor.RightKey = time < processor.releaseRightAt

	processor.lastTime = time
}
