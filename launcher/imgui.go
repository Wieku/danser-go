package launcher

/*
#include <stdlib.h>
*/
import "C"
import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/math32"
	"log"
	"runtime"
	"unsafe"
)

var context *imgui.Context
var ImIO *imgui.IO
var tex *texture.TextureSingle
var rShader *shader.RShader
var vao *buffer.VertexArrayObject

var ibo *buffer.IndexBufferObject
var Font16 *imgui.Font
var Font20 *imgui.Font
var Font24 *imgui.Font
var Font32 *imgui.Font
var Font48 *imgui.Font
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

func SetupImgui(win *glfw.Window) {
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

	//region texture

	quicksandBytes, err := assets.GetBytes("assets/fonts/Quicksand-Bold.ttf")
	if err != nil {
		panic(err)
	}

	quicksandPtr := unsafe.Pointer(&quicksandBytes[0])

	//TODO: switch from multiple fonts to own custom PushFont implementation that sets global scale for each font
	Font16 = ImIO.Fonts().AddFontFromMemoryTTF(uintptr(quicksandPtr), int32(len(quicksandBytes)), 16)
	Font20 = ImIO.Fonts().AddFontFromMemoryTTF(uintptr(quicksandPtr), int32(len(quicksandBytes)), 20)
	Font24 = ImIO.Fonts().AddFontFromMemoryTTF(uintptr(quicksandPtr), int32(len(quicksandBytes)), 24)
	Font32 = ImIO.Fonts().AddFontFromMemoryTTF(uintptr(quicksandPtr), int32(len(quicksandBytes)), 32)
	Font48 = ImIO.Fonts().AddFontFromMemoryTTF(uintptr(quicksandPtr), int32(len(quicksandBytes)), 48)

	fontAwesomeBytes, err := assets.GetBytes("assets/fonts/Font Awesome 6 Free-Solid-900.otf")
	if err != nil {
		panic(err)
	}

	awesomePtr := unsafe.Pointer(&fontAwesomeBytes[0])

	//fontawesome is quite large so for now we will load only needed glyphs
	awesomeBuilder := imgui.NewFontGlyphRangesBuilder()
	awesomeBuilder.AddChar(0xF04B) // play
	awesomeBuilder.AddChar(0xF04D) // stop
	awesomeBuilder.AddChar('+')
	awesomeBuilder.AddChar(0xF068) // minus
	awesomeBuilder.AddChar(0xF0AD) // wrench
	awesomeBuilder.AddChar(0xE163) // display
	awesomeBuilder.AddChar(0xF028) // volume-high
	awesomeBuilder.AddChar(0xF11C) // keyboard
	awesomeBuilder.AddChar(0xF245) // arrow-pointer
	awesomeBuilder.AddChar(0xE599) // worm
	awesomeBuilder.AddChar(0xF0CB) // list-ol
	awesomeBuilder.AddChar(0xF03D) // video
	awesomeBuilder.AddChar(0xF882) // arrow-up-z-a
	awesomeBuilder.AddChar(0xF15D) // arrow-down-a-z
	awesomeBuilder.AddChar(0xF084) // key
	awesomeBuilder.AddChar(0xF7A2) // earth-europe
	awesomeBuilder.AddChar(0xF192) // circle-dot
	awesomeBuilder.AddChar(0xF1E0) // share-nodes
	awesomeBuilder.AddChar(0xF1FC) // paintbrush
	awesomeBuilder.AddChar(0xF43C) // chess-board
	awesomeBuilder.AddChar(0xF188) // bug
	awesomeBuilder.AddChar(0xF2EA) // rotate-left

	awesomeRange := imgui.NewGlyphRange()
	awesomeBuilder.BuildRanges(awesomeRange)

	FontAw = ImIO.Fonts().AddFontFromMemoryTTFV(uintptr(awesomePtr), int32(len(fontAwesomeBytes)), 32, imgui.NewFontConfig(), awesomeRange.Data())

	img0, w0, h0, _ := ImIO.Fonts().TextureDataAsAlpha8()
	img1, _, _, _ := ImIO.Fonts().GetTextureDataAsRGBA32()

	runtime.KeepAlive(quicksandPtr)
	runtime.KeepAlive(awesomePtr)

	tex = texture.NewTextureSingleFormat(int(w0), int(h0), texture.Red, 0) // mip-mapping fails miserably because igui doesn't apply padding to sub-textures

	size := w0 * h0

	pixels := (*[1 << 30]uint8)(img0)[:size:size] // cast from unsafe pointer to uint8 slice

	tex.SetData(0, 0, int(w0), int(h0), pixels)

	C.free(img0) //Reduce some memory, seems that imgui doesn't explode
	C.free(img1) //Reduce some memory, seems that imgui doesn't explode

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

	win.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
		ImIO.AddMouseWheelDelta(float32(xoff), float32(yoff))
	})

	input.Win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		input.CallListeners(w, key, scancode, action, mods)

		if action != glfw.Press && action != glfw.Release {
			return
		}

		if nMods := keyToModifier(key); nMods > 0 {
			if action == glfw.Press {
				mods = mods | nMods
			} else {
				mods = mods & (^nMods)
			}
		}

		updateKeyModifiers(mods)

		iKey := glfwKeyToImGuiKey(key)

		ImIO.AddKeyEvent(iKey, action == glfw.Press)
	})

	input.Win.SetCharCallback(func(w *glfw.Window, char rune) {
		ImIO.AddInputCharactersUTF8(string(char))
	})
}

