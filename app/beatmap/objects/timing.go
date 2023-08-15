package objects

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
	"sort"
)

type TimingPoint struct {
	Time float64

	beatLengthBase float64
	beatLength     float64

	SampleSet    int
	SampleIndex  int
	SampleVolume float64

	Signature int

	Inherited bool

	Kiai             bool
	OmitFirstBarLine bool
}

func (t TimingPoint) GetRatio() float64 {
	if t.beatLength >= 0 || math.IsNaN(t.beatLength) {
		return 1.0
	}

	return float64(float32(mutils.Clamp(-t.beatLength, 10, 1000)) / 100)
}

func (t TimingPoint) GetBaseBeatLength() float64 {
	return t.beatLengthBase
}

func (t TimingPoint) GetBeatLength() float64 {
	return t.beatLengthBase * t.GetRatio()
}

type Timings struct {
	SliderMult float64
	TickRate   float64

	defaultTimingPoint TimingPoint

	points         []TimingPoint
	originalPoints []TimingPoint

	Current TimingPoint

	BaseSet int
}

func NewTimings() *Timings {
	return &Timings{
		defaultTimingPoint: TimingPoint{
			Time:             0,
			beatLengthBase:   60000 / 60,
			beatLength:       60000 / 60,
			SampleSet:        0,
			SampleIndex:      1,
			SampleVolume:     1,
			Signature:        4,
			Inherited:        false,
			Kiai:             false,
			OmitFirstBarLine: false,
		},
		BaseSet: 1,
	}
}

func (tim *Timings) AddPoint(time, beatLength float64, sampleSet, sampleIndex int, sampleVolume float64, signature int, inherited, kiai, omitFirstBarLine bool) {
	point := TimingPoint{
		Time:             time,
		beatLengthBase:   beatLength,
		beatLength:       beatLength,
		SampleSet:        sampleSet,
		SampleIndex:      sampleIndex,
		SampleVolume:     sampleVolume,
		Signature:        signature,
		Inherited:        inherited,
		Kiai:             kiai,
		OmitFirstBarLine: omitFirstBarLine,
	}

	tim.points = append(tim.points, point)
}

func (tim *Timings) FinalizePoints() {
	sort.SliceStable(tim.points, func(i, j int) bool {
		return tim.points[i].Time < tim.points[j].Time
	})

	for i, point := range tim.points {
		if point.Inherited && i > 0 {
			lastPoint := tim.points[i-1]
			point.beatLengthBase = lastPoint.beatLengthBase

			tim.points[i] = point
		} else {
			tim.originalPoints = append(tim.originalPoints, point)
		}
	}
}

func (tim *Timings) Update(time float64) {
	tim.Current = tim.GetPointAt(time)
}

func (tim *Timings) GetDefault() TimingPoint {
	return tim.defaultTimingPoint
}

func (tim *Timings) GetPointAt(time float64) TimingPoint {
	tLen := len(tim.points)

	index := sort.Search(tLen, func(i int) bool {
		return time < tim.points[i].Time
	})

	return tim.points[max(0, index-1)]
}

func (tim *Timings) GetOriginalPointAt(time float64) TimingPoint {
	tLen := len(tim.originalPoints)

	index := sort.Search(tLen, func(i int) bool {
		return time < tim.originalPoints[i].Time
	})

	return tim.originalPoints[max(0, index-1)]
}

func (tim *Timings) GetScoringDistance() float64 {
	return (100 * tim.SliderMult) / tim.TickRate
}

func (tim *Timings) GetSliderTimeP(point TimingPoint, pixelLength float64) float64 {
	return float64(float32(1000.0*pixelLength) / float32(100.0*tim.SliderMult*(1000.0/point.GetBeatLength())))
}

func (tim *Timings) GetVelocity(point TimingPoint) float64 {
	velocity := tim.GetScoringDistance() * tim.TickRate

	beatLength := point.GetBeatLength()

	if beatLength >= 0 {
		velocity *= 1000.0 / beatLength
	}

	return velocity
}

func (tim *Timings) GetTickDistance(point TimingPoint) float64 {
	return tim.GetScoringDistance() / point.GetRatio()
}

func (tim *Timings) HasPoints() bool {
	return len(tim.points) > 0
}

func (tim *Timings) Clear() {
	tim.originalPoints = tim.originalPoints[:0]
	tim.points = tim.points[:0]

	tim.Current = tim.defaultTimingPoint
}

func (tim *Timings) Reset() {
	tim.Current = tim.points[0]
}
