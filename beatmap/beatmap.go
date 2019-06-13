package beatmap

import (
	"github.com/wieku/danser-go/beatmap/objects"
	"strconv"
	"strings"
	"time"
	"github.com/wieku/danser-go/audio"
)

type BeatMap struct {
	Artist,
	ArtistUnicode,
	Name,
	NameUnicode,
	Difficulty,
	Creator,
	Source,
	Tags string

	SliderMultiplier,
	StackLeniency,
	CircleSize,
	AR,
	Preempt,
	FadeIn,
	OverallDifficulty,
	HPDrainRate float64

	Dir,
	File,
	Audio,
	Bg,
	MD5,
	PausesText,
	TimingPoints string

	LastModified, TimeAdded, PlayCount, LastPlayed, PreviewTime int64

	Timings    *objects.Timings
	HitObjects []objects.BaseObject
	Pauses     []objects.BaseObject
	Queue      []objects.BaseObject
}

func NewBeatMap() *BeatMap {
	return &BeatMap{Timings: objects.NewTimings(), StackLeniency: 0.7}
}

func (b *BeatMap) Reset() {
	b.Queue = make([]objects.BaseObject, len(b.HitObjects))
	copy(b.Queue, b.HitObjects)
	b.Timings.Reset()
	for _, o := range b.HitObjects {
		o.SetDifficulty(b.Preempt, b.FadeIn)
	}
}

func (b *BeatMap) Update(time int64) {
	b.Timings.Update(time)
	if len(b.Queue) > 0 {
		for i := 0; i < len(b.Queue); i++ {
			g := b.Queue[i]
			if g.GetBasicData().StartTime > time {
				break
			}

			if isDone := g.Update(time); isDone {
				if i < len(b.Queue)-1 {
					b.Queue = append(b.Queue[:i], b.Queue[i+1:]...)
				} else if i < len(b.Queue) {
					b.Queue = b.Queue[:i]
				}
				i--
			}
		}
	}

}
func (beatMap *BeatMap) GetObjectsCopy() []objects.BaseObject {
	queue := make([]objects.BaseObject, len(beatMap.HitObjects))
	copy(queue, beatMap.HitObjects)
	return queue
}

func (beatMap *BeatMap) LoadTimingPoints() {

	points := strings.Split(beatMap.TimingPoints, "|")

	if len(points) == 1 && points[0] == "" {
		return
	}

	for _, point := range points {
		line := strings.Split(point, ",")
		time, _ := strconv.ParseInt(line[0], 10, 64)
		bpm, _ := strconv.ParseFloat(line[1], 64)
		if len(line) > 3 {
			sampleset, _ := strconv.ParseInt(line[3], 10, 64)
			sampleindex, _ := strconv.ParseInt(line[4], 10, 64)

			samplevolume := int64(100)

			if len(line) > 5 {
				samplevolume, _ = strconv.ParseInt(line[5], 10, 64)
			}

			kiai := false
			if len(line) > 7 {
				ki, _ := strconv.ParseInt(line[7], 10, 64)
				if ki == 1 {
					kiai = true
				}
			}

			beatMap.Timings.LastSet = int(sampleset)
			beatMap.Timings.AddPoint(time, bpm, int(sampleset), int(sampleindex), float64(samplevolume)/100, kiai)
		} else {
			beatMap.Timings.AddPoint(time, bpm, beatMap.Timings.LastSet, 1, 1, false)
		}
	}
}

func (beatMap *BeatMap) LoadCustomSamples() {
	audio.LoadBeatmapSamples(beatMap.Dir)
}

func (beatMap *BeatMap) LoadPauses() {
	points := strings.Split(beatMap.PausesText, ",")

	if len(points) < 2 {
		return
	}

	for i := 0; i < len(points); i += 2 {
		line := []string{"2", points[i], points[i+1]}
		beatMap.Pauses = append(beatMap.Pauses, objects.NewPause(line))
	}
}

func (beatMap *BeatMap) UpdatePlayStats() {
	beatMap.PlayCount += 1
	beatMap.LastPlayed = time.Now().UnixNano() / 1000000
}
