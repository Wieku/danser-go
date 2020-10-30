package bass

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo LDFLAGS: -Wl,-rpath,$ORIGIN -L${SRCDIR} -L${SRCDIR}/../../ -L/usr/lib -lbass -lbass_fx
#include "bass.h"
#include "bass_fx.h"
*/
import "C"

import (
	"fmt"
	"github.com/wieku/danser-go/app/settings"
	"log"
	"runtime"
)

func Init() {
	playbackBufferLength := 500
	deviceBufferLength := 10
	updatePeriod := 5

	if runtime.GOOS != "windows" {
		playbackBufferLength = int(settings.Audio.NonWindows.BassPlaybackBufferLength)
		deviceBufferLength = int(settings.Audio.NonWindows.BassDeviceBufferLength)
		updatePeriod = int(settings.Audio.NonWindows.BassUpdatePeriod)
	}

	// Output data regardless if audio is playing
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_NONSTOP, C.DWORD(1))

	// Worse time resolution but lower latency with DirectSound
	C.BASS_SetConfig(C.BASS_CONFIG_VISTA_TRUEPOS, C.DWORD(0))

	// Smaller buffer length, reduces latency
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_BUFFER, C.DWORD(deviceBufferLength))

	if runtime.GOOS != "windows" {
		C.BASS_SetConfig(C.BASS_CONFIG_BUFFER, C.DWORD(playbackBufferLength))
	}

	// Update BASS more frequently
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_PERIOD, C.DWORD(updatePeriod))

	// BASS_CONFIG_MP3_OLDGAPS
	C.BASS_SetConfig(C.DWORD(68), C.DWORD(1))

	if C.BASS_Init(C.int(-1), C.DWORD(44100), C.DWORD(0), nil, nil) != 0 {
		log.Println("BASS Initialized!")
		log.Println("BASS Version:", parseVersion(int(C.BASS_GetVersion())))
		log.Println("BASS FX Version:", parseVersion(int(C.BASS_FX_GetVersion())))
	} else {
		panic(fmt.Sprintf("Failed to run BASS, error: %d", int(C.BASS_ErrorGetCode())))
	}
}

func parseVersion(version int) string {
	main := version >> 24 & 0xFF
	revision0 := version >> 16 & 0xFF
	revision1 := version >> 8 & 0xFF
	revision2 := version & 0xFF

	return fmt.Sprintf("%d.%d.%d.%d", main, revision0, revision1, revision2)
}
