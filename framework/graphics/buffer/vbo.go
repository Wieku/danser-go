package buffer

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/profiler"
	"runtime"
)

type VertexBufferObject struct {
	handle   uint32
	capacity int
	bound    bool
	disposed bool
	data     []float32
	mode     DrawMode
}

func NewVertexBufferObject(maxFloats int, mapped bool, mode DrawMode) *VertexBufferObject {
	vbo := new(VertexBufferObject)
	vbo.capacity = maxFloats
	vbo.mode = mode

	gl.CreateBuffers(1, &vbo.handle)

	gl.NamedBufferData(vbo.handle, maxFloats*4, gl.Ptr(nil), uint32(mode))

	if mapped {
		vbo.data = make([]float32, maxFloats)
	}

	runtime.SetFinalizer(vbo, (*VertexBufferObject).Dispose)

	return vbo
}

func (vbo *VertexBufferObject) Capacity() int {
	return vbo.capacity
}

func (vbo *VertexBufferObject) SetData(offset int, data []float32) {
	if len(data) == 0 {
		return
	}

	if offset+len(data) > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, offset: %d, capacity: %d", len(data), offset, vbo.capacity))
	}

	if vbo.data != nil {
		copy(vbo.data[offset:], data)
	}

	gl.NamedBufferSubData(vbo.handle, offset*4, len(data)*4, gl.Ptr(data))
	//gl.Flush()
}

func (vbo *VertexBufferObject) Resize(newCapacity int) {
	if vbo.data != nil {
		data := make([]float32, newCapacity)
		copy(data, vbo.data[:min(vbo.capacity, newCapacity)])
		vbo.data = data
	}

	vbo.capacity = newCapacity

	if vbo.data != nil && len(vbo.data) > 0 {
		gl.NamedBufferData(vbo.handle, newCapacity*4, gl.Ptr(vbo.data), uint32(vbo.mode))
	} else {
		gl.NamedBufferData(vbo.handle, newCapacity*4, gl.Ptr(nil), uint32(vbo.mode))
	}
}

func (vbo *VertexBufferObject) Map(size int) MemoryChunk {
	if vbo.data == nil {
		panic("Can't map not mapped VBO")
	}

	if size > vbo.capacity {
		panic(fmt.Sprintf("Data request exceeds VBO's capacity. Requested size: %d, capacity: %d", size, vbo.capacity))
	}

	return MemoryChunk{
		Offset: 0,
		Data:   vbo.data[:size],
	}
}

func (vbo *VertexBufferObject) Unmap(offset, size int) {
	if vbo.data == nil {
		panic("Can't unmap not mapped VBO")
	}

	if size == 0 {
		return
	}

	if offset+size > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, Offset: %d, capacity: %d", size, offset, vbo.capacity))
	}

	gl.NamedBufferSubData(vbo.handle, offset*4, size*4, gl.Ptr(vbo.data[offset:]))
}

func (vbo *VertexBufferObject) Bind() {
	if vbo.disposed {
		panic("Can't bind disposed VBO")
	}

	if vbo.bound {
		panic(fmt.Sprintf("VBO %d is already bound", vbo.handle))
	}

	vbo.bound = true

	history.Push(gl.ARRAY_BUFFER_BINDING, vbo.handle)

	profiler.IncrementStat(profiler.VBOBinds)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
}

func (vbo *VertexBufferObject) Unbind() {
	if !vbo.bound || vbo.disposed {
		return
	}

	vbo.bound = false

	handle := history.Pop(gl.ARRAY_BUFFER_BINDING)

	if handle > 0 {
		profiler.IncrementStat(profiler.VBOBinds)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, handle)
}

func (vbo *VertexBufferObject) Dispose() {
	if !vbo.disposed {
		goroutines.CallNonBlockMain(func() {
			gl.DeleteBuffers(1, &vbo.handle)
		})
	}

	vbo.disposed = true
}

func (vbo *VertexBufferObject) GetID() uint32 {
	return vbo.handle
}
