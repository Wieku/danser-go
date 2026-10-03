package launcher

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"log"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/platform/gcontext"
	"github.com/wieku/danser-go/framework/qpc"
)

var context *imgui.Context
var ImIO *imgui.IO
var rShader *shader.RShader
var vao *buffer.VertexArrayObject

var ibo *buffer.IndexBufferObject

var Font *imgui.Font
var FontAw *imgui.Font

type sCache struct {
	started bool
	blocked bool
	held    bool
	mY      float32
	cId     imgui.ID
	fakeId  imgui.ID
}

var scrCache = sCache{} //make(map[imgui.ID]sCache)

type scrollContainer struct {
	window      *imgui.Window
	fakeId      imgui.ID
	origPos     float32
	speed       float32
	totalScroll float32
}

var scrollDeltas = make(map[imgui.ID]scrollContainer)

func SetupImgui() {
	log.Println("Imgui setup")

	context = imgui.CreateContext()

	imgui.PushStyleVarFloat(imgui.StyleVarPopupRounding, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarGrabRounding, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
	imgui.PushStyleColorVec4(imgui.ColBorder, vec4(1, 1, 1, 1))
	imgui.PushStyleColorVec4(imgui.ColFrameBg, vec4(0.1, 0.1, 0.1, 0.8))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, vec4(0.2, 0.2, 0.2, 0.8))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, vec4(0.4, 0.4, 0.4, 0.8))

	imgui.PushStyleColorVec4(imgui.ColButton, vec4(0.1, 0.1, 0.1, 0.8))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, vec4(0.2, 0.2, 0.2, 0.8))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, vec4(0.4, 0.4, 0.4, 0.8))

	imgui.PushStyleColorVec4(imgui.ColTitleBg, vec4(0.2, 0.2, 0.2, 0.8))
	imgui.PushStyleColorVec4(imgui.ColTitleBgActive, vec4(0.2, 0.2, 0.2, 0.8))
	imgui.PushStyleColorVec4(imgui.ColTitleBgCollapsed, vec4(0.2, 0.2, 0.2, 0.8))

	imgui.SetCurrentContext(context)

	ImIO = imgui.CurrentIO()

	ImIO.SetIniFilename("")

	ImIO.SetBackendFlags(imgui.BackendFlagsRendererHasTextures)

	//region texture

	Font = addFont("Quicksand-Bold.ttf", false)
	FontAw = addFont("Font Awesome 6 Free-Solid-900.otf", false)

	//endregion

	vertexSource, _ := assets.GetString("assets/shaders/imgui.vsh")
	fragmentSource, _ := assets.GetString("assets/shaders/imgui.fsh")

	rShader = shader.NewRShader(shader.NewSource(vertexSource, shader.Vertex), shader.NewSource(fragmentSource, shader.Fragment))

	vao = buffer.NewVertexArrayObject()

	// adding default VBO with initial size of 1000, will be resized when needed
	vao.AddVBO("default", 1000, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_uv", Type: attribute.Vec2},
		{Name: "in_color", Type: attribute.ColorPacked},
	})

	vao.Bind()
	vao.Attach(rShader)
	vao.Unbind()

	ibo = buffer.NewIndexBufferObject(100000)

	gcontext.RegisterListener(func(event gcontext.ScrollEvent) {
		ImIO.AddMouseWheelDelta(-event.X, event.Y)
	})

	gcontext.RegisterListener(func(event gcontext.KeyEvent) {
		if event.Action == gcontext.Repeat {
			return
		}

		sdlUpdateKeyModifiers(event.Mod)

		iKey := sdlKeyToImGuiKey(event.Key, event.Scancode)
		ImIO.AddKeyEvent(iKey, event.Action == gcontext.Press)
	})

	gcontext.RegisterListener(func(event gcontext.CharEvent) {
		ImIO.AddInputCharactersUTF8(event.Text)
	})
}

func addFont(fileName string, merge bool) *imgui.Font {
	fontBytes, err := assets.GetBytes("assets/fonts/" + fileName)
	if err != nil {
		panic(err)
	}

	ftPtr := C.malloc(C.size_t(len(fontBytes)))

	fontPtr := unsafe.Pointer(&fontBytes[0])

	C.memcpy(ftPtr, fontPtr, C.size_t(len(fontBytes)))

	runtime.KeepAlive(fontPtr)

	fc := imgui.NewFontConfig()
	fc.SetMergeMode(merge)
	defer fc.Destroy()

	return ImIO.Fonts().AddFontFromMemoryTTFV(uintptr(ftPtr), int32(len(fontBytes)), -1, fc, nil)
}

