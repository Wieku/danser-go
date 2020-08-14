package audio

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo LDFLAGS: -L${SRCDIR}/../ -L/usr/lib -Wl,-rpath,$ORIGIN -lbass -lbass_fx
#include "bass.h"
#include "bass_fx.h"
 */
import "C"

import (
	"log"
)

func Init() {
	if C.BASS_Init(C.int(-1), C.DWORD(44100), C.DWORD(0), nil, nil) != 0 {
		log.Println("BASS Initialized!")
	} else {
		log.Println("BASS error", int(C.BASS_ErrorGetCode()))
	}
}
