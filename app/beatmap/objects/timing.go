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

	return float64(float32(mutils.ClampF64(-t.beatLength, 10, 1000)) / 100)
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
	LastSet int
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
		LastSet: 1,
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
	tim.Current = tim.GetPoint(time)
}

func (tim *Timings) GetDefault() TimingPoint {
	return tim.defaultTimingPoint
}

func (tim *Timings) GetPoint(time float64) TimingPoint {
	tLen := len(tim.points)

	// We have to search in reverse because sort.Search searches for lowest index at which condition is true, we want the opposite
	index := sort.Search(tLen, func(i int) bool {
		return time >= tim.points[tLen-i-1].Time
	})

	// We have to revert the index to get correct timing point
	return tim.points[mutils.MaxI(0, tLen-index-1)]
}

func (tim *Timings) GetOriginalPoint(time float64) TimingPoint {
	tLen := len(tim.originalPoints)

	// We have to search in reverse because sort.Search searches for lowest index at which condition is true, we want the opposite
	index := sort.Search(tLen, func(i int) bool {
		return time >= tim.originalPoints[tLen-i-1].Time
	})

	// We have to revert the index to get correct timing point
	return tim.originalPoints[mutils.MaxI(0, tLen-index-1)]
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

func (tim *Timings) Reset() {
	tim.Current = tim.points[0]
}
