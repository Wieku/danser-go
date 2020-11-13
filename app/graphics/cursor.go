package graphics

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"sync"
)

var cursorShader *shader.RShader = nil

var cursorFbo *buffer.Framebuffer = nil
var cursorFBOSprite *sprite.Sprite

var cursorSpaceFbo *buffer.Framebuffer = nil
var cursorSpaceFBOSprite *sprite.Sprite

var fboBatch *batch.QuadBatch

var Camera *camera.Camera
var osuRect camera.Rectangle

var useAdditive = false

func initCursor() {
	if settings.Cursor.TrailStyle < 1 || settings.Cursor.TrailStyle > 4 {
		panic("Wrong cursor trail type")
	}

	vert, err := assets.GetString("assets/shaders/cursortrail.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/cursortrail.fsh")
	if err != nil {
		panic(err)
	}

	cursorShader = shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	cursorFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	region := cursorFbo.Texture().GetRegion()
	cursorFBOSprite = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)

	cursorSpaceFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	regionSpace := cursorSpaceFbo.Texture().GetRegion()
	cursorSpaceFBOSprite = sprite.NewSpriteSingle(&regionSpace, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)

	fboBatch = batch.NewQuadBatchSize(1)
	fboBatch.SetCamera(mgl32.Ortho(0, float32(settings.Graphics.GetWidth()), 0, float32(settings.Graphics.GetHeight()), -1, 1))

	osuRect = Camera.GetWorldRect()
}

type Cursor struct {
	Points        []vector.Vector2f
	PointsC       []float64
	removeCounter float64

	scale *animation.Glider

	lastLeftState, lastRightState bool

	LeftButton, RightButton bool
	LeftKey, RightKey       bool
	LeftMouse, RightMouse   bool

	IsReplayFrame    bool // TODO: temporary hacky solution for spinners
	IsPlayer         bool
	LastFrameTime    int64 //
	CurrentFrameTime int64 //
	Position         vector.Vector2f
	LastPos          vector.Vector2f
	VaoPos           vector.Vector2f
	RendPos          vector.Vector2f

	Name string

	vertices  []float32
	vaoSize   int
	maxCap    int
	vaoDirty  bool
	vao       *buffer.VertexArrayObject
	mutex     *sync.Mutex
	hueBase   float64
	vecSize   int
	instances int
}

func NewCursor() *Cursor {
	if cursorShader == nil {
		initCursor()
	}

	points := int(math.Ceil(float64(settings.Cursor.TrailMaxLength) * settings.Cursor.TrailDensity))

	vao := buffer.NewVertexArrayObject()

	vao.AddVBO(
		"default",
		6,
		0,
		attribute.Format{
			{Name: "in_position", Type: attribute.Vec2},
			{Name: "in_tex_coord", Type: attribute.Vec2},
		},
	)

	vao.SetData("default", 0, []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		1, 1, 1, 1,
		-1, 1, 0, 1,
		-1, -1, 0, 0,
	})

	vao.AddVBO(
		"points",
		points,
		1,
		attribute.Format{
			{Name: "in_mid", Type: attribute.Vec2},
			{Name: "hue", Type: attribute.Float},
		},
	)

	vao.Bind()
	vao.Attach(cursorShader)
	vao.Unbind()

	cursor := &Cursor{LastPos: vector.NewVec2f(100, 100), Position: vector.NewVec2f(100, 100), vao: vao, mutex: &sync.Mutex{}, RendPos: vector.NewVec2f(100, 100), vertices: make([]float32, points*3)}
	cursor.scale = animation.NewGlider(1.0)
	cursor.vecSize = 3

	return cursor
}

func (cursor *Cursor) SetPos(pt vector.Vector2f) {
	tmp := pt

	if settings.Cursor.BounceOnEdges && settings.DIVIDES <= 2 {
		for {
			ok1, ok2 := false, false

			if tmp.X < osuRect.MinX {
				tmp.X = 2*osuRect.MinX - tmp.X
			} else if tmp.X > osuRect.MaxX {
				tmp.X = 2*osuRect.MaxX - tmp.X
			} else {
				ok1 = true
			}

			if tmp.Y < osuRect.MinY {
				tmp.Y = 2*osuRect.MinY - tmp.Y
			} else if tmp.Y > osuRect.MaxY {
				tmp.Y = 2*osuRect.MaxY - tmp.Y
			} else {
				ok2 = true
			}

			if ok1 && ok2 {
				break
			}
		}
	}

	cursor.Position = tmp
}

