package sliderrenderer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/curves"
	"github.com/wieku/danser-go/app/render/sprites"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/math32"
	"math"
)

const maxSliderPoints = 10000

type Body struct {
	framebuffer    *buffer.Framebuffer
	bodySprite     *sprites.Sprite
	vao            *buffer.VertexArrayObject
	baseProjection mgl32.Mat4

	maxInstances  int
	previousStart int
	previousEnd   int

	topLeft     bmath.Vector2f
	bottomRight bmath.Vector2f
	points      []bmath.Vector2f

	radius float32

	disposed bool
}

func NewBody(curve *curves.MultiCurve, hitCircleRadius float32) *Body {
	if sliderShader == nil {
		InitRenderer()
	}

	body := new(Body)

	body.radius = hitCircleRadius
	body.previousStart = -1

	body.setupPoints(curve)
	body.setupVAO()

	return body
}

func (body *Body) setupPoints(curve *curves.MultiCurve) {
	length := curve.GetLength()
	numPoints := math32.Min(math32.Ceil(length*float32(settings.Objects.SliderPathLOD)/100.0), maxSliderPoints)

	if numPoints > 0 {
		body.topLeft = curve.PointAt(0)
		body.bottomRight = curve.PointAt(0)

		for i := 0; i <= int(numPoints); i++ {
			point := curve.PointAt(float32(i) / numPoints)

			body.topLeft.X = math32.Min(body.topLeft.X, point.X)
			body.topLeft.Y = math32.Min(body.topLeft.Y, point.Y)

			body.bottomRight.X = math32.Max(body.bottomRight.X, point.X)
			body.bottomRight.Y = math32.Max(body.bottomRight.Y, point.Y)

			multiplier := float32(1.0)
			if settings.Objects.SliderDistortions {
				multiplier = 3.0 //larger allowable area, we want to see distorted sliders "fully"
			}

			if point.X >= -bmath.OsuWidth*multiplier && point.X <= bmath.OsuWidth*2*multiplier && point.Y >= -bmath.OsuHeight*multiplier && point.Y <= bmath.OsuHeight*2*multiplier {
				body.points = append(body.points, point)
			}
		}
	}

	body.maxInstances = len(body.points)
}

func (body *Body) setupVAO() {

	body.vao = buffer.NewVertexArrayObject()

	unitCircle := createUnitCircle(int(settings.Objects.SliderLOD))

	body.vao.AddVBO("default", len(unitCircle)/3, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
	})

	body.vao.SetData("default", 0, unitCircle)

	points := make([]float32, body.maxInstances*2)

	for i := 0; i < body.maxInstances; i++ {
		points[i*2] = body.points[i].X
		points[i*2+1] = body.points[i].Y
	}

	body.vao.AddVBO("points", len(points)/2, 1, attribute.Format{
		{Name: "center", Type: attribute.Vec2},
	})

	body.vao.SetData("points", 0, points)

	body.vao.Bind()
	body.vao.Attach(sliderShader)
	body.vao.Unbind()
}

func (body *Body) DrawBase(head, tail float64, baseProjView mgl32.Mat4) {
	body.ensureFBO(baseProjView)

	// Don't render to nonexistent framebuffer
	if body.framebuffer == nil {
		return
	}

	startInstance := int(math.Ceil(bmath.ClampF64(head, 0.0, 1.0) * float64(body.maxInstances-1)))
	endInstance := int(math.Floor(bmath.ClampF64(tail, 0.0, 1.0) * float64(body.maxInstances-1)))

	if startInstance > endInstance {
		startInstance, endInstance = endInstance, startInstance
	}

	// State is the same so no need to render
	if body.previousStart == startInstance && body.previousEnd == endInstance {
		return
	}

	body.framebuffer.Begin()

	blend.Push()
	blend.Disable()

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	gl.Clear(gl.DEPTH_BUFFER_BIT)

	viewport.Push(int(body.framebuffer.Texture().GetWidth()), int(body.framebuffer.Texture().GetHeight()))

	sliderShader.Bind()

	sliderShader.SetUniform("projection", body.baseProjection)
	sliderShader.SetUniform("scale", body.radius*float32(settings.Beat.BeatScale))
	sliderShader.SetUniform("distort", mgl32.Ident4())

	body.vao.Bind()
	body.vao.DrawInstanced(startInstance, endInstance-startInstance+1)
	body.vao.Unbind()

	sliderShader.Unbind()

	body.framebuffer.End()

	viewport.Pop()

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)

	blend.Pop()

	body.previousStart = startInstance
	body.previousEnd = endInstance
}

