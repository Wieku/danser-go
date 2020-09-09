package viewport

import "github.com/go-gl/gl/v3.3-core/gl"

var stack [][4]int32

func Push(width, height int) {
	PushPos(0, 0, width, height)
}

func PushPos(x, y, width, height int) {
	var previous [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &previous[0])
	stack = append(stack, previous)

	gl.Viewport(int32(x), int32(y), int32(width), int32(height))
	gl.Scissor(int32(x), int32(y), int32(width), int32(height))
}

func Pop() {
	viewport := stack[len(stack)-1]
	stack = stack[:len(stack)-1]

	gl.Viewport(viewport[0], viewport[1], viewport[2], viewport[3])
	gl.Scissor(viewport[0], viewport[1], viewport[2], viewport[3])
}

func ClearStack() {
	stack = stack[:0]
}
