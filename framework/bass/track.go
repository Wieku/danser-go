package bass

/*
#include "bass_util.h"
#include "bass.h"
#include "bass_fx.h"
#include "bassmix.h"
*/
import "C"
import (
	"github.com/wieku/danser-go/app/settings"
	"math"
	"unsafe"
)

const (
	MUSIC_STOPPED = 0
	MUSIC_PLAYING = 1
	MUSIC_STALLED = 2
	MUSIC_PAUSED  = 3
)

type Track struct {
	channel          C.HSTREAM
	offscreenChannel C.HSTREAM
	fft              []float32
	boost            float64
	peak             float64
	leftChannel      float64
	rightChannel     float64
	lowMax           float64
	lastVol          float64
	speed            float64
	pitch            float64
	playing          bool
	addedToMixer     bool
}

func NewTrack(path string) *Track {
	player := new(Track)
	player.speed = 1
	player.pitch = 1
	player.lastVol = -100000
	player.fft = make([]float32, 512)

	normalChannelFlags := C.BASS_STREAM_DECODE | C.BASS_STREAM_PRESCAN
	if !Offscreen {
		normalChannelFlags |= C.BASS_ASYNCFILE
	}

	player.channel = C.CreateBassStream(C.CString(path), C.DWORD(normalChannelFlags))

	if !Offscreen {
		player.channel = C.BASS_FX_TempoCreate(player.channel, C.BASS_FX_FREESOURCE | C.BASS_STREAM_DECODE)
		setupFXChannel(player.channel)

		return player
	}

	offscreenChannel := C.CreateBassStream(C.CString(path), C.BASS_STREAM_DECODE|C.BASS_STREAM_PRESCAN)
	player.offscreenChannel = C.BASS_FX_TempoCreate(offscreenChannel, C.BASS_STREAM_DECODE|C.BASS_FX_FREESOURCE)

	setupFXChannel(player.offscreenChannel)

	return player
}

func setupFXChannel(channel C.HSTREAM) {
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_USE_QUICKALGO, 1)
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_OVERLAP_MS, C.float(4.0))
	C.BASS_ChannelSetAttribute(channel, C.BASS_ATTRIB_TEMPO_OPTION_SEQUENCE_MS, C.float(30.0))
}

func (track *Track) AddSilence(seconds float64) {
	C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_TAIL, C.float(seconds))

	if track.offscreenChannel != 0 {
		C.BASS_ChannelSetAttribute(track.offscreenChannel, C.BASS_ATTRIB_TAIL, C.float(seconds))
	}
}

func (track *Track) Play() {
	track.SetVolume(settings.Audio.GeneralVolume * settings.Audio.MusicVolume)

	track.playing = true

	if !Offscreen {
		C.BASS_Mixer_StreamAddChannel(masterMixer, track.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_MIXER_CHAN_BUFFER)
		track.addedToMixer = true

		return
	}

	trackEvents = append(trackEvents, trackEvent{
		channel:  track.offscreenChannel,
		time:     GlobalTimeMs,
		play:     true,
		delegate: nil,
	})
}

func (track *Track) PlayV(volume float64) {
	track.SetVolume(volume)

	track.playing = true

	if !Offscreen {
		C.BASS_Mixer_StreamAddChannel(masterMixer, track.channel, C.BASS_MIXER_CHAN_NORAMPIN|C.BASS_MIXER_CHAN_BUFFER)
		track.addedToMixer = true

		return
	}

	trackEvents = append(trackEvents, trackEvent{
		channel:  track.offscreenChannel,
		time:     GlobalTimeMs,
		play:     true,
		delegate: nil,
	})
}

func (track *Track) Pause() {
	track.playing = false

	if !Offscreen {
		C.BASS_Mixer_ChannelFlags(track.channel, 0, C.BASS_MIXER_CHAN_PAUSE)

		return
	}

	addNormalEvent(func() {
		C.BASS_Mixer_ChannelFlags(track.offscreenChannel, 0, C.BASS_MIXER_CHAN_PAUSE)
	})
}

func (track *Track) Resume() {
	track.playing = true

	if !Offscreen {
		C.BASS_Mixer_ChannelFlags(track.channel, C.BASS_MIXER_CHAN_PAUSE, C.BASS_MIXER_CHAN_PAUSE)

		return
	}

	addNormalEvent(func() {
		C.BASS_Mixer_ChannelFlags(track.offscreenChannel, C.BASS_MIXER_CHAN_PAUSE, C.BASS_MIXER_CHAN_PAUSE)
	})
}

func (track *Track) Stop() {
	track.playing = false

	if !Offscreen {
		C.BASS_ChannelStop(track.channel)

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelStop(track.offscreenChannel)
	})
}

func (track *Track) SetVolume(vol float64) {
	if !Offscreen {
		C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_VOL, C.float(vol))

		return
	}

	if math.Abs(track.lastVol-vol) > 0.001 {
		addNormalEvent(func() {
			C.BASS_ChannelSetAttribute(track.offscreenChannel, C.BASS_ATTRIB_VOL, C.float(vol))
		})

		track.lastVol = vol
	}
}

