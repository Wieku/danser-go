package bass

/*
#include "bass.h"
#include "bassmix.h"
*/
import "C"

import (
	"github.com/wieku/danser-go/framework/math/mutils"
)

type TrackVirtual struct {
	fft              []float32
	length           float64
	tail             float64
	speed            float64
	pitch            float64
	rFreq            float64
	previousPosition float64
	startTime        float64
	playing          bool
	paused           bool
}

func NewTrackVirtual(length float64) *TrackVirtual {
	player := &TrackVirtual{
		fft:    make([]float32, 512),
		speed:  1,
		pitch:  1,
		length: length,
	}

	return player
}

func (track *TrackVirtual) AddSilence(seconds float64) {
	track.tail = seconds
}

func (track *TrackVirtual) Play() {
	track.playInternal()
}

func (track *TrackVirtual) PlayV(_ float64) {
	track.playInternal()
}

func (track *TrackVirtual) playInternal() {
	track.playing = true

	track.startTime = float64(C.BASS_ChannelBytes2Seconds(masterMixer, C.BASS_ChannelGetPosition(masterMixer, C.BASS_POS_BYTE)))
	track.previousPosition = 0
}

func (track *TrackVirtual) Pause() {
	track.previousPosition = track.GetPosition()
	track.playing = false
	track.paused = true
}

func (track *TrackVirtual) Resume() {
	track.playing = true
	track.paused = false

	track.SetPosition(track.previousPosition)
}

func (track *TrackVirtual) Stop() {
	track.playing = false
	track.previousPosition = 0
}

func (track *TrackVirtual) SetVolume(_ float64) {
}

func (track *TrackVirtual) SetVolumeRelative(_ float64) {
}

func (track *TrackVirtual) GetLength() float64 {
	return track.length
}

func (track *TrackVirtual) SetPosition(pos float64) {
	track.previousPosition = pos
	track.startTime = float64(C.BASS_ChannelBytes2Seconds(masterMixer, C.BASS_ChannelGetPosition(masterMixer, C.BASS_POS_BYTE)))
}

func (track *TrackVirtual) GetPosition() float64 {
	if !track.playing {
		return track.previousPosition
	}

	currentPos := float64(C.BASS_ChannelBytes2Seconds(masterMixer, C.BASS_ChannelGetPosition(masterMixer, C.BASS_POS_BYTE)))

	pos := track.previousPosition + (currentPos-track.startTime)*track.speed*track.rFreq

	return mutils.Clamp(pos, 0, track.length+track.tail)
}

func (track *TrackVirtual) SetTempo(tempo float64) {
	if track.speed == tempo {
		return
	}

	track.previousPosition = track.GetPosition()
	track.startTime = float64(C.BASS_ChannelBytes2Seconds(masterMixer, C.BASS_ChannelGetPosition(masterMixer, C.BASS_POS_BYTE)))

	track.speed = tempo
}

func (track *TrackVirtual) GetTempo() float64 {
	return track.speed
}

func (track *TrackVirtual) SetPitch(pitch float64) {
	track.pitch = pitch
}

func (track *TrackVirtual) GetPitch() float64 {
	return track.pitch
}

func (track *TrackVirtual) SetRelativeFrequency(rFreq float64) {
	if track.rFreq == rFreq {
		return
	}

	track.previousPosition = track.GetPosition()
	track.startTime = float64(C.BASS_ChannelBytes2Seconds(masterMixer, C.BASS_ChannelGetPosition(masterMixer, C.BASS_POS_BYTE)))

	track.rFreq = rFreq
}

func (track *TrackVirtual) GetRelativeFrequency() float64 {
	return track.rFreq
}

func (track *TrackVirtual) GetSpeed() float64 {
	return track.speed * track.rFreq
}

func (track *TrackVirtual) GetState() int {
	if !track.playing {
		if track.paused {
			return MusicPaused
		}

		return MusicStopped
	}

	pos := track.GetPosition()

	if pos == 0 || pos >= track.length+track.tail {
		return MusicStopped
	}

	return MusicPlaying
}

func (track *TrackVirtual) Update() {}

func (track *TrackVirtual) GetFFT() []float32 {
	return track.fft
}

func (track *TrackVirtual) GetPeak() float64 {
	return 0
}

func (track *TrackVirtual) GetLevelCombined() float64 {
	return 0
}

func (track *TrackVirtual) GetLeftLevel() float64 {
	return 0
}

func (track *TrackVirtual) GetRightLevel() float64 {
	return 0
}

func (track *TrackVirtual) GetBoost() float64 {
	return 0
}

func (track *TrackVirtual) GetBeat() float64 {
	return 0
}
