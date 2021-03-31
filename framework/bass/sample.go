package bass

/*
#include "bass_util.h"
#include "bass.h"
*/
import "C"

import (
	"github.com/wieku/danser-go/app/settings"
	"io/ioutil"
	"os"
	"unsafe"
)

type SubSample struct {
	bassSample C.DWORD
	sampleChan C.HCHANNEL
	streamChan C.HSTREAM
}

type Sample struct {
	bassSample C.DWORD
	data       []byte
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

	sample := new(Sample)

	sample.data, err = ioutil.ReadAll(f)
	if err != nil {
		return nil
	}

	sample.bassSample = C.LoadBassSample(C.CString(path), 32, C.BASS_SAMPLE_OVER_POS)

	return sample
}

func NewSampleData(data []byte) *Sample {
	if len(data) == 0 {
		return nil
	}

	sample := new(Sample)
	sample.data = data
	sample.bassSample = C.BASS_SampleLoad(1, unsafe.Pointer(&data[0]), 0, C.DWORD(len(data)), 32, C.BASS_SAMPLE_OVER_POS)

	return sample
}

func (sample *Sample) Play() *SubSample {
	sub := new(SubSample)
	sub.bassSample = sample.bassSample

	if sample.bassSample == 0 {
		return sub
	}

	if !Offscreen {
		channel := C.BASS_SampleGetChannel(C.DWORD(sample.bassSample), 0)
		C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))
		C.BASS_ChannelPlay(channel, 1)

		sub.sampleChan = channel

		return sub
	}

	sample.createPlayEvent(sub, func() {
		C.BASS_ChannelSetAttribute(sub.streamChan, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume))
	})

	return sub
}

func (sample *Sample) PlayLoop() *SubSample {
	sSample := sample.Play()

	setLoop(sSample)

	return sSample
}

func (sample *Sample) PlayV(volume float64) *SubSample {
	sub := new(SubSample)
	sub.bassSample = sample.bassSample

	if sample.bassSample == 0 {
		return sub
	}

	if !Offscreen {
		channel := C.BASS_SampleGetChannel(C.DWORD(sample.bassSample), 0)
		C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(volume))
		C.BASS_ChannelPlay(channel, 1)

		sub.sampleChan = channel

		return sub
	}

	sample.createPlayEvent(sub, func() {
		C.BASS_ChannelSetAttribute(sub.streamChan, C.BASS_ATTRIB_VOL, C.float(volume))
	})

	return sub
}

func (sample *Sample) PlayVLoop(volume float64) *SubSample {
	sSample := sample.PlayV(volume)

	setLoop(sSample)

	return sSample
}

func (sample *Sample) PlayRV(volume float64) *SubSample {
	sub := new(SubSample)
	sub.bassSample = sample.bassSample

	if sample.bassSample == 0 {
		return sub
	}

	if !Offscreen {
		channel := C.BASS_SampleGetChannel(C.DWORD(sample.bassSample), 0)
		C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
		C.BASS_ChannelPlay(channel, 1)

		sub.sampleChan = channel

		return sub
	}

	sample.createPlayEvent(sub, func() {
		C.BASS_ChannelSetAttribute(sub.streamChan, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
	})

	return sub
}

func (sample *Sample) PlayRVLoop(volume float64) *SubSample {
	sSample := sample.PlayRV(volume)

	setLoop(sSample)

	return sSample
}

func (sample *Sample) PlayRVPos(volume float64, balance float64) *SubSample {
	sub := new(SubSample)
	sub.bassSample = sample.bassSample

	if sample.bassSample == 0 {
		return sub
	}

	if !Offscreen {
		channel := C.BASS_SampleGetChannel(C.DWORD(sample.bassSample), 0)
		C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
		C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_PAN, C.float(balance))
		C.BASS_ChannelPlay(channel, 1)

		sub.sampleChan = channel

		return sub
	}

	sample.createPlayEvent(sub, func() {
		C.BASS_ChannelSetAttribute(sub.streamChan, C.BASS_ATTRIB_VOL, C.float(settings.Audio.GeneralVolume*settings.Audio.SampleVolume*volume))
		C.BASS_ChannelSetAttribute(sub.streamChan, C.BASS_ATTRIB_PAN, C.float(balance))
	})

	return sub
}

func (sample *Sample) PlayRVPosLoop(volume float64, balance float64) *SubSample {
	sSample := sample.PlayRVPos(volume, balance)

	setLoop(sSample)

	return sSample
}

func (sample *Sample) createPlayEvent(sSample *SubSample, delegate func()) {
	trackEvents = append(trackEvents, trackEvent{
		channel: 0,
		time:    GlobalTimeMs,
		play:    true,
		delegate: func() C.DWORD {
			sSample.streamChan = C.BASS_StreamCreateFile(1, unsafe.Pointer(&sample.data[0]), C.QWORD(0), C.QWORD(len(sample.data)), C.BASS_STREAM_DECODE)

			if sSample.streamChan != 0 {
				delegate()
			}

			return sSample.streamChan
		},
	})
}

func setLoop(sSample *SubSample) {
	loopingStreams[sSample] = 1

	if sSample.bassSample == 0 {
		return
	}

	if !Offscreen {
		C.BASS_ChannelFlags(C.HCHANNEL(sSample.sampleChan), C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)

		return
	}

	addNormalEvent(func() {
		if sSample.streamChan != 0 {
			C.BASS_ChannelFlags(sSample.streamChan, C.BASS_SAMPLE_LOOP, C.BASS_SAMPLE_LOOP)
		}
	})
}

func SetRate(channel *SubSample, rate float64) {
	if channel.bassSample == 0 {
		return
	}

	if !Offscreen {
		C.BASS_ChannelSetAttribute(C.HCHANNEL(channel.sampleChan), C.BASS_ATTRIB_FREQ, C.float(rate))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelSetAttribute(C.HCHANNEL(channel.streamChan), C.BASS_ATTRIB_FREQ, C.float(rate))
	})
}

func StopSample(channel *SubSample) {
	delete(loopingStreams, channel)

	if channel.bassSample == 0 {
		return
	}

	if !Offscreen {
		C.BASS_ChannelPause(C.HCHANNEL(channel.sampleChan))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelStop(C.HCHANNEL(channel.streamChan))
	})
}

func PauseSample(channel *SubSample) {
	if channel.bassSample == 0 {
		return
	}

	if !Offscreen {
		C.BASS_ChannelPause(C.HCHANNEL(channel.sampleChan))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelPause(C.HCHANNEL(channel.streamChan))
	})
}

func PlaySample(channel *SubSample) {
	if channel.bassSample == 0 {
		return
	}

	if !Offscreen {
		C.BASS_ChannelPlay(C.HCHANNEL(channel.sampleChan), 0)

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelPlay(C.HCHANNEL(channel.streamChan), 0)
	})
}
