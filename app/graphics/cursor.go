package graphics

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/vector"
	"io/ioutil"
	"math"
	"sync"
)

var cursorShader *shader.RShader = nil

var cursorFbo *buffer.Framebuffer = nil
var cursorFBOSprite *sprite.Sprite

var cursorSpaceFbo *buffer.Framebuffer = nil
var cursorSpaceFBOSprite *sprite.Sprite

var fboBatch *sprite.SpriteBatch

var Camera *bmath.Camera
var osuRect bmath.Rectangle

var useAdditive = false

func initCursor() {
	if settings.Cursor.TrailStyle < 1 || settings.Cursor.TrailStyle > 4 {
		panic("Wrong cursor trail type")
	}

	vert, _ := ioutil.ReadFile("assets/shaders/cursortrail.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/cursortrail.fsh")
	cursorShader = shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

	cursorFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	region := cursorFbo.Texture().GetRegion()
	cursorFBOSprite = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)

	cursorSpaceFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	regionSpace := cursorSpaceFbo.Texture().GetRegion()
	cursorSpaceFBOSprite = sprite.NewSpriteSingle(&regionSpace, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), bmath.Origin.Centre)

	fboBatch = sprite.NewSpriteBatchSize(1)
	fboBatch.SetCamera(mgl32.Ortho(0, float32(settings.Graphics.GetWidth()), 0, float32(settings.Graphics.GetHeight()), -1, 1))

	osuRect = Camera.GetWorldRect()
}

type Cursor struct {
	Points        []vector.Vector2f
	PointsC       []float64
	removeCounter float64

	LeftButton, RightButton bool
	IsReplayFrame           bool // TODO: temporary hacky solution for spinners
	IsPlayer                bool
	LastFrameTime           int64 //
	CurrentFrameTime        int64 //
	Position                vector.Vector2f
	LastPos                 vector.Vector2f
	VaoPos                  vector.Vector2f
	RendPos                 vector.Vector2f

	Name string

	vertices  []float32
	vaoSize   int
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
	cursor.vecSize = 3

	return cursor
}

func (cr *Cursor) SetPos(pt vector.Vector2f) {
	tmp := pt

	if settings.Cursor.BounceOnEdges {
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

	cr.Position = tmp
}

func (cr *Cursor) SetScreenPos(pt vector.Vector2f) {
	cr.SetPos(Camera.Unproject(pt.Copy64()).Copy32())
}

func (cr *Cursor) Update(delta float64) {
	delta = math.Abs(delta)

	if settings.Cursor.TrailStyle == 3 {
		cr.hueBase += settings.Cursor.Style23Speed / 360.0 * delta
		if cr.hueBase > 1.0 {
			cr.hueBase -= 1.0
		} else if cr.hueBase < 0 {
			cr.hueBase += 1.0
		}
	}

	points := cr.Position.Dst(cr.LastPos)
	density := float32(1.0 / settings.Cursor.TrailDensity)

	dirtyLocal := false

	if int(points/density) > 0 {
		temp := cr.LastPos
		for i := density; i < points; i += density {
			temp = cr.Position.Sub(cr.LastPos).Scl(i / points).Add(cr.LastPos)
			cr.Points = append(cr.Points, temp)
			cr.PointsC = append(cr.PointsC, cr.hueBase)

			if settings.Cursor.TrailStyle == 2 {
				cr.hueBase += settings.Cursor.Style23Speed / 360.0 * float64(density)
				if cr.hueBase > 1.0 {
					cr.hueBase -= 1.0
				} else if cr.hueBase < 0 {
					cr.hueBase += 1.0
				}
			}
		}
		dirtyLocal = true
		cr.LastPos = temp
	}

	if len(cr.Points) > 0 {
		cr.removeCounter += float64(len(cr.Points)+3) / (360.0 / delta) * settings.Cursor.TrailRemoveSpeed
		times := int(math.Floor(cr.removeCounter))
		lengthAdjusted := int(float64(settings.Cursor.TrailMaxLength) * float64(density))

		if len(cr.Points) > lengthAdjusted {
			cr.Points = cr.Points[len(cr.Points)-lengthAdjusted:]
			cr.PointsC = cr.PointsC[len(cr.PointsC)-lengthAdjusted:]
			cr.removeCounter = 0
			dirtyLocal = true
		} else if times > 0 {
			times = bmath.MinI(times, len(cr.Points))

			cr.Points = cr.Points[times:]
			cr.PointsC = cr.PointsC[times:]
			cr.removeCounter -= float64(times)

			dirtyLocal = true
		}
	}

	if dirtyLocal {
		cr.mutex.Lock()

		for i, o := range cr.Points {
			inv := float32(len(cr.Points) - i - 1)

			hue := float32(cr.PointsC[i])
			if settings.Cursor.TrailStyle == 4 {
				hue = float32(settings.Cursor.Style4Shift) * inv / float32(len(cr.Points))
			}

			index := i * 3
			cr.vertices[index] = o.X
			cr.vertices[index+1] = o.Y
			cr.vertices[index+2] = hue
		}

		cr.vaoSize = len(cr.Points)
		cr.VaoPos = cr.Position
		cr.vaoDirty = true
		cr.mutex.Unlock()
	}
}

func (cursor *Cursor) UpdateRenderer() {
	cursor.mutex.Lock()
	if cursor.vaoDirty {
		cursor.vao.SetData("points", 0, cursor.vertices[0:cursor.vaoSize*3])
		cursor.instances = cursor.vaoSize
		cursor.RendPos = cursor.VaoPos
		cursor.vaoDirty = false
	}
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

func (cursor *Cursor) Draw(scale float64, batch *sprite.SpriteBatch, color mgl32.Vec4, hueshift float64) {
	cursor.DrawM(scale, batch, color, color, hueshift)
}

func (cursor *Cursor) DrawM(scale float64, batch *sprite.SpriteBatch, color mgl32.Vec4, color2 mgl32.Vec4, hueshift float64) {
	colorD := color
	colorD2 := color2

	if settings.Cursor.TrailStyle > 1 {
		colorD = mgl32.Vec4{1.0, 1.0, 1.0, color.W()}
		colorD2 = mgl32.Vec4{1.0, 1.0, 1.0, color2.W()}
	}

	if useAdditive {
		cursorFbo.Bind()
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}

	siz := settings.Cursor.CursorSize

	if settings.Cursor.EnableCustomTrailGlowOffset {
		color2 = utils.GetColorShifted(color, settings.Cursor.TrailGlowOffset)
		colorD2 = color2
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
			cursorShader.SetUniform("hueshift", float32((hueshift-36)/360))
		}
		cursor.vao.DrawInstanced(0, cursor.instances)
	}

	if settings.Cursor.TrailStyle > 1 {
		cursorShader.SetUniform("hueshift", float32(hueshift/360))
	}
	cursorShader.SetUniform("col_tint", colorD)
	cursorShader.SetUniform("scale", cursorScl*float32(settings.Cursor.TrailScale))
	cursorShader.SetUniform("points", float32(len(cursor.Points))*innerLengthMult)
	cursorShader.SetUniform("endScale", float32(settings.Cursor.TrailEndScale))

	cursor.vao.DrawInstanced(0, cursor.instances)

	cursor.vao.Unbind()

	cursorShader.Unbind()

	batch.Begin()

	batch.SetTranslation(cursor.RendPos.Copy64())
	batch.SetScale(siz*scale, siz*scale)
	batch.SetSubScale(1, 1)

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), float64(color[3]))
	batch.DrawUnit(*CursorTex)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color[3])))
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
