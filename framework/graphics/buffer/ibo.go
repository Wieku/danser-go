package buffer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/buffer/history"
	"github.com/wieku/danser-go/framework/statistic"
	"log"
)

type IndexBufferObject struct {
	handle   uint32
	capacity int
}

func NewIndexBufferObject(maxIndices int) *IndexBufferObject {
	ibo := &IndexBufferObject{}
	ibo.capacity = maxIndices
	log.Println(&ibo.handle)
	gl.GenBuffers(1, &ibo.handle)

	ibo.Bind()
	emptyData := make([]uint16, maxIndices)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(emptyData)*2, gl.Ptr(emptyData), gl.DYNAMIC_DRAW)
	ibo.Unbind()

	return ibo
}

func (ibo *IndexBufferObject) Capacity() int {
	return ibo.capacity
}

func (ibo *IndexBufferObject) SetData(offset int, data []uint16) {
	if offset+len(data) > ibo.capacity {
		panic("Data exceeds Index Buffer Object's capacity.")
	}

	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, offset, len(data)*2, gl.Ptr(data))
}

func (ibo *IndexBufferObject) Draw() {
	statistic.Add(statistic.VerticesDrawn, int64(ibo.capacity))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawElements(gl.TRIANGLES, int32(ibo.capacity), gl.UNSIGNED_SHORT, gl.PtrOffset(0))
}

func (ibo *IndexBufferObject) DrawInstanced(baseInstance, instanceCount int) {
	statistic.Add(statistic.VerticesDrawn, int64(ibo.capacity*instanceCount))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawElementsInstancedBaseInstance(gl.TRIANGLES, int32(ibo.capacity), gl.UNSIGNED_SHORT, gl.PtrOffset(0), int32(instanceCount), uint32(baseInstance))
}

func (ibo *IndexBufferObject) DrawPart(offset, length int) {
	if offset+length > ibo.capacity {
		panic("Draw exceeds Index Buffer Object's capacity.")
	}

	statistic.Add(statistic.VerticesDrawn, int64(length))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawElements(gl.TRIANGLES, int32(length), gl.UNSIGNED_SHORT, gl.PtrOffset(offset))
}

func (ibo *IndexBufferObject) DrawPartInstanced(offset, length, baseInstance, instanceCount int) {
	if offset+length > ibo.capacity {
		panic("Draw exceeds Index Buffer Object's capacity.")
	}

	statistic.Add(statistic.VerticesDrawn, int64(length*instanceCount))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawElementsInstancedBaseInstance(gl.TRIANGLES, int32(length), gl.UNSIGNED_SHORT, gl.PtrOffset(offset), int32(instanceCount), uint32(baseInstance))
}

func (ibo *IndexBufferObject) Bind() {
	history.Push(gl.ELEMENT_ARRAY_BUFFER_BINDING)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo.handle)
}

func (ibo *IndexBufferObject) Unbind() {
	handle := history.Pop(gl.ELEMENT_ARRAY_BUFFER_BINDING)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, handle)
}
