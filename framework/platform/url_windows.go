//go:build windows

package platform

/*
#cgo LDFLAGS: -lole32
#include "winstuff.h"
*/
import "C"
import (
	"fmt"
	"log"
	"os/exec"
	"unicode/utf16"
)

func ShowFileInManager(path string) {
	wFile := utf16.Encode([]rune(path))
	wFile = append(wFile, 0) // NULL terminated string

	if res := C.openInExplorer((*C.wchar_t)(&wFile[0])); res != 0 {
		log.Println(fmt.Sprintf("WINAPI failed: 0x%08x", int(res)))
		log.Println("Trying an alternative method...")

		// Failsafe
		err := exec.Command("explorer", "/select,", path).Start()
		if err != nil {
			panic(err)
		}
	}
}
