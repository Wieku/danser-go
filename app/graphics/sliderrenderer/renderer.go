package sliderrenderer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	batch2 "github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

var capShader *shader.RShader
var lineShader *shader.RShader

var colorShader *shader.RShader

var colorVAO *buffer.VertexArrayObject

var framebuffer *buffer.Framebuffer

var fboSprite *sprite.Sprite
var batch *batch2.QuadBatch

func InitRenderer() {
	capsSource, err := assets.GetString("assets/shaders/slidercaps.vsh")

	if err != nil {
		panic(err)
	}

	capShader = shader.NewRShader(shader.NewSource(capsSource, shader.Vertex))

	linesSource, err := assets.GetString("assets/shaders/sliderlines.vsh")

	if err != nil {
		panic(err)
	}

	lineShader = shader.NewRShader(shader.NewSource(linesSource, shader.Vertex))

	colorVSource, err := assets.GetString("assets/shaders/slidercolor.vsh")

	if err != nil {
		panic(err)
	}

	colorFSource, err := assets.GetString("assets/shaders/slidercolor.fsh")

	if err != nil {
		panic(err)
	}

	colorShader = shader.NewRShader(shader.NewSource(colorVSource, shader.Vertex), shader.NewSource(colorFSource, shader.Fragment))

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

	colorVAO.Attach(colorShader)

	framebuffer = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false, true)
	region := framebuffer.Texture().GetRegion()
	fboSprite = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), vector.Centre)
	batch = batch2.NewQuadBatchSize(1)
}

func BeginRenderer() {
	if capShader == nil {
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

	framebuffer.Bind()
	framebuffer.ClearColor(0, 0, 0, 0)
	framebuffer.ClearDepth()

	blend.Push()
	blend.Disable()
}

func EndRendererMerge() {
	blend.Pop()

	framebuffer.Unbind()

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)

	EndRenderer()

	batch.Begin()
	batch.SetCamera(mgl32.Ortho(0, float32(settings.Graphics.GetWidth()), 0, float32(settings.Graphics.GetHeight()), -1, 1))

	fboSprite.Draw(0, batch)
	batch.End()
}

func drawSlider(sprite *sprite.Sprite, stackOffset vector.Vector2f, scale float32, text texture.Texture, bodyInner, bodyOuter, borderInner, borderOuter color2.Color, projection mgl32.Mat4) {
	colorShader.SetUniform("projection", projection)

	colorShader.SetUniform("col_border", borderInner)
	colorShader.SetUniform("col_border1", borderOuter)

	colorShader.SetUniform("col_body", bodyInner)
	colorShader.SetUniform("col_body1", bodyOuter)

	text.Bind(0)

	colorShader.SetUniform("tex", int32(0))
	colorShader.SetUniform("position", mgl32.Vec2{sprite.GetPosition().X32() + stackOffset.X, sprite.GetPosition().Y32() + stackOffset.Y})
	colorShader.SetUniform("size", mgl32.Vec2{sprite.GetScale().X32() * float32(text.GetWidth()), sprite.GetScale().Y32() * float32(text.GetHeight())})
	colorShader.SetUniform("cutoff", scale/float32(settings.Audio.BeatScale))
	colorShader.SetUniform("borderWidth", mutils.Clamp(float32(settings.Objects.Sliders.BorderWidth), 0.0, 10.0))

	colorVAO.Draw()
}
