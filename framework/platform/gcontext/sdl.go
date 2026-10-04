package gcontext

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
)

type OptionalProps struct {
	IconName       string
	BuiltinMSAA    bool
	Resizable      bool
	ScaleToMonitor bool
	Hidden         bool
	Fullscreen     bool
}

var (
	sdlWindow      *sdl.Window
	sdlShouldClose bool
	offscreenCtx   bool
	hovered        bool
	keyMap         = make(map[sdl.Keycode]bool)
	kMutex         sync.RWMutex
	fullscreen     bool
)

func Initialize(offscreen bool) error {
	libPath := filepath.Join(env.LibDir(), "SDL3.dll")
	if runtime.GOOS != "windows" {
		libPath = filepath.Join(env.LibDir(), "libSDL3.so")
	}

	if err := sdl.LoadLibrary(libPath); err != nil {
		return fmt.Errorf("sdl: couldn't load library: %w", err)
	}

	_ = sdl.SetHint(sdl.HINT_MOUSE_FOCUS_CLICKTHROUGH, "1")

	if err := sdl.SetHint("SDL_WINDOWS_DPI_AWARENESS", "derptest"); err != nil {
		return fmt.Errorf(`sdl: couldn't set hint "SDL_WINDOWS_DPI_AWARENESS": %w`, err)
	} // we set garbage value here so we can set proper one just before creating the window

	if offscreen && runtime.GOOS != "windows" {
		if err := sdl.SetHint(sdl.HINT_VIDEO_DRIVER, "offscreen"); err != nil {
			return fmt.Errorf(`sdl: couldn't set hint "%s": %w`, sdl.HINT_VIDEO_DRIVER, err)
		}

		offscreenCtx = true
	}

	return sdl.Init(sdl.INIT_VIDEO) // preinitialize to get access to display info - it will be reinitialized during window creation
}

func SDLCreateWindow(width, height int, title string, props OptionalProps) {

	sdl.QuitSubSystem(sdl.INIT_VIDEO)

	if props.ScaleToMonitor {
		// We want the window to be rescaled
		if err := sdl.SetHint("SDL_WINDOWS_DPI_AWARENESS", "unaware"); err != nil {
			return
		}
	} else {
		if err := sdl.SetHint("SDL_WINDOWS_DPI_AWARENESS", ""); err != nil {
			return
		}
	}

	err2 := sdl.InitSubSystem(sdl.INIT_VIDEO)
	if err2 != nil {
		log.Fatal(err2)
	}

	_ = sdl.GL_SetAttribute(sdl.GL_FRAMEBUFFER_SRGB_CAPABLE, 1)
	_ = sdl.GL_SetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	_ = sdl.GL_SetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	_ = sdl.GL_SetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	if props.BuiltinMSAA {
		_ = sdl.GL_SetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
		_ = sdl.GL_SetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	}

	var flags sdl.WindowFlags

	if props.Resizable {
		flags |= sdl.WINDOW_RESIZABLE
	}

	var err error
	sdlWindow, err = sdl.CreateWindow(title, width, height, sdl.WINDOW_OPENGL|sdl.WINDOW_HIDDEN|flags)

	if err != nil {
		panic(err)
	}

	if props.Fullscreen {
		md, err := sdl.GetPrimaryDisplay().CurrentDisplayMode()
		if err != nil {
			panic(err)
		}

		md.W = int32(width)
		md.H = int32(height)

		if err = sdlWindow.SetFullscreenMode(md); err != nil {
			panic(err)
		}

		if err = sdlWindow.SetFullscreen(true); err != nil {
			panic(err)
		}

		fullscreen = true
	}

	if !props.Hidden {
		if err = sdlWindow.Show(); err != nil {
			return
		}
	}

	if !offscreenCtx {
		if err = sdlWindow.StartTextInput(); err != nil {
			panic(err)
		}
	}

	if props.IconName != "" && !offscreenCtx {
		loadIconsSDL(props.IconName)
	}

	_, err = sdl.GL_CreateContext(sdlWindow)
	if err != nil {
		panic(err)
	}
}

func GetFramebufferSize() (int, int) {
	w, h, err := sdlWindow.Size()
	if err != nil {
		panic(err)
	}

	return int(w), int(h)
}

func IsHovered() bool {
	return hovered
}

func IsFocused() bool {
	return sdlWindow.Flags()&sdl.WINDOW_INPUT_FOCUS > 0
}

func Focus() {
	if err := sdlWindow.Raise(); err != nil {
		panic(err)
	}
}

func IsMinimized() bool {
	return sdlWindow.Flags()&sdl.WINDOW_MINIMIZED > 0
}

func GetCursorPosition() (float32, float32) {
	if sdlWindow.RelativeMouseMode() {
		_, x, y := sdl.GetMouseState()

		return x, y
	}

	xW, yW, _ := sdlWindow.Position()
	_, xG, yG := sdl.GetGlobalMouseState()

	return xG - float32(xW), yG - float32(yW)
}

