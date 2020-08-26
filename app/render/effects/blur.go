package effects

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/render/framebuffer"
	"github.com/wieku/danser-go/app/render/texture"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/glhf"
	"io/ioutil"
	"math"
)

type BlurEffect struct {
	blurShader *glhf.Shader
	fbo1       *framebuffer.Framebuffer
	fbo2       *framebuffer.Framebuffer
	kernelSize mgl32.Vec2
	sigma      mgl32.Vec2
	size       mgl32.Vec2
	fboSlice   *glhf.VertexSlice
}

func NewBlurEffect(width, height int) *BlurEffect {
	effect := &BlurEffect{}
	vertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}

	uniformFormat := glhf.AttrFormat{
		{Name: "tex", Type: glhf.Int},
		{Name: "kernelSize", Type: glhf.Vec2},
		{Name: "direction", Type: glhf.Vec2},
		{Name: "sigma", Type: glhf.Vec2},
		{Name: "size", Type: glhf.Vec2},
	}

	var err error
	vert, _ := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/blur.fsh")
	effect.blurShader, err = glhf.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic("Blur: " + err.Error())
	}

	effect.fboSlice = glhf.MakeVertexSlice(effect.blurShader, 6, 6)
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

	effect.fbo1 = framebuffer.NewFrame(width, height, true, false)
	effect.fbo2 = framebuffer.NewFrame(width, height, true, false)
	effect.SetBlur(0, 0)
	effect.size = mgl32.Vec2{float32(width), float32(height)}
	return effect
}

func (effect *BlurEffect) SetBlur(blurX, blurY float64) {
	sigmaX, sigmaY := float32(blurX)*25, float32(blurY)*25
	kX := kernelSize(sigmaX)
	if kX == 0 {
		sigmaX = 1.0
	}
	kY := kernelSize(sigmaY)
	if kY == 0 {
		sigmaY = 1.0
	}
	effect.kernelSize = mgl32.Vec2{float32(kX), float32(kY)}
	effect.sigma = mgl32.Vec2{sigmaX, sigmaY}
}

func gauss(x int, sigma float32) float32 {
	factor := float32(0.398942)
	return factor * float32(math.Exp(-0.5*float64(x*x)/float64(sigma*sigma))) / sigma
}

func kernelSize(sigma float32) int {
	if sigma == 0 {
		return 0
	}
	baseG := gauss(0, sigma) * 0.1
	max := 200

	for i := 1; i <= max; i++ {
		if gauss(i, sigma) < baseG {
			return i - 1
		}
	}
	return max
}

func (effect *BlurEffect) Begin() {
	effect.fbo1.Begin()
	glhf.Clear(0, 0, 0, 1)
	gl.Viewport(0, 0, int32(effect.fbo1.Texture().GetWidth()), int32(effect.fbo1.Texture().GetHeight()))
}

func (effect *BlurEffect) EndAndProcess() texture.Texture {
	effect.fbo1.End()

	effect.blurShader.Begin()
	effect.blurShader.SetUniformAttr(0, int32(0))
	effect.blurShader.SetUniformAttr(1, effect.kernelSize)
	effect.blurShader.SetUniformAttr(2, mgl32.Vec2{1, 0})
	effect.blurShader.SetUniformAttr(3, effect.sigma)
	effect.blurShader.SetUniformAttr(4, effect.size)

	effect.fboSlice.Begin()

	effect.fbo2.Begin()
	glhf.Clear(0, 0, 0, 0)

	effect.fbo1.Texture().Bind(0)

	effect.fboSlice.Draw()

	effect.fbo2.End()

	effect.fbo1.Begin()
	glhf.Clear(0, 0, 0, 0)

	effect.fbo2.Texture().Bind(0)

	effect.blurShader.SetUniformAttr(2, mgl32.Vec2{0, 1})
	effect.fboSlice.Draw()

	effect.fbo1.End()

	effect.fboSlice.End()
	effect.blurShader.End()
	gl.Viewport(0, 0, int32(settings.Graphics.GetWidth()), int32(settings.Graphics.GetHeight()))
	return effect.fbo1.Texture()
}