func sdlKeyToImGuiKey(keycode sdl.Keycode, scancode sdl.Scancode) imgui.Key {
	// Keypad doesn't have individual key values in SDL3
	switch scancode {
	case sdl.SCANCODE_KP_0:
		return imgui.KeyKeypad0
	case sdl.SCANCODE_KP_1:
		return imgui.KeyKeypad1
	case sdl.SCANCODE_KP_2:
		return imgui.KeyKeypad2
	case sdl.SCANCODE_KP_3:
		return imgui.KeyKeypad3
	case sdl.SCANCODE_KP_4:
		return imgui.KeyKeypad4
	case sdl.SCANCODE_KP_5:
		return imgui.KeyKeypad5
	case sdl.SCANCODE_KP_6:
		return imgui.KeyKeypad6
	case sdl.SCANCODE_KP_7:
		return imgui.KeyKeypad7
	case sdl.SCANCODE_KP_8:
		return imgui.KeyKeypad8
	case sdl.SCANCODE_KP_9:
		return imgui.KeyKeypad9
	case sdl.SCANCODE_KP_PERIOD:
		return imgui.KeyKeypadDecimal
	case sdl.SCANCODE_KP_DIVIDE:
		return imgui.KeyKeypadDivide
	case sdl.SCANCODE_KP_MULTIPLY:
		return imgui.KeyKeypadMultiply
	case sdl.SCANCODE_KP_MINUS:
		return imgui.KeyKeypadSubtract
	case sdl.SCANCODE_KP_PLUS:
		return imgui.KeyKeypadAdd
	case sdl.SCANCODE_KP_ENTER:
		return imgui.KeyKeypadEnter
	case sdl.SCANCODE_KP_EQUALS:
		return imgui.KeyKeypadEqual
	default:
		break
	}
	switch keycode {
	case sdl.K_TAB:
		return imgui.KeyTab
	case sdl.K_LEFT:
		return imgui.KeyLeftArrow
	case sdl.K_RIGHT:
		return imgui.KeyRightArrow
	case sdl.K_UP:
		return imgui.KeyUpArrow
	case sdl.K_DOWN:
		return imgui.KeyDownArrow
	case sdl.K_PAGEUP:
		return imgui.KeyPageUp
	case sdl.K_PAGEDOWN:
		return imgui.KeyPageDown
	case sdl.K_HOME:
		return imgui.KeyHome
	case sdl.K_END:
		return imgui.KeyEnd
	case sdl.K_INSERT:
		return imgui.KeyInsert
	case sdl.K_DELETE:
		return imgui.KeyDelete
	case sdl.K_BACKSPACE:
		return imgui.KeyBackspace
	case sdl.K_SPACE:
		return imgui.KeySpace
	case sdl.K_RETURN:
		return imgui.KeyEnter
	case sdl.K_ESCAPE:
		return imgui.KeyEscape
	case sdl.K_COMMA:
		return imgui.KeyComma

	case sdl.K_PERIOD:
		return imgui.KeyPeriod

	case sdl.K_SEMICOLON:
		return imgui.KeySemicolon

	case sdl.K_CAPSLOCK:
		return imgui.KeyCapsLock
	case sdl.K_SCROLLLOCK:
		return imgui.KeyScrollLock
	case sdl.K_NUMLOCKCLEAR:
		return imgui.KeyNumLock
	case sdl.K_PRINTSCREEN:
		return imgui.KeyPrintScreen
	case sdl.K_PAUSE:
		return imgui.KeyPause
	case sdl.K_LCTRL:
		return imgui.KeyLeftCtrl
	case sdl.K_LSHIFT:
		return imgui.KeyLeftShift
	case sdl.K_LALT:
		return imgui.KeyLeftAlt
	case sdl.K_LGUI:
		return imgui.KeyLeftSuper
	case sdl.K_RCTRL:
		return imgui.KeyRightCtrl
	case sdl.K_RSHIFT:
		return imgui.KeyRightShift
	case sdl.K_RALT:
		return imgui.KeyRightAlt
	case sdl.K_RGUI:
		return imgui.KeyRightSuper
	case sdl.K_APPLICATION:
		return imgui.KeyMenu
	case sdl.K_0:
		return imgui.Key0
	case sdl.K_1:
		return imgui.Key1
	case sdl.K_2:
		return imgui.Key2
	case sdl.K_3:
		return imgui.Key3
	case sdl.K_4:
		return imgui.Key4
	case sdl.K_5:
		return imgui.Key5
	case sdl.K_6:
		return imgui.Key6
	case sdl.K_7:
		return imgui.Key7
	case sdl.K_8:
		return imgui.Key8
	case sdl.K_9:
		return imgui.Key9
	case sdl.K_A:
		return imgui.KeyA
	case sdl.K_B:
		return imgui.KeyB
	case sdl.K_C:
		return imgui.KeyC
	case sdl.K_D:
		return imgui.KeyD
	case sdl.K_E:
		return imgui.KeyE
	case sdl.K_F:
		return imgui.KeyF
	case sdl.K_G:
		return imgui.KeyG
	case sdl.K_H:
		return imgui.KeyH
	case sdl.K_I:
		return imgui.KeyI
	case sdl.K_J:
		return imgui.KeyJ
	case sdl.K_K:
		return imgui.KeyK
	case sdl.K_L:
		return imgui.KeyL
	case sdl.K_M:
		return imgui.KeyM
	case sdl.K_N:
		return imgui.KeyN
	case sdl.K_O:
		return imgui.KeyO
	case sdl.K_P:
		return imgui.KeyP
	case sdl.K_Q:
		return imgui.KeyQ
	case sdl.K_R:
		return imgui.KeyR
	case sdl.K_S:
		return imgui.KeyS
	case sdl.K_T:
		return imgui.KeyT
	case sdl.K_U:
		return imgui.KeyU
	case sdl.K_V:
		return imgui.KeyV
	case sdl.K_W:
		return imgui.KeyW
	case sdl.K_X:
		return imgui.KeyX
	case sdl.K_Y:
		return imgui.KeyY
	case sdl.K_Z:
		return imgui.KeyZ
	case sdl.K_F1:
		return imgui.KeyF1
	case sdl.K_F2:
		return imgui.KeyF2
	case sdl.K_F3:
		return imgui.KeyF3
	case sdl.K_F4:
		return imgui.KeyF4
	case sdl.K_F5:
		return imgui.KeyF5
	case sdl.K_F6:
		return imgui.KeyF6
	case sdl.K_F7:
		return imgui.KeyF7
	case sdl.K_F8:
		return imgui.KeyF8
	case sdl.K_F9:
		return imgui.KeyF9
	case sdl.K_F10:
		return imgui.KeyF10
	case sdl.K_F11:
		return imgui.KeyF11
	case sdl.K_F12:
		return imgui.KeyF12
	case sdl.K_F13:
		return imgui.KeyF13
	case sdl.K_F14:
		return imgui.KeyF14
	case sdl.K_F15:
		return imgui.KeyF15
	case sdl.K_F16:
		return imgui.KeyF16
	case sdl.K_F17:
		return imgui.KeyF17
	case sdl.K_F18:
		return imgui.KeyF18
	case sdl.K_F19:
		return imgui.KeyF19
	case sdl.K_F20:
		return imgui.KeyF20
	case sdl.K_F21:
		return imgui.KeyF21
	case sdl.K_F22:
		return imgui.KeyF22
	case sdl.K_F23:
		return imgui.KeyF23
	case sdl.K_F24:
		return imgui.KeyF24
	case sdl.K_AC_BACK:
		return imgui.KeyAppBack
	case sdl.K_AC_FORWARD:
		return imgui.KeyAppForward
	default:
		break
	}

	// Fallback to scancode
	switch scancode {
	case sdl.SCANCODE_GRAVE:
		return imgui.KeyGraveAccent
	case sdl.SCANCODE_MINUS:
		return imgui.KeyMinus
	case sdl.SCANCODE_EQUALS:
		return imgui.KeyEqual
	case sdl.SCANCODE_LEFTBRACKET:
		return imgui.KeyLeftBracket
	case sdl.SCANCODE_RIGHTBRACKET:
		return imgui.KeyRightBracket
		//case sdl.SCANCODE_NONUSBACKSLASH: return imgui.KeyOem102;
	case sdl.SCANCODE_BACKSLASH:
		return imgui.KeyBackslash
	case sdl.SCANCODE_SEMICOLON:
		return imgui.KeySemicolon
	case sdl.SCANCODE_APOSTROPHE:
		return imgui.KeyApostrophe
	case sdl.SCANCODE_COMMA:
		return imgui.KeyComma
	case sdl.SCANCODE_PERIOD:
		return imgui.KeyPeriod
	case sdl.SCANCODE_SLASH:
		return imgui.KeySlash
	default:
		break
	}
	return imgui.KeyNone
}