func (body *Body) DrawNormal(projection mgl32.Mat4, stackOffset bmath.Vector2f, color mgl32.Vec4, prev mgl32.Vec4) {
	drawSlider(body.bodySprite, stackOffset, body.framebuffer.Texture(), color, prev, projection)
}

func (body *Body) ensureFBO(baseProjView mgl32.Mat4) {
	if body.framebuffer != nil || body.disposed {
		return
	}

	invProjView := baseProjView.Inv()

	multiplierX := float32(1.0)
	multiplierY := float32(1.0)

	aspect := float32(settings.Graphics.GetAspectRatio())

	if aspect > 1 {
		multiplierY = 1 / aspect
	} else {
		multiplierX = aspect
	}

	overdraw := math32.Sqrt(multiplierX*multiplierX + multiplierY*multiplierY)

	multiplierX = 1 / multiplierX * overdraw
	multiplierY = 1 / multiplierY * overdraw

	tLW := invProjView.Mul4x1(mgl32.Vec4{-multiplierX, multiplierY, 0.0, 1.0})
	bRW := invProjView.Mul4x1(mgl32.Vec4{multiplierX, -multiplierY, 0.0, 1.0})

	var topLeftScreenE bmath.Vector2f
	var bottomRightScreenE bmath.Vector2f

	topLeftScreenE.X = math32.Max(tLW.X(), body.topLeft.X-body.radius*float32(settings.Beat.BeatScale))
	topLeftScreenE.Y = math32.Max(tLW.Y(), body.topLeft.Y-body.radius*float32(settings.Beat.BeatScale))

	bottomRightScreenE.X = math32.Min(bRW.X(), body.bottomRight.X+body.radius*float32(settings.Beat.BeatScale))
	bottomRightScreenE.Y = math32.Min(bRW.Y(), body.bottomRight.Y+body.radius*float32(settings.Beat.BeatScale))

	tLS := baseProjView.Mul4x1(mgl32.Vec4{topLeftScreenE.X, topLeftScreenE.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)
	bRS := baseProjView.Mul4x1(mgl32.Vec4{bottomRightScreenE.X, bottomRightScreenE.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)

	topLeftScreen := bmath.NewVec2f(tLS.X(), tLS.Y()).Mult(bmath.NewVec2f(float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())))
	bottomRightScreen := bmath.NewVec2f(bRS.X(), bRS.Y()).Mult(bmath.NewVec2f(float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())))

	dimensions := bottomRightScreen.Sub(topLeftScreen).Abs()
	body.framebuffer = buffer.NewFrameDepth(int(dimensions.X), int(dimensions.Y), true)

	tex := body.framebuffer.Texture().GetRegion()

	body.bodySprite = sprites.NewSpriteSingle(&tex, 0, bottomRightScreenE.Sub(topLeftScreenE).Scl(0.5).Add(topLeftScreenE).Copy64(), bmath.Origin.Centre)
	body.bodySprite.SetScale(float64((bottomRightScreenE.X - topLeftScreenE.X) / dimensions.X))
	body.bodySprite.SetVFlip(true)

	body.baseProjection = mgl32.Ortho(topLeftScreenE.X, bottomRightScreenE.X, bottomRightScreenE.Y, topLeftScreenE.Y, 1, -1)
}