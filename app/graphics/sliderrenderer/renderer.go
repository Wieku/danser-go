package sliderrenderer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"io/ioutil"
)

var sliderShader *shader.RShader
var colorShader *shader.RShader

var colorVAO *buffer.VertexArrayObject

var framebuffer *buffer.Framebuffer

var fboSprite *sprite.Sprite
var batch *batches.SpriteBatch

func InitRenderer() {

	passSource, err := ioutil.ReadFile("assets/shaders/sliderpass.vsh")

	if err != nil {
		panic(err)
	}

	sliderShader = shader.NewRShader(shader.NewSource(string(passSource), shader.Vertex))

	colorVSource, err := ioutil.ReadFile("assets/shaders/slidercolor.vsh")

	if err != nil {
		panic(err)
	}

	colorFSource, err := ioutil.ReadFile("assets/shaders/slidercolor.fsh")

	if err != nil {
		panic(err)
	}

	colorShader = shader.NewRShader(shader.NewSource(string(colorVSource), shader.Vertex), shader.NewSource(string(colorFSource), shader.Fragment))

	colorVAO = buffer.NewVertexArrayObject()

	colorVAO.AddVBO(
		"default",
		6,
		0,
		attribute.Format{
			{Name: "in_position", Type: attribute.Vec2},
			{Name: "in_tex_coord", Type: attribute.Vec2},
		},
	)

	colorVAO.SetData("default", 0, []float32{
		-0.5, -0.5, 0, 1,
		0.5, -0.5, 1, 1,
		0.5, 0.5, 1, 0,
		0.5, 0.5, 1, 0,
		-0.5, 0.5, 0, 0,
		-0.5, -0.5, 0, 1,
	})

	colorVAO.Bind()
	colorVAO.Attach(colorShader)
	colorVAO.Unbind()

	framebuffer = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false, true)
	region := framebuffer.Texture().GetRegion()
	fboSprite = sprite.NewSpriteSingle(&region, 0, bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)
	batch = batches.NewSpriteBatchSize(1)
	batch.SetCamera(mgl32.Ortho(0, float32(settings.Graphics.GetWidth()), 0, float32(settings.Graphics.GetHeight()), -1, 1))
}

func BeginRenderer() {
	if sliderShader == nil {
		InitRenderer()
	}

	colorShader.Bind()
	colorVAO.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.SrcAlpha, blend.OneMinusSrcAlpha)
}

func EndRenderer() {
	colorVAO.Unbind()
	colorShader.Unbind()

	blend.Pop()
}

func BeginRendererMerge() {
	BeginRenderer()

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	framebuffer.Begin()
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	blend.Push()
	blend.Enable()
	blend.SetFunctionSeparate(blend.SrcAlpha, blend.OneMinusSrcAlpha, blend.One, blend.OneMinusSrcAlpha)
}

func EndRendererMerge() {
	blend.Pop()

	framebuffer.End()

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)

	EndRenderer()

	batch.Begin()
	fboSprite.Draw(0, batch)
	batch.End()
}

func drawSlider(sprite *sprite.Sprite, stackOffset bmath.Vector2f, scale float32, text texture.Texture, color mgl32.Vec4, prev mgl32.Vec4, projection mgl32.Mat4) {
	colorShader.SetUniform("projection", projection)
	colorShader.SetUniform("col_border", color)
	if settings.Objects.EnableCustomSliderBorderGradientOffset {
		colorShader.SetUniform("col_border1", utils.GetColorShifted(color, settings.Objects.SliderBorderGradientOffset))
	} else {
		colorShader.SetUniform("col_border1", prev)
	}

	text.Bind(0)

	colorShader.SetUniform("tex", int32(0))
	colorShader.SetUniform("position", mgl32.Vec2{sprite.GetPosition().X32() + stackOffset.X, sprite.GetPosition().Y32() + stackOffset.Y})
	colorShader.SetUniform("size", mgl32.Vec2{sprite.GetScale().X32() * float32(text.GetWidth()), sprite.GetScale().Y32() * float32(text.GetHeight())})
	colorShader.SetUniform("cutoff", scale/float32(settings.Audio.BeatScale))

	colorVAO.Draw()
}
