package buffer

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/hacks"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/profiler"
	"runtime"
)

type bufferHolder struct {
	buffer  StreamingBuffer
	divisor int
	format  attribute.Format
	binding int
}

type VertexArrayObject struct {
	handle uint32

	buffers map[string]*bufferHolder

	capacity int

	bound    bool
	disposed bool

	ibo *IndexBufferObject
}

func NewVertexArrayObject() *VertexArrayObject {
	vao := new(VertexArrayObject)
	vao.buffers = make(map[string]*bufferHolder)

	gl.CreateVertexArrays(1, &vao.handle)

	runtime.SetFinalizer(vao, (*VertexArrayObject).Dispose)

	return vao
}

func (vao *VertexArrayObject) AddVBO(name string, maxVertices int, divisor int, format attribute.Format) {
	vao.addVBO(name, maxVertices, divisor, false, format)
}

func (vao *VertexArrayObject) AddMappedVBO(name string, maxVertices int, divisor int, format attribute.Format) {
	vao.addVBO(name, maxVertices, divisor, true, format)
}

func (vao *VertexArrayObject) addVBO(name string, maxVertices int, divisor int, mapped bool, format attribute.Format) {
	if _, exists := vao.buffers[name]; exists {
		panic(fmt.Sprintf("VBO with name \"%s\" already exists", name))
	}

	holder := &bufferHolder{
		buffer:  NewVertexBufferObject(maxVertices*format.Size()/4, mapped, DynamicDraw),
		divisor: divisor,
		format:  format,
		binding: -1,
	}

	if divisor == 0 {
		vao.capacity = maxVertices
	}

	vao.buffers[name] = holder
}

func (vao *VertexArrayObject) AddPersistentVBO(name string, maxVertices int, divisor int, format attribute.Format) {
	if _, exists := vao.buffers[name]; exists {
		panic(fmt.Sprintf("VBO with name \"%s\" already exists", name))
	}

	holder := &bufferHolder{
		buffer:  NewPersistentBufferObject(maxVertices / 100 * format.Size() / 4),
		divisor: divisor,
		format:  format,
	}

	holder.buffer.Resize(maxVertices * format.Size() / 4)

	if divisor == 0 {
		vao.capacity = maxVertices
	}

	vao.buffers[name] = holder
}

func (vao *VertexArrayObject) GetVBOFormat(name string) attribute.Format {
	if holder, exists := vao.buffers[name]; exists {
		return holder.format
	}

	panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
}

func (vao *VertexArrayObject) GetVBO(name string) StreamingBuffer {
	if holder, exists := vao.buffers[name]; exists {
		return holder.buffer
	}

	panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
}

func (vao *VertexArrayObject) Resize(name string, maxVertices int) {
	if holder, exists := vao.buffers[name]; exists {
		size := maxVertices * holder.format.Size() / 4
		if holder.buffer.Capacity() != size {
			holder.buffer.Resize(size)

			// If we have persistent buffer object that was bound we want to bind it again because new object was created on resize
			if _, ok := holder.buffer.(*PersistentBufferObject); ok && holder.binding >= 0 {
				gl.VertexArrayVertexBuffer(vao.handle, uint32(holder.binding), holder.buffer.GetID(), 0, int32(holder.format.Size()))
			}
		}

		return
	}

	panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
}

func (vao *VertexArrayObject) Attach(s *shader.RShader) {
	var index int
	for _, holder := range vao.buffers {
		var offset int
		for _, attr := range holder.format {
			location := s.GetAttributeInfo(attr.Name).Location

			gl.EnableVertexArrayAttrib(vao.handle, uint32(location))

			gl.VertexArrayAttribBinding(vao.handle, uint32(location), uint32(index))

			gl.VertexArrayAttribFormat(
				vao.handle,
				uint32(location),
				int32(attr.Type.Components()),
				uint32(attr.Type.InternalType()),
				attr.Type.Normalize(),
				uint32(offset),
			)

			offset += attr.Type.Size()
		}

		holder.binding = index
		gl.VertexArrayVertexBuffer(vao.handle, uint32(index), holder.buffer.GetID(), 0, int32(holder.format.Size()))
		gl.VertexArrayBindingDivisor(vao.handle, uint32(index), uint32(holder.divisor))

		index++
	}
}