func (track *Track) SetVolumeRelative(vol float64) {
	combined := settings.Audio.GeneralVolume * settings.Audio.MusicVolume * vol

	if !Offscreen {
		C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_VOL, C.float(combined))

		return
	}

	if math.Abs(track.lastVol-combined) > 0.001 {
		addNormalEvent(func() {
			C.BASS_ChannelSetAttribute(track.offscreenChannel, C.BASS_ATTRIB_VOL, C.float(combined))
		})

		track.lastVol = combined
	}
}

func (track *Track) GetLength() float64 {
	return float64(C.BASS_ChannelBytes2Seconds(track.channel, C.BASS_ChannelGetLength(track.channel, C.BASS_POS_BYTE)))
}

func (track *Track) SetPosition(pos float64) {
	if !Offscreen {
		if track.addedToMixer {
			C.BASS_Mixer_ChannelSetPosition(track.channel, C.BASS_ChannelSeconds2Bytes(track.channel, C.double(pos)), C.BASS_POS_BYTE)
		} else {
			C.BASS_ChannelSetPosition(track.channel, C.BASS_ChannelSeconds2Bytes(track.channel, C.double(pos)), C.BASS_POS_BYTE)
		}

		return
	}

	addNormalEvent(func() {
		C.BASS_Mixer_ChannelSetPosition(track.offscreenChannel, C.BASS_ChannelSeconds2Bytes(track.offscreenChannel, C.double(pos)), C.BASS_POS_BYTE|C.BASS_POS_DECODETO)
	})
}

func (track *Track) SetPositionF(pos float64) {
	C.BASS_ChannelSetPosition(track.channel, C.BASS_ChannelSeconds2Bytes(track.channel, C.double(pos)), C.BASS_POS_BYTE)
}

func (track *Track) GetPosition() float64 {
	var bassPos float64

	if track.addedToMixer {
		bassPos = float64(C.BASS_ChannelBytes2Seconds(track.channel, C.BASS_Mixer_ChannelGetPosition(track.channel, C.BASS_POS_BYTE)))
	} else {
		bassPos = float64(C.BASS_ChannelBytes2Seconds(track.channel, C.BASS_ChannelGetPosition(track.channel, C.BASS_POS_BYTE)))
	}

	return bassPos
}

func (track *Track) SetTempo(tempo float64) {
	if track.speed == tempo {
		return
	}

	track.speed = tempo

	if !Offscreen {
		C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelSetAttribute(track.offscreenChannel, C.BASS_ATTRIB_TEMPO, C.float((tempo-1.0)*100))
	})
}

func (track *Track) GetTempo() float64 {
	return track.speed
}

func (track *Track) SetPitch(tempo float64) {
	if track.pitch == tempo {
		return
	}

	track.pitch = tempo

	if !Offscreen {
		C.BASS_ChannelSetAttribute(track.channel, C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*14.4))

		return
	}

	addNormalEvent(func() {
		C.BASS_ChannelSetAttribute(track.offscreenChannel, C.BASS_ATTRIB_TEMPO_PITCH, C.float((tempo-1.0)*14.4))
	})
}

func (track *Track) GetPitch() float64 {
	return track.pitch
}

func (track *Track) GetState() int {
	if !track.addedToMixer {
		return MUSIC_STOPPED
	}

	state := int(C.BASS_ChannelIsActive(track.channel))

	if state == MUSIC_PLAYING && track.addedToMixer && C.BASS_Mixer_ChannelFlags(track.channel, 0, 0) & C.BASS_MIXER_CHAN_PAUSE > 0 {
		return MUSIC_PAUSED
	}

	return state
}

func (track *Track) Update() {
	if track.playing {
		if track.addedToMixer {
			C.BASS_Mixer_ChannelGetData(track.channel, unsafe.Pointer(&track.fft[0]), C.BASS_DATA_FFT1024)
		} else {
			C.BASS_ChannelGetData(track.channel, unsafe.Pointer(&track.fft[0]), C.BASS_DATA_FFT1024)
		}
	} else {
		for i := range track.fft {
			track.fft[i] = 0
		}
	}

	toPeak := 0.0
	beatAv := 0.0

	for i, g := range track.fft {
		h := math.Abs(float64(g))

		toPeak = math.Max(toPeak, h)

		if i > 0 && i < 5 {
			beatAv = math.Max(beatAv, float64(g))
		}
	}

	boost := 0.0

	for i := 0; i < 10; i++ {
		boost += float64(track.fft[i]*track.fft[i]) * float64(10-i) / float64(10)
	}

	track.lowMax = beatAv
	track.boost = boost
	track.peak = toPeak

	var level int

	if track.addedToMixer {
		level = int(C.BASS_Mixer_ChannelGetLevel(track.channel))
	} else {
		level = int(C.BASS_ChannelGetLevel(track.channel))
	}

	left := level & 65535
	right := level >> 16

	track.leftChannel = float64(left) / 32768
	track.rightChannel = float64(right) / 32768
}

func (track *Track) GetFFT() []float32 {
	return track.fft
}

func (track *Track) GetPeak() float64 {
	return track.peak
}

func (track *Track) GetLevelCombined() float64 {
	return (track.leftChannel + track.rightChannel) / 2
}

func (track *Track) GetLeftLevel() float64 {
	return track.leftChannel
}

func (track *Track) GetRightLevel() float64 {
	return track.rightChannel
}

func (track *Track) GetBoost() float64 {
	return track.boost
}

func (track *Track) GetBeat() float64 {
	return track.lowMax
}
