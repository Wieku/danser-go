package input

import "github.com/go-gl/glfw/v3.3/glfw"

var Win *glfw.Window

type KeyListener glfw.KeyCallback

var listeners []KeyListener

func RegisterListener(listener KeyListener) {
	listeners = append(listeners, listener)
}

func CallListeners(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	for _, l := range listeners {
		l(w, key, scancode, action, mods)
	}
}