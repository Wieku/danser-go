package audio

/*
#include "bass.h"
 */
import "C"

import (
	"unsafe"
	"github.com/wieku/danser-go/settings"
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
	han := C.BASS_SampleLoad(0, unsafe.Pointer(C.CString(path)), 0, 0, 32, C.BASS_SAMPLE_OVER_POS)
	player.channel = han
	return player
}

func (wv *Sample) Play() {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))
	C.BASS_ChannelPlay(channel, 1)
}

func (wv *Sample) PlayV(volume float64) {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(volume))
	C.BASS_ChannelPlay(channel, 1)
}

func (wv *Sample) PlayRV(volume float64) {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
	C.BASS_ChannelPlay(channel, 1)
}
