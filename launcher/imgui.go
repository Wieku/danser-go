package launcher

/*
#include <stdlib.h>
*/
import "C"
import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"log"
)

var context *imgui.Context
var ImIO imgui.IO
var tex *texture.TextureSingle
var rShader *shader.RShader
var vao *buffer.VertexArrayObject

var ibo *buffer.IndexBufferObject
var Font16 imgui.Font
var Font20 imgui.Font
var Font24 imgui.Font
var Font32 imgui.Font
var Font48 imgui.Font
var FontAw imgui.Font

func SetupImgui(win *glfw.Window) {
	log.Println("Imgui setup")

	context = imgui.CreateContext(nil)

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarGrabRounding, 5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
	imgui.PushStyleColor(imgui.StyleColorBorder, imgui.Vec4{1, 1, 1, 1})
	imgui.PushStyleColor(imgui.StyleColorFrameBg, imgui.Vec4{0.1, 0.1, 0.1, 0.8})
	imgui.PushStyleColor(imgui.StyleColorFrameBgActive, imgui.Vec4{0.2, 0.2, 0.2, 0.8})
	imgui.PushStyleColor(imgui.StyleColorFrameBgHovered, imgui.Vec4{0.4, 0.4, 0.4, 0.8})

	imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{0.1, 0.1, 0.1, 0.8})
	imgui.PushStyleColor(imgui.StyleColorButtonActive, imgui.Vec4{0.2, 0.2, 0.2, 0.8})
	imgui.PushStyleColor(imgui.StyleColorButtonHovered, imgui.Vec4{0.4, 0.4, 0.4, 0.8})

	imgui.PushStyleColor(imgui.StyleColorTitleBg, imgui.Vec4{0.2, 0.2, 0.2, 0.8})
	imgui.PushStyleColor(imgui.StyleColorTitleBgActive, imgui.Vec4{0.2, 0.2, 0.2, 0.8})
	imgui.PushStyleColor(imgui.StyleColorTitleBgCollapsed, imgui.Vec4{0.2, 0.2, 0.2, 0.8})

	//imgui.PushStyleColor(imgui.StyleColorFrameBg)

	context.SetCurrent()

	ImIO = imgui.CurrentIO()

	ImIO.SetIniFilename("")

	//region texture

	//ImIO.Fonts().AddFontDefault()

	quicksandBytes, err := assets.GetBytes("assets/fonts/Quicksand-Bold.ttf")
	if err != nil {
		panic(err)
	}

	//TODO: switch from multiple fonts to own custom PushFont implementation that sets global scale for each font
	Font16 = ImIO.Fonts().AddFontFromMemoryTTF(quicksandBytes, 16)
	Font20 = ImIO.Fonts().AddFontFromMemoryTTF(quicksandBytes, 20)
	Font24 = ImIO.Fonts().AddFontFromMemoryTTF(quicksandBytes, 24)
	Font32 = ImIO.Fonts().AddFontFromMemoryTTF(quicksandBytes, 32)
	Font48 = ImIO.Fonts().AddFontFromMemoryTTF(quicksandBytes, 48)

	fontAwesomeBytes, err := assets.GetBytes("assets/fonts/Font Awesome 6 Free-Solid-900.otf")
	if err != nil {
		panic(err)
	}

	//fontawesome is quite large so for now we will load only needed glyphs
	awesomeBuilder := &imgui.GlyphRangesBuilder{}
	awesomeBuilder.Add(0xF04B, 0xF04B)
	awesomeBuilder.Add(0xF04D, 0xF04D)
	awesomeBuilder.Add('+', '+')
	awesomeBuilder.Add(0xF068, 0xF068)

	awesomeBuilder.Add(0xF0AD, 0xF0AD)
	awesomeBuilder.Add(0xF108, 0xF108)
	awesomeBuilder.Add(0xF028, 0xF028)
	awesomeBuilder.Add(0xF11C, 0xF11C)
	awesomeBuilder.Add(0xF140, 0xF140)
	awesomeBuilder.Add(0xF53F, 0xF53F)
	awesomeBuilder.Add(0xF245, 0xF245)
	awesomeBuilder.Add(0xF1CD, 0xF1CD)
	awesomeBuilder.Add(0xF853, 0xF853)
	awesomeBuilder.Add(0xF5B7, 0xF5B7)
	awesomeBuilder.Add(0xE599, 0xE599)
	awesomeBuilder.Add(0xF0CB, 0xF0CB)
	awesomeBuilder.Add(0xF03D, 0xF03D)
	awesomeBuilder.Add(0xF882, 0xF882)
	awesomeBuilder.Add(0xF15D, 0xF15D)
	awesomeBuilder.Add(0xF084, 0xF084)

	//awesomeBuilder.Add(0x0020, 0xffff)

	awesomeRange := awesomeBuilder.Build()

	FontAw = ImIO.Fonts().AddFontFromMemoryTTFV(fontAwesomeBytes, 32, imgui.DefaultFontConfig, awesomeRange.GlyphRanges)

	img0 := ImIO.Fonts().TextureDataAlpha8()
	img1 := ImIO.Fonts().TextureDataRGBA32()

	tex = texture.NewTextureSingleFormat(img1.Width, img1.Height, texture.RGBA, 0) // mip-mapping fails miserably because igui doesn't apply padding to sub-textures

	size := img1.Width * img1.Height * 4

	pixels := (*[1 << 30]uint8)(img1.Pixels)[:size:size] // cast from unsafe pointer to uint8 slice

	tex.SetData(0, 0, img1.Width, img1.Height, pixels)

	C.free(img0.Pixels) //Reduce some memory, seems that imgui doesn't explode
	C.free(img1.Pixels) //Reduce some memory, seems that imgui doesn't explode

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
		ImIO.AddInputCharacters(string(char))
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
	ImIO.AddKeyEvent(imgui.KeyModCtrl, (mods&glfw.ModControl) != 0)
	ImIO.AddKeyEvent(imgui.KeyModShift, (mods&glfw.ModShift) != 0)
	ImIO.AddKeyEvent(imgui.KeyModAlt, (mods&glfw.ModAlt) != 0)
	ImIO.AddKeyEvent(imgui.KeyModSuper, (mods&glfw.ModSuper) != 0)
}

