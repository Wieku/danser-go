package sliderrenderer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
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
	"sort"
)

const LevelOfDetail = 50

type lineSection struct {
	curves.Linear

	length float32

	prevLength  float32
	prevIndices int
}

func (section *lineSection) pointAtLen(length float32) vector.Vector2f {
	return section.PointAt(section.capLen(length-section.prevLength) / section.length)
}

func (section *lineSection) capLen(length float32) float32 {
	return mutils.Clamp(length, 0, section.length)
}

type Body struct {
	framebuffer *buffer.Framebuffer
	bodySprite  *sprite.Sprite

	lineVAO *buffer.VertexArrayObject
	fillVAO *buffer.VertexArrayObject
	capVAO  *buffer.VertexArrayObject

	baseProjection mgl32.Mat4

	headLength float32
	tailLength float32

	headIndex int
	tailIndex int

	topLeft     vector.Vector2f
	bottomRight vector.Vector2f

	radius float32

	disposed         bool
	distortionMatrix mgl32.Mat4

	sections []*lineSection

	totalLength float32

	lineBuffer []float32
	capBuffer  []float32
}

func NewBody(curve *curves.MultiCurve, hardRock bool, hitCircleRadius float32) *Body {
	if capShader == nil {
		InitRenderer()
	}

	body := &Body{
		distortionMatrix: mgl32.Ident4(),
		radius:           hitCircleRadius,
		headLength:       -1,
		tailLength:       -1,
		headIndex:        -1,
		tailIndex:        -1,
		lineBuffer:       make([]float32, 3),
		capBuffer:        make([]float32, 4),
	}

	body.setupLinesAndBounds(curve, hardRock)

	if body.sections != nil && len(body.sections) > 0 {
		body.setupLineVAO()
		body.setupFillVAO()
		body.setupCapVAO()
	}

	return body
}

func (body *Body) setupLinesAndBounds(curve *curves.MultiCurve, hardRock bool) {
	lines := curve.GetLines()
	if lines == nil || len(lines) == 0 {
		return
	}

	body.topLeft = vector.NewVec2f(math.MaxFloat32, math.MaxFloat32)
	body.bottomRight = vector.NewVec2f(-math.MaxFloat32, -math.MaxFloat32)

	for _, line := range lines {
		if hardRock {
			line.Point1.Y = 384 - line.Point1.Y
			line.Point2.Y = 384 - line.Point2.Y
		}

		length := line.GetLength()

		if length == 0 {
			continue
		}

		body.topLeft.X = min(body.topLeft.X, min(line.Point1.X, line.Point2.X))
		body.topLeft.Y = min(body.topLeft.Y, min(line.Point1.Y, line.Point2.Y))

		body.bottomRight.X = max(body.bottomRight.X, max(line.Point1.X, line.Point2.X))
		body.bottomRight.Y = max(body.bottomRight.Y, max(line.Point1.Y, line.Point2.Y))

		body.sections = append(body.sections, &lineSection{
			Linear:     line,
			prevLength: body.totalLength,
			length:     length,
		})

		body.totalLength += length
	}
}

func (body *Body) setupLineVAO() {
	body.lineVAO = buffer.NewVertexArrayObject()

	unitLine, indices := createUnitLine()

	body.lineVAO.AddVBO("default", len(unitLine)/3, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
	})

	body.lineVAO.SetData("default", 0, unitLine)

	ibo := buffer.NewIndexBufferObject(len(indices))

	ibo.SetData(0, indices)

	body.lineVAO.AttachIBO(ibo)

	instances := make([]float32, len(body.sections)*5)

	for i, section := range body.sections {
		b := i * 5

		instances[b], instances[b+1] = section.Point1.X, section.Point1.Y
		instances[b+2] = section.length

		agl := -section.GetEndAngle()

		instances[b+3], instances[b+4] = math32.Sin(agl), math32.Cos(agl)
	}

	body.lineVAO.AddVBO("lines", len(instances)/5, 1, attribute.Format{
		{Name: "pos1", Type: attribute.Vec2},
		{Name: "lineLength", Type: attribute.Float},
		{Name: "sinCos", Type: attribute.Vec2},
	})

	body.lineVAO.SetData("lines", 0, instances)

	body.lineVAO.Attach(lineShader)
}

