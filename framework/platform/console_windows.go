// +build windows

package platform

/*
#include <windows.h>

DWORD quickMode = 0;

void DisableQuickEdit() {
	HANDLE hInput;
    DWORD prev_mode;

    hInput = GetStdHandle(STD_INPUT_HANDLE);
	if (hInput == NULL || hInput == INVALID_HANDLE_VALUE) {
		return;
	}

    GetConsoleMode(hInput, &prev_mode);

	quickMode = prev_mode & ENABLE_QUICK_EDIT_MODE;

    SetConsoleMode(hInput, (prev_mode & ~ENABLE_QUICK_EDIT_MODE) | ENABLE_EXTENDED_FLAGS);
}

void EnableQuickEdit() {
	HANDLE hInput;
    DWORD prev_mode;

    hInput = GetStdHandle(STD_INPUT_HANDLE);
	if (hInput == NULL || hInput == INVALID_HANDLE_VALUE) {
		return;
	}

    GetConsoleMode(hInput, &prev_mode);
    SetConsoleMode(hInput, prev_mode | quickMode | ENABLE_EXTENDED_FLAGS);
}
*/
import "C"

func DisableQuickEdit() {
	C.DisableQuickEdit()
}

func EnableQuickEdit() {
	C.EnableQuickEdit()
}
