package effects

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"math"
)

type BlurEffect struct {
	blurShader *shader.RShader
	fbo1       *buffer.Framebuffer
	fbo2       *buffer.Framebuffer
	kernelSize mgl32.Vec2
	sigma      mgl32.Vec2
	size       mgl32.Vec2
	vao        *buffer.VertexArrayObject
}

func NewBlurEffect(width, height int) *BlurEffect {
	effect := new(BlurEffect)
	effect.size = mgl32.Vec2{float32(width), float32(height)}
	effect.SetBlur(0, 0)

	vert, err := assets.GetString("assets/shaders/fbopass.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/blur.fsh")
	if err != nil {
		panic(err)
	}

	effect.blurShader = shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

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
	effect.vao.Attach(effect.blurShader)
	effect.vao.Unbind()

	effect.fbo1 = buffer.NewFrame(width, height, true, false)
	effect.fbo2 = buffer.NewFrame(width, height, true, false)

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
	effect.fbo1.Bind()
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.Viewport(0, 0, int32(effect.fbo1.Texture().GetWidth()), int32(effect.fbo1.Texture().GetHeight()))
}

func (effect *BlurEffect) EndAndProcess() texture.Texture {
	effect.fbo1.Unbind()

	effect.blurShader.Bind()
	effect.blurShader.SetUniform("tex", int32(0))
	effect.blurShader.SetUniform("kernelSize", effect.kernelSize)
	effect.blurShader.SetUniform("direction", mgl32.Vec2{1, 0})
	effect.blurShader.SetUniform("sigma", effect.sigma)
	effect.blurShader.SetUniform("size", effect.size)

	effect.vao.Bind()

	effect.fbo2.Bind()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	effect.fbo1.Texture().Bind(0)

	effect.vao.Draw()

	effect.fbo2.Unbind()

	effect.fbo1.Bind()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	effect.fbo2.Texture().Bind(0)

	effect.blurShader.SetUniform("direction", mgl32.Vec2{0, 1})
	effect.vao.Draw()

	effect.fbo1.Unbind()

	effect.vao.Unbind()
	effect.blurShader.Unbind()
	gl.Viewport(0, 0, int32(settings.Graphics.GetWidth()), int32(settings.Graphics.GetHeight()))
	return effect.fbo1.Texture()
}
