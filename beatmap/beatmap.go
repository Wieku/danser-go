package beatmap

import (
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/movers"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/bmath"
	"strings"
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
	movers                          []func() movers.Mover
	moversC                         []movers.Mover
	cursors 						[]*render.Cursor
}

var MoverId = 2

func SetMover(name string) {
	name = strings.ToLower(name)

	if name == "bezier" {
		MoverId = 0
	} else if name == "circular" {
		MoverId = 1
	} else if name == "linear" {
		MoverId = 3
	} else {
		MoverId = 2
	}

}

func NewBeatMap() *BeatMap {
	return &BeatMap{timings: objects.NewTimings(), movers: []func() movers.Mover {movers.NewBezierMover, movers.NewCircularMover, movers.NewFlowerBezierMover, movers.NewLinearMover}, StackLeniency: 0.7}
}

func (b *BeatMap) Reset() {
	b.Queue = make([]objects.BaseObject, len(b.HitObjects))
	copy(b.Queue, b.HitObjects)
	b.timings.Reset()

	b.moversC = make([]movers.Mover, len(b.cursors))

	for i:=0; i < len(b.cursors); i++ {
		b.moversC[i] = b.movers[MoverId]()
		index := 0
		for j, o := range b.Queue {
			if o.GetBasicData().Number % int64(len(b.cursors)) == int64(i) || o.GetBasicData().Number == -1 {
				index = j
				break
			}
		}
		b.moversC[i].SetObjects(objects.DummyCircle(bmath.NewVec2d(100, 100), 0), b.Queue[index])
	}
}

func (b *BeatMap) SetCursors(cursors []*render.Cursor) {
	b.cursors = cursors
}

func (b *BeatMap) Update(time int64, cursor *render.Cursor) {
	b.timings.Update(time)
	if len(b.Queue) > 0 {
		any := make([]bool, len(b.cursors))
		for i:=0; i < len(b.Queue); i++ {
			g := b.Queue[i]
			if g.GetBasicData().StartTime > time {
				break
			}

			if _, ok := g.(*objects.Pause); ok {
				if i < len(b.Queue) -1 {
					b.Queue = append(b.Queue[:i], b.Queue[i+1:]...)
				} else if i < len(b.Queue) {
					b.Queue = b.Queue[:i]
				}

				i--
				continue
			}

			any[int(g.GetBasicData().Number) % len(b.cursors)] = true

			if isDone := g.Update(time, b.cursors[int(g.GetBasicData().Number) % len(b.cursors)]); isDone {
				if i < len(b.Queue) -1 {
					b.Queue = append(b.Queue[:i], b.Queue[i+1:]...)
				} else if i < len(b.Queue) {
					b.Queue = b.Queue[:i]
				}
				i--
				if len(b.Queue) > 0 && i < len(b.Queue) {

					for _, o := range b.Queue {
						if o.GetBasicData().Number != -1 && o.GetBasicData().StartTime > g.GetBasicData().EndTime &&((int(o.GetBasicData().Number) % len(b.cursors)) == (int(g.GetBasicData().Number) % len(b.cursors)) ) {
							b.moversC[int(g.GetBasicData().Number) % len(b.cursors)].SetObjects(g, o)
							any[int(g.GetBasicData().Number) % len(b.cursors)] = true
							break
						}
					}

				}
			}
		}

		for i, g := range b.cursors {
			if !any[i] {
				b.moversC[i].Update(time, g)
			}
		}

	}

}