func (cursor *Cursor) SetScreenPos(pt vector.Vector2f) {
	cursor.SetPos(Camera.Unproject(pt.Copy64()).Copy32())
}

func (cursor *Cursor) Update(delta float64) {
	delta = math.Abs(delta)

	leftState := cursor.LeftKey || cursor.LeftMouse
	rightState := cursor.RightKey || cursor.RightMouse
	if cursor.lastLeftState != leftState || cursor.lastRightState != rightState {
		if (leftState || rightState) && settings.Cursor.CursorExpand {
			cursor.scale.AddEventS(cursor.scale.GetTime(), cursor.scale.GetTime()+100, 1.0, 1.3)
		} else {
			cursor.scale.AddEventS(cursor.scale.GetTime(), cursor.scale.GetTime()+100, cursor.scale.GetValue(), 1.0)
		}

		cursor.lastLeftState = leftState
		cursor.lastRightState = rightState
	}

	cursor.scale.UpdateD(delta)

	if settings.Cursor.TrailStyle == 3 {
		cursor.hueBase += settings.Cursor.Style23Speed / 360.0 * delta
		if cursor.hueBase > 1.0 {
			cursor.hueBase -= 1.0
		} else if cursor.hueBase < 0 {
			cursor.hueBase += 1.0
		}
	}

	lengthAdjusted := int(float64(settings.Cursor.TrailMaxLength) * settings.Cursor.TrailDensity)

	points := cursor.Position.Dst(cursor.LastPos)
	distance := float32(1.0 / settings.Cursor.TrailDensity)

	dirtyLocal := false

	if int(points/distance) > 0 {
		temp := cursor.LastPos
		for i := distance; i < points; i += distance {
			temp = cursor.Position.Sub(cursor.LastPos).Scl(i / points).Add(cursor.LastPos)
			cursor.Points = append(cursor.Points, temp)
			cursor.PointsC = append(cursor.PointsC, cursor.hueBase)

			if settings.Cursor.TrailStyle == 2 {
				cursor.hueBase += settings.Cursor.Style23Speed / 360.0 * float64(distance)
				if cursor.hueBase > 1.0 {
					cursor.hueBase -= 1.0
				} else if cursor.hueBase < 0 {
					cursor.hueBase += 1.0
				}
			}
		}
		dirtyLocal = true
		cursor.LastPos = temp
	}

	if len(cursor.Points) > 0 {
		cursor.removeCounter += float64(len(cursor.Points)+3) / (360.0 / delta) * settings.Cursor.TrailRemoveSpeed
		times := int(math.Floor(cursor.removeCounter))

		if len(cursor.Points) > lengthAdjusted {
			cursor.Points = cursor.Points[len(cursor.Points)-lengthAdjusted:]
			cursor.PointsC = cursor.PointsC[len(cursor.PointsC)-lengthAdjusted:]
			cursor.removeCounter = 0
			dirtyLocal = true
		} else if times > 0 {
			times = bmath.MinI(times, len(cursor.Points))

			cursor.Points = cursor.Points[times:]
			cursor.PointsC = cursor.PointsC[times:]
			cursor.removeCounter -= float64(times)

			dirtyLocal = true
		}
	}

	cursor.mutex.Lock()
	if dirtyLocal {
		if len(cursor.vertices) != lengthAdjusted*3 {
			cursor.vertices = make([]float32, lengthAdjusted*3)
		}

		for i, o := range cursor.Points {
			inv := float32(len(cursor.Points) - i - 1)

			hue := float32(cursor.PointsC[i])
			if settings.Cursor.TrailStyle == 4 {
				hue = float32(settings.Cursor.Style4Shift) * inv / float32(len(cursor.Points))
			}

			index := i * 3
			cursor.vertices[index] = o.X
			cursor.vertices[index+1] = o.Y
			cursor.vertices[index+2] = hue
		}

		cursor.maxCap = lengthAdjusted
		cursor.vaoSize = len(cursor.Points)

		cursor.vaoDirty = true
	}
	cursor.VaoPos = cursor.Position
	cursor.mutex.Unlock()
}

func (cursor *Cursor) UpdateRenderer() {
	cursor.mutex.Lock()
	if cursor.vaoDirty {
		cursor.vao.Resize("points", cursor.maxCap)
		cursor.vao.SetData("points", 0, cursor.vertices[0:cursor.vaoSize*3])
		cursor.instances = cursor.vaoSize
		cursor.vaoDirty = false
	}
	cursor.RendPos = cursor.VaoPos
	cursor.mutex.Unlock()
}

