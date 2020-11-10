package buffer

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/statistic"
	"runtime"
)

type vboHolder struct {
	vbo     StreamingBuffer
	divisor int
	format  attribute.Format
}

type VertexArrayObject struct {
	handle uint32

	vbos map[string]*vboHolder

	capacity int

	bound    bool
	disposed bool
}

func NewVertexArrayObject() *VertexArrayObject {
	vao := new(VertexArrayObject)
	vao.vbos = make(map[string]*vboHolder)

	gl.GenVertexArrays(1, &vao.handle)

	vao.Bind()
	vao.Unbind()

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
	if _, exists := vao.vbos[name]; exists {
		panic(fmt.Sprintf("VBO with name \"%s\" already exists", name))
	}

	holder := &vboHolder{
		vbo:     NewVertexBufferObject(maxVertices*format.Size()/4, mapped, DynamicDraw),
		divisor: divisor,
		format:  format,
	}

	if divisor == 0 {
		vao.capacity = maxVertices
	}

	vao.vbos[name] = holder
}

func (vao *VertexArrayObject) AddPersistentVBO(name string, maxVertices int, divisor int, format attribute.Format) {
	if _, exists := vao.vbos[name]; exists {
		panic(fmt.Sprintf("VBO with name \"%s\" already exists", name))
	}

	holder := &vboHolder{
		vbo:     NewPersistentBufferObject(maxVertices * format.Size() / 4),
		divisor: divisor,
		format:  format,
	}

	if divisor == 0 {
		vao.capacity = maxVertices
	}

	vao.vbos[name] = holder
}

func (vao *VertexArrayObject) GetVBOFormat(name string) attribute.Format {
	if holder, exists := vao.vbos[name]; exists {
		return holder.format
	}

	panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
}

func (vao *VertexArrayObject) GetVBO(name string) StreamingBuffer {
	if holder, exists := vao.vbos[name]; exists {
		return holder.vbo
	}

	panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
}

func (vao *VertexArrayObject) Resize(name string, maxVertices int) {
	if holder, exists := vao.vbos[name]; exists {
		size := maxVertices * holder.format.Size() / 4
		if holder.vbo.Capacity() != size {
			holder.vbo.Bind()
			holder.vbo.Resize(size)
			holder.vbo.Unbind()
		}

		return
	}

	panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
}

func (vao *VertexArrayObject) Attach(s *shader.RShader) {
	currentVAO := history.GetCurrent(gl.VERTEX_ARRAY_BINDING)
	if currentVAO != vao.handle {
		panic(fmt.Sprintf("VAO mismatch. Target VAO: %d, current: %d", vao.handle, currentVAO))
	}

	for _, holder := range vao.vbos {

		holder.vbo.Bind()

		var offset int
		for _, attr := range holder.format {

			location := s.GetAttributeInfo(attr.Name).Location

			gl.VertexAttribPointer(
				uint32(location),
				int32(attr.Type.Components()),
				uint32(attr.Type.InternalType()),
				attr.Type.Normalize(),
				int32(holder.format.Size()),
				gl.PtrOffset(offset),
			)

			gl.VertexAttribDivisor(uint32(location), uint32(holder.divisor))
			gl.EnableVertexAttribArray(uint32(location))

			offset += attr.Type.Size()
		}

		holder.vbo.Unbind()

	}
}

func (vao *VertexArrayObject) SetData(name string, offset int, data []float32) {
	holder, exists := vao.vbos[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	holder.vbo.Bind()
	holder.vbo.SetData(offset, data)
	holder.vbo.Unbind()

	statistic.Add(statistic.VertexUpload, int64(len(data)*4/holder.format.Size()))
}

func (vao *VertexArrayObject) MapVBO(name string, size int) MemoryChunk {
	holder, exists := vao.vbos[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	return holder.vbo.Map(size)
}

func (vao *VertexArrayObject) UnmapVBO(name string, offset int, size int) {
	holder, exists := vao.vbos[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	holder.vbo.Bind()
	holder.vbo.Unmap(offset, size)
	holder.vbo.Unbind()

	statistic.Add(statistic.VertexUpload, int64(size*4/holder.format.Size()))
}

func (vao *VertexArrayObject) Draw() {
	vao.DrawPart(0, vao.capacity)
}

func (vao *VertexArrayObject) DrawInstanced(baseInstance, instanceCount int) {
	vao.DrawPartInstanced(0, vao.capacity, baseInstance, instanceCount)
}

func (vao *VertexArrayObject) DrawPart(offset, length int) {
	vao.check(offset, length)

	statistic.Add(statistic.VerticesDrawn, int64(length))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawArrays(gl.TRIANGLES, int32(offset), int32(length))
}

func (vao *VertexArrayObject) DrawPartInstanced(offset, length, baseInstance, instanceCount int) {
	vao.check(offset, length)

	statistic.Add(statistic.VerticesDrawn, int64(length*instanceCount))
	statistic.Increment(statistic.DrawCalls)

	gl.DrawArraysInstancedBaseInstance(gl.TRIANGLES, int32(offset), int32(length), int32(instanceCount), uint32(baseInstance))
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

	history.Push(gl.VERTEX_ARRAY_BINDING)

	statistic.Increment(statistic.VAOBinds)

	gl.BindVertexArray(vao.handle)
}

func (vao *VertexArrayObject) Unbind() {
	if !vao.bound || vao.disposed {
		return
	}

	vao.bound = false

	handle := history.Pop(gl.VERTEX_ARRAY_BINDING)

	if handle > 0 {
		statistic.Increment(statistic.VAOBinds)
	}

	gl.BindVertexArray(handle)
}

func (vao *VertexArrayObject) Dispose() {
	if !vao.disposed {

		for _, holder := range vao.vbos {
			holder.vbo.Dispose()
		}

		mainthread.CallNonBlock(func() {
			gl.DeleteVertexArrays(1, &vao.handle)
		})
	}

	vao.disposed = true
}
