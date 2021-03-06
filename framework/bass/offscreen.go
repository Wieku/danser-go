package bass

/*
#include <stdint.h>
#include "bass.h"
#include "bassmix.h"
#include "bassenc.h"

extern void goCallback(int);

static inline void CALLBACK EventSyncProc(HSYNC handle, DWORD channel, DWORD data, void *user)
{
	int e = (intptr_t)user; // the event number
	goCallback(e);
}

static inline void SetSync(HSTREAM stream, QWORD pos, int eNum) {
	BASS_ChannelSetSync(stream, BASS_SYNC_POS|BASS_SYNC_MIXTIME, pos, EventSyncProc, (void*)(intptr_t)eNum);
}
*/
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

var GlobalTimeMs = 0.0
var Offscreen = false

var mixStream C.HSTREAM

type trackEvent struct {
	channel  C.DWORD
	time     float64
	play     bool
	delegate func() C.DWORD
	called   bool
}

var trackEvents = make([]trackEvent, 0)

func addNormalEvent(delegate func()) {
	addDelayedEvent(0, delegate)
}

func addDelayedEvent(delay float64, delegate func()) {
	trackEvents = append(trackEvents, trackEvent{
		channel: 0,
		time:    GlobalTimeMs + delay,
		play:    false,
		delegate: func() C.DWORD {
			delegate()
			return C.DWORD(0)
		},
	})
}

func SaveToFile(file string) {
	mixStream = C.BASS_Mixer_StreamCreate(48000, 2, C.BASS_STREAM_DECODE|C.BASS_MIXER_END|C.BASS_SAMPLE_FLOAT)

	log.Println("Audio mixing stream created, adding events for processing...")

	wasPlay := false

	for i, e := range trackEvents {
		//Push music event early to prevent mix stream from closing too early (BASS_MIXER_END flag)
		if e.play && e.channel != 0 && !wasPlay {
			goCallback(C.int(i))

			wasPlay = true

			continue
		}

		pos := C.BASS_ChannelSeconds2Bytes(mixStream, C.double(e.time/1000)) // get start position in bytes
		C.SetSync(mixStream, pos, C.int(i))
	}

	log.Println("Events added, starting encoding audio...")

	C.BASS_Encode_Start(mixStream, C.CString(file), C.BASS_ENCODE_PCM, (*C.ENCODEPROC)(nil), unsafe.Pointer(nil)) // set a WAV writer on the mixer

	buffer := make([]byte, 512)

	var ret int32
	for ret != -1 {
		ret = int32(C.BASS_ChannelGetData(mixStream, unsafe.Pointer(&buffer[0]), C.DWORD(len(buffer)))) // process the mixer
	}

	C.BASS_Encode_Stop(mixStream) // close the WAV writer

	log.Println("Encoding finished!")
}

//export goCallback
func goCallback(i C.int) {
	eventIndex := int(i)

	if trackEvents[eventIndex].called {
		return
	}

	minIndex := eventIndex

	for i := eventIndex - 1; i >= 0; i-- {
		if trackEvents[i].called {
			break
		}

		minIndex = i
	}

	for i := minIndex; i <= eventIndex; i++ {
		processEvent(i)
	}
}

func processEvent(eventIndex int) {
	event := trackEvents[eventIndex]

	var ret C.DWORD
	if event.delegate != nil {
		ret = event.delegate()
	}

	if event.play {
		if ret != 0 { //add samples to the queue
			C.BASS_Mixer_StreamAddChannel(mixStream, ret, C.BASS_STREAM_AUTOFREE|C.BASS_MIXER_CHAN_NORAMPIN)
		} else { //push main music to the queue
			pos := C.BASS_ChannelSeconds2Bytes(mixStream, C.double(event.time/1000))
			C.BASS_Mixer_StreamAddChannelEx(mixStream, event.channel, C.BASS_STREAM_AUTOFREE|C.BASS_MIXER_CHAN_NORAMPIN, pos, C.QWORD(0))
		}
	}

	errCode := GetError()
	if errCode != 0 {
		log.Println(fmt.Sprintf("BASS encountered an error: %d (%s) at: %f, event id: %d", errCode, errCode.Message(), event.time, eventIndex))
	}

	event.called = true
	trackEvents[eventIndex] = event
}