func (vao *VertexArrayObject) SetData(name string, offset int, data []float32) {
	holder, exists := vao.buffers[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	holder.buffer.SetData(offset, data)

	profiler.AddStat(profiler.VertexUpload, int64(len(data)*4/holder.format.Size()))
}

func (vao *VertexArrayObject) MapVBO(name string, size int) MemoryChunk {
	holder, exists := vao.buffers[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	return holder.buffer.Map(size)
}

func (vao *VertexArrayObject) UnmapVBO(name string, offset int, size int) {
	holder, exists := vao.buffers[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	holder.buffer.Unmap(offset, size)

	profiler.AddStat(profiler.VertexUpload, int64(size*4/holder.format.Size()))
}

func (vao *VertexArrayObject) Draw() {
	if vao.ibo != nil {
		vao.check(0, 0)
		vao.ibo.Draw()
	} else {
		vao.DrawPart(0, vao.capacity)
	}
}

func (vao *VertexArrayObject) DrawInstanced(baseInstance, instanceCount int) {
	if vao.ibo != nil {
		vao.check(0, 0)
		vao.ibo.DrawInstanced(baseInstance, instanceCount)
	} else {
		vao.DrawPartInstanced(0, vao.capacity, baseInstance, instanceCount)
	}
}

func (vao *VertexArrayObject) DrawPart(offset, length int) {
	if vao.ibo != nil {
		vao.check(0, 0)
		vao.ibo.DrawPart(offset, length)
	} else {
		vao.check(offset, length)

		profiler.AddStat(profiler.VerticesDrawn, int64(length))
		profiler.IncrementStat(profiler.DrawCalls)

		gl.DrawArrays(gl.TRIANGLES, int32(offset), int32(length))

		if hacks.IsIntel {
			gl.Flush()
		}
	}
}

func (vao *VertexArrayObject) DrawPartInstanced(offset, length, baseInstance, instanceCount int) {
	if vao.ibo != nil {
		vao.check(0, 0)
		vao.ibo.DrawPartInstanced(offset, length, baseInstance, instanceCount)
	} else {
		vao.check(offset, length)

		profiler.AddStat(profiler.VerticesDrawn, int64(length*instanceCount))
		profiler.IncrementStat(profiler.DrawCalls)

		gl.DrawArraysInstancedBaseInstance(gl.TRIANGLES, int32(offset), int32(length), int32(instanceCount), uint32(baseInstance))

		if hacks.IsIntel {
			gl.Flush()
		}
	}
}

func (vao *VertexArrayObject) check(offset, length int) {
	currentVAO := history.GetCurrent(gl.VERTEX_ARRAY_BINDING)
	if currentVAO != vao.handle {
		panic(fmt.Sprintf("VAO mismatch. Target VAO: %d, current: %d", vao.handle, currentVAO))
	}

	if offset+length > vao.capacity {
		panic(fmt.Sprintf("Draw exceeds VAO's capacity. Draw length: %d, offset: %d, capacity: %d", length, offset, vao.capacity))
	}
}

func (vao *VertexArrayObject) Bind() {
	if vao.disposed {
		panic("Can't bind disposed VAO")
	}

	if vao.bound {
		panic(fmt.Sprintf("VAO %d is already bound", vao.handle))
	}

	vao.bound = true

	history.Push(gl.VERTEX_ARRAY_BINDING, vao.handle)

	profiler.IncrementStat(profiler.VAOBinds)

	gl.BindVertexArray(vao.handle)
}

func (vao *VertexArrayObject) Unbind() {
	if !vao.bound || vao.disposed {
		return
	}

	vao.bound = false

	handle := history.Pop(gl.VERTEX_ARRAY_BINDING)

	if handle > 0 {
		profiler.IncrementStat(profiler.VAOBinds)
	}

	gl.BindVertexArray(handle)
}

func (vao *VertexArrayObject) Dispose() {
	if !vao.disposed {
		for _, holder := range vao.buffers {
			holder.buffer.Dispose()
		}

		goroutines.CallNonBlockMain(func() {
			gl.DeleteVertexArrays(1, &vao.handle)
		})
	}

	vao.disposed = true
}

func (vao *VertexArrayObject) AttachIBO(ibo *IndexBufferObject) {
	ibo.attached = true
	vao.ibo = ibo
	gl.VertexArrayElementBuffer(vao.handle, ibo.handle)
}
