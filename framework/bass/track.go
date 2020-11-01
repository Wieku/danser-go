package bass

/*
#include "bass_util.h"
#include "bass.h"
#include "bass_fx.h"
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
	channel      C.DWORD
	fft          []float32
	boost        float64
	peak         float64
	leftChannel  float64
	rightChannel float64
	lowMax       float64
}

func NewTrack(path string) *Track {
	player := new(Track)

	channel := C.CreateBassStream(C.CString(path), C.BASS_ASYNCFILE|C.BASS_STREAM_DECODE|C.BASS_STREAM_PRESCAN)

	player.channel = C.BASS_FX_TempoCreate(channel, C.BASS_FX_FREESOURCE)
	player.fft = make([]float32, 512)
	return player
}

func (wv *Track) Play() {
	wv.SetVolume(settings.Audio.GeneralVolume * settings.Audio.MusicVolume)
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 1)
}

func (wv *Track) PlayV(volume float64) {
	wv.SetVolume(volume)
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 0)
}

func (wv *Track) Pause() {
	C.BASS_ChannelPause(C.DWORD(wv.channel))
}

func (wv *Track) Resume() {
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 0)
}

func (wv *Track) Stop() {
	C.BASS_ChannelStop(C.DWORD(wv.channel))
}

func (wv *Track) SetVolume(vol float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(vol))
}

func (wv *Track) SetVolumeRelative(vol float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.MusicVolume*vol))
}

func (wv *Track) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetLength(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Track) SetPosition(pos float64) {
	C.BASS_ChannelSetPosition(wv.channel, C.BASS_ChannelSeconds2Bytes(wv.channel, C.double(pos)), C.BASS_POS_BYTE)
}

func (wv *Track) GetPosition() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetPosition(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Track) SetTempo(tempo float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))
}

func (wv *Track) SetPitch(tempo float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*12))
}

func (wv *Track) GetState() int {
	return int(C.BASS_ChannelIsActive(wv.channel))
}

func (wv *Track) Update() {
	C.BASS_ChannelGetData(wv.channel, unsafe.Pointer(&wv.fft[0]), C.BASS_DATA_FFT1024)

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
