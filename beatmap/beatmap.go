package beatmap

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/movers"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/bmath"
)

type BeatMap struct {
	Artist, Name, Difficulty, Creator        string
	SliderMultiplier, StackLeniency, CircleSize, AR, ARms float64
	Path 							string
	Audio 							string
	Bg 								string
	timings                         *objects.Timings
	HitObjects                      []objects.BaseObject
	Queue                      		[]objects.BaseObject
	movers                          []movers.Mover
}

const MoverId = 2

func NewBeatMap() *BeatMap {
	return &BeatMap{timings: objects.NewTimings(), movers: []movers.Mover{movers.NewBezierMover(), movers.NewCircularMover(), movers.NewFlowerBezierMover()}, StackLeniency: 0.7}
}

func (b *BeatMap) Reset() {
	b.Queue = make([]objects.BaseObject, len(b.HitObjects))
	copy(b.Queue, b.HitObjects)
	b.timings.Reset()
	b.movers[MoverId].Reset()
	b.movers[MoverId].SetObjects(objects.DummyCircle(bmath.NewVec2d(100, 100), 0), b.Queue[0])
}

func (b *BeatMap) Update(time int64, cursor *render.Cursor) {
	b.timings.Update(time)
	if len(b.Queue) > 0 {
		any := false
		for i, g := range b.Queue {
			if g.GetBasicData().StartTime > time {
				break
			}

			any = true

			if isDone := g.Update(time, cursor); isDone {
				if i < len(b.Queue) -1 {
					b.Queue = append(b.Queue[:i], b.Queue[i+1:]...)
				} else if i < len(b.Queue) {
					b.Queue = b.Queue[:i]
				}

				if len(b.Queue) > 0 && i < len(b.Queue) {
					b.movers[MoverId].SetObjects(g, b.Queue[i])
				}
			}
		}

		if !any {
			b.movers[MoverId].Update(time, cursor)
		}

		/*if p := b.Queue[0]; p.GetBasicData().StartTime <= time {
			if isDone := p.Update(time, cursor); isDone {
				b.Queue = b.Queue[1:]
				if len(b.Queue) > 0 {
					b.movers[MoverId].SetObjects(p, b.Queue[0])
				}
			}
		} else {
			b.movers[MoverId].Update(time, cursor)
		}*/
	}
}


