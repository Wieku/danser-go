package viewport

import "github.com/go-gl/gl/v3.3-core/gl"

var viewportStack [][4]int32
var scissorStack [][4]int32

func Push(width, height int) {
	PushPos(0, 0, width, height)
}

func PushPos(x, y, width, height int) {
	var previous [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &previous[0])
	viewportStack = append(viewportStack, previous)

	gl.Viewport(int32(x), int32(y), int32(width), int32(height))

	PushScissorPos(x, y, width, height)
}

func PushScissor(width, height int) {
	PushScissorPos(0, 0, width, height)
}

func PushScissorPos(x, y, width, height int) {
	var previous [4]int32
	gl.GetIntegerv(gl.SCISSOR_BOX, &previous[0])
	scissorStack = append(scissorStack, previous)

	gl.Scissor(int32(x), int32(y), int32(width), int32(height))
}

func PopScissor() {
	scissor := scissorStack[len(scissorStack)-1]
	scissorStack = scissorStack[:len(scissorStack)-1]

	gl.Scissor(scissor[0], scissor[1], scissor[2], scissor[3])
}

func Pop() {
	PopScissor()

	viewport := viewportStack[len(viewportStack)-1]
	viewportStack = viewportStack[:len(viewportStack)-1]

	gl.Viewport(viewport[0], viewport[1], viewport[2], viewport[3])
}

func ClearStack() {
	viewportStack = viewportStack[:0]
}
