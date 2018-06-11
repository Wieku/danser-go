package utils

type FPSCounter struct {
	samples []float64
	index int
	FPS float64
}

func NewFPSCounter(samples int) *FPSCounter {
	return &FPSCounter{make([]float64, samples), -1, 0}
}

func (prof *FPSCounter) PutSample(fps float64) {
	prof.index++
	if prof.index >= len(prof.samples) {
		prof.index = 0
	}
	prof.samples[prof.index] = fps
}

func (prof *FPSCounter) GetFPS() float64 {
	sum := 0.0
	for _, g := range prof.samples {
		sum +=g
	}
	return sum / float64(len(prof.samples))
}