var lastTime float64

func Begin() {
	x, y := input.Win.GetCursorPos()

	ImIO.AddMousePosEvent(imgui.Vec2{X: float32(x), Y: float32(y)})
	ImIO.AddMouseButtonEvent(0, input.Win.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press)
	ImIO.AddMouseButtonEvent(1, input.Win.GetMouseButton(glfw.MouseButtonRight) == glfw.Press)

	w, h := input.Win.GetFramebufferSize()

	ImIO.SetDisplaySize(imgui.Vec2{X: float32(w), Y: float32(h)})

	time := glfw.GetTime()

	delta := float32(time - lastTime)

	lastTime = time

	ImIO.SetDeltaTime(delta)

	imgui.NewFrame()
}

func DrawImgui() {
	imgui.Render()

	drawData := imgui.RenderedDrawData()

	if len(drawData.CommandLists()) == 0 {
		return
	}

	rShader.Bind()

	w, h := input.Win.GetFramebufferSize()

	rShader.SetUniform("proj", mgl32.Ortho(0, float32(w), float32(h), 0, -1, 1))

	tex.Bind(0)
	rShader.SetUniform("tex", 0)

	lastBound := imgui.TextureID(0)

	vao.Bind()
	ibo.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.SrcAlpha, blend.OneMinusSrcAlpha)

	for _, list := range drawData.CommandLists() {
		vertexBuffer, vertexBufferSize := list.VertexBuffer()
		vertexBufferSize /= 4 // convert size in bytes to size in float32

		vertices := (*[1 << 30]float32)(vertexBuffer)[:vertexBufferSize:vertexBufferSize] // cast from unsafe to float32 slice

		if vao.GetVBO("default").Capacity() < vertexBufferSize {
			vao.Resize("default", vertexBufferSize) // resize, if necessary
		}

		vao.SetData("default", 0, vertices)

		indexBuffer, indexBufferSize := list.IndexBuffer()
		indexBufferSize /= 2

		indices := (*[1 << 30]uint16)(indexBuffer)[:indexBufferSize:indexBufferSize]

		ibo.SetData(0, indices)

		for _, cmd := range list.Commands() {
			cId := cmd.TextureID()
			if cId != lastBound {
				if cId == 0 {
					tex.Bind(0)
				} else {
					gl.BindTextureUnit(0, uint32(cId))
				}

				lastBound = cId
			}

			if cmd.HasUserCallback() {
				cmd.CallUserCallback(list)
			} else {
				clipRect := cmd.ClipRect()
				viewport.PushScissorPos(int(clipRect.X), int(settings.Graphics.GetHeight())-int(clipRect.W), int(clipRect.Z-clipRect.X), int(clipRect.W-clipRect.Y))

				ibo.DrawPart(cmd.IndexOffset(), cmd.ElementCount())

				viewport.PopScissor()
			}
		}
	}

	blend.Pop()

	ibo.Unbind()
	vao.Unbind()
	rShader.Unbind()
}
