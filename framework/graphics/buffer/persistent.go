package buffer

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/statistic"
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
	vbo := new(PersistentBufferObject)
	vbo.capacity = maxFloats

	gl.GenBuffers(1, &vbo.handle)

	vbo.Bind()

	gl.BufferStorage(gl.ARRAY_BUFFER, maxFloats*4, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT)

	pt := gl.MapBufferRange(gl.ARRAY_BUFFER, 0, maxFloats*4, gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT|gl.MAP_FLUSH_EXPLICIT_BIT)

	vbo.data = (*[1 << 30]float32)(pt)[:maxFloats:maxFloats]

	vbo.Unbind()

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

	currentVBO := history.GetCurrent(gl.ARRAY_BUFFER_BINDING)
	if currentVBO != vbo.handle {
		panic(fmt.Sprintf("VBO mismatch. Target VBO: %d, current: %d", vbo.handle, currentVBO))
	}

	if offset+len(data) > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, offset: %d, capacity: %d", len(data), offset, vbo.capacity))
	}

	if vbo.data != nil {
		copy(vbo.data[offset:], data)
	}

	gl.BufferSubData(gl.ARRAY_BUFFER, offset*4, len(data)*4, gl.Ptr(data[offset:]))
}

func (vbo *PersistentBufferObject) Resize(newCapacity int) {
	currentVBO := history.GetCurrent(gl.ARRAY_BUFFER_BINDING)
	if currentVBO != vbo.handle {
		panic(fmt.Sprintf("VBO mismatch. Target VBO: %d, current: %d", vbo.handle, currentVBO))
	}

	vbo.capacity = newCapacity

	gl.BufferStorage(gl.ARRAY_BUFFER, newCapacity*4, gl.Ptr(nil), gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT)

	pt := gl.MapBufferRange(gl.ARRAY_BUFFER, 0, newCapacity*4, gl.MAP_PERSISTENT_BIT|gl.MAP_WRITE_BIT|gl.MAP_FLUSH_EXPLICIT_BIT)

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

	currentVBO := history.GetCurrent(gl.ARRAY_BUFFER_BINDING)
	if currentVBO != vbo.handle {
		panic(fmt.Sprintf("VBO mismatch. Target VBO: %d, current: %d", vbo.handle, currentVBO))
	}

	if vbo.offset+offset+size > vbo.capacity {
		panic(fmt.Sprintf("Data exceeds VBO's capacity. Data length: %d, Offset: %d, capacity: %d", size, vbo.offset+offset, vbo.capacity))
	}

	gl.FlushMappedBufferRange(gl.ARRAY_BUFFER, (vbo.offset+offset)*4, size*4)

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

	history.Push(gl.ARRAY_BUFFER_BINDING)

	statistic.Increment(statistic.VBOBinds)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
}

func (vbo *PersistentBufferObject) Unbind() {
	if !vbo.bound || vbo.disposed {
		return
	}

	vbo.bound = false

	handle := history.Pop(gl.ARRAY_BUFFER_BINDING)

	if handle > 0 {
		statistic.Increment(statistic.VBOBinds)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, handle)
}

func (vbo *PersistentBufferObject) Dispose() {
	if !vbo.disposed {
		mainthread.CallNonBlock(func() {
			gl.DeleteBuffers(1, &vbo.handle)
		})
	}

	vbo.disposed = true
}
