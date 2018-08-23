package audio
/*
#include "bass.h"
 */
import "C"

import (
	"unsafe"
	"github.com/wieku/danser/settings"
	"os"
)

type Sample struct {
	channel C.DWORD
}

func NewSample(path string) *Sample {
	f, err := os.Open(path)

	if os.IsNotExist(err) {
		return nil
	}
	f.Close()

	player := &Sample{}
	han := C.BASS_SampleLoad(0, unsafe.Pointer(C.CString(path)), 0, 0, 10, 0)
	ch1 := C.BASS_SampleGetChannel(han, 0)
	player.channel = ch1
	return player
}

func (wv *Sample) Play() {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 1)
}

func (wv *Sample) PlayV(volume float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(volume))
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 1)
}

func (wv *Sample) PlayRV(volume float64) {
	C.BASS_ChannelSetAttribute(C.DWORD(wv.channel), C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
	C.BASS_ChannelPlay(C.DWORD(wv.channel), 1)
}