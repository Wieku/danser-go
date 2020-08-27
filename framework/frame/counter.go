package frame

import (
	"log"
	"math"
)

type Counter struct {
	samples []float64
	index   int
	FPS     float64
	sum     float64
	log     bool
}

func NewCounter(samples int, log bool) *Counter {
	return &Counter{make([]float64, samples), -1, 0, 0.0, log}
}

func (prof *Counter) PutSample(fps float64) {
	prof.index++
	if prof.index >= len(prof.samples) {
		prof.index = 0
	}
	prof.samples[prof.index] = fps
	prof.sum += fps
	if prof.sum >= 1.0 && prof.log {
		log.Println("FPS:", prof.GetFPS())
		prof.sum = 0.0
	}
}

func (prof *Counter) GetFPS() float64 {
	sum := 0.0
	count := 0
	for _, g := range prof.samples {
		if g > 0.01 {
			sum += g
			count++
		}
	}

	if count == 0 {
		return math.NaN()
	}

	return 1000 / (sum / float64(count))
}