func BeginCursorRender() {
	CursorTrail.Bind(1)

	useAdditive = settings.Cursor.AdditiveBlending && (settings.PLAYERS > 1 || settings.DIVIDES > 1 || settings.TAG > 1)

	if useAdditive {
		cursorSpaceFbo.Bind()
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}

	blend.Push()
	blend.Enable()
	blend.SetFunctionSeparate(blend.SrcAlpha, blend.OneMinusSrcAlpha, blend.One, blend.OneMinusSrcAlpha)
}

func EndCursorRender() {
	if useAdditive {
		cursorSpaceFbo.Unbind()

		fboBatch.Begin()
		cursorSpaceFBOSprite.Draw(0, fboBatch)
		fboBatch.End()
	}

	blend.Pop()
}

func (cursor *Cursor) Draw(scale float64, batch *batch.QuadBatch, color color2.Color) {
	cursor.DrawM(scale, batch, color, color)
}

func (cursor *Cursor) DrawM(scale float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color) {
	hueShift := color.GetHue()

	if useAdditive {
		cursorFbo.Bind()
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}

	siz := settings.Cursor.CursorSize * cursor.scale.GetValue()

	if settings.Cursor.EnableCustomTrailGlowOffset {
		colorGlow = color.Shift(float32(settings.Cursor.TrailGlowOffset), 0, 0)
	}

	glowShift := colorGlow.GetHue()

	if settings.Cursor.TrailStyle > 1 {
		color = color.Shift(float32(cursor.hueBase*360), 0, 0)
		colorGlow = colorGlow.Shift(float32(cursor.hueBase*360), 0, 0)
	}

	colorD := color
	colorD2 := colorGlow

	if settings.Cursor.TrailStyle > 1 {
		colorD = color2.NewLA(1.0, color.A)
		colorD2 = color2.NewLA(1.0, colorGlow.A)
	}

	cursorShader.Bind()
	cursorShader.SetUniform("tex", int32(1))
	cursorShader.SetUniform("proj", batch.Projection)
	cursorShader.SetUniform("points", float32(cursor.instances))
	cursorShader.SetUniform("instances", float32(cursor.instances))

	if settings.Cursor.TrailStyle == 1 {
		cursorShader.SetUniform("saturation", float32(0.0))
	} else {
		cursorShader.SetUniform("saturation", float32(1.0))
	}

	cursor.vao.Bind()

	innerLengthMult := float32(1.0)

	cursorScl := float32(siz * (16.0 / 18) * scale)

	if settings.Cursor.EnableTrailGlow {
		cursorScl = float32(siz * (12.0 / 18) * scale)
		innerLengthMult = float32(settings.Cursor.InnerLengthMult)

		cursorShader.SetUniform("col_tint", colorD2)
		cursorShader.SetUniform("scale", float32(siz*(16.0/18)*scale*settings.Cursor.TrailScale))
		cursorShader.SetUniform("endScale", float32(settings.Cursor.GlowEndScale))
		if settings.Cursor.TrailStyle > 1 {
			cursorShader.SetUniform("hueshift", glowShift/360)
		}

		cursor.vao.DrawInstanced(0, cursor.instances)
	}

	cursorShader.SetUniform("col_tint", colorD)
	cursorShader.SetUniform("scale", cursorScl*float32(settings.Cursor.TrailScale))
	cursorShader.SetUniform("points", float32(len(cursor.Points))*innerLengthMult)
	cursorShader.SetUniform("endScale", float32(settings.Cursor.TrailEndScale))
	if settings.Cursor.TrailStyle > 1 {
		cursorShader.SetUniform("hueshift", hueShift/360)
	}

	cursor.vao.DrawInstanced(0, cursor.instances)

	cursor.vao.Unbind()

	cursorShader.Unbind()

	batch.Begin()

	position := cursor.RendPos
	if settings.PLAY {
		position = cursor.Position
	}

	batch.ResetTransform()

	batch.SetTranslation(position.Copy64())
	batch.SetScale(siz*scale, siz*scale)
	batch.SetSubScale(1, 1)

	batch.SetColorM(color)
	batch.DrawUnit(*CursorTex)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color.A)))
	batch.DrawUnit(*CursorTop)

	batch.End()

	if useAdditive {
		cursorFbo.Unbind()

		fboBatch.Begin()

		blend.Push()
		blend.SetFunction(blend.SrcAlpha, blend.One)

		cursorFBOSprite.Draw(0, fboBatch)
		fboBatch.Flush()

		blend.Pop()

		fboBatch.End()
	}
}