func GetRelativePosition() (float32, float32) {
	_, xG, yG := sdl.GetRelativeMouseState()

	return xG, yG
}

func getWindowBounds() (tl, br vector.Vector2f) {
	xW, yW, _ := sdlWindow.Position()
	w, h, _ := sdlWindow.SizeInPixels()

	return vector.NewVec2f(float32(xW), float32(yW)), vector.NewVec2f(float32(xW)+float32(w), float32(yW)+float32(h))
}

func SetCursorPosition(x, y float32) {
	tl, br := getWindowBounds()

	if x < 0 || x > br.X-tl.X || y < 0 || y > br.Y-tl.Y {
		hovered = false
	} else {
		hovered = true
	}

	err := sdl.WarpMouseGlobal(x+tl.X, y+tl.Y)
	if err != nil {
		panic(err)
	}
}

func SetWindowCursorPosition(x, y float32) {
	sdlWindow.WarpMouseIn(x, y)
}

func SetRawInput(on bool) {
	if err := sdlWindow.SetRelativeMouseMode(on); err != nil {
		panic(err)
	}
}

func SetCursorVisible(visible bool) {
	if visible {
		if err := sdl.ShowCursor(); err != nil {
			return
		}
	} else {
		if err := sdl.HideCursor(); err != nil {
			return
		}
	}
}

func GetLeftClick() bool {
	flags, _, _ := sdl.GetMouseState()

	return flags&(1<<(sdl.BUTTON_LEFT-1)) != 0
}

func GetRightClick() bool {
	flags, _, _ := sdl.GetMouseState()

	return flags&(1<<(sdl.BUTTON_RIGHT-1)) != 0
}

func GetKeyState(key sdl.Keycode) Action {
	kMutex.RLock()
	defer kMutex.RUnlock()

	if keyMap[key] {
		return Press
	}

	return Release
}

func Minimize() {
	if err := sdlWindow.Minimize(); err != nil {
		panic(err)
	}
}

func Restore() {
	if err := sdlWindow.Restore(); err != nil {
		panic(err)
	}
}

func ShouldClose() bool {
	return sdlShouldClose
}

func SetShouldClose(shouldClose bool) {
	sdlShouldClose = shouldClose
}

func StartProgress() {
	if err := sdlWindow.SetProgressState(sdl.PROGRESS_STATE_NORMAL); err != nil {
		panic(err)
	}
}

func StopProgress() {
	if err := sdlWindow.SetProgressState(sdl.PROGRESS_STATE_NONE); err != nil {
		panic(err)
	}
}

func ErrorProgress() {
	if err := sdlWindow.SetProgressState(sdl.PROGRESS_STATE_ERROR); err != nil {
		panic(err)
	}
}

func SetProgress(value float32) {
	if err := sdlWindow.SetProgressValue(value); err != nil {
		panic(err)
	}
}

func loadIconsSDL(name string) {
	var iconSizes = []int{128, 64, 48, 32, 24, 16}
	if runtime.GOOS == "windows" { // windows looks broken with higher res icons, so 32px one is the first
		iconSizes = []int{32, 128, 64, 48, 24, 16}
	}

	var mainIcon *sdl.Surface

	var toDispose []*texture.Pixmap

	for i, size := range iconSizes {
		pxMap, _ := assets.GetPixmap("assets/textures/" + strings.Replace(name, "*", strconv.Itoa(size), 1) + ".png")

		icon, err := sdl.CreateSurfaceFrom(size, size, sdl.PIXELFORMAT_RGBA32, pxMap.Data, size*4)

		if err != nil {
			panic(err)
		}

		if i == 0 {
			mainIcon = icon
		} else if err = mainIcon.AddAlternateImage(icon); err != nil {
			panic(err)
		}

		toDispose = append(toDispose, pxMap)
	}

	err := sdlWindow.SetIcon(mainIcon)
	if err != nil {
		panic(err)
	}

	for _, pxMap := range toDispose {
		pxMap.Dispose()
	}
}

func GetPrimaryRefreshRate() float32 {
	if offscreenCtx {
		return 60
	}

	data, err := sdl.GetPrimaryDisplay().CurrentDisplayMode()
	if err != nil {
		panic(err)
	}

	return data.RefreshRate
}

func GetPrimaryVideoMode() sdl.DisplayMode {
	if offscreenCtx {
		return sdl.DisplayMode{
			W:           3840,
			H:           2160,
			RefreshRate: 60,
		}
	}

	data, err := sdl.GetPrimaryDisplay().CurrentDisplayMode()
	if err != nil {
		panic(err)
	}

	return *data
}

func AddToClipboard(text string) {
	if err := sdl.SetClipboardText(text); err != nil {
		panic(err)
	}
}

func IsMainThread() bool {
	return sdl.IsMainThread()
}
