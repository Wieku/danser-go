package audio

/*
#include "bass.h"

extern void musicCallback(DWORD);

static inline void SyncFunc(HSYNC handle, DWORD channel, DWORD data, void *user) {
  musicCallback(channel);
}

static inline void setSync(HCHANNEL channel) {
	BASS_ChannelSetSync(channel, BASS_SYNC_END | BASS_SYNC_MIXTIME, 0, SyncFunc, 0);
}
 */
import "C"
import "unsafe"

const (
	MUSIC_STOPPED = 0
	MUSIC_PLAYING = 1
	MUSIC_STALLED = 2
	MUSIC_PAUSED = 3
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
	channel C.DWORD
}

func NewMusic(path string) *Music {
	player := &Music{}
	ch := C.BASS_StreamCreateFile(0, unsafe.Pointer(C.CString(path)), 0, 0, C.BASS_ASYNCFILE | C.BASS_STREAM_AUTOFREE)
	player.channel = ch
	return player
}

func (wv *Music) Play() {
	wv.SetVolume(0.1)
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

func (wv *Music) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetLength(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Music) SetPosition(pos float64) {
	C.BASS_ChannelSetPosition(wv.channel, C.BASS_ChannelSeconds2Bytes(wv.channel, C.double(pos)), C.BASS_POS_BYTE)
}

func (wv *Music) GetPosition() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(wv.channel, C.BASS_ChannelGetPosition(wv.channel, C.BASS_POS_BYTE)))
}

func (wv *Music) GetState() int {
	 return int(C.BASS_ChannelIsActive(wv.channel))
}
