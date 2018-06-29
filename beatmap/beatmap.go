package beatmap

import (
	"github.com/wieku/danser/beatmap/objects"
)

type BeatMap struct {
	Artist, Name, Difficulty, Creator        string
	SliderMultiplier, StackLeniency, CircleSize, AR, ARms float64
	Path 							string
	Audio 							string
	Bg 								string
	timings                         *objects.Timings
	HitObjects                      []objects.BaseObject
	Pauses                      	[]objects.BaseObject
	Queue                      		[]objects.BaseObject
}

func NewBeatMap() *BeatMap {
	return &BeatMap{timings: objects.NewTimings(), StackLeniency: 0.7}
}

func (b *BeatMap) Reset() {
	b.Queue = make([]objects.BaseObject, len(b.HitObjects))
	copy(b.Queue, b.HitObjects)
	b.timings.Reset()
}

func (b *BeatMap) Update(time int64) {
	b.timings.Update(time)
	if len(b.Queue) > 0 {
		for i:=0; i < len(b.Queue); i++ {
			g := b.Queue[i]
			if g.GetBasicData().StartTime > time {
				break
			}

			if isDone := g.Update(time); isDone {
				if i < len(b.Queue) -1 {
					b.Queue = append(b.Queue[:i], b.Queue[i+1:]...)
				} else if i < len(b.Queue) {
					b.Queue = b.Queue[:i]
				}
				i--
			}
		}
	}

}


