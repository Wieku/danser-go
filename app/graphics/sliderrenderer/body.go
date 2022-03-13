package sliderrenderer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const maxSliderPoints = 10000

type Body struct {
	framebuffer    *buffer.Framebuffer
	bodySprite     *sprite.Sprite
	vao            *buffer.VertexArrayObject
	baseProjection mgl32.Mat4

	maxInstances  int
	previousStart int
	previousEnd   int

	topLeft     vector.Vector2f
	bottomRight vector.Vector2f
	points      []vector.Vector2f

	radius float32

	disposed         bool
	distortionMatrix mgl32.Mat4
}

func NewBody(curve *curves.MultiCurve, hardRock bool, hitCircleRadius float32) *Body {
	if sliderShader == nil {
		InitRenderer()
	}

	body := new(Body)

	body.distortionMatrix = mgl32.Ident4()
	body.radius = hitCircleRadius
	body.previousStart = -1

	body.setupPoints(curve, hardRock)

	if len(body.points) > 0 {
		body.setupVAO()
	}

	return body
}

func pointAt(curve *curves.MultiCurve, t float32, hardRock bool) vector.Vector2f {
	point := curve.PointAt(t)
	if hardRock {
		point.Y = 384 - point.Y
	}

	return point
}

func (body *Body) setupPoints(curve *curves.MultiCurve, hardRock bool) {
	length := curve.GetLength()
	numPoints := math32.Min(math32.Ceil(length*float32(settings.Objects.Sliders.Quality.PathLevelOfDetail)/100.0), maxSliderPoints)

	if numPoints > 0 {
		body.topLeft = pointAt(curve, 0, hardRock)
		body.bottomRight = pointAt(curve, 0, hardRock)

		for i := 0; i <= int(numPoints); i++ {
			point := pointAt(curve, float32(i)/numPoints, hardRock)

			body.topLeft.X = math32.Min(body.topLeft.X, point.X)
			body.topLeft.Y = math32.Min(body.topLeft.Y, point.Y)

			body.bottomRight.X = math32.Max(body.bottomRight.X, point.X)
			body.bottomRight.Y = math32.Max(body.bottomRight.Y, point.Y)

			multiplier := float32(1.0)
			if settings.Objects.Sliders.SliderDistortions {
				multiplier = 3.0 //larger allowable area, we want to see distorted sliders "fully"
			}

			if point.X >= -camera.OsuWidth*multiplier && point.X <= camera.OsuWidth*2*multiplier && point.Y >= -camera.OsuHeight*multiplier && point.Y <= camera.OsuHeight*2*multiplier {
				body.points = append(body.points, point)
			}
		}
	}

	body.maxInstances = len(body.points)
}

func (body *Body) setupVAO() {
	body.vao = buffer.NewVertexArrayObject()

	unitCircle := createUnitCircle(int(settings.Objects.Sliders.Quality.CircleLevelOfDetail))

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

	body.vao.Attach(sliderShader)
}

func (body *Body) DrawBase(head, tail float64, baseProjView mgl32.Mat4) {
	if len(body.points) == 0 {
		return
	}

	body.ensureFBO(baseProjView)

	// Don't render to nonexistent framebuffer
	if body.framebuffer == nil {
		return
	}

	startInstance := int(math.Ceil(mutils.ClampF(head, 0.0, 1.0) * float64(body.maxInstances-1)))
	endInstance := int(math.Floor(mutils.ClampF(tail, 0.0, 1.0) * float64(body.maxInstances-1)))

	if startInstance > endInstance {
		startInstance, endInstance = endInstance, startInstance
	}

	// State is the same so no need to render
	if body.previousStart == startInstance && body.previousEnd == endInstance {
		return
	}

	body.framebuffer.Bind()
	viewport.Push(int(body.framebuffer.Texture().GetWidth()), int(body.framebuffer.Texture().GetHeight()))

	//TODO: Make depth manager
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	body.framebuffer.ClearDepth()

	sliderShader.Bind()

	sliderShader.SetUniform("projection", body.baseProjection)
	sliderShader.SetUniform("scale", body.radius*float32(settings.Audio.BeatScale))
	sliderShader.SetUniform("distort", body.distortionMatrix)

	body.vao.Bind()
	body.vao.DrawInstanced(startInstance, endInstance-startInstance+1)
	body.vao.Unbind()

	sliderShader.Unbind()

	body.framebuffer.Unbind()

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)

	viewport.Pop()

	body.previousStart = startInstance
	body.previousEnd = endInstance
}

