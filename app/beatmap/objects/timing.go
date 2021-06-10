package objects

import (
	"github.com/wieku/danser-go/app/bmath"
	"math"
)

type TimingPoint struct {
	Time float64

	beatLengthBase float64
	beatLength     float64

	SampleSet    int
	SampleIndex  int
	SampleVolume float64

	Kiai bool
}

func (t TimingPoint) GetRatio() float64 {
	if t.beatLength >= 0 || math.IsNaN(t.beatLength) {
		return 1.0
	}

	return float64(float32(math.Max(10, math.Min(-t.beatLength, 1000))) / 100)
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

	Points  []TimingPoint
	queue   []TimingPoint
	Current TimingPoint

	BaseSet int
	LastSet int
}

func NewTimings() *Timings {
	return &Timings{BaseSet: 1, LastSet: 1}
}

func (tim *Timings) AddPoint(time, beatLength float64, sampleSet, sampleIndex int, sampleVolume float64, inherited, kiai bool) {
	point := TimingPoint{
		Time:           time,
		beatLengthBase: beatLength,
		beatLength:     beatLength,
		SampleSet:      sampleSet,
		SampleIndex:    sampleIndex,
		SampleVolume:   sampleVolume,
		Kiai:           kiai,
	}

	if inherited && len(tim.Points) > 0 {
		lastPoint := tim.Points[len(tim.Points)-1]
		point.beatLengthBase = lastPoint.beatLengthBase
	}

	tim.Points = append(tim.Points, point)
	tim.queue = append(tim.queue, point)
}

func (tim *Timings) Update(time float64) {
	if len(tim.queue) > 0 {
		p := tim.queue[0]
		if p.Time <= time {
			tim.queue = tim.queue[1:]
			tim.Current = p
		}
	}
}

func (tim *Timings) GetPoint(time float64) TimingPoint {
	for i, pt := range tim.Points {
		if time < pt.Time {
			return tim.Points[bmath.ClampI(i-1, 0, len(tim.Points)-1)]
		}
	}

	return tim.Points[len(tim.Points)-1]
}

func (tim *Timings) GetScoringDistance() float64 {
	return (100 * tim.SliderMult) / tim.TickRate
}

func (tim *Timings) GetSliderTimeP(point TimingPoint, pixelLength float64) float64 {
	return float64(float32(1000.0*pixelLength) / float32(100.0*tim.SliderMult*(1000.0/point.GetBeatLength())))
}

func (tim *Timings) GetVelocity(point TimingPoint) float64 {
	return tim.GetScoringDistance() * tim.TickRate * (1000.0 / point.GetBeatLength())
}

func (tim *Timings) GetTickDistance(point TimingPoint) float64 {
	return tim.GetScoringDistance() / point.GetRatio()
}

func (tim *Timings) Reset() {
	tim.queue = make([]TimingPoint, len(tim.Points))
	copy(tim.queue, tim.Points)

	tim.Current = tim.queue[0]
}
