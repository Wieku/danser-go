package effects

import (
	"github.com/wieku/glhf"
	"io/ioutil"
	"github.com/go-gl/gl/v3.3-core/gl"
)

type BloomEffect struct {
	colFilter *glhf.Shader
	combineShader *glhf.Shader
	fbo *glhf.Frame

	blurEffect *BlurEffect

	blur, threshold, power float64
	fboSlice *glhf.VertexSlice
}

func NewBloomEffect(width, height int) *BloomEffect {
	effect := &BloomEffect{}
	vertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}

	uniformFormat := glhf.AttrFormat{
		{Name: "tex", Type: glhf.Int},
		{Name: "threshold", Type: glhf.Float},
	}

	var err error
	vert , _ := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	frag , _ := ioutil.ReadFile("assets/shaders/brightfilter.fsh")
	effect.colFilter, err = glhf.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic("BloomFilter: " + err.Error())
	}

	uniformFormat = glhf.AttrFormat{
		{Name: "tex", Type: glhf.Int},
		{Name: "tex2", Type: glhf.Int},
		{Name: "power", Type: glhf.Float},
	}
	frag , _ = ioutil.ReadFile("assets/shaders/combine.fsh")
	effect.combineShader, err = glhf.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic("BloomCombine: " + err.Error())
	}
	
	effect.fboSlice = glhf.MakeVertexSlice(effect.colFilter, 6, 6)
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

	effect.fbo = glhf.NewFrame(width, height, true, false)
	effect.fbo.Texture().Begin()
	effect.fbo.Texture().SetWrap(glhf.CLAMP_TO_EDGE)
	effect.fbo.Texture().End()
	
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
	glhf.Clear(0, 0, 0, 0)
}

func (effect *BloomEffect) EndAndRender() {
	effect.fbo.End()

	effect.blurEffect.Begin()

	effect.colFilter.Begin()
	effect.colFilter.SetUniformAttr(0, int32(0))
	effect.colFilter.SetUniformAttr(1, float32(effect.threshold))

	gl.ActiveTexture(gl.TEXTURE0)
	effect.fbo.Texture().Begin()

	effect.fboSlice.Begin()
	effect.fboSlice.Draw()
	effect.fboSlice.End()

	effect.fbo.Texture().End()

	effect.colFilter.End()

	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	texture := effect.blurEffect.EndAndProcess()

	effect.combineShader.Begin()
	effect.combineShader.SetUniformAttr(0, int32(0))
	effect.combineShader.SetUniformAttr(1, int32(1))
	effect.combineShader.SetUniformAttr(2, float32(effect.power))

	gl.ActiveTexture(gl.TEXTURE0)
	effect.fbo.Texture().Begin()

	gl.ActiveTexture(gl.TEXTURE1)
	texture.Begin()


	effect.fboSlice.Begin()
	effect.fboSlice.Draw()
	effect.fboSlice.End()

	texture.End()
	effect.fbo.Texture().End()

	effect.combineShader.End()

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}