func glfwKeyToImGuiKey(k glfw.Key) imgui.Key {
	switch k {
	case glfw.KeyTab:
		return imgui.KeyTab
	case glfw.KeyLeft:
		return imgui.KeyLeftArrow
	case glfw.KeyRight:
		return imgui.KeyRightArrow
	case glfw.KeyUp:
		return imgui.KeyUpArrow
	case glfw.KeyDown:
		return imgui.KeyDownArrow
	case glfw.KeyPageUp:
		return imgui.KeyPageUp
	case glfw.KeyPageDown:
		return imgui.KeyPageDown
	case glfw.KeyHome:
		return imgui.KeyHome
	case glfw.KeyEnd:
		return imgui.KeyEnd
	case glfw.KeyInsert:
		return imgui.KeyInsert
	case glfw.KeyDelete:
		return imgui.KeyDelete
	case glfw.KeyBackspace:
		return imgui.KeyBackspace
	case glfw.KeySpace:
		return imgui.KeySpace
	case glfw.KeyEnter:
		return imgui.KeyEnter
	case glfw.KeyEscape:
		return imgui.KeyEscape
	case glfw.KeyApostrophe:
		return imgui.KeyApostrophe
	case glfw.KeyComma:
		return imgui.KeyComma
	case glfw.KeyMinus:
		return imgui.KeyMinus
	case glfw.KeyPeriod:
		return imgui.KeyPeriod
	case glfw.KeySlash:
		return imgui.KeySlash
	case glfw.KeySemicolon:
		return imgui.KeySemicolon
	case glfw.KeyEqual:
		return imgui.KeyEqual
	case glfw.KeyLeftBracket:
		return imgui.KeyLeftBracket
	case glfw.KeyBackslash:
		return imgui.KeyBackslash
	case glfw.KeyRightBracket:
		return imgui.KeyRightBracket
	case glfw.KeyGraveAccent:
		return imgui.KeyGraveAccent
	case glfw.KeyCapsLock:
		return imgui.KeyCapsLock
	case glfw.KeyScrollLock:
		return imgui.KeyScrollLock
	case glfw.KeyNumLock:
		return imgui.KeyNumLock
	case glfw.KeyPrintScreen:
		return imgui.KeyPrintScreen
	case glfw.KeyPause:
		return imgui.KeyPause
	case glfw.KeyKP0:
		return imgui.KeyKeypad0
	case glfw.KeyKP1:
		return imgui.KeyKeypad1
	case glfw.KeyKP2:
		return imgui.KeyKeypad2
	case glfw.KeyKP3:
		return imgui.KeyKeypad3
	case glfw.KeyKP4:
		return imgui.KeyKeypad4
	case glfw.KeyKP5:
		return imgui.KeyKeypad5
	case glfw.KeyKP6:
		return imgui.KeyKeypad6
	case glfw.KeyKP7:
		return imgui.KeyKeypad7
	case glfw.KeyKP8:
		return imgui.KeyKeypad8
	case glfw.KeyKP9:
		return imgui.KeyKeypad9
	case glfw.KeyKPDecimal:
		return imgui.KeyKeypadDecimal
	case glfw.KeyKPDivide:
		return imgui.KeyKeypadDivide
	case glfw.KeyKPMultiply:
		return imgui.KeyKeypadMultiply
	case glfw.KeyKPSubtract:
		return imgui.KeyKeypadSubtract
	case glfw.KeyKPAdd:
		return imgui.KeyKeypadAdd
	case glfw.KeyKPEnter:
		return imgui.KeyKeypadEnter
	case glfw.KeyKPEqual:
		return imgui.KeyKeypadEqual
	case glfw.KeyLeftShift:
		return imgui.KeyLeftShift
	case glfw.KeyLeftControl:
		return imgui.KeyLeftCtrl
	case glfw.KeyLeftAlt:
		return imgui.KeyLeftAlt
	case glfw.KeyLeftSuper:
		return imgui.KeyLeftSuper
	case glfw.KeyRightShift:
		return imgui.KeyRightShift
	case glfw.KeyRightControl:
		return imgui.KeyRightCtrl
	case glfw.KeyRightAlt:
		return imgui.KeyRightAlt
	case glfw.KeyRightSuper:
		return imgui.KeyRightSuper
	case glfw.KeyMenu:
		return imgui.KeyMenu
	case glfw.Key0:
		return imgui.Key0
	case glfw.Key1:
		return imgui.Key1
	case glfw.Key2:
		return imgui.Key2
	case glfw.Key3:
		return imgui.Key3
	case glfw.Key4:
		return imgui.Key4
	case glfw.Key5:
		return imgui.Key5
	case glfw.Key6:
		return imgui.Key6
	case glfw.Key7:
		return imgui.Key7
	case glfw.Key8:
		return imgui.Key8
	case glfw.Key9:
		return imgui.Key9
	case glfw.KeyA:
		return imgui.KeyA
	case glfw.KeyB:
		return imgui.KeyB
	case glfw.KeyC:
		return imgui.KeyC
	case glfw.KeyD:
		return imgui.KeyD
	case glfw.KeyE:
		return imgui.KeyE
	case glfw.KeyF:
		return imgui.KeyF
	case glfw.KeyG:
		return imgui.KeyG
	case glfw.KeyH:
		return imgui.KeyH
	case glfw.KeyI:
		return imgui.KeyI
	case glfw.KeyJ:
		return imgui.KeyJ
	case glfw.KeyK:
		return imgui.KeyK
	case glfw.KeyL:
		return imgui.KeyL
	case glfw.KeyM:
		return imgui.KeyM
	case glfw.KeyN:
		return imgui.KeyN
	case glfw.KeyO:
		return imgui.KeyO
	case glfw.KeyP:
		return imgui.KeyP
	case glfw.KeyQ:
		return imgui.KeyQ
	case glfw.KeyR:
		return imgui.KeyR
	case glfw.KeyS:
		return imgui.KeyS
	case glfw.KeyT:
		return imgui.KeyT
	case glfw.KeyU:
		return imgui.KeyU
	case glfw.KeyV:
		return imgui.KeyV
	case glfw.KeyW:
		return imgui.KeyW
	case glfw.KeyX:
		return imgui.KeyX
	case glfw.KeyY:
		return imgui.KeyY
	case glfw.KeyZ:
		return imgui.KeyZ
	case glfw.KeyF1:
		return imgui.KeyF1
	case glfw.KeyF2:
		return imgui.KeyF2
	case glfw.KeyF3:
		return imgui.KeyF3
	case glfw.KeyF4:
		return imgui.KeyF4
	case glfw.KeyF5:
		return imgui.KeyF5
	case glfw.KeyF6:
		return imgui.KeyF6
	case glfw.KeyF7:
		return imgui.KeyF7
	case glfw.KeyF8:
		return imgui.KeyF8
	case glfw.KeyF9:
		return imgui.KeyF9
	case glfw.KeyF10:
		return imgui.KeyF10
	case glfw.KeyF11:
		return imgui.KeyF11
	case glfw.KeyF12:
		return imgui.KeyF12
	default:
		return imgui.KeyNone
	}
}