func sdlUpdateKeyModifiers(mods sdl.Keymod) {
	ImIO.AddKeyEvent(imgui.KeyReservedForModCtrl, (mods&sdl.KMOD_CTRL) != 0)
	ImIO.AddKeyEvent(imgui.KeyReservedForModShift, (mods&sdl.KMOD_SHIFT) != 0)
	ImIO.AddKeyEvent(imgui.KeyReservedForModAlt, (mods&sdl.KMOD_ALT) != 0)
	ImIO.AddKeyEvent(imgui.KeyReservedForModSuper, (mods&sdl.KMOD_GUI) != 0)
}

var lastTime float64

func Begin() {
	sliderSledLastFrame = sliderSledThisFrame
	sliderSledThisFrame = false

	x, y := gcontext.GetCursorPosition()

	w, h := int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()) //input.Win.GetFramebufferSize()
	_, h1 := gcontext.GetFramebufferSize()

	scaling := float32(h1) / float32(h)

	ImIO.AddMousePosEvent(x/scaling, y/scaling)
	ImIO.AddMouseButtonEvent(0, gcontext.GetLeftClick())
	ImIO.AddMouseButtonEvent(1, gcontext.GetRightClick())

	ImIO.SetDisplaySize(imgui.Vec2{X: float32(w), Y: float32(h)})

	time := qpc.GetMilliTimeF() / 1000

	delta := float32(time - lastTime)

	lastTime = time

	ImIO.SetDeltaTime(delta)

	delta60 := delta / 0.0166667

	for k, v := range scrollDeltas {
		if v.window.Scroll().Y != math32.Round(v.origPos+v.totalScroll) || math32.Abs(v.speed) < 1 {
			delete(scrollDeltas, k)

			continue
		}

		v.speed *= math32.Pow(0.95, delta60)
		v.totalScroll += v.speed

		v.window.SetScroll(vec2(v.window.Scroll().X, math32.Round(v.origPos+v.totalScroll)))

		imgui.InternalKeepAliveID(v.fakeId)
		imgui.InternalSetActiveID(v.fakeId, v.window)

		scrollDeltas[k] = v
	}

	imgui.NewFrame()
}

