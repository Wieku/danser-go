package shader

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"runtime"
)

type RShader struct {
	handle uint32

	attributes map[string]attribute.VertexAttribute
	uniforms   map[string]attribute.VertexAttribute

	bound    bool
	disposed bool

	iCache *int32
	fCache *float32

	v2Cache *mgl32.Vec2
	v3Cache *mgl32.Vec3
	v4Cache *mgl32.Vec4

	m2Cache  *mgl32.Mat2
	m23Cache *mgl32.Mat2x3
	m24Cache *mgl32.Mat2x4

	m3Cache  *mgl32.Mat3
	m32Cache *mgl32.Mat3x2
	m34Cache *mgl32.Mat3x4

	m4Cache  *mgl32.Mat4
	m42Cache *mgl32.Mat4x2
	m43Cache *mgl32.Mat4x3

	cache []float32
}

func NewRShader(sources ...*Source) *RShader {
	for _, src := range sources {
		if !src.success {
			panic(fmt.Sprintf("Failed to build %s: %s", src.srcType.Name(), src.log))
		}
	}

	s := new(RShader)

	s.iCache = new(int32)
	s.fCache = new(float32)

	s.v2Cache = new(mgl32.Vec2)
	s.v3Cache = new(mgl32.Vec3)
	s.v4Cache = new(mgl32.Vec4)

	s.m2Cache = new(mgl32.Mat2)
	s.m23Cache = new(mgl32.Mat2x3)
	s.m24Cache = new(mgl32.Mat2x4)

	s.m3Cache = new(mgl32.Mat3)
	s.m32Cache = new(mgl32.Mat3x2)
	s.m34Cache = new(mgl32.Mat3x4)

	s.m4Cache = new(mgl32.Mat4)
	s.m42Cache = new(mgl32.Mat4x2)
	s.m43Cache = new(mgl32.Mat4x3)

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
	var attribs int32
	gl.GetProgramiv(s.handle, gl.ACTIVE_ATTRIBUTES, &attribs)

	for i := int32(0); i < attribs; i++ {
		var maxLength int32
		gl.GetProgramiv(s.handle, gl.ACTIVE_ATTRIBUTE_MAX_LENGTH, &maxLength)

		var nameB = make([]uint8, maxLength)

		var length, size int32
		var xtype uint32

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
	var uniforms int32
	gl.GetProgramiv(s.handle, gl.ACTIVE_UNIFORMS, &uniforms)

	for i := int32(0); i < uniforms; i++ {

		var maxLength int32
		gl.GetProgramiv(s.handle, gl.ACTIVE_UNIFORM_MAX_LENGTH, &maxLength)

		var nameB = make([]uint8, maxLength)

		var length, size int32
		var xtype uint32

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

func (s *RShader) GetUniformInfo(name string) attribute.VertexAttribute {
	attr, exists := s.uniforms[name]
	if !exists {
		panic(fmt.Sprintf("Uniform %s doesn't exist", name))
	}

	return attr
}

func (s *RShader) SetUniform(name string, value any) {
	uniform, exists := s.uniforms[name]
	if !exists {
		log.Println(s.uniforms)
		panic(fmt.Sprintf("Uniform %s doesn't exist", name))
	}

	s.setUniformInternal(uniform.Location, uniform.Type, value)
}

func (s *RShader) SetUniformArr(name string, offset int, value interface{}) {
	name += "[0]"

	uniform, exists := s.uniforms[name]
	if !exists {
		log.Println(s.uniforms)
		panic(fmt.Sprintf("Uniform %s doesn't exist", name))
	}

	s.setUniformInternal(uniform.Location+int32(offset), uniform.Type, value)
}

func (s *RShader) setUniformInternal(location int32, uType attribute.Type, value any) {
	switch uType {
	case attribute.Float:
		*s.fCache = value.(float32)
		gl.ProgramUniform1fv(s.handle, location, 1, s.fCache)
	case attribute.Vec2:
		if c, ok := value.(vector.Vector2f); ok {
			s.v2Cache[0] = c.X
			s.v2Cache[1] = c.Y
		} else if c2, ok2 := value.(vector.Vector2d); ok2 {
			s.v2Cache[0] = c2.X32()
			s.v2Cache[1] = c2.Y32()
		} else {
			*s.v2Cache = value.(mgl32.Vec2)
		}

		gl.ProgramUniform2fv(s.handle, location, 1, &s.v2Cache[0])
	case attribute.Vec3:
		*s.v3Cache = value.(mgl32.Vec3)
		gl.ProgramUniform3fv(s.handle, location, 1, &s.v3Cache[0])
	case attribute.Vec4:
		if c, ok := value.(color.Color); ok {
			s.v4Cache[0] = c.R
			s.v4Cache[1] = c.G
			s.v4Cache[2] = c.B
			s.v4Cache[3] = c.A
		} else {
			*s.v4Cache = value.(mgl32.Vec4)
		}

		gl.ProgramUniform4fv(s.handle, location, 1, &s.v4Cache[0])
	case attribute.Mat2:
		*s.m2Cache = value.(mgl32.Mat2)
		gl.ProgramUniformMatrix2fv(s.handle, location, 1, false, &s.m2Cache[0])
	case attribute.Mat23:
		*s.m23Cache = value.(mgl32.Mat2x3)
		gl.ProgramUniformMatrix2x3fv(s.handle, location, 1, false, &s.m23Cache[0])
	case attribute.Mat24:
		*s.m24Cache = value.(mgl32.Mat2x4)
		gl.ProgramUniformMatrix2x4fv(s.handle, location, 1, false, &s.m24Cache[0])
	case attribute.Mat3:
		*s.m3Cache = value.(mgl32.Mat3)
		gl.ProgramUniformMatrix3fv(s.handle, location, 1, false, &s.m3Cache[0])
	case attribute.Mat32:
		*s.m32Cache = value.(mgl32.Mat3x2)
		gl.ProgramUniformMatrix3x2fv(s.handle, location, 1, false, &s.m32Cache[0])
	case attribute.Mat34:
		*s.m34Cache = value.(mgl32.Mat3x4)
		gl.ProgramUniformMatrix3x4fv(s.handle, location, 1, false, &s.m34Cache[0])
	case attribute.Mat4:
		*s.m4Cache = value.(mgl32.Mat4)
		gl.ProgramUniformMatrix4fv(s.handle, location, 1, false, &s.m4Cache[0])
	case attribute.Mat42:
		*s.m42Cache = value.(mgl32.Mat4x2)
		gl.ProgramUniformMatrix4x2fv(s.handle, location, 1, false, &s.m42Cache[0])
	case attribute.Mat43:
		*s.m43Cache = value.(mgl32.Mat4x3)
		gl.ProgramUniformMatrix4x3fv(s.handle, location, 1, false, &s.m43Cache[0])
	default: // We assume that uniform is of type int or sampler
		if vI, ok := value.(int); ok {
			*s.iCache = int32(vI)
		} else {
			*s.iCache = value.(int32)
		}

		gl.ProgramUniform1iv(s.handle, location, 1, s.iCache)
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

	history.Push(gl.CURRENT_PROGRAM, s.handle)
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
		goroutines.CallNonBlockMain(func() {
			gl.DeleteProgram(s.handle)
		})
	}

	s.disposed = true
}