func (body *Body) setupFillVAO() {
	lodAngle := (2 * math32.Pi) / float32(LevelOfDetail)

	vertexBuffer := make([]float32, 0)
	indexBuffer := make([]uint32, 0)

	// Temporary buffers
	cvBuffer := make([]float32, LevelOfDetail*5+10)
	ciBuffer := make([]uint32, LevelOfDetail*3)

	for i := 0; i < len(body.sections)-1; i++ {
		cSection := body.sections[i]
		nSection := body.sections[i+1]

		cSection.prevIndices = len(indexBuffer)

		ePoint := cSection.Point2

		cAngle := cSection.GetEndAngle()
		nAngle := nSection.GetEndAngle()

		diff := mutils.SanitizeAngleArc(nAngle - cAngle) // Switch back to atan2 angles

		dir := -mutils.Signum(diff) // We need to go the opposite direction

		nAngle += dir * math32.Pi / 2

		tCount := int(math32.Ceil(math32.Abs(diff) / lodAngle))

		if tCount > 0 {
			iBase := uint32(len(vertexBuffer) / 5)

			cvBuffer[3], cvBuffer[4] = ePoint.X, ePoint.Y

			for j := 0; j <= tCount; j++ {
				p := vector.NewVec2fRad(nAngle+dir*lodAngle*float32(j), 1)

				bv := 5 + j*5

				cvBuffer[bv], cvBuffer[bv+1], cvBuffer[bv+2] = p.X, p.Y, 1.0
				cvBuffer[bv+3], cvBuffer[bv+4] = ePoint.X, ePoint.Y

				if j < tCount {
					bi := j * 3
					ii := uint32(j) + iBase

					ciBuffer[bi], ciBuffer[bi+1], ciBuffer[bi+2] = iBase, ii+1, ii+2
				}
			}

			vertexBuffer = append(vertexBuffer, cvBuffer[:tCount*5+10]...)
			indexBuffer = append(indexBuffer, ciBuffer[:tCount*3]...)
		}
	}

	body.sections[len(body.sections)-1].prevIndices = len(indexBuffer)

	body.fillVAO = buffer.NewVertexArrayObject()

	body.fillVAO.AddVBO("default", len(vertexBuffer)/5, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
		{Name: "center", Type: attribute.Vec2},
	})

	body.fillVAO.SetData("default", 0, vertexBuffer)

	ibo := buffer.NewIndexBufferObjectInt(len(indexBuffer))

	ibo.SetDataI(0, indexBuffer)

	body.fillVAO.AttachIBO(ibo)

	body.fillVAO.Attach(capShader)
}

func (body *Body) setupCapVAO() {
	body.capVAO = buffer.NewVertexArrayObject()

	unitCircle, indexes := createUnitCircle(LevelOfDetail)

	body.capVAO.AddVBO("default", len(unitCircle)/3, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
	})

	body.capVAO.SetData("default", 0, unitCircle)

	ibo := buffer.NewIndexBufferObject(len(indexes))

	ibo.SetData(0, indexes)

	body.capVAO.AttachIBO(ibo)

	body.capVAO.AddVBO("points", 2, 1, attribute.Format{
		{Name: "center", Type: attribute.Vec2},
	})

	body.capVAO.Attach(capShader)
}

