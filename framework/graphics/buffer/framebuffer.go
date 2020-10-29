package buffer

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/history"
	"github.com/wieku/danser-go/framework/statistic"
	"runtime"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/texture"
)

// Framebuffer is a fixed resolution texture that you can draw on.
type Framebuffer struct {
	obj           uint32
	last          int32
	tex           *texture.TextureSingle
	multisampled  bool
	helperObj     uint32
	helperTexture uint32
	depth         uint32
}

// NewFrame creates a new fully transparent Framebuffer with given dimensions in pixels.
func NewFrame(width, height int, smooth, depth bool) *Framebuffer {
	f := new(Framebuffer)

	f.tex = texture.NewTextureSingle(width, height, 0)

	gl.GenFramebuffers(1, &f.obj)

	f.Bind()
	f.tex.Bind(0)
	gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, f.tex.GetID(), 0, 0)

	if depth {
		var depthRenderBuffer uint32
		gl.GenRenderbuffers(1, &depthRenderBuffer)
		gl.BindRenderbuffer(gl.RENDERBUFFER, depthRenderBuffer)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthRenderBuffer)
	}

	f.Unbind()

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameDepth(width, height int, smooth bool) *Framebuffer {
	f := new(Framebuffer)

	f.tex = texture.NewTextureSingleFormat(width, height, texture.Depth, 0)

	gl.GenFramebuffers(1, &f.obj)

	f.Bind()
	f.tex.Bind(0)
	gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, f.tex.GetID(), 0, 0)

	f.Unbind()

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	return f
}

func NewFrameMultisample(width, height int, smooth, depth bool) *Framebuffer {
	f := new(Framebuffer)

	f.tex = texture.NewTextureSingle(width, height, 0)

	gl.GenFramebuffers(1, &f.helperObj)

	history.Push(gl.FRAMEBUFFER_BINDING)
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.helperObj)

	f.tex.Bind(0)
	gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, f.tex.GetID(), 0, 0)

	previous := history.Pop(gl.FRAMEBUFFER_BINDING)
	gl.BindFramebuffer(gl.FRAMEBUFFER, previous)

	gl.GenFramebuffers(1, &f.obj)

	f.Bind()

	gl.GenTextures(1, &f.helperTexture)
	gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, f.helperTexture)

	gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, settings.Graphics.MSAA, gl.RGBA8, int32(width), int32(height), true)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D_MULTISAMPLE, f.helperTexture, 0)

	if depth {
		var depthRenderBuffer uint32
		gl.GenRenderbuffers(1, &depthRenderBuffer)
		gl.BindRenderbuffer(gl.RENDERBUFFER, depthRenderBuffer)
		gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, settings.Graphics.MSAA, gl.DEPTH_COMPONENT, int32(width), int32(height))
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthRenderBuffer)
		f.depth = depthRenderBuffer
	}

	f.Unbind()

	runtime.SetFinalizer(f, (*Framebuffer).Dispose)

	f.multisampled = true
	return f
}

func (f *Framebuffer) Dispose() {
	mainthread.CallNonBlock(func() {
		f.tex.Dispose()
		if f.depth > 0 {
			gl.DeleteRenderbuffers(1, &f.depth)
		}

		if f.helperObj > 0 {
			gl.DeleteFramebuffers(1, &f.helperObj)
		}

		if f.helperTexture > 0 {
			gl.DeleteTextures(1, &f.helperTexture)
		}

		gl.DeleteFramebuffers(1, &f.obj)
	})
}

// ID returns the OpenGL framebuffer ID of this Framebuffer.
func (f *Framebuffer) ID() uint32 {
	return f.obj
}

// Bind binds the Framebuffer. All draw operations will target this Framebuffer until Unbind is called.
func (f *Framebuffer) Bind() {
	history.Push(gl.FRAMEBUFFER_BINDING)
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.obj)
	statistic.Increment(statistic.FBOBinds)
}

// Unbind unbinds the Framebuffer. All draw operations will go to whatever was bound before this Framebuffer.
func (f *Framebuffer) Unbind() {
	if f.multisampled {
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, f.obj)
		gl.ReadBuffer(gl.COLOR_ATTACHMENT0)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, f.helperObj)
		gl.DrawBuffer(gl.COLOR_ATTACHMENT0)

		gl.BlitFramebuffer(0, 0, f.tex.GetWidth(), f.tex.GetHeight(), 0, 0, f.tex.GetWidth(), f.tex.GetHeight(), gl.COLOR_BUFFER_BIT, gl.LINEAR)

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}

	handle := history.Pop(gl.FRAMEBUFFER_BINDING)
	if handle != 0 {
		statistic.Increment(statistic.FBOBinds)
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, handle)
}

// Texture returns the Framebuffer's underlying Texture that the Framebuffer draws on.
func (f *Framebuffer) Texture() texture.Texture {
	return f.tex
}
