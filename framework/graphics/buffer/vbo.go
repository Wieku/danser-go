package buffer

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/statistic"
	"runtime"
)

type VertexBufferObject struct {
	handle   uint32
	capacity int
	bound    bool
	disposed bool
}

func NewVertexBufferObject(maxFloats int, mode DrawMode) *VertexBufferObject {
	vbo := new(VertexBufferObject)
	vbo.capacity = maxFloats

	gl.GenBuffers(1, &vbo.handle)

	vbo.Bind()
	gl.BufferData(gl.ARRAY_BUFFER, maxFloats*4, gl.Ptr(nil), uint32(mode))
	vbo.Unbind()

	runtime.SetFinalizer(vbo, (*VertexBufferObject).Dispose)

	return vbo
}

func (vbo *VertexBufferObject) Capacity() int {
	return vbo.capacity
}

func (vbo *VertexBufferObject) SetData(offset int, data []float32) {
	currentVBO := history.GetCurrent(gl.ARRAY_BUFFER_BINDING)
	if currentVBO != vbo.handle {
		panic(fmt.Sprintf("VBO mismatch. Target VBO: %d, current: %d", vbo.handle, currentVBO))
	}

	if offset+len(data) > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, offset: %d, capacity: %d", len(data), offset, vbo.capacity))
	}

	gl.BufferSubData(gl.ARRAY_BUFFER, offset, len(data)*4, gl.Ptr(data))
}

func (vbo *VertexBufferObject) Bind() {
	if vbo.disposed {
		panic("Can't bind disposed VBO")
	}

	if vbo.bound {
		panic(fmt.Sprintf("VBO %d is already bound", vbo.handle))
	}

	vbo.bound = true

	history.Push(gl.ARRAY_BUFFER)

	statistic.Increment(statistic.VBOBinds)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
}

func (vbo *VertexBufferObject) Unbind() {
	if !vbo.bound || vbo.disposed {
		return
	}

	vbo.bound = false

	handle := history.Pop(gl.ARRAY_BUFFER)

	if handle > 0 {
		statistic.Increment(statistic.VBOBinds)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, handle)
}

func (vbo *VertexBufferObject) Dispose() {
	if !vbo.disposed {
		mainthread.CallNonBlock(func() {
			gl.DeleteBuffers(1, &vbo.handle)
		})
	}

	vbo.disposed = true
}
