package effects

import (
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
)

type RGBYUV struct {
	width  int
	height int

	fbo          *buffer.Framebuffer
	yuvFBO       *buffer.Framebuffer
	subsampleFBO *buffer.Framebuffer

	yuvShader       *shader.RShader
	subsampleShader *shader.RShader

	vao *buffer.VertexArrayObject

	subsample bool
}

func NewRGBYUV(width, height int, subsample bool) *RGBYUV {
	effect := new(RGBYUV)
	effect.width = width
	effect.height = height
	effect.subsample = subsample

	vert, err := assets.GetString("assets/shaders/fbopass.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/rgbyuv.fsh")
	if err != nil {
		panic(err)
	}

	fpass, err := assets.GetString("assets/shaders/rgbyuv_scale.fsh")
	if err != nil {
		panic(err)
	}

	effect.yuvShader = shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	effect.subsampleShader = shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(fpass, shader.Fragment))

	effect.vao = buffer.NewVertexArrayObject()

	effect.vao.AddVBO("default", 6, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
		{Name: "in_tex_coord", Type: attribute.Vec2},
	})

	effect.vao.SetData("default", 0, []float32{
		-1, -1, 0, 0, 0,
		1, -1, 0, 1, 0,
		-1, 1, 0, 0, 1,
		1, -1, 0, 1, 0,
		1, 1, 0, 1, 1,
		-1, 1, 0, 0, 1,
	})

	effect.vao.Attach(effect.yuvShader)

	effect.fbo = buffer.NewFrame(width, height, false, false)

	effect.yuvFBO = buffer.NewFrameYUV(width, height)

	effect.subsampleFBO = buffer.NewFrameYUVSmall((width+1)/2, (height+1)/2)

	return effect
}

func (effect *RGBYUV) Begin() {
	effect.fbo.Bind()
	viewport.Push(effect.width, effect.height)
}

func (effect *RGBYUV) End() {
	viewport.Pop()
	effect.fbo.Unbind()
}

func (effect *RGBYUV) Draw() (yuv, uv []texture.Texture) {
	blend.Push()
	blend.Disable()

	effect.fbo.Texture().Bind(5)

	effect.yuvShader.SetUniform("tex", 5)

	effect.vao.Bind()

	effect.yuvFBO.Bind()

	viewport.Push(effect.width, effect.height)

	effect.yuvShader.Bind()

	effect.vao.Draw()

	effect.yuvShader.Unbind()

	viewport.Pop()

	effect.yuvFBO.Unbind()

	if effect.subsample {
		effect.yuvFBO.Textures()[1].Bind(6)
		effect.yuvFBO.Textures()[2].Bind(7)

		effect.subsampleShader.SetUniform("texU", 6)
		effect.subsampleShader.SetUniform("texV", 7)

		effect.subsampleFBO.Bind()

		viewport.Push(effect.subsampleFBO.GetWidth(), effect.subsampleFBO.GetHeight())

		effect.subsampleShader.Bind()

		effect.vao.Draw()

		effect.subsampleShader.Unbind()

		viewport.Pop()

		effect.subsampleFBO.Unbind()
	}

	effect.vao.Unbind()

	blend.Pop()

	return effect.yuvFBO.Textures(), effect.subsampleFBO.Textures()
}
