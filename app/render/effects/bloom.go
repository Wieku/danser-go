package effects

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"io/ioutil"
)

type BloomEffect struct {
	colFilter     *shader.RShader
	combineShader *shader.RShader
	fbo           *buffer.Framebuffer

	blurEffect *BlurEffect

	blur, threshold, power float64
	fboSlice               *buffer.VertexArrayObject
}

func NewBloomEffect(width, height int) *BloomEffect {
	effect := new(BloomEffect)

	vert, err := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := ioutil.ReadFile("assets/shaders/brightfilter.fsh")
	if err != nil {
		panic(err)
	}

	effect.colFilter = shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

	frag, err = ioutil.ReadFile("assets/shaders/combine.fsh")
	if err != nil {
		panic(err)
	}

	effect.combineShader = shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

	effect.fboSlice = buffer.NewVertexArrayObject()

	effect.fboSlice.AddVBO("default", 6, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
		{Name: "in_tex_coord", Type: attribute.Vec2},
	})

	effect.fboSlice.SetData("default", 0, []float32{
		-1, -1, 0, 0, 0,
		1, -1, 0, 1, 0,
		-1, 1, 0, 0, 1,
		1, -1, 0, 1, 0,
		1, 1, 0, 1, 1,
		-1, 1, 0, 0, 1,
	})

	effect.fboSlice.Bind()
	effect.fboSlice.Attach(effect.colFilter)
	effect.fboSlice.Unbind()

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
	effect.fbo.Begin()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (effect *BloomEffect) EndAndRender() {
	effect.fbo.End()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.SrcAlpha, blend.OneMinusSrcAlpha)

	effect.blurEffect.Begin()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	effect.colFilter.Bind()
	effect.colFilter.SetUniform("tex", int32(0))
	effect.colFilter.SetUniform("threshold", float32(effect.threshold))

	effect.fbo.Texture().Bind(0)

	effect.fboSlice.Bind()
	effect.fboSlice.Draw()

	effect.colFilter.Unbind()

	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)

	texture := effect.blurEffect.EndAndProcess()

	effect.combineShader.Bind()
	effect.combineShader.SetUniform("tex", int32(0))
	effect.combineShader.SetUniform("tex2", int32(1))
	effect.combineShader.SetUniform("power", float32(effect.power))

	effect.fbo.Texture().Bind(0)

	texture.Bind(1)

	effect.fboSlice.Draw()
	effect.fboSlice.Unbind()

	effect.combineShader.Unbind()

	blend.Pop()
}
