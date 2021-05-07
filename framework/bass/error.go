package bass

/*
#include "bass.h"
*/
import "C"

type Error int

const (
	Ok                 Error = 0
	ErrorMem           Error = 1
	ErrorFileOpen      Error = 2
	ErrorDriver        Error = 3
	ErrorBufLost       Error = 4
	ErrorHandle        Error = 5
	ErrorFormat        Error = 6
	ErrorPosition      Error = 7
	ErrorInit          Error = 8
	ErrorStart         Error = 9
	ErrorNoCD          Error = 12
	ErrorCDTrack       Error = 13
	ErrorAlready       Error = 14
	ErrorNoPause       Error = 16
	ErrorNotAudio      Error = 17
	ErrorNoChan        Error = 18
	ErrorIllType       Error = 19
	ErrorIllParam      Error = 20
	ErrorNo3D          Error = 21
	ErrorNoEAX         Error = 22
	ErrorDevice        Error = 23
	ErrorNoPlay        Error = 24
	ErrorFreq          Error = 25
	ErrorNotFile       Error = 27
	ErrorNoHW          Error = 29
	ErrorEmpty         Error = 31
	ErrorNoNet         Error = 32
	ErrorCreate        Error = 33
	ErrorNoFX          Error = 34
	ErrorPlaying       Error = 35
	ErrorNotAvail      Error = 37
	ErrorDecode        Error = 38
	ErrorDX            Error = 39
	ErrorTimeOut       Error = 40
	ErrorFileForm      Error = 41
	ErrorSpeaker       Error = 42
	ErrorVersion       Error = 43
	ErrorCodec         Error = 44
	ErrorEnded         Error = 45
	ErrorBusy          Error = 46
	ErrorUnstreamable  Error = 47
	ErrorUnknown       Error = -1
	ErrorWMALicense    Error = 1000
	ErrorWWM9          Error = 1001
	ErrorWMADenied     Error = 1002
	ErrorWMACodec      Error = 1003
	ErrorWMAIndividual Error = 1004
	ErrorACMCancel     Error = 2000
	ErrorCastDenied    Error = 2100
	ErrorVSTNoInputs   Error = 3000
	ErrorVSTNoOutputs  Error = 3001
	ErrorVSTNoRealTime Error = 3002
	ErrorWASAPI        Error = 5000
	ErrorMP4NoStream   Error = 6000
)

func (e Error) Message() string {
	switch e {
	case Ok:
		return "All is OK"
	case ErrorMem:
		return "Memory error"
	case ErrorFileOpen:
		return "Can't open the file"
	case ErrorDriver:
		return "Can't find a free/valid driver"
	case ErrorBufLost:
		return "The sample buffer was lost"
	case ErrorHandle:
		return "Invalid handle"
	case ErrorFormat:
		return "Unsupported sample format"
	case ErrorPosition:
		return "Invalid playback position"
	case ErrorInit:
		return "BASS_Init has not been successfully called"
	case ErrorStart:
		return "BASS_Start has not been successfully called"
	case ErrorNoCD:
		return "No CD in drive"
	case ErrorCDTrack:
		return "Invalid track number"
	case ErrorAlready:
		return "Already initialized/paused/whatever"
	case ErrorNoPause:
		return "Not paused"
	case ErrorNotAudio:
		return "Not an audio track"
	case ErrorNoChan:
		return "Can't get a free channel"
	case ErrorIllType:
		return "An illegal type was specified"
	case ErrorIllParam:
		return "An illegal parameter was specified"
	case ErrorNo3D:
		return "No 3D support"
	case ErrorNoEAX:
		return "No EAX support"
	case ErrorDevice:
		return "Illegal device number"
	case ErrorNoPlay:
		return "Not playing"
	case ErrorFreq:
		return "Illegal sample rate"
	case ErrorNotFile:
		return "The stream is not a file stream"
	case ErrorNoHW:
		return "No hardware voices available"
	case ErrorEmpty:
		return "The MOD music has no sequence data"
	case ErrorNoNet:
		return "No internet connection could be opened"
	case ErrorCreate:
		return "Couldn't create the file"
	case ErrorNoFX:
		return "Effects are not available"
	case ErrorPlaying:
		return "The channel is playing"
	case ErrorNotAvail:
		return "Requested data is not available"
	case ErrorDecode:
		return "The channel is a 'decoding channel'"
	case ErrorDX:
		return "A sufficient DirectX version is not installed"
	case ErrorTimeOut:
		return "Connection timed out"
	case ErrorFileForm:
		return "Unsupported file format"
	case ErrorSpeaker:
		return "Unavailable speaker"
	case ErrorVersion:
		return "Invalid BASS version (used by add-ons)"
	case ErrorCodec:
		return "Codec is not available/supported"
	case ErrorEnded:
		return "The channel/file has ended"
	case ErrorBusy:
		return "The device is busy (eg. in \"exclusive\" use by another process)"
	case ErrorUnstreamable:
		return "The file is unstreamable"
	case ErrorUnknown:
		return "Some other mystery error"
	case ErrorWMALicense:
		return "BassWma: the file is protected"
	case ErrorWWM9:
		return "BassWma: WM9 is required"
	case ErrorWMADenied:
		return "BassWma: access denied (user/pass is invalid)"
	case ErrorWMACodec:
		return "BassWma: no appropriate codec is installed"
	case ErrorWMAIndividual:
		return "BassWma: individualization is needed"
	case ErrorACMCancel:
		return "BassEnc: ACM codec selection cancelled"
	case ErrorCastDenied:
		return "BassEnc: Access denied (invalid password)"
	case ErrorVSTNoInputs:
		return "BassVst: the given effect has no inputs and is probably a VST instrument and no effect"
	case ErrorVSTNoOutputs:
		return "BassVst: the given effect has no outputs"
	case ErrorVSTNoRealTime:
		return "BassVst: the given effect does not support realtime processing"
	case ErrorWASAPI:
		return "BASSWASAPI: no WASAPI available"
	case ErrorMP4NoStream:
		return "BASS_AAC: non-streamable due to MP4 atom order ('mdat' before 'moov')"
	}

	return "Unknown error code"
}

func GetError() Error {
	return Error(C.BASS_ErrorGetCode())
}