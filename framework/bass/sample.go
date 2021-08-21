package bass

/*
#include "bass.h"
#include "bassmix.h"
*/
import "C"

import (
	"github.com/wieku/danser-go/app/settings"
	"io/ioutil"
	"os"
	"unsafe"
)

type SubSample struct {
	source     C.HSAMPLE
	sampleChan C.HSTREAM
	streamChan C.HSTREAM
}

type Sample struct {
	bassSample C.DWORD
}

var loopingStreams = make(map[*SubSample]int)

func StopLoops() {
	for k := range loopingStreams {
		StopSample(k)
	}
}

func NewSample(path string) *Sample {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil
	}

	return NewSampleData(data)
}

func NewSampleData(data []byte) *Sample {
	if len(data) == 0 {
		return nil
	}

	sample := new(Sample)
	sample.bassSample = C.BASS_SampleLoad(1, unsafe.Pointer(&data[0]), 0, C.DWORD(len(data)), 32, C.BASS_SAMPLE_OVER_POS)

	return sample
}

func (sample *Sample) Play() *SubSample {
	sub := &SubSample{source: sample.bassSample}

	if sample.bassSample == 0 {
		return sub
	}

	sub.sampleChan = C.BASS_SampleGetChannel(sample.bassSample, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if sub.sampleChan != 0 {
		C.BASS_ChannelSetAttribute(sub.sampleChan, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))

		C.BASS_Mixer_StreamAddChannel(masterMixer, sub.sampleChan, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return sub
}

func (sample *Sample) PlayLoop() *SubSample {
	sSample := sample.Play()

	setLoop(sSample)

	return sSample
}

func (sample *Sample) PlayV(volume float64) *SubSample {
	sub := &SubSample{source: sample.bassSample}

	if sample.bassSample == 0 {
		return sub
	}

	sub.sampleChan = C.BASS_SampleGetChannel(sample.bassSample, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if sub.sampleChan != 0 {
		C.BASS_ChannelSetAttribute(sub.sampleChan, C.BASS_ATTRIB_VOL, C.float(volume))

		C.BASS_Mixer_StreamAddChannel(masterMixer, sub.sampleChan, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return sub
}

func (sample *Sample) PlayVLoop(volume float64) *SubSample {
	sSample := sample.PlayV(volume)

	setLoop(sSample)

	return sSample
}

func (sample *Sample) PlayRV(volume float64) *SubSample {
	sub := &SubSample{source: sample.bassSample}

	if sample.bassSample == 0 {
		return sub
	}

	sub.sampleChan = C.BASS_SampleGetChannel(sample.bassSample, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if sub.sampleChan != 0 {
		C.BASS_ChannelSetAttribute(sub.sampleChan, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))

		C.BASS_Mixer_StreamAddChannel(masterMixer, sub.sampleChan, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return sub
}

func (sample *Sample) PlayRVLoop(volume float64) *SubSample {
	sSample := sample.PlayRV(volume)

	setLoop(sSample)

	return sSample
}

func (sample *Sample) PlayRVPos(volume float64, balance float64) *SubSample {
	sub := &SubSample{source: sample.bassSample}

	if sample.bassSample == 0 {
		return sub
	}

	sub.sampleChan = C.BASS_SampleGetChannel(sample.bassSample, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if sub.sampleChan != 0 {
		C.BASS_ChannelSetAttribute(sub.sampleChan, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
		C.BASS_ChannelSetAttribute(sub.sampleChan, C.BASS_ATTRIB_PAN, C.float(balance))

		C.BASS_Mixer_StreamAddChannel(masterMixer, sub.sampleChan, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return sub
}

func (sample *Sample) PlayRVPosLoop(volume float64, balance float64) *SubSample {
	sSample := sample.PlayRVPos(volume, balance)

	setLoop(sSample)

	return sSample
}

func setLoop(sSample *SubSample) {
	loopingStreams[sSample] = 1

	if sSample.source == 0 {
		return
	}

	if sSample.sampleChan != 0 {
		C.BASS_ChannelFlags(sSample.sampleChan, C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
	}
}

func SetRate(channel *SubSample, rate float64) {
	if channel.source == 0 {
		return
	}

	if channel.sampleChan != 0 {
		C.BASS_ChannelSetAttribute(channel.sampleChan, C.BASS_ATTRIB_FREQ, C.float(rate))
	}
}

func StopSample(channel *SubSample) {
	delete(loopingStreams, channel)

	if channel.source == 0 {
		return
	}

	if channel.sampleChan != 0 {
		C.BASS_Mixer_ChannelRemove(channel.sampleChan)

		C.BASS_ChannelFree(channel.sampleChan)
	}
}

func PauseSample(channel *SubSample) {
	if channel.source == 0 {
		return
	}

	if channel.sampleChan != 0 {
		C.BASS_Mixer_ChannelFlags(channel.sampleChan, C.BASS_MIXER_CHAN_PAUSE, C.BASS_MIXER_CHAN_PAUSE)
	}
}

func PlaySample(channel *SubSample) {
	if channel.source == 0 {
		return
	}

	if channel.sampleChan != 0 {
		C.BASS_Mixer_ChannelFlags(channel.sampleChan, 0, C.BASS_MIXER_CHAN_PAUSE)
	}
}
