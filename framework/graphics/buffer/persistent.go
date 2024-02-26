package buffer

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/profiler"
	"runtime"
)

type PersistentBufferObject struct {
	handle   uint32
	capacity int
	bound    bool
	disposed bool
	offset   int
	data     []float32
}

func NewPersistentBufferObject(maxFloats int) *PersistentBufferObject {
	if !glfw.ExtensionSupported("GL_ARB_buffer_storage") {
		panic("Your GPU does not support one or more required OpenGL extensions: [GL_ARB_buffer_storage]. Please update your graphics drivers or upgrade your GPU.")
	}

	vbo := new(PersistentBufferObject)
	vbo.capacity = maxFloats

	gl.CreateBuffers(1, &vbo.handle)

	gl.NamedBufferStorage(vbo.handle, maxFloats*4, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT|gl.MAP_COHERENT_BIT)

	pt := gl.MapNamedBufferRange(vbo.handle, 0, maxFloats*4, gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT|gl.MAP_COHERENT_BIT)

	vbo.data = (*[1 << 30]float32)(pt)[:maxFloats:maxFloats]

	runtime.SetFinalizer(vbo, (*PersistentBufferObject).Dispose)

	return vbo
}

func (vbo *PersistentBufferObject) Capacity() int {
	return vbo.capacity
}

func (vbo *PersistentBufferObject) SetData(offset int, data []float32) {
	if len(data) == 0 {
		return
	}

	if offset+len(data) > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, offset: %d, capacity: %d", len(data), offset, vbo.capacity))
	}

	if vbo.data != nil {
		copy(vbo.data[offset:], data)
	}

	gl.NamedBufferSubData(vbo.handle, offset*4, len(data)*4, gl.Ptr(data[offset:]))
}

func (vbo *PersistentBufferObject) Resize(newCapacity int) {
	vbo.capacity = newCapacity

	gl.DeleteBuffers(1, &vbo.handle)

	gl.CreateBuffers(1, &vbo.handle)

	gl.NamedBufferStorage(vbo.handle, newCapacity*4, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT|gl.MAP_COHERENT_BIT)

	pt := gl.MapNamedBufferRange(vbo.handle, 0, newCapacity*4, gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT|gl.MAP_COHERENT_BIT)

	vbo.data = (*[1 << 30]float32)(pt)[:newCapacity:newCapacity]

	vbo.offset = 0
}

func (vbo *PersistentBufferObject) Map(size int) MemoryChunk {
	if size > vbo.capacity {
		panic(fmt.Sprintf("Data request exceeds VBO's capacity. Requested size: %d, capacity: %d", size, vbo.capacity))
	}

	if vbo.offset+size >= vbo.capacity {
		fence := gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)
		gl.ClientWaitSync(fence, gl.SYNC_FLUSH_COMMANDS_BIT, gl.TIMEOUT_IGNORED)
		gl.DeleteSync(fence)

		vbo.offset = 0
	}

	return MemoryChunk{Offset: vbo.offset, Data: vbo.data[vbo.offset : vbo.offset+size]}
}

func (vbo *PersistentBufferObject) Unmap(offset, size int) {
	if size == 0 {
		return
	}

	if vbo.offset+offset+size > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, Offset: %d, capacity: %d", size, vbo.offset+offset, vbo.capacity))
	}

	vbo.offset += offset + size
}

func (vbo *PersistentBufferObject) Bind() {
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

func (vbo *PersistentBufferObject) Unbind() {
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

func (vbo *PersistentBufferObject) Dispose() {
	if !vbo.disposed {
		goroutines.CallNonBlockMain(func() {
			gl.DeleteBuffers(1, &vbo.handle)
		})
	}

	vbo.disposed = true
}

func (vbo *PersistentBufferObject) GetID() uint32 {
	return vbo.handle
}