func (body *Body) DrawBase(headProgress, tailProgress float64, baseProjView mgl32.Mat4) {
	if body.lineVAO == nil || len(body.sections) == 0 {
		return
	}

	body.ensureFBO(baseProjView)

	// Don't render to nonexistent framebuffer
	if body.framebuffer == nil || body.disposed {
		return
	}

	headLength := mutils.Clamp(float32(headProgress), 0.0, 1.0) * body.totalLength
	tailLength := mutils.Clamp(float32(tailProgress), 0.0, 1.0) * body.totalLength

	if !settings.RECORD {
		// In watch mode re-render on changes bigger than 1 o!px in any direction to conserve FPS a bit
		// As a result very slow sliders may look choppy but whatever

		headLength = min(math32.Ceil(headLength), body.totalLength) // Make sure snake-out doesn't go outside the path

		if tailLength != body.totalLength { // Don't clamp at snake-in end to have the spot-on position
			tailLength = math32.Floor(tailLength)
		}
	}

	if headLength > tailLength {
		headLength, tailLength = tailLength, headLength
	}

	// State is the same so no need to render
	// TODO: make it more smooth
	if body.headLength == headLength && body.tailLength == tailLength {
		return
	}

	headIndex := sort.Search(len(body.sections), func(i int) bool {
		return headLength < body.sections[i].prevLength
	}) - 1

	tailIndex := sort.Search(len(body.sections), func(i int) bool {
		return tailLength < body.sections[i].prevLength
	}) - 1

	head := body.sections[headIndex]
	tail := body.sections[tailIndex]

	if tailIndex == headIndex { // Special case if snake ends on the same section
		body.modifyBuffer(headIndex, head.pointAtLen(headLength), tailLength-headLength)
	} else {
		forceUpdateHead := false

		if body.tailLength != tailLength {
			if body.tailIndex > -1 && body.tailIndex != tailIndex { // Using not equals because less caused missing sections on time jumps
				if body.tailIndex == headIndex {
					forceUpdateHead = true // Update the head if previous tail collides with the current head
				} else {
					prevTail := body.sections[body.tailIndex]

					body.modifyBuffer(body.tailIndex, prevTail.Point1, prevTail.length)
				}
			}

			body.modifyBuffer(tailIndex, tail.Point1, tail.capLen(tailLength-tail.prevLength))
		}

		if body.headLength != headLength || forceUpdateHead {
			body.modifyBuffer(headIndex, head.pointAtLen(headLength), head.capLen(head.length+head.prevLength-headLength))
		}
	}

	hPoint := head.pointAtLen(headLength)
	tPoint := tail.pointAtLen(tailLength)

	body.capBuffer[0], body.capBuffer[1] = hPoint.X, hPoint.Y
	body.capBuffer[2], body.capBuffer[3] = tPoint.X, tPoint.Y

	body.capVAO.SetData("points", 0, body.capBuffer)

	prevIndices := body.sections[headIndex].prevIndices
	currIndices := body.sections[tailIndex].prevIndices

	body.framebuffer.Bind()
	viewport.Push(int(body.framebuffer.Texture().GetWidth()), int(body.framebuffer.Texture().GetHeight()))

	//TODO: Make depth manager
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	body.framebuffer.ClearDepth()

	lineShader.Bind()

	lineShader.SetUniform("projection", body.baseProjection)
	lineShader.SetUniform("scale", body.radius*float32(settings.Audio.BeatScale))
	lineShader.SetUniform("distort", body.distortionMatrix)

	body.lineVAO.Bind()
	body.lineVAO.DrawInstanced(headIndex, tailIndex-headIndex+1)
	body.lineVAO.Unbind()

	lineShader.Unbind()

	capShader.Bind()

	capShader.SetUniform("projection", body.baseProjection)
	capShader.SetUniform("scale", body.radius*float32(settings.Audio.BeatScale))
	capShader.SetUniform("distort", body.distortionMatrix)

	body.fillVAO.Bind()
	body.fillVAO.DrawPart(prevIndices, currIndices-prevIndices)
	body.fillVAO.Unbind()

	body.capVAO.Bind()
	body.capVAO.DrawInstanced(0, 2)
	body.capVAO.Unbind()

	capShader.Unbind()

	body.framebuffer.Unbind()

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)

	viewport.Pop()

	body.headLength = headLength
	body.tailLength = tailLength

	body.headIndex = headIndex
	body.tailIndex = tailIndex
}

func (body *Body) modifyBuffer(index int, point vector.Vector2f, length float32) {
	body.lineBuffer[0], body.lineBuffer[1] = point.X, point.Y
	body.lineBuffer[2] = length

	body.lineVAO.SetData("lines", index*5, body.lineBuffer)
}

func (body *Body) DrawNormal(projection mgl32.Mat4, stackOffset vector.Vector2f, scale float32, bodyInner, bodyOuter, borderInner, borderOuter color2.Color) {
	if body.framebuffer == nil || body.disposed || len(body.sections) == 0 {
		return
	}

	drawSlider(body.bodySprite, stackOffset, scale, body.framebuffer.Texture(), bodyInner, bodyOuter, borderInner, borderOuter, projection)
}