func keyToModifier(key glfw.Key) glfw.ModifierKey {
	switch {
	case key == glfw.KeyLeftControl || key == glfw.KeyRightControl:
		return glfw.ModControl
	case key == glfw.KeyLeftShift || key == glfw.KeyRightShift:
		return glfw.ModShift
	case key == glfw.KeyLeftAlt || key == glfw.KeyRightAlt:
		return glfw.ModAlt
	case key == glfw.KeyLeftSuper || key == glfw.KeyRightSuper:
		return glfw.ModSuper
	}

	return 0
}

func updateKeyModifiers(mods glfw.ModifierKey) {
	ImIO.AddKeyEvent(imgui.KeyReservedForModCtrl, (mods&glfw.ModControl) != 0)
	ImIO.AddKeyEvent(imgui.KeyReservedForModShift, (mods&glfw.ModShift) != 0)
	ImIO.AddKeyEvent(imgui.KeyReservedForModAlt, (mods&glfw.ModAlt) != 0)
	ImIO.AddKeyEvent(imgui.KeyReservedForModSuper, (mods&glfw.ModSuper) != 0)
}

var lastTime float64

func Begin() {
	sliderSledLastFrame = sliderSledThisFrame
	sliderSledThisFrame = false

	x, y := input.Win.GetCursorPos()

	w, h := int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()) //input.Win.GetFramebufferSize()
	_, h1 := glfw.GetCurrentContext().GetFramebufferSize()

	scaling := float32(h1) / float32(h)

	ImIO.AddMousePosEvent(float32(x)/scaling, float32(y)/scaling)
	ImIO.AddMouseButtonEvent(0, input.Win.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press)
	ImIO.AddMouseButtonEvent(1, input.Win.GetMouseButton(glfw.MouseButtonRight) == glfw.Press)

	ImIO.SetDisplaySize(imgui.Vec2{X: float32(w), Y: float32(h)})

	time := glfw.GetTime()

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

	tex.Bind(0)
	rShader.SetUniform("tex", 0)
	rShader.SetUniform("texRGBA", 0)

	lastBound := imgui.TextureID{Data: 0}

	vao.Bind()
	ibo.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.SrcAlpha, blend.OneMinusSrcAlpha)

	_, h1 := glfw.GetCurrentContext().GetFramebufferSize()

	scaling := float32(h1) / float32(h)

	for _, list := range drawData.CommandLists() {
		vertexBuffer, vertexBufferSize := list.GetVertexBuffer()
		vertexBufferSize /= 4 // convert size in bytes to size in float32

		vertices := (*[1 << 30]float32)(vertexBuffer)[:vertexBufferSize:vertexBufferSize] // cast from unsafe to float32 slice

		if vao.GetVBO("default").Capacity() < vertexBufferSize {
			vao.Resize("default", vertexBufferSize) // resize, if necessary
		}

		vao.SetData("default", 0, vertices)

		indexBuffer, indexBufferSize := list.GetIndexBuffer()
		indexBufferSize /= 2

		indices := (*[1 << 30]uint16)(indexBuffer)[:indexBufferSize:indexBufferSize]

		ibo.SetData(0, indices)

		for _, cmd := range list.Commands() {
			cId := cmd.TextureId()
			if cId != lastBound {
				if cId.Data == 0 {
					rShader.SetUniform("texRGBA", 0)
					tex.Bind(0)
				} else {
					rShader.SetUniform("texRGBA", 1)
					gl.BindTextureUnit(0, uint32(cId.Data))
				}

				lastBound = cId
			}

			if cmd.HasUserCallback() {
				cmd.CallUserCallback(list)
			} else {
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
