package shader

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/history"
	"runtime"
)

type RShader struct {
	handle uint32

	attributes map[string]attribute.VertexAttribute
	uniforms   map[string]attribute.VertexAttribute

	bound    bool
	disposed bool
}

func NewRShader(sources ...*Source) *RShader {
	for _, src := range sources {
		if !src.success {
			panic(fmt.Sprintf("Failed to build %s: %s", src.srcType.Name(), src.log))
		}
	}

	s := new(RShader)
	s.attributes = make(map[string]attribute.VertexAttribute)
	s.uniforms = make(map[string]attribute.VertexAttribute)

	s.handle = gl.CreateProgram()

	for _, src := range sources {
		gl.AttachShader(s.handle, src.handle)
	}

	gl.LinkProgram(s.handle)

	var success int32
	gl.GetProgramiv(s.handle, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(s.handle, gl.INFO_LOG_LENGTH, &logLen)

		infoLog := make([]byte, logLen)
		gl.GetProgramInfoLog(s.handle, logLen, nil, &infoLog[0])

		panic(fmt.Sprintf("Can't link shader program: %s", string(infoLog)))
	}

	s.fetchAttributes()
	s.fetchUniforms()

	for _, src := range sources {
		src.Dispose()
	}

	runtime.SetFinalizer(s, (*RShader).Dispose)

	return s
}

func (s *RShader) fetchAttributes() {
	var max int32
	gl.GetProgramiv(s.handle, gl.ACTIVE_ATTRIBUTES, &max)

	for i := int32(0); i < max; i++ {

		var maxLength int32
		gl.GetProgramiv(s.handle, gl.ACTIVE_ATTRIBUTE_MAX_LENGTH, &maxLength)

		var length, size int32
		var xtype uint32
		var nameB = make([]uint8, maxLength)

		gl.GetActiveAttrib(s.handle, uint32(i), maxLength, &length, &size, &xtype, &nameB[0])

		name := string(nameB[:length])

		location := gl.GetAttribLocation(s.handle, &nameB[0])

		s.attributes[name] = attribute.VertexAttribute{
			Name:     name,
			Type:     Type(xtype),
			Location: uint32(location),
		}
	}
}

func (s *RShader) fetchUniforms() {
	var max int32
	gl.GetProgramiv(s.handle, gl.ACTIVE_UNIFORMS, &max)

	for i := int32(0); i < max; i++ {

		var maxLength int32
		gl.GetProgramiv(s.handle, gl.ACTIVE_UNIFORM_MAX_LENGTH, &maxLength)

		var length, size int32
		var xtype uint32
		var nameB = make([]uint8, maxLength)

		gl.GetActiveUniform(s.handle, uint32(i), maxLength, &length, &size, &xtype, &nameB[0])

		name := string(nameB[:length])

		location := gl.GetUniformLocation(s.handle, &nameB[0])

		s.uniforms[name] = attribute.VertexAttribute{
			Name:     name,
			Type:     Type(xtype),
			Location: uint32(location),
		}
	}
}

func (s *RShader) GetAttributeInfo(name string) attribute.VertexAttribute {
	attr, exists := s.attributes[name]
	if !exists {
		panic(fmt.Sprintf("Attribute %s doesn't exist", name))
	}

	return attr
}

func (s *RShader) GetUnformInfo(name string) attribute.VertexAttribute {
	attr, exists := s.uniforms[name]
	if !exists {
		panic(fmt.Sprintf("Uniform %s doesn't exist", name))
	}

	return attr
}

func (s *RShader) Bind() {
	if s.disposed {
		panic("Can't bind disposed shader")
	}

	if s.bound {
		panic(fmt.Sprintf("Shader %d is already bound", s.handle))
	}

	s.bound = true

	history.Push(gl.CURRENT_PROGRAM)
	gl.UseProgram(s.handle)
}

// End unbinds the Shader program and restores the previous one.
func (s *RShader) Unbind() {
	if s.disposed || !s.bound {
		return
	}

	s.bound = false

	handle := history.Pop(gl.CURRENT_PROGRAM)
	gl.UseProgram(handle)
}

func (s *RShader) Dispose() {
	if !s.disposed {
		mainthread.CallNonBlock(func() {
			gl.DeleteProgram(s.handle)
		})
	}

	s.disposed = true
}
