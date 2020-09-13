package effects

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"io/ioutil"
)

type BloomEffect struct {
	colFilter     *shader.Shader
	combineShader *shader.Shader
	fbo           *buffer.Framebuffer

	blurEffect *BlurEffect

	blur, threshold, power float64
	fboSlice               *buffer.VertexSlice
}

func NewBloomEffect(width, height int) *BloomEffect {
	effect := &BloomEffect{}
	vertexFormat := shader.AttrFormat{
		{Name: "in_position", Type: shader.Vec3},
		{Name: "in_tex_coord", Type: shader.Vec2},
	}

	uniformFormat := shader.AttrFormat{
		{Name: "tex", Type: shader.Int},
		{Name: "threshold", Type: shader.Float},
	}

	var err error
	vert, _ := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/brightfilter.fsh")
	effect.colFilter, err = shader.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic("BloomFilter: " + err.Error())
	}

	uniformFormat = shader.AttrFormat{
		{Name: "tex", Type: shader.Int},
		{Name: "tex2", Type: shader.Int},
		{Name: "power", Type: shader.Float},
	}
	frag, _ = ioutil.ReadFile("assets/shaders/combine.fsh")
	effect.combineShader, err = shader.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic("BloomCombine: " + err.Error())
	}

	effect.fboSlice = buffer.MakeVertexSlice(effect.colFilter, 6, 6)
	effect.fboSlice.Begin()
	effect.fboSlice.SetVertexData([]float32{
		-1, -1, 0, 0, 0,
		1, -1, 0, 1, 0,
		-1, 1, 0, 0, 1,
		1, -1, 0, 1, 0,
		1, 1, 0, 1, 1,
		-1, 1, 0, 0, 1,
	})
	effect.fboSlice.End()

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

	effect.colFilter.Begin()
	effect.colFilter.SetUniformAttr(0, int32(0))
	effect.colFilter.SetUniformAttr(1, float32(effect.threshold))

	effect.fbo.Texture().Bind(0)

	effect.fboSlice.Begin()
	effect.fboSlice.Draw()
	effect.fboSlice.End()

	effect.colFilter.End()

	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)

	texture := effect.blurEffect.EndAndProcess()

	effect.combineShader.Begin()
	effect.combineShader.SetUniformAttr(0, int32(0))
	effect.combineShader.SetUniformAttr(1, int32(1))
	effect.combineShader.SetUniformAttr(2, float32(effect.power))

	effect.fbo.Texture().Bind(0)

	texture.Bind(1)

	effect.fboSlice.Begin()
	effect.fboSlice.Draw()
	effect.fboSlice.End()

	effect.combineShader.End()

	blend.Pop()
}
