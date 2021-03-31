package effects

import (
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
)

type Blend struct {
	width        int
	height       int
	layers       int
	head         int
	fbos         []*buffer.Framebuffer
	blendShader  *shader.RShader
	vao          *buffer.VertexArrayObject
	multiTexture *texture.TextureMultiLayer
}

func NewBlend(width, height, frames int, weights []float32) *Blend {
	if frames != len(weights) {
		panic("Wrong number of weights")
	}

	effect := new(Blend)
	effect.width = width
	effect.height = height
	effect.layers = frames

	vert, err := assets.GetString("assets/shaders/fbopass.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/blend.fsh")
	if err != nil {
		panic(err)
	}

	effect.blendShader = shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

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

	effect.vao.Attach(effect.blendShader)

	effect.blendShader.SetUniform("layers", frames)

	var sum float32
	for _, v := range weights {
		sum += v
	}

	for i, v := range weights {
		effect.blendShader.SetUniformArr("weights", i, v/sum)
	}

	effect.multiTexture = texture.NewTextureMultiLayerFormat(width, height, texture.RGB, 0, frames)

	for i := 0; i < frames; i++ {
		effect.fbos = append(effect.fbos, buffer.NewFrameLayer(effect.multiTexture, i))
	}

	return effect
}

func (effect *Blend) Begin() {
	effect.head = (effect.head + 1) % effect.layers
	effect.fbos[effect.head].Bind()
	effect.fbos[effect.head].ClearColor(0, 0, 0, 1)
	viewport.Push(effect.width, effect.height)
}

func (effect *Blend) End() {
	viewport.Pop()
	effect.fbos[effect.head].Unbind()
}

func (effect *Blend) Blend() {
	effect.multiTexture.Bind(0)
	effect.blendShader.SetUniform("tex", 0)
	effect.blendShader.SetUniform("head", effect.head)

	viewport.Push(effect.width, effect.height)

	effect.blendShader.Bind()
	effect.vao.Bind()
	effect.vao.Draw()
	effect.vao.Unbind()
	effect.blendShader.Unbind()

	viewport.Pop()
}