func DrawImgui() {
	imgui.Render()

	drawData := imgui.CurrentDrawData()

	if len(drawData.CommandLists()) == 0 {
		return
	}

	rShader.Bind()

	w, h := int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()) //input.Win.GetFramebufferSize()

	rShader.SetUniform("proj", mgl32.Ortho(0, float32(w), float32(h), 0, -1, 1))

	imTextures := drawData.Textures()
	if imTextures.Size() > 0 {
		for _, tex := range imTextures.Slice() {
			if tex.Status() != imgui.TextureStatusOK {
				handleTexture(tex)
			}
		}
	}

	rShader.SetUniform("tex", 0)

	lastBound := imgui.TextureID(0)

	vao.Bind()
	ibo.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.SrcAlpha, blend.OneMinusSrcAlpha)

	_, h1 := gcontext.GetFramebufferSize()

	scaling := float32(h1) / float32(h)

	for _, list := range drawData.CommandLists() {
		vertexBuffer, vertexBufferSize := list.GetVertexBuffer()
		vertexBufferSize /= 4 // convert size in bytes to size in float32

		vertices := unsafe.Slice((*float32)(vertexBuffer), vertexBufferSize) // cast from unsafe to float32 slice

		if vao.GetVBO("default").Capacity() < vertexBufferSize {
			vao.Resize("default", vertexBufferSize) // resize, if necessary
		}

		vao.SetData("default", 0, vertices)

		indexBuffer, indexBufferSize := list.GetIndexBuffer()
		indexBufferSize /= 2

		indices := unsafe.Slice((*uint16)(indexBuffer), indexBufferSize)

		ibo.SetData(0, indices)

		for _, cmd := range list.Commands() {
			if cmd.HasUserCallback() {
				cmd.CallUserCallback(list)
			} else {
				cId := cmd.TexID()
				if cId != lastBound {
					gl.BindTextureUnit(0, uint32(cId))

					lastBound = cId
				}

				clipRect := cmd.ClipRect() //.Times(scaling)
				clipRect.X *= scaling
				clipRect.Y *= scaling
				clipRect.Z *= scaling
				clipRect.W *= scaling

				viewport.PushScissorPos(int(clipRect.X), int(float32(h1)-clipRect.W), int(clipRect.Z-clipRect.X), int(clipRect.W-clipRect.Y))

				ibo.DrawPart(int(cmd.IdxOffset()), int(cmd.ElemCount()))

				viewport.PopScissor()
			}
		}
	}

	blend.Pop()

	ibo.Unbind()
	vao.Unbind()
	rShader.Unbind()
}

var textures = make(map[uint32]*texture.TextureSingle)

