package audio

var Samples [4][3]*Sample

func LoadSamples() {

	Samples[0][0] = NewSample("assets/sounds/normal-hitnormal.wav")
	Samples[1][0] = NewSample("assets/sounds/normal-hitwhistle.wav")
	Samples[2][0] = NewSample("assets/sounds/normal-hitfinish.wav")
	Samples[3][0] = NewSample("assets/sounds/normal-hitclap.wav")

	Samples[0][1] = NewSample("assets/sounds/soft-hitnormal.wav")
	Samples[1][1] = NewSample("assets/sounds/soft-hitwhistle.wav")
	Samples[2][1] = NewSample("assets/sounds/soft-hitfinish.wav")
	Samples[3][1] = NewSample("assets/sounds/soft-hitclap.wav")

	Samples[0][2] = NewSample("assets/sounds/drum-hitnormal.wav")
	Samples[1][2] = NewSample("assets/sounds/drum-hitwhistle.wav")
	Samples[2][2] = NewSample("assets/sounds/drum-hitfinish.wav")
	Samples[3][2] = NewSample("assets/sounds/drum-hitclap.wav")
}

func PlaySample(sampleSet int, hitsound int) {
	Samples[0][sampleSet-1].Play()
	if hitsound&2 > 0 {
		Samples[1][sampleSet-1].Play()
	}
	if hitsound&4 > 0 {
		Samples[2][sampleSet-1].Play()
	}
	if hitsound&8 > 0 {
		Samples[3][sampleSet-1].Play()
	}
}