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

	rtBase := delta
	if rtBase < 0.2 {
		rtBase = rtBase*0.5 + 0.1
	}

	rate := 1 - math.Pow(0.1, rtBase/1000)
	counter.avg = counter.avg + (delta-counter.avg)*rate
}

func (counter *Counter) GetAverage() float64 {
	return counter.avg
}

func (counter *Counter) GetFPS() float64 {
	return 1000 / counter.avg
}
