//go:build windows

package platform

/*
#cgo LDFLAGS: -lole32
#include <stdint.h>
#include <shobjidl.h>
#include "winstuff.h"
*/
import "C"
import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"unsafe"
)

func getWindow(win *glfw.Window) C.HWND {
	return C.HWND(unsafe.Pointer(win.GetWin32Window())) // Converting from glfw.HWND to this package HWND
}

func StopProgress(win *glfw.Window) {
	C.setState(getWindow(win), C.TBPF_NOPROGRESS)
}

func StartProgress(win *glfw.Window) {
	C.setState(getWindow(win), C.TBPF_NORMAL)
}

func SetProgress(win *glfw.Window, current, max int) {
	C.setProgress(getWindow(win), C.int32_t(current), C.int32_t(max))
}
