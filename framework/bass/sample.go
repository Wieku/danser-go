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

var emptyData = make([]byte, 1024)

type SampleChannel struct {
	source  C.HSAMPLE
	channel C.HSTREAM
}

type Sample struct {
	bassSample C.DWORD
}

var loopingStreams = make(map[*SampleChannel]int)

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
	sample := new(Sample)

	if len(data) < 1024 { // If we have useless data, create ~10ms empty sample, simpler solution than creating a flag and checking it later
		sample.bassSample = C.BASS_SampleCreate(1024, 44100, 2, 32, C.BASS_SAMPLE_OVER_POS)

		C.BASS_SampleSetData(sample.bassSample, unsafe.Pointer(&emptyData[0]))
	} else {
		sample.bassSample = C.BASS_SampleLoad(1, unsafe.Pointer(&data[0]), 0, C.DWORD(len(data)), 32, C.BASS_SAMPLE_OVER_POS)
	}

	return sample
}

func (sample *Sample) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(sample.bassSample, C.BASS_ChannelGetLength(sample.bassSample, C.BASS_POS_BYTE)))
}

func (sample *Sample) Play() *SampleChannel {
	channel := &SampleChannel{source: sample.bassSample}

	if channel.source == 0 {
		return channel
	}

	channel.channel = C.BASS_SampleGetChannel(channel.source, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if channel.channel != 0 {
		C.BASS_ChannelSetAttribute(channel.channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))

		C.BASS_Mixer_StreamAddChannel(masterMixer, channel.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return channel
}

func (sample *Sample) PlayLoop() *SampleChannel {
	channel := sample.Play()

	setLoop(channel)

	return channel
}

func (sample *Sample) PlayV(volume float64) *SampleChannel {
	channel := &SampleChannel{source: sample.bassSample}

	if channel.source == 0 {
		return channel
	}

	channel.channel = C.BASS_SampleGetChannel(channel.source, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if channel.channel != 0 {
		C.BASS_ChannelSetAttribute(channel.channel, C.BASS_ATTRIB_VOL, C.float(volume))

		C.BASS_Mixer_StreamAddChannel(masterMixer, channel.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return channel
}

func (sample *Sample) PlayVLoop(volume float64) *SampleChannel {
	channel := sample.PlayV(volume)

	setLoop(channel)

	return channel
}

func (sample *Sample) PlayRV(volume float64) *SampleChannel {
	channel := &SampleChannel{source: sample.bassSample}

	if channel.source == 0 {
		return channel
	}

	channel.channel = C.BASS_SampleGetChannel(channel.source, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if channel.channel != 0 {
		C.BASS_ChannelSetAttribute(channel.channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))

		C.BASS_Mixer_StreamAddChannel(masterMixer, channel.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return channel
}

func (sample *Sample) PlayRVLoop(volume float64) *SampleChannel {
	channel := sample.PlayRV(volume)

	setLoop(channel)

	return channel
}

func (sample *Sample) PlayRVPos(volume float64, balance float64) *SampleChannel {
	channel := &SampleChannel{source: sample.bassSample}

	if channel.source == 0 {
		return channel
	}

	channel.channel = C.BASS_SampleGetChannel(channel.source, C.BASS_SAMCHAN_STREAM|C.BASS_STREAM_DECODE)

	if channel.channel != 0 {
		C.BASS_ChannelSetAttribute(channel.channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
		C.BASS_ChannelSetAttribute(channel.channel, C.BASS_ATTRIB_PAN, C.float(balance))

		C.BASS_Mixer_StreamAddChannel(masterMixer, channel.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_STREAM_AUTOFREE)
	}

	return channel
}

func (sample *Sample) PlayRVPosLoop(volume float64, balance float64) *SampleChannel {
	channel := sample.PlayRVPos(volume, balance)

	setLoop(channel)

	return channel
}

func setLoop(channel *SampleChannel) {
	loopingStreams[channel] = 1

	if channel.channel != 0 {
		C.BASS_ChannelFlags(channel.channel, C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
	}
}

func SetRate(channel *SampleChannel, rate float64) {
	if channel.channel != 0 {
		C.BASS_ChannelSetAttribute(channel.channel, C.BASS_ATTRIB_FREQ, C.float(rate))
	}
}

func StopSample(channel *SampleChannel) {
	delete(loopingStreams, channel)

	if channel.channel != 0 {
		C.BASS_Mixer_ChannelRemove(channel.channel)

		C.BASS_ChannelFree(channel.channel)
	}
}

func PauseSample(channel *SampleChannel) {
	if channel.channel != 0 {
		C.BASS_Mixer_ChannelFlags(channel.channel, C.BASS_MIXER_CHAN_PAUSE, C.BASS_MIXER_CHAN_PAUSE)
	}
}

func PlaySample(channel *SampleChannel) {
	if channel.channel != 0 {
		C.BASS_Mixer_ChannelFlags(channel.channel, 0, C.BASS_MIXER_CHAN_PAUSE)
	}
}
