package gcontext

import (
	"reflect"

	"github.com/Zyko0/go-sdl3/sdl"
)

type Action uint

const (
	None Action = iota
	Release
	Press
	Repeat
)

type Event interface {
	KeyEvent | CharEvent | ScrollEvent | CloseEvent | DropEvent
}

type CharEvent struct {
	Text string
}

type KeyEvent struct {
	Key      sdl.Keycode
	Scancode sdl.Scancode
	Mod      sdl.Keymod
	Name     string
	Action   Action
}

type ScrollEvent struct {
	X, Y float32
}

type CloseEvent struct {
}

type DropEvent struct {
	Names []string
}

var sdlListeners = make(map[reflect.Type][]any)

func RegisterListener[T Event](listener func(event T)) {
	rType := reflect.TypeFor[T]()

	if _, ok := sdlListeners[rType]; !ok {
		sdlListeners[rType] = make([]any, 0)
	}

	sdlListeners[rType] = append(sdlListeners[rType], listener)
}

func callListeners[T Event](event T) {
	rType := reflect.TypeFor[T]()

	if listeners, ok := sdlListeners[rType]; ok {
		for _, listener := range listeners {
			listener.(func(T))(event)
		}
	}
}

var fileList []string

func HandleEvents() {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		switch event.Type {
		case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
			kEvent := event.KeyboardEvent()

			kn, _ := GetKeyName(kEvent.Key, kEvent.Scancode)

			kMutex.Lock()

			action := Release
			if kEvent.Repeat {
				action = Repeat
			} else if kEvent.Down {
				action = Press
				keyMap[kEvent.Key] = true
			} else {
				keyMap[kEvent.Key] = false
			}

			kMutex.Unlock()

			callListeners(KeyEvent{
				Key:      kEvent.Key,
				Scancode: kEvent.Scancode,
				Mod:      kEvent.Mod,
				Name:     kn,
				Action:   action,
			})
		case sdl.EVENT_TEXT_INPUT:
			callListeners(CharEvent{
				Text: event.TextInputEvent().Text,
			})
		case sdl.EVENT_MOUSE_WHEEL:
			wEvent := event.MouseWheelEvent()
			callListeners(ScrollEvent{
				X: wEvent.X,
				Y: wEvent.Y,
			})
		case sdl.EVENT_QUIT, sdl.EVENT_WINDOW_CLOSE_REQUESTED:
			SetShouldClose(true)
			callListeners(CloseEvent{})
		case sdl.EVENT_DROP_BEGIN:
			fileList = fileList[:0]
		case sdl.EVENT_DROP_FILE:
			dEvent := event.DropEvent()

			fileList = append(fileList, dEvent.Data)
		case sdl.EVENT_DROP_COMPLETE:
			if len(fileList) > 0 {
				callListeners(DropEvent{
					Names: fileList,
				})
			}

			fileList = fileList[:0]
		case sdl.EVENT_WINDOW_MOUSE_ENTER:
			hovered = true
		case sdl.EVENT_WINDOW_MOUSE_LEAVE:
			hovered = false
		}
	}
}
