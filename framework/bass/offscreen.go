package bass

/*
#include <stdint.h>
#include "bass.h"
#include "bassmix.h"
#include "bassenc.h"
*/
import "C"
import (
	"log"
	"unsafe"
)

func StartEncoding(file string) {
	log.Println("Starting encoder...")

	C.BASS_Encode_Start(masterMixer, C.CString(file), C.BASS_ENCODE_PCM, (*C.ENCODEPROC)(nil), unsafe.Pointer(nil)) // set a WAV writer on the mixer
}

var lastChunkSize float64
var mixerBuffer []byte

func EncodePart(chunkSize float64) {
	if lastChunkSize != chunkSize {
		lastChunkSize = chunkSize

		bufferSize := int(C.BASS_ChannelSeconds2Bytes(masterMixer, C.double(chunkSize)))
		mixerBuffer = make([]byte, bufferSize)
	}

	C.BASS_ChannelGetData(masterMixer, unsafe.Pointer(&mixerBuffer[0]), C.DWORD(len(mixerBuffer)))
}

func StopEncoding() {
	C.BASS_Encode_Stop(masterMixer)
}
