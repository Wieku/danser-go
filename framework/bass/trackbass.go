package bass

/*
#include "bass_util.h"
#include "bass.h"
#include "bass_fx.h"
#include "bassmix.h"
*/
import "C"

import (
	"github.com/wieku/danser-go/app/settings"
	"math"
	"unsafe"
)

type TrackBass struct {
	channel      C.HSTREAM
	fft          []float32
	boost        float64
	peak         float64
	leftChannel  float64
	rightChannel float64
	lowMax       float64
	speed        float64
	pitch        float64
	playing      bool
	addedToMixer bool
}

func NewTrack(path string) *TrackBass {
	player := &TrackBass{
		fft:   make([]float32, 512),
		speed: 1,
		pitch: 1,
	}

	player.channel = C.CreateBassStream(C.CString(path), C.BASS_STREAM_DECODE|C.BASS_STREAM_PRESCAN|C.BASS_ASYNCFILE)

	if player.channel == 0 {
		return nil
	}

	player.channel = C.BASS_FX_TempoCreate(player.channel, C.BASS_FX_FREESOURCE|C.BASS_STREAM_DECODE)

	setupFXChannel(player.channel)

	return player
}

func setupFXChannel(channel C.HSTREAM) {
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_USE_QUICKALGO, 1)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_OVERLAP_MS, C.float(4.0))
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_SEQUENCE_MS, C.float(30.0))
}

func (track *TrackBass) AddSilence(seconds float64) {
	C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_TAIL, C.float(seconds))
}

func (track *TrackBass) Play() {
	track.SetVolume(settings.Audio.GeneralVolume * settings.Audio.MusicVolume)

	C.BASS_Mixer_StreamAddChannel(masterMixer, track.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_MIXER_CHAN_BUFFER)

	track.playing = true
	track.addedToMixer = true
}

func (track *TrackBass) PlayV(volume float64) {
	track.SetVolume(volume)

	track.playing = true

	C.BASS_Mixer_StreamAddChannel(masterMixer, track.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_MIXER_CHAN_BUFFER)
	track.addedToMixer = true
}

func (track *TrackBass) Pause() {
	track.playing = false

	C.BASS_Mixer_ChannelFlags(track.channel, C.BASS_MIXER_CHAN_PAUSE, C.BASS_MIXER_CHAN_PAUSE)
}

func (track *TrackBass) Resume() {
	track.playing = true

	C.BASS_Mixer_ChannelFlags(track.channel, 0, C.BASS_MIXER_CHAN_PAUSE)
}

func (track *TrackBass) Stop() {
	track.playing = false
	track.addedToMixer = false

	C.BASS_ChannelStop(track.channel)
}

func (track *TrackBass) SetVolume(vol float64) {
	C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_VOL, C.float(vol))
}

func (track *TrackBass) SetVolumeRelative(vol float64) {
	combined := settings.Audio.GeneralVolume * settings.Audio.MusicVolume * vol

	C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_VOL, C.float(combined))
}

func (track *TrackBass) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(track.channel, C.BASS_ChannelGetLength(track.channel, C.BASS_POS_BYTE)))
}

func (track *TrackBass) SetPosition(pos float64) {
	if track.addedToMixer {
		C.BASS_Mixer_ChannelSetPosition(track.channel, C.BASS_ChannelSeconds2Bytes(track.channel, C.double(pos)), C.BASS_POS_BYTE)
	} else {
		C.BASS_ChannelSetPosition(track.channel, C.BASS_ChannelSeconds2Bytes(track.channel, C.double(pos)), C.BASS_POS_BYTE)
	}
}

func (track *TrackBass) GetPosition() float64 {
	var bassPos float64

	if track.addedToMixer {
		bassPos = float64(C.BASS_ChannelBytes2Seconds(track.channel, C.BASS_Mixer_ChannelGetPosition(track.channel, C.BASS_POS_BYTE)))
	} else {
		bassPos = float64(C.BASS_ChannelBytes2Seconds(track.channel, C.BASS_ChannelGetPosition(track.channel, C.BASS_POS_BYTE)))
	}

	return bassPos
}

func (track *TrackBass) SetTempo(tempo float64) {
	if track.speed == tempo {
		return
	}

	track.speed = tempo

	C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))
}

func (track *TrackBass) GetTempo() float64 {
	return track.speed
}

func (track *TrackBass) SetPitch(tempo float64) {
	if track.pitch == tempo {
		return
	}

	track.pitch = tempo

	C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*14.4))
}

func (track *TrackBass) GetPitch() float64 {
	return track.pitch
}

func (track *TrackBass) GetState() int {
	if !track.addedToMixer {
		return MUSIC_STOPPED
	}

	state := int(C.BASS_ChannelIsActive(track.channel))

	if state == MUSIC_PLAYING && track.addedToMixer && C.BASS_Mixer_ChannelFlags(track.channel, 0, 0)&C.BASS_MIXER_CHAN_PAUSE > 0 {
		return MUSIC_PAUSED
	}

	return state
}

func (track *TrackBass) Update() {
	if track.playing {
		if track.addedToMixer {
			C.BASS_Mixer_ChannelGetData(track.channel, unsafe.Pointer(&track.fft[0]), C.BASS_DATA_FFT1024)
		} else {
			C.BASS_ChannelGetData(track.channel, unsafe.Pointer(&track.fft[0]), C.BASS_DATA_FFT1024)
		}
	} else {
		for i := range track.fft {
			track.fft[i] = 0
		}
	}

	toPeak := 0.0
	beatAv := 0.0

	for i, g := range track.fft {
		h := math.Abs(float64(g))

		toPeak = math.Max(toPeak, h)

		if i > 0 && i < 5 {
			beatAv = math.Max(beatAv, float64(g))
		}
	}

	boost := 0.0

	for i := 0; i < 10; i++ {
		boost += float64(track.fft[i]*track.fft[i]) * float64(10-i) / float64(10)
	}

	track.lowMax = beatAv
	track.boost = boost
	track.peak = toPeak

	var level int

	if track.addedToMixer {
		level = int(C.BASS_Mixer_ChannelGetLevel(track.channel))
	} else {
		level = int(C.BASS_ChannelGetLevel(track.channel))
	}

	left := level & 65535
	right := level >> 16

	track.leftChannel = float64(left) / 32768
	track.rightChannel = float64(right) / 32768
}

func (track *TrackBass) GetFFT() []float32 {
	return track.fft
}

func (track *TrackBass) GetPeak() float64 {
	return track.peak
}

func (track *TrackBass) GetLevelCombined() float64 {
	return (track.leftChannel + track.rightChannel) / 2
}

func (track *TrackBass) GetLeftLevel() float64 {
	return track.leftChannel
}

func (track *TrackBass) GetRightLevel() float64 {
	return track.rightChannel
}

func (track *TrackBass) GetBoost() float64 {
	return track.boost
}

func (track *TrackBass) GetBeat() float64 {
	return track.lowMax
}
