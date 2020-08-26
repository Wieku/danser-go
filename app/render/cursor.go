package render

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/render/framebuffer"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/glhf"
	"io/ioutil"
	"math"
	"sync"
)

var cursorShader *glhf.Shader = nil
var cursorFbo *framebuffer.Framebuffer = nil
var cursorSpaceFbo *framebuffer.Framebuffer = nil
var Camera *bmath.Camera
var osuRect bmath.Rectangle

var useAdditive = false

func initCursor() {

	if settings.Cursor.TrailStyle < 1 || settings.Cursor.TrailStyle > 4 {
		panic("Wrong cursor trail type")
	}

	vertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_mid", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
		{Name: "in_index", Type: glhf.Float},
		{Name: "hue", Type: glhf.Float},
	}

	uniformFormat := glhf.AttrFormat{
		{Name: "col_tint", Type: glhf.Vec4},
		{Name: "tex", Type: glhf.Int},
		{Name: "proj", Type: glhf.Mat4},
		{Name: "points", Type: glhf.Float},
		{Name: "scale", Type: glhf.Float},
		{Name: "endScale", Type: glhf.Float},
		{Name: "hueshift", Type: glhf.Float},
		{Name: "saturation", Type: glhf.Float},
	}

	var err error

	vert, _ := ioutil.ReadFile("assets/shaders/cursortrail.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/cursortrail.fsh")
	cursorShader, err = glhf.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic("Cursor: " + err.Error())
	}

	cursorFbo = framebuffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	cursorFbo.Texture().Bind(30)
	cursorSpaceFbo = framebuffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	cursorSpaceFbo.Texture().Bind(18)
	osuRect = Camera.GetWorldRect()
}

type Cursor struct {
	Points        []bmath.Vector2f
	PointsC       []float64
	removeCounter float64

	LeftButton, RightButton bool
	IsReplayFrame           bool // TODO: temporary hacky solution for spinners
	IsPlayer                bool
	LastFrameTime           int64 //
	CurrentFrameTime        int64 //
	Position                bmath.Vector2f
	LastPos                 bmath.Vector2f
	VaoPos                  bmath.Vector2f
	RendPos                 bmath.Vector2f

	Name string

	vertices []float32
	vaoSize  int
	vaoDirty bool
	vao      *glhf.VertexSlice
	subVao   *glhf.VertexSlice
	mutex    *sync.Mutex
	hueBase  float64
	vecSize  int
}

func NewCursor() *Cursor {
	if cursorShader == nil {
		initCursor()
	}

	length := int(math.Ceil(float64(settings.Cursor.TrailMaxLength)*settings.Cursor.TrailDensity) * 6)
	vao := glhf.MakeVertexSlice(cursorShader, length, length)
	cursor := &Cursor{LastPos: bmath.NewVec2f(100, 100), Position: bmath.NewVec2f(100, 100), vao: vao, subVao: vao.Slice(0, 0), mutex: &sync.Mutex{}, RendPos: bmath.NewVec2f(100, 100)}
	cursor.vecSize = 10
	return cursor
}

