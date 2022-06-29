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

	SetID int64
	ID    int64

	LastModified, TimeAdded, PlayCount, LastPlayed, PreviewTime int64

	Stars        float64
	StarsVersion int

	Length   int
	Circles  int
	Sliders  int
	Spinners int

	MinBPM float64
	MaxBPM float64

	Timings    *objects.Timings
	HitObjects []objects.IHitObject
	Pauses     []*Pause
	Queue      []objects.IHitObject
	Version    int

	ARSpecified bool

	LocalOffset int
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
	//beatMap.Diff.SetMods(difficulty.HardRock|difficulty.Hidden)
	return beatMap
}

func (beatMap *BeatMap) Reset() {
	beatMap.Queue = beatMap.GetObjectsCopy()
	beatMap.Timings.Reset()

	for _, o := range beatMap.HitObjects {
		o.SetDifficulty(beatMap.Diff)
	}
}

func (beatMap *BeatMap) Clear() {
	beatMap.HitObjects = make([]objects.IHitObject, 0)
	beatMap.Timings = objects.NewTimings()
}

func (beatMap *BeatMap) Update(time float64) {
	beatMap.Timings.Update(time)

	for i := 0; i < len(beatMap.Queue); i++ {
		g := beatMap.Queue[i]
		if g.GetStartTime()-beatMap.Diff.Preempt > time {
			break
		}

		g.Update(time)

		if time >= g.GetEndTime()+difficulty.HitFadeOut+float64(beatMap.Diff.Hit50) {
			if i < len(beatMap.Queue)-1 {
				beatMap.Queue = append(beatMap.Queue[:i], beatMap.Queue[i+1:]...)
			} else if i < len(beatMap.Queue) {
				beatMap.Queue = beatMap.Queue[:i]
			}
			i--
		}
	}
}

func (beatMap *BeatMap) GetObjectsCopy() []objects.IHitObject {
	objs := make([]objects.IHitObject, len(beatMap.HitObjects))
	copy(objs, beatMap.HitObjects)

	return objs
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

	signature := 4
	sampleSet := beatMap.Timings.BaseSet
	sampleIndex := 1
	sampleVolume := 1.0
	inherited := false
	kiai := false
	omitFirstBarLine := false

	if len(line) > 2 {
		signature, _ = strconv.Atoi(line[2])
		if signature == 0 {
			signature = 4
		}
	}

	if len(line) > 3 {
		sampleSet, _ = strconv.Atoi(line[3])
	}

	if len(line) > 4 {
		sampleIndex, _ = strconv.Atoi(line[4])
	}

	if len(line) > 5 {
		sV, _ := strconv.Atoi(line[5])
		sampleVolume = float64(sV) / 100
	}

	if len(line) > 6 {
		inh, _ := strconv.Atoi(line[6])
		inherited = inh == 0
	}

	if len(line) > 7 {
		ki, _ := strconv.Atoi(line[7])
		kiai = (ki & 1) > 0
		omitFirstBarLine = (ki & 8) > 0
	}

	beatMap.Timings.AddPoint(float64(pointTime), bpm, sampleSet, sampleIndex, sampleVolume, signature, inherited, kiai, omitFirstBarLine)
}

func (beatMap *BeatMap) FinalizePoints() {
	beatMap.Timings.FinalizePoints()
}

func (beatMap *BeatMap) LoadCustomSamples() {
	audio.LoadBeatmapSamples(beatMap.Dir)
}

func (beatMap *BeatMap) UpdatePlayStats() {
	beatMap.PlayCount += 1
	beatMap.LastPlayed = time.Now().UnixNano() / 1000000
}
