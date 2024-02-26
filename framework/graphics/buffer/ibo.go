package buffer

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/hacks"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/profiler"
	"runtime"
)

type IndexBufferObject struct {
	handle   uint32
	capacity int
	bound    bool
	disposed bool
	attached bool

	xtype uint32
	xsize int
}

func NewIndexBufferObject(maxIndices int) *IndexBufferObject {
	ibo := new(IndexBufferObject)
	ibo.capacity = maxIndices
	ibo.xtype = gl.UNSIGNED_SHORT
	ibo.xsize = 2

	gl.CreateBuffers(1, &ibo.handle)

	gl.NamedBufferData(ibo.handle, maxIndices*2, gl.Ptr(nil), gl.DYNAMIC_DRAW)

	runtime.SetFinalizer(ibo, (*IndexBufferObject).Dispose)

	return ibo
}

func NewIndexBufferObjectInt(maxIndices int) *IndexBufferObject {
	ibo := new(IndexBufferObject)
	ibo.capacity = maxIndices
	ibo.xtype = gl.UNSIGNED_INT
	ibo.xsize = 4

	gl.CreateBuffers(1, &ibo.handle)

	gl.NamedBufferData(ibo.handle, maxIndices*4, gl.Ptr(nil), gl.DYNAMIC_DRAW)

	runtime.SetFinalizer(ibo, (*IndexBufferObject).Dispose)

	return ibo
}

func (ibo *IndexBufferObject) Capacity() int {
	return ibo.capacity
}

// reflect.TypeOf(s).Elem().Size()
func (ibo *IndexBufferObject) SetData(offset int, data []uint16) {
	if len(data) == 0 {
		return
	}

	if offset+len(data) > ibo.capacity {
		panic(fmt.Sprintf("Data exceeds IBO's capacity. Data length: %d, offset: %d, capacity: %d", len(data), offset, ibo.capacity))
	}

	gl.NamedBufferSubData(ibo.handle, offset, len(data)*2, gl.Ptr(data))
}

func (ibo *IndexBufferObject) SetDataI(offset int, data []uint32) {
	if len(data) == 0 {
		return
	}

	if offset+len(data) > ibo.capacity {
		panic(fmt.Sprintf("Data exceeds IBO's capacity. Data length: %d, offset: %d, capacity: %d", len(data), offset, ibo.capacity))
	}

	gl.NamedBufferSubData(ibo.handle, offset, len(data)*4, gl.Ptr(data))
}

func (ibo *IndexBufferObject) Draw() {
	ibo.DrawPart(0, ibo.capacity)
}

func (ibo *IndexBufferObject) DrawInstanced(baseInstance, instanceCount int) {
	ibo.DrawPartInstanced(0, ibo.capacity, baseInstance, instanceCount)
}

func (ibo *IndexBufferObject) DrawPart(offset, length int) {
	ibo.check(offset, length)

	profiler.AddStat(profiler.VerticesDrawn, int64(length))
	profiler.IncrementStat(profiler.DrawCalls)

	gl.DrawElements(gl.TRIANGLES, int32(length), ibo.xtype, gl.PtrOffset(offset*ibo.xsize))

	if hacks.IsIntel {
		gl.Flush()
	}
}

func (ibo *IndexBufferObject) DrawPartInstanced(offset, length, baseInstance, instanceCount int) {
	ibo.check(offset, length)

	profiler.AddStat(profiler.VerticesDrawn, int64(length*instanceCount))
	profiler.IncrementStat(profiler.DrawCalls)

	gl.DrawElementsInstancedBaseInstance(gl.TRIANGLES, int32(length), ibo.xtype, gl.PtrOffset(offset*ibo.xsize), int32(instanceCount), uint32(baseInstance))

	if hacks.IsIntel {
		gl.Flush()
	}
}

func (ibo *IndexBufferObject) check(offset, length int) {
	if !ibo.attached {
		currentIBO := history.GetCurrent(gl.ELEMENT_ARRAY_BUFFER_BINDING)
		if currentIBO != ibo.handle {
			panic(fmt.Sprintf("IBO mismatch. Target IBO: %d, current: %d", ibo.handle, currentIBO))
		}
	}

	if offset+length > ibo.capacity {
		panic(fmt.Sprintf("Draw exceeds IBO's capacity. Draw length: %d, offset: %d, capacity: %d", length, offset, ibo.capacity))
	}
}

func (ibo *IndexBufferObject) Bind() {
	if ibo.disposed {
		panic("Can't bind disposed IBO")
	}

	if ibo.bound {
		panic(fmt.Sprintf("IBO %d is already bound", ibo.handle))
	}

	ibo.bound = true

	history.Push(gl.ELEMENT_ARRAY_BUFFER_BINDING, ibo.handle)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo.handle)
}

func (ibo *IndexBufferObject) Unbind() {
	if !ibo.bound || ibo.disposed {
		return
	}

	ibo.bound = false

	handle := history.Pop(gl.ELEMENT_ARRAY_BUFFER_BINDING)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, handle)
}

func (ibo *IndexBufferObject) Dispose() {
	if !ibo.disposed {
		goroutines.CallNonBlockMain(func() {
			gl.DeleteBuffers(1, &ibo.handle)
		})
	}

	ibo.disposed = true
}