func (body *Body) ensureFBO(baseProjView mgl32.Mat4) {
	if body.framebuffer != nil || body.disposed {
		return
	}

	invProjView := baseProjView.Inv()

	body.calculateDistortionMatrix(baseProjView, invProjView)

	multiplierX := float32(1.0)
	multiplierY := float32(1.0)

	aspect := float32(settings.Graphics.GetAspectRatio())

	if aspect > 1 {
		multiplierY = 1 / aspect
	} else {
		multiplierX = aspect
	}

	// Overdraw is used for mandalas when base sliders are out of bounds
	overdraw := math32.Sqrt(multiplierX*multiplierX + multiplierY*multiplierY)

	multiplierX = overdraw / multiplierX
	multiplierY = overdraw / multiplierY

	topLeftWorldRelative := invProjView.Mul4x1(mgl32.Vec4{-multiplierX, multiplierY, 0.0, 1.0})
	bottomRightWorldRelative := invProjView.Mul4x1(mgl32.Vec4{multiplierX, -multiplierY, 0.0, 1.0})

	var topLeftWorld, bottomRightWorld vector.Vector2f

	topLeftWorld.X = max(topLeftWorldRelative.X(), body.topLeft.X-body.radius*float32(settings.Audio.BeatScale))
	topLeftWorld.Y = max(topLeftWorldRelative.Y(), body.topLeft.Y-body.radius*float32(settings.Audio.BeatScale))

	bottomRightWorld.X = min(bottomRightWorldRelative.X(), body.bottomRight.X+body.radius*float32(settings.Audio.BeatScale))
	bottomRightWorld.Y = min(bottomRightWorldRelative.Y(), body.bottomRight.Y+body.radius*float32(settings.Audio.BeatScale))

	tLS := baseProjView.Mul4x1(mgl32.Vec4{topLeftWorld.X, topLeftWorld.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)
	bRS := baseProjView.Mul4x1(mgl32.Vec4{bottomRightWorld.X, bottomRightWorld.Y, 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)

	topLeftScreen := vector.NewVec2f(tLS.X(), tLS.Y()).Mult(vector.NewVec2f(float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())))
	bottomRightScreen := vector.NewVec2f(bRS.X(), bRS.Y()).Mult(vector.NewVec2f(float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())))

	dimensions := bottomRightScreen.Sub(topLeftScreen).Abs()

	body.framebuffer = buffer.NewFrameDepth(int(dimensions.X), int(dimensions.Y), false)

	tex := body.framebuffer.Texture().GetRegion()

	body.bodySprite = sprite.NewSpriteSingle(&tex, 0, bottomRightWorld.Sub(topLeftWorld).Scl(0.5).Add(topLeftWorld).Copy64(), vector.Centre)
	body.bodySprite.SetScale(float64((bottomRightWorld.X - topLeftWorld.X) / dimensions.X))
	body.bodySprite.SetVFlip(true)

	body.baseProjection = mgl32.Ortho(topLeftWorld.X, bottomRightWorld.X, bottomRightWorld.Y, topLeftWorld.Y, 1, -1)
}

func (body *Body) calculateDistortionMatrix(baseProjView, invProjView mgl32.Mat4) {
	distortions := settings.Objects.Sliders.Distortions

	if distortions.Enabled {
		distortionBase := [2]int32{int32(distortions.ViewportSize), int32(distortions.ViewportSize)}
		if distortionBase[0] <= 0 {
			gl.GetIntegerv(gl.MAX_VIEWPORT_DIMS, &distortionBase[0])
		}

		screenSizeX, screenSizeY := settings.Graphics.GetWidthF(), settings.Graphics.GetHeightF()
		if distortions.UseCustomResolution {
			screenSizeX = float64(distortions.CustomResolutionX)
			screenSizeY = float64(distortions.CustomResolutionY)
		}

		relativeViewportWidth := float32(float64(distortionBase[0]) / screenSizeX)
		relativeViewportHeight := float32(float64(distortionBase[1]) / screenSizeY)

		r2 := body.radius * 1.15 // for some reason stable uses bigger slider size for calculations

		relativeTopLeft := baseProjView.Mul4x1(mgl32.Vec4{math32.Floor(body.topLeft.X - r2), math32.Floor(body.topLeft.Y - r2), 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)
		relativeBottomRight := baseProjView.Mul4x1(mgl32.Vec4{math32.Floor(body.bottomRight.X + r2), math32.Floor(body.bottomRight.Y + r2), 0, 1}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)

		leftCapped := max(relativeTopLeft.X(), 0)
		topCapped := min(relativeTopLeft.Y(), 1.0)

		scale := vector.NewVec2f(1, 1)

		if -relativeTopLeft.X()+relativeBottomRight.X() > relativeViewportWidth {
			scale.X = relativeViewportWidth / (-relativeTopLeft.X() + relativeBottomRight.X())
		}

		if relativeBottomRight.Y()-topCapped < -relativeViewportHeight {
			scale.Y = -relativeViewportHeight / (relativeBottomRight.Y() - topCapped)
		}

		offset := invProjView.Mul4x1(mgl32.Vec4{leftCapped*2 - 1, topCapped*2 - 1, 0.0, 1.0}) // relative to osu! coordinates conversion
		body.distortionMatrix = mgl32.Translate3D(offset.X(), offset.Y(), 0).Mul4(mgl32.Scale3D(scale.X, scale.Y, 1)).Mul4(mgl32.Translate3D(-offset.X(), -offset.Y(), 0))
	}
}

func (body *Body) Dispose() {
	if body.disposed || body.framebuffer == nil {
		return
	}

	body.disposed = true
	body.framebuffer.Dispose()
}
