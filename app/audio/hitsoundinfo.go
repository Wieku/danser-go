package audio

type HitSoundInfo struct {
	SampleSet    int
	AdditionSet  int
	CustomIndex  int
	CustomVolume float64
}

type HitSound struct {
	Sample int
	Info   HitSoundInfo
}
