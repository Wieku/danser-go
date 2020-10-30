package bass

/*
#include "bass_util.h"
#include "bass.h"
*/
import "C"

import (
	"github.com/wieku/danser-go/app/settings"
	"os"
	"unsafe"
)

type SubSample C.HCHANNEL

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

func NewSampleData(data []byte) *Sample {
	player := &Sample{}
	player.channel = C.BASS_SampleLoad(1, unsafe.Pointer(&data[0]), 0, C.DWORD(len(data)), 32, C.BASS_SAMPLE_OVER_POS)

	return player
}

func (wv *Sample) Play() SubSample {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))
	C.BASS_ChannelPlay(channel, 1)

	return SubSample(channel)
}

func (wv *Sample) PlayLoop() SubSample {
	sSample := wv.Play()
	C.BASS_ChannelFlags(C.HCHANNEL(sSample), C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
	return sSample
}

func (wv *Sample) PlayV(volume float64) SubSample {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(volume))
	C.BASS_ChannelPlay(channel, 1)

	return SubSample(channel)
}

func (wv *Sample) PlayVLoop(volume float64) SubSample {
	sSample := wv.PlayV(volume)
	C.BASS_ChannelFlags(C.HCHANNEL(sSample), C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
	return sSample
}

func (wv *Sample) PlayRV(volume float64) SubSample {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
	C.BASS_ChannelPlay(channel, 1)

	return SubSample(channel)
}

func (wv *Sample) PlayRVLoop(volume float64) SubSample {
	sSample := wv.PlayRV(volume)
	C.BASS_ChannelFlags(C.HCHANNEL(sSample), C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
	return sSample
}

func (wv *Sample) PlayRVPos(volume float64, balance float64) SubSample {
	channel := C.BASS_SampleGetChannel(C.DWORD(wv.channel), 0)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_PAN, C.float(balance))
	C.BASS_ChannelPlay(channel, 1)

	return SubSample(channel)
}

func (wv *Sample) PlayRVPosLoop(volume float64, balance float64) SubSample {
	sSample := wv.PlayRVPos(volume, balance)
	C.BASS_ChannelFlags(C.HCHANNEL(sSample), C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
	return sSample
}

func SetRate(channel SubSample, rate float64) {
	C.BASS_ChannelSetAttribute(C.HCHANNEL(channel), C.BASS_ATTRIB_FREQ, C.float(rate))
}

func StopSample(channel SubSample) {
	C.BASS_ChannelStop(C.HCHANNEL(channel))
}

func PauseSample(channel SubSample) {
	C.BASS_ChannelPause(C.HCHANNEL(channel))
}

func PlaySample(channel SubSample) {
	C.BASS_ChannelPlay(C.HCHANNEL(channel), 0)
}
