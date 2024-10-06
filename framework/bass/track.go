package bass

const (
	MusicStopped = 0
	MusicPlaying = 1
	MusicStalled = 2
	MusicPaused  = 3
)

type ITrack interface {
	AddSilence(seconds float64)
	Play()
	PlayV(volume float64)
	Pause()
	Resume()
	Stop()
	SetVolume(vol float64)
	SetVolumeRelative(vol float64)
	GetLength() float64
	SetPosition(pos float64)
	GetPosition() float64
	SetTempo(tempo float64)
	GetTempo() float64
	SetPitch(pitch float64)
	GetPitch() float64
	SetRelativeFrequency(rFreq float64)
	GetRelativeFrequency() float64
	GetSpeed() float64
	GetState() int
	Update()
	GetFFT() []float32
	GetPeak() float64
	GetLevelCombined() float64
	GetLeftLevel() float64
	GetRightLevel() float64
	GetBoost() float64
	GetBeat() float64
}
