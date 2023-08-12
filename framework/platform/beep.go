package platform

/*
#include <stdint.h>

#ifdef _WIN32

#include <windows.h>
#include <winuser.h>

void system_beep(uint32_t t) {
	MessageBeep(t);
}

#else

void system_beep(uint32_t t) {}

#endif
*/
import "C"

type BeepType uint

const (
	Question BeepType = 0x20
	Error    BeepType = 0x10
	Info     BeepType = 0x40
	Warning  BeepType = 0x30
	Ok       BeepType = 0
)

func Beep(t BeepType) {
	C.system_beep(C.uint32_t(t))
}
