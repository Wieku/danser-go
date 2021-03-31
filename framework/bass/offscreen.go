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
	mixStream = C.BASS_Mixer_StreamCreate(44100, 2, C.BASS_STREAM_DECODE|C.BASS_MIXER_END)

	log.Println("Mixer stream created, adding events for processing...")

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

	log.Println("Events added, starting encoding...")

	C.BASS_Encode_Start(mixStream, C.CString(file), C.BASS_ENCODE_PCM, (*C.ENCODEPROC)(nil), unsafe.Pointer(nil)) // set a WAV writer on the mixer

	// TODO: test if buffer length affects latency
	buffer := make([]byte, 512)

	for {
		ret := C.BASS_ChannelGetData(mixStream, unsafe.Pointer(&buffer[0]), C.DWORD(len(buffer))) // process the mixer
		if int32(ret) == -1 {
			break
		}
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

	errCode := C.BASS_ErrorGetCode()
	if errCode != 0 {
		log.Println("BASS encountered an error: ", errCode, " at: ", event.time, ret)
	}

	event.called = true
	trackEvents[eventIndex] = event
}
