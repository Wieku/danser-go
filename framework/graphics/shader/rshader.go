package shader

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/math/color"
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
			Type:     attribute.Type(xtype),
			Location: location,
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
			Type:     attribute.Type(xtype),
			Location: location,
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

func (s *RShader) SetUniform(name string, value interface{}) {
	uniform, exists := s.uniforms[name]
	if !exists {
		panic(fmt.Sprintf("Uniform %s doesn't exist", name))
	}

	switch uniform.Type {
	case attribute.Float:
		value := value.(float32)
		gl.Uniform1fv(uniform.Location, 1, &value)
	case attribute.Vec2:
		value := value.(mgl32.Vec2)
		gl.Uniform2fv(uniform.Location, 1, &value[0])
	case attribute.Vec3:
		value := value.(mgl32.Vec3)
		gl.Uniform3fv(uniform.Location, 1, &value[0])
	case attribute.Vec4:
		if c, ok := value.(color.Color); ok {
			gl.Uniform4fv(uniform.Location, 1, &c.ToArray()[0])

			break
		}

		value := value.(mgl32.Vec4)
		gl.Uniform4fv(uniform.Location, 1, &value[0])
	case attribute.Mat2:
		value := value.(mgl32.Mat2)
		gl.UniformMatrix2fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat23:
		value := value.(mgl32.Mat2x3)
		gl.UniformMatrix2x3fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat24:
		value := value.(mgl32.Mat2x4)
		gl.UniformMatrix2x4fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat3:
		value := value.(mgl32.Mat3)
		gl.UniformMatrix3fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat32:
		value := value.(mgl32.Mat3x2)
		gl.UniformMatrix3x2fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat34:
		value := value.(mgl32.Mat3x4)
		gl.UniformMatrix3x4fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat4:
		value := value.(mgl32.Mat4)
		gl.UniformMatrix4fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat42:
		value := value.(mgl32.Mat4x2)
		gl.UniformMatrix4x2fv(uniform.Location, 1, false, &value[0])
	case attribute.Mat43:
		value := value.(mgl32.Mat4x3)
		gl.UniformMatrix4x3fv(uniform.Location, 1, false, &value[0])
	default: // We assume that uniform is of type int or sampler
		if value, ok := value.(int); ok {
			value := int32(value)
			gl.Uniform1iv(uniform.Location, 1, &value)

			break
		}

		value := value.(int32)
		gl.Uniform1iv(uniform.Location, 1, &value)
	}
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

// Unbind unbinds the Shader program and restores the previous one.
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