func (cr *Cursor) SetPos(pt bmath.Vector2f) {
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

func (cr *Cursor) SetScreenPos(pt bmath.Vector2f) {
	cr.SetPos(Camera.Unproject(pt.Copy64()).Copy32())
}

func (cr *Cursor) Update(tim float64) {
	tim = math.Abs(tim)

	if settings.Cursor.TrailStyle == 3 {
		cr.hueBase += settings.Cursor.Style23Speed / 360.0 * tim
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
		cr.removeCounter += float64(len(cr.Points)+3) / (360.0 / tim) * settings.Cursor.TrailRemoveSpeed
		times := int(math.Floor(cr.removeCounter))
		lengthAdjusted := int(float64(settings.Cursor.TrailMaxLength) / float64(density))

		if len(cr.Points) > lengthAdjusted {
			cr.Points = cr.Points[len(cr.Points)-lengthAdjusted:]
			cr.PointsC = cr.PointsC[len(cr.PointsC)-lengthAdjusted:]
			cr.removeCounter = 0
			dirtyLocal = true
		} else if times > 0 {
			if times < len(cr.Points) {
				cr.Points = cr.Points[times:]
				cr.PointsC = cr.PointsC[times:]
				cr.removeCounter -= float64(times)
			} else {
				cr.Points = cr.Points[len(cr.Points):]
				cr.PointsC = cr.PointsC[len(cr.PointsC):]
				cr.removeCounter = 0
			}

			dirtyLocal = true
		}
	}

	if dirtyLocal {
		cr.mutex.Lock()

		if len(cr.vertices) < len(cr.Points)*6*cr.vecSize {
			cr.vertices = make([]float32, len(cr.Points)*6*cr.vecSize)
		}

		for i, o := range cr.Points {
			offset := i * 6 * cr.vecSize
			inv := float32(len(cr.Points) - i - 1)

			hue := float32(cr.PointsC[i])
			if settings.Cursor.TrailStyle == 4 {
				hue = float32(settings.Cursor.Style4Shift) * inv / float32(len(cr.Points))
			}

			fillArray(cr.vertices, offset, -1+o.X, -1+o.Y, 0, o.X, o.Y, 0, 0, 0, inv, hue)
			fillArray(cr.vertices, offset+cr.vecSize, 1+o.X, -1+o.Y, 0, o.X, o.Y, 0, 1, 0, inv, hue)
			fillArray(cr.vertices, offset+cr.vecSize*2, -1+o.X, 1+o.Y, 0, o.X, o.Y, 0, 0, 1, inv, hue)
			fillArray(cr.vertices, offset+cr.vecSize*3, 1+o.X, -1+o.Y, 0, o.X, o.Y, 0, 1, 0, inv, hue)
			fillArray(cr.vertices, offset+cr.vecSize*4, 1+o.X, 1+o.Y, 0, o.X, o.Y, 0, 1, 1, inv, hue)
			fillArray(cr.vertices, offset+cr.vecSize*5, -1+o.X, 1+o.Y, 0, o.X, o.Y, 0, 0, 1, inv, hue)
		}

		cr.vaoSize = len(cr.Points) * 6 * cr.vecSize
		cr.VaoPos = cr.Position
		cr.vaoDirty = true
		cr.mutex.Unlock()
	}
}

func fillArray(dst []float32, index int, values ...float32) {
	for i, j := range values {
		dst[index+i] = j
	}
}

func (cursor *Cursor) UpdateRenderer() {
	cursor.mutex.Lock()
	if cursor.vaoDirty {
		cursor.subVao = cursor.vao.Slice(0, cursor.vaoSize/cursor.vecSize)
		cursor.subVao.Begin()
		cursor.subVao.SetVertexData(cursor.vertices[0:cursor.vaoSize])
		cursor.subVao.End()
		cursor.RendPos = cursor.VaoPos
		cursor.vaoDirty = false
	}
	cursor.mutex.Unlock()
}

func BeginCursorRender() {
	CursorTrail.Bind(1)

	useAdditive = settings.Cursor.AdditiveBlending && (settings.PLAYERS > 1 || settings.DIVIDES > 1 || settings.TAG > 1)

	if useAdditive {
		cursorSpaceFbo.Begin()
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}
}

func EndCursorRender() {
	if useAdditive {
		cursorSpaceFbo.End()
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		fboShader.Begin()
		fboShader.SetUniformAttr(0, int32(18))
		fboSlice.BeginDraw()
		fboSlice.Draw()
		fboSlice.EndDraw()
		fboShader.End()
	}
}

func (cursor *Cursor) Draw(scale float64, batch *batches.SpriteBatch, color mgl32.Vec4, hueshift float64) {
	cursor.DrawM(scale, batch, color, color, hueshift)
}

func (cursor *Cursor) DrawM(scale float64, batch *batches.SpriteBatch, color mgl32.Vec4, color2 mgl32.Vec4, hueshift float64) {
	gl.Disable(gl.DEPTH_TEST)

	if settings.Cursor.TrailStyle > 1 {
		color = mgl32.Vec4{1.0, 1.0, 1.0, color.W()}
		color2 = mgl32.Vec4{1.0, 1.0, 1.0, color2.W()}
	}

	if useAdditive {
		cursorFbo.Begin()
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	siz := settings.Cursor.CursorSize

	if settings.Cursor.EnableCustomTrailGlowOffset {
		color2 = utils.GetColorShifted(color, settings.Cursor.TrailGlowOffset)
	}

	cursorShader.Begin()
	cursorShader.SetUniformAttr(1, int32(1))
	cursorShader.SetUniformAttr(2, batch.Projection)
	cursorShader.SetUniformAttr(3, float32(len(cursor.Points)))

	if settings.Cursor.TrailStyle == 1 {
		cursorShader.SetUniformAttr(7, float32(0.0))
	} else {
		cursorShader.SetUniformAttr(7, float32(1.0))
	}

	cursor.subVao.BeginDraw()

	innerLengthMult := float32(1.0)

	cursorScl := float32(siz * (16.0 / 18) * scale)

	if settings.Cursor.EnableTrailGlow {
		cursorScl = float32(siz * (12.0 / 18) * scale)
		innerLengthMult = float32(settings.Cursor.InnerLengthMult)
		cursorShader.SetUniformAttr(0, color2)
		cursorShader.SetUniformAttr(4, float32(siz*(16.0/18)*scale*settings.Cursor.TrailScale))
		cursorShader.SetUniformAttr(5, float32(settings.Cursor.GlowEndScale))
		if settings.Cursor.TrailStyle > 1 {
			cursorShader.SetUniformAttr(6, float32((hueshift-36)/360))
		}
		cursor.subVao.Draw()
	}

	if settings.Cursor.TrailStyle > 1 {
		cursorShader.SetUniformAttr(6, float32(hueshift/360))
	}
	cursorShader.SetUniformAttr(0, color)
	cursorShader.SetUniformAttr(4, cursorScl*float32(settings.Cursor.TrailScale))
	cursorShader.SetUniformAttr(3, float32(len(cursor.Points))*innerLengthMult)
	cursorShader.SetUniformAttr(5, float32(settings.Cursor.TrailEndScale))

	cursor.subVao.Draw()

	cursor.subVao.EndDraw()

	cursorShader.End()

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
		cursorFbo.End()

		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)

		fboShader.Begin()
		fboShader.SetUniformAttr(0, int32(30))
		fboSlice.BeginDraw()
		fboSlice.Draw()
		fboSlice.EndDraw()
		fboShader.End()
	}

}
