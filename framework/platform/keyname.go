package platform

import "C"
import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"strconv"
	"strings"
)

func GetKeyName(key glfw.Key, scancode int) (name string, found bool) {
	if key < 0 {
		return "", false
	} else if key >= glfw.KeyF1 && key <= glfw.KeyF25 {
		name = "F" + strconv.Itoa(int(key-glfw.KeyF1)+1)
	} else if key >= glfw.KeyKP0 && key <= glfw.KeyKP9 {
		name = "NUMPAD" + strconv.Itoa(int(key-glfw.KeyKP0))
	} else {
		switch key {
		case glfw.KeyKPDecimal:
			name = "NUMPADDECIMAL"
		case glfw.KeyKPDivide:
			name = "NUMPADDIVIDE"
		case glfw.KeyKPMultiply:
			name = "NUMPADMULTIPLY"
		case glfw.KeyKPSubtract:
			name = "NUMPADSUBTRACT"
		case glfw.KeyKPAdd:
			name = "NUMPADADD"
		case glfw.KeyKPEnter:
			name = "NUMPADENTER"
		case glfw.KeyKPEqual:
			name = "NUMPADEQUAL"
		case glfw.KeyEscape:
			name = "ESCAPE"
		case glfw.KeyEnter:
			name = "ENTER"
		case glfw.KeyTab:
			name = "TAB"
		case glfw.KeyBackspace:
			name = "BACKSPACE"
		case glfw.KeyInsert:
			name = "INSERT"
		case glfw.KeyDelete:
			name = "DELETE"
		case glfw.KeyRight:
			name = "RIGHT"
		case glfw.KeyLeft:
			name = "LEFT"
		case glfw.KeyDown:
			name = "DOWN"
		case glfw.KeyUp:
			name = "UP"
		case glfw.KeyPageUp:
			name = "PAGEUP"
		case glfw.KeyPageDown:
			name = "PAGEDOWN"
		case glfw.KeyHome:
			name = "HOME"
		case glfw.KeyEnd:
			name = "END"
		case glfw.KeyCapsLock:
			name = "CAPS"
		case glfw.KeyScrollLock:
			name = "SCROLLLOCK"
		case glfw.KeyNumLock:
			name = "NUMLOCK"
		case glfw.KeyPrintScreen:
			name = "PRINTSCREEN"
		case glfw.KeyPause:
			name = "PAUSE"
		case glfw.KeyLeftShift:
			name = "LSHIFT"
		case glfw.KeyLeftControl:
			name = "LCTRL"
		case glfw.KeyLeftAlt:
			name = "LALT"
		case glfw.KeyLeftSuper:
			name = "LSUPER"
		case glfw.KeyRightShift:
			name = "RSHIFT"
		case glfw.KeyRightControl:
			name = "RCTRL"
		case glfw.KeyRightAlt:
			name = "RALT"
		case glfw.KeyRightSuper:
			name = "RSUPER"
		case glfw.KeySpace:
			name = "SPACE"
		default:
			name = strings.ToUpper(glfw.GetKeyName(key, scancode))
		}
	}

	return name, true
}
