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
	movers                          []func() movers.Mover
	moversC                         []movers.Mover
	cursors 						[]*render.Cursor
}

const MoverId = 2

func NewBeatMap() *BeatMap {
	return &BeatMap{timings: objects.NewTimings(), movers: []func() movers.Mover {movers.NewBezierMover, movers.NewCircularMover, movers.NewFlowerBezierMover}, StackLeniency: 0.7}
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

	//b.movers[MoverId].Reset()
	//b.movers[MoverId].SetObjects(objects.DummyCircle(bmath.NewVec2d(100, 100), 0), b.Queue[0])
}

func (b *BeatMap) SetCursors(cursors []*render.Cursor) {
	b.cursors = cursors
}

func (b *BeatMap) Update(time int64, cursor *render.Cursor) {
	b.timings.Update(time)
	if len(b.Queue) > 0 {
		any := make([]bool, len(b.cursors))
		//log.Println(any)
		for /*i, g := range b.Queue*/i:=0; i < len(b.Queue); i++ {
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
						if /*j >= i && */o.GetBasicData().Number != -1 && o.GetBasicData().StartTime > g.GetBasicData().EndTime &&((int(o.GetBasicData().Number) % len(b.cursors)) == (int(g.GetBasicData().Number) % len(b.cursors)) /*|| o.GetBasicData().Number == -1*/) {
							//log.Println(i, j, g.GetBasicData().Number, o.GetBasicData().Number, int(g.GetBasicData().Number) % len(b.cursors), int(o.GetBasicData().Number) % len(b.cursors))
							//log.Println(o)
							b.moversC[int(g.GetBasicData().Number) % len(b.cursors)].SetObjects(g, o)
							any[int(g.GetBasicData().Number) % len(b.cursors)] = true
							break
						}
					}

					//b.movers[MoverId].SetObjects(g, b.Queue[i])
				}
			}
		}


		for i, g := range b.cursors {
			if !any[i] {
				b.moversC[i].Update(time, g)
			}
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

	/*
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
//}

}