func handleTexture(tex imgui.TextureData) {
	if tex.Status() == imgui.TextureStatusWantCreate {
		if tex.TexID() != 0 {
			panic("invalid texture state: want to create existing texture")
		}

		if tex.Format() != imgui.TextureFormatRGBA32 {
			panic("invalid texture state: want to create non rgba32 texture")
		}

		tW := int(tex.Width())
		tH := int(tex.Height())

		// go vet workaround
		data := unsafe.Slice(*(**byte)(unsafe.Pointer(new(tex.Pixels()))), 4*tW*tH)

		gTex := texture.NewTextureSingle(tW, tH, 0)
		gTex.SetData(0, 0, tW, tH, data)

		tex.SetTexID(imgui.TextureID(gTex.GetID()))
		tex.SetStatus(imgui.TextureStatusOK)

		textures[gTex.GetID()] = gTex
	} else if tex.Status() == imgui.TextureStatusWantUpdates {
		texId := uint32(tex.TexID())

		gTex, ok := textures[texId]
		if !ok {
			panic("invalid texture state: missing gl texture")
		}

		for _, rc := range tex.Updates().Slice() {
			uW := int(rc.W())
			uH := int(rc.H())

			rPtr := tex.PixelsAt(int32(rc.X()), int32(rc.Y()))

			gTex.SetDataBuf(int(rc.X()), int(rc.Y()), uW, uH, int(tex.Width()), rPtr)
		}

		tex.SetStatus(imgui.TextureStatusOK)
	} else if tex.Status() == imgui.TextureStatusWantDestroy {
		texId := uint32(tex.TexID())

		gTex, ok := textures[texId]
		if !ok {
			panic("invalid texture state: gl texture already destroyed")
		}

		gTex.Dispose()

		delete(textures, texId)

		tex.SetTexID(imgui.TextureID(0))
		tex.SetStatus(imgui.TextureStatusDestroyed)
	}
}

func handleDragScroll() (ret bool) {
	window := context.CurrentWindow()
	wId := window.ID()

	if imgui.IsMouseDown(imgui.MouseButtonLeft) && (scrCache.cId == 0 || scrCache.cId == wId) {
		_, wasScrolling := scrollDeltas[wId]
		delete(scrollDeltas, wId)

		if scrCache.blocked || isAnyScrollbarActive() {
			return
		}

		intRect := window.InternalRect()

		if scrCache.held || ((&intRect).InternalContainsVec2(ImIO.MousePos()) && imgui.IsWindowHoveredV(imgui.HoveredFlagsAllowWhenBlockedByActiveItem)) {
			ret = true

			if !scrCache.started { // capture the first hold position
				scrCache.started = true
				scrCache.mY = ImIO.MousePos().Y
				scrCache.cId = wId
			}

			if sliderSledLastFrame { // prevent scrolling if slider changed valu
				scrCache.blocked = true
			} else if wasScrolling || math32.Abs(ImIO.MousePos().Y-scrCache.mY) > 5 { // start scrolling if mouse goes over the threshold
				scrCache.held = true
				scrCache.fakeId = imgui.IDStr("#scrollcontainer" + window.Name())
			}

			if scrCache.held {
				imgui.InternalKeepAliveID(scrCache.fakeId)
				imgui.InternalSetActiveID(scrCache.fakeId, window)
			}

			window.SetScroll(vec2(window.Scroll().X, window.Scroll().Y-ImIO.MouseDelta().Y))
		}
	} else if scrCache.cId == wId {
		scrollDeltas[wId] = scrollContainer{window: window, speed: -ImIO.MouseDelta().Y, fakeId: scrCache.fakeId, origPos: window.Scroll().Y}
		scrCache = sCache{}
	}

	return
}

func isAnyScrollbarActive() bool {
	activeWindow := context.ActiveIdWindow()

	return activeWindow.CData != nil && imgui.InternalActiveID() == imgui.InternalWindowScrollbarID(activeWindow, imgui.AxisY)
}

func sliderDoubleV(label string, v *float64, vMin float64, vMax float64, format string, flags imgui.SliderFlags) bool {
	pinner := &runtime.Pinner{}

	ptrV := unsafe.Pointer(v)
	ptrMin := unsafe.Pointer(&vMin)
	ptrMax := unsafe.Pointer(&vMax)

	pinner.Pin(ptrV)
	pinner.Pin(ptrMin)
	pinner.Pin(ptrMax)

	defer func() {
		pinner.Unpin()
	}()

	return imgui.SliderScalarV(label, imgui.DataTypeDouble, uintptr(ptrV), uintptr(ptrMin), uintptr(ptrMax), format, flags)
}
