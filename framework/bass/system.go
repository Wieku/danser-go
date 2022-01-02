package bass

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo LDFLAGS: -Wl,-rpath,$ORIGIN -L${SRCDIR} -L${SRCDIR}/../../ -L/usr/lib/danser -L/usr/lib -lbass -lbass_fx -lbassmix
#include "bass.h"
#include "bass_fx.h"
#include "bassmix.h"
*/
import "C"

import (
	"fmt"
	"github.com/wieku/danser-go/app/settings"
	"log"
	"runtime"
)

var masterMixer C.HSTREAM

var sampleRate = 44100

func Init(offscreen bool) {
	log.Println("Initializing BASS...")

	playbackBufferLength := 100
	deviceBufferLength := 10
	updatePeriod := 5
	devUpdatePeriod := 10

	if runtime.GOOS != "windows" {
		playbackBufferLength = int(settings.Audio.NonWindows.BassPlaybackBufferLength)
		deviceBufferLength = int(settings.Audio.NonWindows.BassDeviceBufferLength)
		updatePeriod = int(settings.Audio.NonWindows.BassUpdatePeriod)
		devUpdatePeriod = int(settings.Audio.NonWindows.BassDeviceUpdatePeriod)
	}

	// Output data regardless if audio is playing
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_NONSTOP, C.DWORD(1))

	// Worse time resolution but lower latency with DirectSound
	C.BASS_SetConfig(C.BASS_CONFIG_VISTA_TRUEPOS, C.DWORD(0))

	// Smaller stream buffer length, reduces latency
	C.BASS_SetConfig(C.BASS_CONFIG_BUFFER, C.DWORD(playbackBufferLength))

	// Update BASS stream buffer more frequently
	C.BASS_SetConfig(C.BASS_CONFIG_UPDATEPERIOD, C.DWORD(updatePeriod))

	// Smaller device buffer length, reduces latency
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_BUFFER, C.DWORD(deviceBufferLength))

	// Update BASS device buffer more frequently
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_PERIOD, C.DWORD(devUpdatePeriod))

	// BASS_CONFIG_MP3_OLDGAPS
	C.BASS_SetConfig(C.DWORD(68), C.DWORD(1))

	deviceId := -1 //default audio device
	mixerFlags := C.BASS_MIXER_NONSTOP

	if offscreen {
		sampleRate = 48000
		deviceId = 0 //If we're rendering, we don't want BASS to be tied to specific device, especially in headless system
		mixerFlags |= C.BASS_SAMPLE_FLOAT | C.BASS_STREAM_DECODE
	}

	if C.BASS_Init(C.int(deviceId), C.DWORD(sampleRate), C.DWORD(0), nil, nil) != 0 {
		log.Println("BASS Initialized!")
		log.Println("BASS Version:       ", parseVersion(int(C.BASS_GetVersion())))
		log.Println("BASS FX Version:    ", parseVersion(int(C.BASS_FX_GetVersion())))
		log.Println("BASS Mix Version:   ", parseVersion(int(C.BASS_Mixer_GetVersion())))

		// We're not interested in BASSEnc in onscreen mode, show audio device instead
		if !offscreen {
			log.Println("BASS Audio Device:  ", getDeviceName())
			log.Println("BASS Audio Latency: ", fmt.Sprintf("%dms", getLatency()))
		}

		masterMixer = C.BASS_Mixer_StreamCreate(C.DWORD(sampleRate), 2, C.DWORD(mixerFlags))
		C.BASS_ChannelSetAttribute(masterMixer, C.BASS_ATTRIB_BUFFER, 0)
		C.BASS_ChannelSetDevice(masterMixer, C.BASS_GetDevice())

		if !offscreen {
			C.BASS_ChannelPlay(masterMixer, 0)
		}
	} else {
		err := GetError()
		panic(fmt.Sprintf("Failed to run BASS, error id: %d, message: %s", err, err.Message()))
	}
}

func parseVersion(version int) string {
	main := version >> 24 & 0xFF
	revision0 := version >> 16 & 0xFF
	revision1 := version >> 8 & 0xFF
	revision2 := version & 0xFF

	return fmt.Sprintf("%d.%d.%d.%d", main, revision0, revision1, revision2)
}

func getDeviceName() string {
	var info C.BASS_DEVICEINFO

	C.BASS_GetDeviceInfo(C.BASS_GetDevice(), &info)

	return C.GoString(info.name)
}

func getLatency() int {
	var info C.BASS_INFO

	C.BASS_GetInfo(&info)

	return int(info.latency)
}
