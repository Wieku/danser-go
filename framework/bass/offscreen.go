package bass

/*
#include <stdint.h>
#include "bass.h"
#include "bassmix.h"
*/
import "C"
import (
	"unsafe"
)

func GetMixerRequiredBufferSize(seconds float64) int {
	return int(C.BASS_ChannelSeconds2Bytes(masterMixer, C.double(seconds)))
}

func ProcessMixer(buffer []byte) {
	C.BASS_ChannelGetData(masterMixer, unsafe.Pointer(&buffer[0]), C.DWORD(len(buffer)))
}
