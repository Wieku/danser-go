package shader

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/goroutines"
)

type Type uint32

const (
	Vertex   = Type(gl.VERTEX_SHADER)
	Geometry = Type(gl.GEOMETRY_SHADER)
	Fragment = Type(gl.FRAGMENT_SHADER)
)

func (t Type) Name() string {
	switch t {
	case Vertex:
		return "Vertex shader"
	case Geometry:
		return "Geometry shader"
	case Fragment:
		return "Fragment shader"
	default:
		panic("Invalid type")
	}
}

type Source struct {
	handle  uint32
	success bool
	log     string
	srcType Type
}

func NewSource(source string, srcType Type) *Source {
	src := new(Source)
	src.srcType = srcType

	src.handle = gl.CreateShader(uint32(src.srcType))

	srcC, free := gl.Strs(source)
	defer free()

	length := int32(len(source))
	gl.ShaderSource(src.handle, 1, srcC, &length)
	gl.CompileShader(src.handle)

	var success int32
	gl.GetShaderiv(src.handle, gl.COMPILE_STATUS, &success)

	src.success = success == gl.TRUE

	if !src.success {
		var logLen int32
		gl.GetShaderiv(src.handle, gl.INFO_LOG_LENGTH, &logLen)

		infoLog := make([]byte, logLen)
		gl.GetShaderInfoLog(src.handle, logLen, nil, &infoLog[0])

		src.log = string(infoLog)
	}

	return src
}

func (src *Source) Dispose() {
	goroutines.CallNonBlockMain(func() {
		gl.DeleteShader(src.handle)
	})
}
