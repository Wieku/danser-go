package frame

import (
	"math"
)

type Counter struct {
	avg float64
}

func NewCounter() *Counter {
	return &Counter{
		avg: 0,
	}
}

func (counter *Counter) PutSample(delta float64) {
	if counter.avg == 0 {
		counter.avg = delta
	}

	rate := 1 - math.Pow(0.1, delta/1000)
	counter.avg = counter.avg + (delta-counter.avg)*rate
}

func (counter *Counter) GetAverage() float64 {
	return counter.avg
}

func (counter *Counter) GetFPS() float64 {
	return 1000 / counter.avg
}
