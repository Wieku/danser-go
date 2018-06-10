package objects

import "log"

type TimingPoint struct {
	Time int64
	BaseBpm, Bpm float64
	SampleSet int
}

func (t TimingPoint) GetRatio() float64 {
	return t.Bpm / t.BaseBpm
}

type Timings struct {
	points []TimingPoint
	queue []TimingPoint
	SliderMult float64
	Current TimingPoint
	fullBPM, partBPM float64
	BaseSet int
	LastSet int
	TickRate float64
}

func NewTimings() *Timings {
	return &Timings{}
}

func (tim *Timings) AddPoint(time int64, bpm float64, sampleset int) {
	point := TimingPoint{Time: time, Bpm: bpm, SampleSet: sampleset}
	if point.Bpm > 0 {
		tim.fullBPM = point.Bpm
	} else {
		point.Bpm = tim.fullBPM / ( -100.0 / point.Bpm)
	}
	point.BaseBpm = tim.fullBPM
	tim.points = append(tim.points, point)
	tim.queue = append(tim.queue, point)
}

func (tim *Timings) Update(time int64) {
	if len(tim.queue) > 0 {
		p := tim.queue[0]
		if p.Time <= time {
			tim.queue = tim.queue[1:]
			tim.partBPM = p.Bpm
			tim.Current = p
		}
	}
}

func clamp(a, min, max int) int {
	if a > max {
		return max
	}
	if a < min {
		return min
	}
	return a
}

func (tim *Timings) GetPoint(time int64) TimingPoint {
	for i, pt := range tim.points {
		if time < pt.Time {
			return tim.points[clamp(i-1, 0, len(tim.points)-1)]
		}
	}
	return tim.points[len(tim.points)-1]
}

func (tim Timings) GetSliderTimeS(time int64, pixelLength float64) int64 {
	res := int64(tim.GetPoint(time).Bpm * pixelLength / (100.0 * tim.SliderMult))
	if res < 0 {
		log.Println("E?", tim.GetPoint(time).Bpm, pixelLength, tim.SliderMult)
	}
	return res
}

func (tim Timings) GetSliderTime(pixelLength float64) int64 {
	return int64(tim.partBPM * pixelLength / (100.0 * tim.SliderMult))
}

func (tim Timings) GetSliderTimeP(point TimingPoint, pixelLength float64) int64 {
	return int64(point.Bpm * pixelLength / (100.0 * tim.SliderMult))
}

func (tim *Timings) Reset() {
	tim.queue = make([]TimingPoint, len(tim.points))
	copy(tim.queue, tim.points)
	tim.Current = tim.queue[0]
}