func (body *Body) DrawNormal(projection mgl32.Mat4, stackOffset vector.Vector2f, scale float32, bodyInner, bodyOuter, borderInner, borderOuter color2.Color) {
	if body.framebuffer == nil || body.disposed || len(body.points) == 0 {
		return
	}

	drawSlider(body.bodySprite, stackOffset, scale, body.framebuffer.Texture(), bodyInner, bodyOuter, borderInner, borderOuter, projection)
}

func (body *Body) ensureFBO(baseProjView mgl32.Mat4) {
	if body.framebuffer != nil || body.disposed {
		return
	}

	invProjView := baseProjView.Inv()

	scaleX := 1.0
	scaleY := 1.0

	if settings.Objects.Sliders.SliderDistortions {
		tLS := baseProjView.Mul4x1(mgl32.Vec4{body.topLeft.X, body.topLeft.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)
		bRS := baseProjView.Mul4x1(mgl32.Vec4{body.bottomRight.X, body.bottomRight.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)

		wS := float32(32768 / (settings.Graphics.GetWidthF()))
		hS := float32(32768 / (settings.Graphics.GetHeightF()))

		if -tLS.X()+bRS.X() > wS {
			scaleX = float64(wS / (-tLS.X() + bRS.X()))
		}

		if bRS.Y() < -hS {
			scaleY = float64(-hS / bRS.Y())
		}
	}

	tS := invProjView.Mul4x1(mgl32.Vec4{-1, 1, 0.0, 1.0})
	body.distortionMatrix = mgl32.Translate3D(tS.X()*float32(scaleX), tS.Y()*float32(scaleY), 0).Mul4(mgl32.Scale3D(float32(scaleX), float32(scaleY), 1)).Mul4(mgl32.Translate3D(-tS.X(), -tS.Y(), 0))

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

	var topLeftScreenE vector.Vector2f
	var bottomRightScreenE vector.Vector2f

	topLeftScreenE.X = math32.Max(tLW.X(), body.topLeft.X-body.radius*float32(settings.Audio.BeatScale))
	topLeftScreenE.Y = math32.Max(tLW.Y(), body.topLeft.Y-body.radius*float32(settings.Audio.BeatScale))

	// make upper left part of fbo bigger to fit distorted sliders
	topLeftScreenE.X = (topLeftScreenE.X-tLW.X())*float32(scaleX) + tLW.X()
	topLeftScreenE.Y = (topLeftScreenE.Y-tLW.Y())*float32(scaleY) + tLW.Y()

	bottomRightScreenE.X = math32.Min(bRW.X(), body.bottomRight.X+body.radius*float32(settings.Audio.BeatScale))
	bottomRightScreenE.Y = math32.Min(bRW.Y(), body.bottomRight.Y+body.radius*float32(settings.Audio.BeatScale))

	tLS := baseProjView.Mul4x1(mgl32.Vec4{topLeftScreenE.X, topLeftScreenE.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)
	bRS := baseProjView.Mul4x1(mgl32.Vec4{bottomRightScreenE.X, bottomRightScreenE.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)

	topLeftScreen := vector.NewVec2f(tLS.X(), tLS.Y()).Mult(vector.NewVec2f(float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())))
	bottomRightScreen := vector.NewVec2f(bRS.X(), bRS.Y()).Mult(vector.NewVec2f(float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())))

	dimensions := bottomRightScreen.Sub(topLeftScreen).Abs()
	body.framebuffer = buffer.NewFrameDepth(int(dimensions.X), int(dimensions.Y), true)

	tex := body.framebuffer.Texture().GetRegion()

	body.bodySprite = sprite.NewSpriteSingle(&tex, 0, bottomRightScreenE.Sub(topLeftScreenE).Scl(0.5).Add(topLeftScreenE).Copy64(), vector.Centre)
	body.bodySprite.SetScale(float64((bottomRightScreenE.X - topLeftScreenE.X) / dimensions.X))
	body.bodySprite.SetVFlip(true)

	body.baseProjection = mgl32.Ortho(topLeftScreenE.X, bottomRightScreenE.X, bottomRightScreenE.Y, topLeftScreenE.Y, 1, -1)
}

func (body *Body) Dispose() {
	if body.disposed || body.framebuffer == nil {
		return
	}

	body.disposed = true
	body.framebuffer.Dispose()
}
