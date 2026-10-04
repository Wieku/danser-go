package gcontext

import "C"
import (
	"strconv"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

func GetKeyName(keycode sdl.Keycode, scancode sdl.Scancode) (name string, found bool) {
	if keycode < 0 {
		return "", false
	} else if keycode >= sdl.K_KP_1 && keycode <= sdl.K_KP_9 {
		name = "NUMPAD" + strconv.Itoa(int(keycode-sdl.K_KP_1))
	} else {
		switch keycode {
		case sdl.K_KP_0:
			name = "NUMPAD0"
		case sdl.K_KP_00:
			name = "NUMPAD00"
		case sdl.K_KP_000:
			name = "NUMPAD000"
		case sdl.K_KP_DECIMAL:
			name = "NUMPADDECIMAL"
		case sdl.K_KP_DIVIDE:
			name = "NUMPADDIVIDE"
		case sdl.K_KP_MULTIPLY:
			name = "NUMPADMULTIPLY"
		case sdl.K_KP_MINUS:
			name = "NUMPADSUBTRACT"
		case sdl.K_KP_PLUS:
			name = "NUMPADADD"
		case sdl.K_KP_ENTER:
			name = "NUMPADENTER"
		case sdl.K_KP_EQUALS:
			name = "NUMPADEQUAL"
		case sdl.K_ESCAPE:
			name = "ESCAPE"
		case sdl.K_RETURN:
			name = "ENTER"
		case sdl.K_TAB:
			name = "TAB"
		case sdl.K_BACKSPACE:
			name = "BACKSPACE"
		case sdl.K_INSERT:
			name = "INSERT"
		case sdl.K_DELETE:
			name = "DELETE"
		case sdl.K_RIGHT:
			name = "RIGHT"
		case sdl.K_LEFT:
			name = "LEFT"
		case sdl.K_DOWN:
			name = "DOWN"
		case sdl.K_UP:
			name = "UP"
		case sdl.K_PAGEUP:
			name = "PAGEUP"
		case sdl.K_PAGEDOWN:
			name = "PAGEDOWN"
		case sdl.K_HOME:
			name = "HOME"
		case sdl.K_END:
			name = "END"
		case sdl.K_CAPSLOCK:
			name = "CAPS"
		case sdl.K_SCROLLLOCK:
			name = "SCROLLLOCK"
		case sdl.K_NUMLOCKCLEAR:
			name = "NUMLOCK"
		case sdl.K_PRINTSCREEN:
			name = "PRINTSCREEN"
		case sdl.K_PAUSE:
			name = "PAUSE"
		case sdl.K_LSHIFT:
			name = "LSHIFT"
		case sdl.K_LCTRL:
			name = "LCTRL"
		case sdl.K_LALT:
			name = "LALT"
		case sdl.K_LGUI:
			name = "LSUPER"
		case sdl.K_RSHIFT:
			name = "RSHIFT"
		case sdl.K_RCTRL:
			name = "RCTRL"
		case sdl.K_RALT:
			name = "RALT"
		case sdl.K_RGUI:
			name = "RSUPER"
		case sdl.K_SPACE:
			name = "SPACE"
		default:
			name = strings.ToUpper(keycode.KeyName())
		}
	}

	return name, true
}
