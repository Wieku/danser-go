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

	uniformFormat attribute.Format

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

	for _, src := range sources {
		src.Dispose()
	}

	runtime.SetFinalizer(s, (*RShader).Dispose)

	return s
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
