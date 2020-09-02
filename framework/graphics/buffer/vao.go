package buffer

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/statistic"
	"runtime"
)

type vboHolder struct {
	vbo     *VertexBufferObject
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
	if _, exists := vao.vbos[name]; exists {
		panic(fmt.Sprintf("VBO with name \"%s\" already exists", name))
	}

	holder := &vboHolder{
		vbo:     NewVertexBufferObject(maxVertices*format.Size()/4, DynamicDraw),
		divisor: divisor,
		format:  format,
	}

	if divisor == 0 {
		vao.capacity = maxVertices
	}

	vao.vbos[name] = holder
}

func (vao *VertexArrayObject) SetData(name string, offset int, data []float32) {
	holder, exists := vao.vbos[name]
	if !exists {
		panic(fmt.Sprintf("VBO with name \"%s\" doesn't exist", name))
	}

	holder.vbo.Bind()
	holder.vbo.SetData(offset, data)
	holder.vbo.Unbind()
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
