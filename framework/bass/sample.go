package bass

/*
#include "bass_util.hpp"
#include "bass.h"
*/
import "C"

import (
	"github.com/wieku/danser-go/app/settings"
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
	player.channel = C.LoadBassSample(C.CString(path), 32, C.BASS_SAMPLE_OVER_POS)

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

func (wv *Sample) PlayRVPos(volume float64, balance float64) {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_PAN, C.float(balance))
	C.BASS_ChannelPlay(channel, 1)
}
