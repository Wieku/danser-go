package audio

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo LDFLAGS: -L${SRCDIR}/../ -L/usr/lib -Wl,-rpath,$ORIGIN -lbass -lbass_fx -lstdc++
#include "bass.h"
#include "bass_fx.h"
*/
import "C"

import (
	"log"
)

func Init() {

	// Output data regardless if audio is playing
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_NONSTOP, C.DWORD(1))

	// Worse time resolution but lower latency with DirectSound
	C.BASS_SetConfig(C.BASS_CONFIG_VISTA_TRUEPOS, C.DWORD(0))

	// Smaller buffer length, reduces latency
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_BUFFER, C.DWORD(10))

	// Update BASS more frequently
	C.BASS_SetConfig(C.BASS_CONFIG_DEV_PERIOD, C.DWORD(5))

	if C.BASS_Init(C.int(-1), C.DWORD(44100), C.DWORD(0), nil, nil) != 0 {
		log.Println("BASS Initialized!")
	} else {
		log.Println("BASS error", int(C.BASS_ErrorGetCode()))
	}
}
