package beatmap

import (
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
	"strconv"
	"strings"
	"time"
)

type BeatMap struct {
	Artist        string
	ArtistUnicode string
	Name          string
	NameUnicode   string
	Difficulty    string
	Creator       string
	Source        string
	Tags          string

	Mode int64

	SliderMultiplier float64
	StackLeniency    float64

	Diff *difficulty.Difficulty

	Dir   string
	File  string
	Audio string
	Bg    string
	MD5   string

	LastModified, TimeAdded, PlayCount, LastPlayed, PreviewTime int64

	Stars float64

	Length   int
	Circles  int
	Sliders  int
	Spinners int

	MinBPM float64
	MaxBPM float64

	Timings    *objects.Timings
	HitObjects []objects.BaseObject
	Pauses     []objects.BaseObject
	Queue      []objects.BaseObject
}

func NewBeatMap() *BeatMap {
	beatMap := &BeatMap{
		Timings:       objects.NewTimings(),
		StackLeniency: 0.7,
		Diff:          difficulty.NewDifficulty(5, 5, 5, 5),
		Stars:         -1,
		MinBPM:        math.Inf(0),
		MaxBPM:        0,
	}
	//beatMap.Diff.SetMods(difficulty.Hidden)
	return beatMap
}

func (b *BeatMap) Reset() {
	b.Queue = make([]objects.BaseObject, len(b.HitObjects))
	copy(b.Queue, b.HitObjects)
	b.Timings.Reset()
	for _, o := range b.HitObjects {
		o.SetDifficulty(b.Diff)
	}
}

func (b *BeatMap) Update(time int64) {
	b.Timings.Update(time)

	for i := 0; i < len(b.Queue); i++ {
		g := b.Queue[i]
		if g.GetBasicData().StartTime-int64(b.Diff.Preempt) > time {
			break
		}

		g.Update(time)

		if time >= g.GetBasicData().EndTime+difficulty.HitFadeOut+b.Diff.Hit50 {
			if i < len(b.Queue)-1 {
				b.Queue = append(b.Queue[:i], b.Queue[i+1:]...)
			} else if i < len(b.Queue) {
				b.Queue = b.Queue[:i]
			}
			i--
		}
	}
}

func (beatMap *BeatMap) GetObjectsCopy() []objects.BaseObject {
	queue := make([]objects.BaseObject, len(beatMap.HitObjects))
	copy(queue, beatMap.HitObjects)
	return queue
}

func (beatMap *BeatMap) ParsePoint(point string) {
	line := strings.Split(point, ",")
	pointTime, _ := strconv.ParseInt(line[0], 10, 64)
	bpm, _ := strconv.ParseFloat(line[1], 64)

	if !math.IsNaN(bpm) && bpm >= 0 {
		rBPM := 60000 / bpm
		beatMap.MinBPM = math.Min(beatMap.MinBPM, rBPM)
		beatMap.MaxBPM = math.Max(beatMap.MaxBPM, rBPM)
	}

	if len(line) > 3 {
		sampleset, _ := strconv.ParseInt(line[3], 10, 64)
		sampleindex, _ := strconv.ParseInt(line[4], 10, 64)

		samplevolume := int64(100)

		if len(line) > 5 {
			samplevolume, _ = strconv.ParseInt(line[5], 10, 64)
		}

		inherited := false
		if len(line) > 6 {
			inh, _ := strconv.ParseInt(line[6], 10, 64)
			inherited = inh == 0
		}

		kiai := false
		if len(line) > 7 {
			ki, _ := strconv.ParseInt(line[7], 10, 64)
			kiai = ki == 1
		}

		beatMap.Timings.LastSet = int(sampleset)
		beatMap.Timings.AddPoint(pointTime, bpm, int(sampleset), int(sampleindex), float64(samplevolume)/100, inherited, kiai)
	} else {
		beatMap.Timings.AddPoint(pointTime, bpm, beatMap.Timings.LastSet, 1, 1, false, false)
	}
}

func (beatMap *BeatMap) LoadCustomSamples() {
	audio.LoadBeatmapSamples(beatMap.Dir)
}

func (beatMap *BeatMap) UpdatePlayStats() {
	beatMap.PlayCount += 1
	beatMap.LastPlayed = time.Now().UnixNano() / 1000000
}
