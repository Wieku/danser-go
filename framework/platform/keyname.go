package platform

import "C"
import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"strconv"
	"strings"
)

func GetKeyName(key glfw.Key, scancode int) string {
	if key >= glfw.KeyF1 && key <= glfw.KeyF25 {
		return "F" + strconv.Itoa(int(key-glfw.KeyF1)+1)
	} else if key >= glfw.KeyKP0 && key <= glfw.KeyKP9 {
		return "NUMPAD" + strconv.Itoa(int(key-glfw.KeyKP0))
	} else {
		switch key {
		case glfw.KeyKPDecimal:
			return "NUMPADDECIMAL"
		case glfw.KeyKPDivide:
			return "NUMPADDIVIDE"
		case glfw.KeyKPMultiply:
			return "NUMPADMULTIPLY"
		case glfw.KeyKPSubtract:
			return "NUMPADSUBTRACT"
		case glfw.KeyKPAdd:
			return "NUMPADADD"
		case glfw.KeyKPEnter:
			return "NUMPADENTER"
		case glfw.KeyKPEqual:
			return "NUMPADEQUAL"
		case glfw.KeyEscape:
			return "ESCAPE"
		case glfw.KeyEnter:
			return "ENTER"
		case glfw.KeyTab:
			return "TAB"
		case glfw.KeyBackspace:
			return "BACKSPACE"
		case glfw.KeyInsert:
			return "INSERT"
		case glfw.KeyDelete:
			return "DELETE"
		case glfw.KeyRight:
			return "RIGHT"
		case glfw.KeyLeft:
			return "LEFT"
		case glfw.KeyDown:
			return "DOWN"
		case glfw.KeyUp:
			return "UP"
		case glfw.KeyPageUp:
			return "PAGEUP"
		case glfw.KeyPageDown:
			return "PAGEDOWN"
		case glfw.KeyHome:
			return "HOME"
		case glfw.KeyEnd:
			return "END"
		case glfw.KeyCapsLock:
			return "CAPS"
		case glfw.KeyScrollLock:
			return "SCROLLLOCK"
		case glfw.KeyNumLock:
			return "NUMLOCK"
		case glfw.KeyPrintScreen:
			return "PRINTSCREEN"
		case glfw.KeyPause:
			return "PAUSE"
		case glfw.KeyLeftShift:
			return "LSHIFT"
		case glfw.KeyLeftControl:
			return "LCTRL"
		case glfw.KeyLeftAlt:
			return "LALT"
		case glfw.KeyLeftSuper:
			return "LSUPER"
		case glfw.KeyRightShift:
			return "RSHIFT"
		case glfw.KeyRightControl:
			return "RCTRL"
		case glfw.KeyRightAlt:
			return "RALT"
		case glfw.KeyRightSuper:
			return "RSUPER"
		case glfw.KeySpace:
			return "SPACE"
		default:
			return strings.ToUpper(glfw.GetKeyName(key, scancode))
		}
	}
}
