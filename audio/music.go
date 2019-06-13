package audio

/*
#include "bass.h"
#include "bass_fx.h"

extern void musicCallback(DWORD);

static inline void SyncFunc(HSYNC handle, DWORD channel, DWORD data, void *user) {
  musicCallback(channel);
}

static inline void setSync(HCHANNEL channel) {
	BASS_ChannelSetSync(channel, BASS_SYNC_END | BASS_SYNC_MIXTIME, 0, SyncFunc, 0);
}
 */
import "C"
import (
	"unsafe"
	"github.com/wieku/danser-go/settings"
	//"log"
	"math"
)

const (
	MUSIC_STOPPED = 0
	MUSIC_PLAYING = 1
	MUSIC_STALLED = 2
	MUSIC_PAUSED  = 3
)

type Callback func()

var callbacks = make(map[C.DWORD][]Callback)

//export musicCallback
func musicCallback(channel C.DWORD) {
	for _, f := range callbacks[channel] {
		f()
	}
}

func registerEndCallback(channel C.DWORD, f func()) {
	arr := callbacks[channel]
	arr = append(arr, Callback(f))
	callbacks[channel] = arr
	C.setSync(channel)
}

func unregisterEndCallback(channel C.DWORD, f func()) {
	arr := callbacks[channel]
	var cb Callback = f
	for i, v := range arr {
		if &v == &cb {
			arr = append(arr[:i], arr[i+1:]...)
			break
		}
	}
	callbacks[channel] = arr
}

type Music struct {
	channel      C.DWORD
	fft          []float32
	beat         float64
	peak         float64
	leftChannel  float64
	rightChannel float64
}

func NewMusic(path string) *Music {
	player := new(Music)
	channel := C.BASS_StreamCreateFile(0, unsafe.Pointer(C.CString(path)), 0, 0, C.BASS_ASYNCFILE|C.BASS_STREAM_DECODE)
	player.channel = C.BASS_FX_TempoCreate(channel, C.BASS_FX_FREESOURCE)
	player.fft = make([]float32, 512)
	return player
}

func (wv *Music) Play() {
	wv.SetVolume(settings.Audio.GeneralVolume * settings.Audio.MusicVolume)
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 1)
}

func (wv *Music) PlayV(volume float64) {
	wv.SetVolume(volume)
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 0)
}

func (wv *Music) RegisterCallback(f func()) {
	registerEndCallback(wv.channel, f)
}

func (wv *Music) UnregisterCallback(f func()) {
	unregisterEndCallback(wv.channel, f)
}

func (wv *Music) Pause() {
	C.BASS_ChannelPause(C.DWORD(wv.channel))
}

func (wv *Music) Resume() {
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 0)
}

func (wv *Music) Stop() {
	C.BASS_ChannelStop(C.DWORD(wv.channel))
}

func (wv *Music) SetVolume(vol float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(vol))
}

func (wv *Music) SetVolumeRelative(vol float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.MusicVolume*vol))
}

func (wv *Music) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetLength(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Music) SetPosition(pos float64) {
	C.BASS_ChannelSetPosition(wv.channel, C.BASS_ChannelSeconds2Bytes(wv.channel, C.double(pos)), C.BASS_POS_BYTE)
}

func (wv *Music) GetPosition() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetPosition(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Music) SetTempo(tempo float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))
}

func (wv *Music) SetPitch(tempo float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*12))
}

func (wv *Music) GetState() int {
	return int(C.BASS_ChannelIsActive(wv.channel))
}

func (wv *Music) Update() {
	C.BASS_ChannelGetData(wv.channel, unsafe.Pointer(&wv.fft[0]), C.BASS_DATA_FFT1024)
	toPeak := -1.0
	beatAv := 0.0
	for i, g := range wv.fft {
		h := math.Abs(float64(g))
		if toPeak < h {
			toPeak = h
		}
		if i > 0 && i < 5 {
			beatAv = math.Max(beatAv, float64(g))
		}
		//toAv += math.Abs(float64(g))
	}
	//beatAv /= 5.0
	//toAv /= 512
	wv.beat = beatAv
	wv.peak = toPeak

	level := int(C.BASS_ChannelGetLevel(wv.channel))
	left := int(level & 65535)
	right := int(level >> 16)

	wv.leftChannel = float64(left) / 32768
	wv.rightChannel = float64(right) / 32768
}

func (wv *Music) GetFFT() []float32 {
	return wv.fft
}

func (wv *Music) GetPeak() float64 {
	return wv.peak
}

func (wv *Music) GetLevelCombined() float64 {
	return (wv.leftChannel + wv.rightChannel) / 2
}

func (wv *Music) GetLeftLevel() float64 {
	return wv.leftChannel
}

func (wv *Music) GetRightLevel() float64 {
	return wv.rightChannel
}

func (wv *Music) GetBeat() float64 {
	return wv.beat
}
