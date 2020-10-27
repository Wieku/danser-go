package effects

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
)

type BloomEffect struct {
	filterShader  *shader.RShader
	combineShader *shader.RShader
	fbo           *buffer.Framebuffer

	blurEffect *BlurEffect

	blur, threshold, power float64
	vao                    *buffer.VertexArrayObject
}

func NewBloomEffect(width, height int) *BloomEffect {
	effect := new(BloomEffect)

	vert, err := assets.GetString("assets/shaders/fbopass.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/brightfilter.fsh")
	if err != nil {
		panic(err)
	}

	effect.filterShader = shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

	frag, err = assets.GetString("assets/shaders/combine.fsh")
	if err != nil {
		panic(err)
	}

	effect.combineShader = shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

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

	effect.vao.Bind()
	effect.vao.Attach(effect.filterShader)
	effect.vao.Unbind()

	effect.fbo = buffer.NewFrame(width, height, true, false)

	effect.threshold = 0.7
	effect.blur = 0.3
	effect.power = 1.2

	effect.blurEffect = NewBlurEffect(width, height)
	effect.blurEffect.SetBlur(effect.blur, effect.blur)

	return effect
}

func (effect *BloomEffect) SetThreshold(threshold float64) {
	effect.threshold = threshold
}

func (effect *BloomEffect) SetBlur(blur float64) {
	effect.blur = blur
	effect.blurEffect.SetBlur(blur, blur)
}

func (effect *BloomEffect) SetPower(power float64) {
	effect.power = power
}

func (effect *BloomEffect) Begin() {
	effect.fbo.Bind()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (effect *BloomEffect) EndAndRender() {
	effect.fbo.Unbind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.SrcAlpha, blend.OneMinusSrcAlpha)

	effect.blurEffect.Begin()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	effect.filterShader.Bind()
	effect.filterShader.SetUniform("tex", int32(0))
	effect.filterShader.SetUniform("threshold", float32(effect.threshold))

	effect.fbo.Texture().Bind(0)

	effect.vao.Bind()
	effect.vao.Draw()

	effect.filterShader.Unbind()

	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)

	texture := effect.blurEffect.EndAndProcess()

	effect.combineShader.Bind()
	effect.combineShader.SetUniform("tex", int32(0))
	effect.combineShader.SetUniform("tex2", int32(1))
	effect.combineShader.SetUniform("power", float32(effect.power))

	effect.fbo.Texture().Bind(0)

	texture.Bind(1)

	effect.vao.Draw()
	effect.vao.Unbind()

	effect.combineShader.Unbind()

	blend.Pop()
}
