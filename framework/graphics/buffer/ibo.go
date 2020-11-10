package buffer

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/statistic"
	"runtime"
)

type IndexBufferObject struct {
	handle   uint32
	capacity int
	bound    bool
	disposed bool
}

func NewIndexBufferObject(maxIndices int) *IndexBufferObject {
	ibo := new(IndexBufferObject)
	ibo.capacity = maxIndices

	gl.GenBuffers(1, &ibo.handle)

	ibo.Bind()
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, maxIndices*2, gl.Ptr(nil), gl.DYNAMIC_DRAW)
	ibo.Unbind()

	runtime.SetFinalizer(ibo, (*IndexBufferObject).Dispose)

	return ibo
}

func (ibo *IndexBufferObject) Capacity() int {
	return ibo.capacity
}

func (ibo *IndexBufferObject) SetData(offset int, data []uint16) {
	if len(data) == 0 {
		return
	}

	ibo.check(offset, len(data), "Data")

	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, offset, len(data)*2, gl.Ptr(data))
}

func (ibo *IndexBufferObject) Draw() {
	ibo.DrawPart(0, ibo.capacity)
}

func (ibo *IndexBufferObject) DrawInstanced(baseInstance, instanceCount int) {
	ibo.DrawPartInstanced(0, ibo.capacity, baseInstance, instanceCount)
}

func (ibo *IndexBufferObject) DrawPart(offset, length int) {
	ibo.check(offset, length, "Draw")

	statistic.Add(statistic.VerticesDrawn, int64(length))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawElements(gl.TRIANGLES, int32(length), gl.UNSIGNED_SHORT, gl.PtrOffset(offset*2))
}

func (ibo *IndexBufferObject) DrawPartInstanced(offset, length, baseInstance, instanceCount int) {
	ibo.check(offset, length, "Draw")

	statistic.Add(statistic.VerticesDrawn, int64(length*instanceCount))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawElementsInstancedBaseInstance(gl.TRIANGLES, int32(length), gl.UNSIGNED_SHORT, gl.PtrOffset(offset), int32(instanceCount), uint32(baseInstance))
}

func (ibo *IndexBufferObject) check(offset, length int, checkTarget string) {
	currentIBO := history.GetCurrent(gl.ELEMENT_ARRAY_BUFFER_BINDING)
	if currentIBO != ibo.handle {
		panic(fmt.Sprintf("IBO mismatch. Target IBO: %d, current: %d", ibo.handle, currentIBO))
	}

	if offset+length > ibo.capacity {
		panic(fmt.Sprintf("%[1]s exceeds IBO's capacity. %[1]s length: %d, offset: %d, capacity: %d", checkTarget, length, offset, ibo.capacity))
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

	history.Push(gl.ELEMENT_ARRAY_BUFFER_BINDING)

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
		mainthread.CallNonBlock(func() {
			gl.DeleteBuffers(1, &ibo.handle)
		})
	}

	ibo.disposed = true
}
