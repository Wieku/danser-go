package buffer

import (
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/history"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/profiler"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/texture"
)

// Framebuffer is a fixed resolution texture that you can draw on.
type Framebuffer struct {
	handle uint32

	width  int
	height int

	tex             *texture.TextureSingle
	multisampled    bool
	helperHandle    uint32
	helperTexture   uint32
	depth           uint32
	texRenderbuffer uint32

	disposed bool
}

// NewFrame creates a new fully transparent Framebuffer with given dimensions in pixels.
func NewFrame(width, height int, smooth, depth bool) *Framebuffer {
	f := new(Framebuffer)
	f.width = width
	f.height = height

	f.tex = texture.NewTextureSingle(width, height, 0)

	gl.CreateFramebuffers(1, &f.handle)

	gl.NamedFramebufferTextureLayer(f.handle, gl.COLOR_ATTACHMENT0, f.tex.GetID(), 0, 0)

	if depth {
		gl.CreateRenderbuffers(1, &f.depth)

		gl.NamedRenderbufferStorage(f.depth, gl.DEPTH_COMPONENT, int32(width), int32(height))
		gl.NamedFramebufferRenderbuffer(f.handle, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, f.depth)
	}

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameF(width, height int) *Framebuffer {
	f := new(Framebuffer)
	f.width = width
	f.height = height

	f.tex = texture.NewTextureSingleFormat(width, height, texture.RGBA32F, 0)
	f.tex.SetFiltering(texture.Filtering.Nearest, texture.Filtering.Nearest)

	gl.CreateFramebuffers(1, &f.handle)

	gl.NamedFramebufferTextureLayer(f.handle, gl.COLOR_ATTACHMENT0, f.tex.GetID(), 0, 0)

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameLayer(texture texture.Texture, layer int) *Framebuffer {
	f := new(Framebuffer)
	f.width = int(texture.GetWidth())
	f.height = int(texture.GetHeight())

	gl.CreateFramebuffers(1, &f.handle)

	gl.NamedFramebufferTextureLayer(f.handle, gl.COLOR_ATTACHMENT0, texture.GetID(), 0, int32(layer))

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameDepth(width, height int, smooth bool) *Framebuffer {
	f := new(Framebuffer)
	f.width = width
	f.height = height

	f.tex = texture.NewTextureSingleFormat(width, height, texture.Depth, 0)

	gl.CreateFramebuffers(1, &f.handle)

	gl.NamedFramebufferTextureLayer(f.handle, gl.DEPTH_ATTACHMENT, f.tex.GetID(), 0, 0)

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameMultisample(width, height int, samples int) *Framebuffer {
	f := new(Framebuffer)
	f.width = width
	f.height = height
	f.multisampled = true

	gl.CreateFramebuffers(1, &f.handle)

	gl.CreateRenderbuffers(1, &f.texRenderbuffer)
	gl.NamedRenderbufferStorageMultisample(f.texRenderbuffer, int32(samples), texture.RGBA.InternalFormat(), int32(width), int32(height))
	gl.NamedFramebufferRenderbuffer(f.handle, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, f.texRenderbuffer)

	f.tex = texture.NewTextureSingle(width, height, 0)

	gl.CreateFramebuffers(1, &f.helperHandle)
	gl.NamedFramebufferTextureLayer(f.helperHandle, gl.COLOR_ATTACHMENT0, f.tex.GetID(), 0, 0)

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameMultisampleScreen(width, height int, depth bool, samples int) *Framebuffer {
	f := new(Framebuffer)
	f.width = width
	f.height = height
	f.multisampled = true

	gl.CreateFramebuffers(1, &f.handle)

	gl.CreateRenderbuffers(1, &f.texRenderbuffer)
	gl.NamedRenderbufferStorageMultisample(f.texRenderbuffer, int32(samples), texture.RGBA.InternalFormat(), int32(width), int32(height))
	gl.NamedFramebufferRenderbuffer(f.handle, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, f.texRenderbuffer)

	if depth {
		gl.CreateRenderbuffers(1, &f.depth)
		gl.NamedRenderbufferStorageMultisample(f.depth, int32(samples), gl.DEPTH_COMPONENT, int32(width), int32(height))
		gl.NamedFramebufferRenderbuffer(f.handle, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, f.depth)
	}

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func (f *Framebuffer) Dispose() {
	if !f.disposed {
		goroutines.CallNonBlockMain(func() {
			if f.tex != nil {
				f.tex.Dispose()
			}

			if f.depth > 0 {
				gl.DeleteRenderbuffers(1, &f.depth)
			}

			if f.texRenderbuffer > 0 {
				gl.DeleteRenderbuffers(1, &f.texRenderbuffer)
			}

			if f.helperHandle > 0 {
				gl.DeleteFramebuffers(1, &f.helperHandle)
			}

			if f.helperTexture > 0 {
				gl.DeleteTextures(1, &f.helperTexture)
			}

			gl.DeleteFramebuffers(1, &f.handle)
		})
	}

	f.disposed = true
}

// GetID returns the OpenGL framebuffer ID of this Framebuffer.
func (f *Framebuffer) GetID() uint32 {
	return f.handle
}

// Bind binds the Framebuffer. All draw operations will target this Framebuffer until Unbind is called.
func (f *Framebuffer) Bind() {
	history.Push(gl.FRAMEBUFFER_BINDING, f.handle)
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.handle)
	profiler.IncrementStat(profiler.FBOBinds)
}

// Unbind unbinds the Framebuffer. All draw operations will go to whatever was bound before this Framebuffer.
func (f *Framebuffer) Unbind() {
	handle := history.Pop(gl.FRAMEBUFFER_BINDING)

	if f.multisampled {
		hHandle := f.helperHandle
		if hHandle == 0 {
			hHandle = handle
		}

		gl.NamedFramebufferReadBuffer(f.handle, gl.COLOR_ATTACHMENT0)

		if hHandle > 0 {
			gl.NamedFramebufferDrawBuffer(hHandle, gl.COLOR_ATTACHMENT0)
		}

		gl.BlitNamedFramebuffer(f.handle, hHandle, 0, 0, int32(f.width), int32(f.height), 0, 0, int32(f.width), int32(f.height), gl.COLOR_BUFFER_BIT, gl.LINEAR)
	}

	if handle != 0 {
		profiler.IncrementStat(profiler.FBOBinds)
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, handle)
}

// Texture returns the Framebuffer's underlying Texture that the Framebuffer draws on.
func (f *Framebuffer) Texture() texture.Texture {
	return f.tex
}

func (f *Framebuffer) GetWidth() int {
	return f.width
}

func (f *Framebuffer) GetHeight() int {
	return f.height
}

func (f *Framebuffer) ClearColor(r, g, b, a float32) {
	col := []float32{r, g, b, a}
	gl.ClearNamedFramebufferfv(f.handle, gl.COLOR, 0, &col[0])
}

func (f *Framebuffer) ClearColorI(index int, r, g, b, a float32) {
	col := []float32{r, g, b, a}
	gl.ClearNamedFramebufferfv(f.handle, gl.COLOR, int32(index), &col[0])
}

func (f *Framebuffer) ClearColorM(color color2.Color) {
	gl.ClearNamedFramebufferfv(f.handle, gl.COLOR, 0, &color.ToArray()[0])
}

func (f *Framebuffer) ClearColorIM(index int, color color2.Color) {
	gl.ClearNamedFramebufferfv(f.handle, gl.COLOR, int32(index), &color.ToArray()[0])
}

func (f *Framebuffer) ClearDepthV(v float32) {
	gl.ClearNamedFramebufferfv(f.handle, gl.DEPTH, 0, &v)
}

func (f *Framebuffer) ClearDepth() {
	f.ClearDepthV(1)
}
