package frame

import (
	"math"
)

type Counter struct {
	samples []float64
	index   int
}

func NewCounter(samples int) *Counter {
	return &Counter{
		samples: make([]float64, samples),
		index:   0,
	}
}

func (counter *Counter) PutSample(delta float64) {
	counter.samples[counter.index] = delta
	counter.index = (counter.index + 1) % len(counter.samples)
}

func (counter *Counter) GetAverage() float64 {
	sum := 0.0
	count := 0
	for _, g := range counter.samples {
		if g > 0.01 {
			sum += g
			count++
		}
	}

	if count == 0 {
		return math.NaN()
	}

	return sum / float64(count)
}

func (counter *Counter) GetFPS() float64 {
	return 1000 / counter.GetAverage()
}
