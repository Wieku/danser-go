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
	"unsafe"
	//"log"
	"math"
)

const (
	MUSIC_STOPPED = 0
	MUSIC_PLAYING = 1
	MUSIC_STALLED = 2
	MUSIC_PAUSED  = 3
)

type Track struct {
	channel          C.DWORD
	offscreenChannel C.DWORD
	fft              []float32
	boost            float64
	peak             float64
	leftChannel      float64
	rightChannel     float64
	lowMax           float64
	lastVol          float64
	speed            float64
	pitch            float64
	playing          bool
}

func NewTrack(path string) *Track {
	player := new(Track)
	player.speed = 1
	player.pitch = 1
	player.lastVol = -100000
	player.fft = make([]float32, 512)

	player.channel = C.CreateBassStream(C.CString(path), C.BASS_ASYNCFILE|C.BASS_STREAM_DECODE|C.BASS_STREAM_PRESCAN)
	if !Offscreen {
		player.channel = C.BASS_FX_TempoCreate(player.channel, C.BASS_FX_FREESOURCE)
		setupFXChannel(player.channel)

		return player
	}

	second := C.CreateBassStream(C.CString(path), C.BASS_STREAM_DECODE|C.BASS_STREAM_PRESCAN)
	player.offscreenChannel = C.BASS_FX_TempoCreate(second, C.BASS_STREAM_DECODE|C.BASS_FX_FREESOURCE)

	setupFXChannel(player.offscreenChannel)

	return player
}

func setupFXChannel(channel C.HSTREAM) {
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_USE_QUICKALGO, 1)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_OVERLAP_MS, C.float(4.0))
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_SEQUENCE_MS, C.float(30.0))
}

func (wv *Track) Play() {
	wv.SetVolume(settings.Audio.GeneralVolume * settings.Audio.MusicVolume)

	wv.playing = true

	if !Offscreen {
		C.BASS_ChannelPlay(C.DWORD(wv.channel), 1)

		return
	}

	trackEvents = append(trackEvents, trackEvent{
		channel:  wv.offscreenChannel,
		time:     GlobalTimeMs,
		play:     true,
		delegate: nil,
	})
}

func (wv *Track) PlayV(volume float64) {
	wv.SetVolume(volume)

	wv.playing = true

	if !Offscreen {
		C.BASS_ChannelPlay(C.DWORD(wv.channel), 0)

		return
	}

	trackEvents = append(trackEvents, trackEvent{
		channel:  wv.offscreenChannel,
		time:     GlobalTimeMs,
		play:     true,
		delegate: nil,
	})
}

func (wv *Track) Pause() {
	wv.playing = false

	if !Offscreen {
		C.BASS_ChannelPause(C.DWORD(wv.channel))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelPause(C.DWORD(wv.offscreenChannel))
	})
}

func (wv *Track) Resume() {
	wv.playing = true

	if !Offscreen {
		C.BASS_ChannelPlay(C.DWORD(wv.channel), 0)

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelPlay(C.DWORD(wv.offscreenChannel), 0)
	})
}

func (wv *Track) Stop() {
	wv.playing = false

	if !Offscreen {
		C.BASS_ChannelStop(C.DWORD(wv.channel))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelStop(C.DWORD(wv.offscreenChannel))
	})
}

func (wv *Track) SetVolume(vol float64) {
	if !Offscreen {
		C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(vol))

		return
	}

	if math.Abs(wv.lastVol-vol) > 0.001 {
		addNormalEvent(func() {
			C.BASS_ChannelSetAttribute(C.DWORD(wv.offscreenChannel), C.BASS_ATTRIB_VOL, C.float(vol))
		})

		wv.lastVol = vol
	}
}

func (wv *Track) SetVolumeRelative(vol float64) {
	combined := settings.Audio.GeneralVolume * settings.Audio.MusicVolume * vol
	if !Offscreen {
		C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(combined))

		return
	}

	if math.Abs(wv.lastVol-combined) > 0.001 {
		addNormalEvent(func() {
			C.BASS_ChannelSetAttribute(C.DWORD(wv.offscreenChannel), C.BASS_ATTRIB_VOL, C.float(combined))
		})

		wv.lastVol = combined
	}
}

func (wv *Track) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetLength(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Track) SetPosition(pos float64) {
	if !Offscreen {
		C.BASS_ChannelSetPosition(wv.channel, C.BASS_ChannelSeconds2Bytes(wv.channel, C.double(pos)), C.BASS_POS_BYTE)

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelSetPosition(wv.offscreenChannel, C.BASS_ChannelSeconds2Bytes(wv.offscreenChannel, C.double(pos /*+tMs/1000*/)), C.BASS_POS_BYTE|C.BASS_POS_DECODETO)
	})
}

func (wv *Track) SetPositionF(pos float64) {
	C.BASS_ChannelSetPosition(wv.channel, C.BASS_ChannelSeconds2Bytes(wv.channel, C.double(pos)), C.BASS_POS_BYTE)
}

func (wv *Track) GetPosition() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetPosition(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Track) SetTempo(tempo float64) {
	if wv.speed == tempo {
		return
	}

	wv.speed = tempo

	if !Offscreen {
		C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelSetAttribute(C.DWORD(wv.offscreenChannel), C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))
	})
}

func (wv *Track) GetTempo() float64 {
	return wv.speed
}

func (wv *Track) SetPitch(tempo float64) {
	if wv.pitch == tempo {
		return
	}

	wv.pitch = tempo

	if !Offscreen {
		C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*14.4))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelSetAttribute(C.DWORD(wv.offscreenChannel), C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*14.4))
	})
}

func (wv *Track) GetPitch() float64 {
	return wv.pitch
}

func (wv *Track) GetState() int {
	return int(C.BASS_ChannelIsActive(wv.channel))
}

func (wv *Track) Update() {
	if wv.playing {
		C.BASS_ChannelGetData(wv.channel, unsafe.Pointer(&wv.fft[0]), C.BASS_DATA_FFT1024)
	} else {
		for i := range wv.fft {
			wv.fft[i] = 0
		}
	}

	toPeak := 0.0
	beatAv := 0.0

	for i, g := range wv.fft {
		h := math.Abs(float64(g))

		toPeak = math.Max(toPeak, h)

		if i > 0 && i < 5 {
			beatAv = math.Max(beatAv, float64(g))
		}
	}

	boost := 0.0

	for i := 0; i < 10; i++ {
		boost += float64(wv.fft[i]*wv.fft[i]) * float64(10-i) / float64(10)
	}

	wv.lowMax = beatAv
	wv.boost = boost
	wv.peak = toPeak

	level := int(C.BASS_ChannelGetLevel(wv.channel))

	left := level & 65535
	right := level >> 16

	wv.leftChannel = float64(left) / 32768
	wv.rightChannel = float64(right) / 32768
}

func (wv *Track) GetFFT() []float32 {
	return wv.fft
}

func (wv *Track) GetPeak() float64 {
	return wv.peak
}

func (wv *Track) GetLevelCombined() float64 {
	return (wv.leftChannel + wv.rightChannel) / 2
}

func (wv *Track) GetLeftLevel() float64 {
	return wv.leftChannel
}

func (wv *Track) GetRightLevel() float64 {
	return wv.rightChannel
}

func (wv *Track) GetBoost() float64 {
	return wv.boost
}

func (wv *Track) GetBeat() float64 {
	return wv.lowMax
}
