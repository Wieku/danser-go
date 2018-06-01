package beatmap

import (
	"danser/beatmap/objects"
	"danser/movers"
	"danser/render"
	"danser/bmath"
)

type BeatMap struct {
	Artist, Name, Difficulty        string
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
	return &BeatMap{timings: objects.NewTimings(), movers: []movers.Mover{movers.NewBezierMover(), movers.NewCircularMover(), movers.NewFlowerBezierMover()}}
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
		if p := b.Queue[0]; p.GetBasicData().StartTime <= time {
			if isDone := p.Update(time, cursor); isDone {
				b.Queue = b.Queue[1:]
				if len(b.Queue) > 0 {
					b.movers[MoverId].SetObjects(p, b.Queue[0])
				}
			}
		} else {
			b.movers[MoverId].Update(time, cursor)
		}
	}
}


