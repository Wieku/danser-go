package render

import (
	"github.com/wieku/danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v3.3-core/gl"
	"math"
	"github.com/wieku/glhf"
	"sync"
	"github.com/wieku/danser/settings"
	"github.com/wieku/danser/utils"
	"io/ioutil"
)

var cursorShader *glhf.Shader = nil
var cursorFbo *glhf.Frame = nil
var Camera *bmath.Camera
var osuRect bmath.Rectangle
func initCursor() {

	vertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_mid", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
		{Name: "in_index", Type: glhf.Float},
	}

	uniformFormat := glhf.AttrFormat{
		{Name: "col_tint", Type: glhf.Vec4},
		{Name: "tex", Type: glhf.Int},
		{Name: "proj", Type: glhf.Mat4},
		{Name: "points", Type: glhf.Float},
		{Name: "scale", Type: glhf.Float},
		{Name: "endScale", Type: glhf.Float},
	}

	var err error
	vert , _ := ioutil.ReadFile("assets/shaders/cursortrail.vsh")
	frag , _ := ioutil.ReadFile("assets/shaders/cursortrail.fsh")
	cursorShader, err = glhf.NewShader(vertexFormat, uniformFormat, string(vert), string(frag))

	if err != nil {
		panic(err)
	}

	cursorFbo = glhf.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	osuRect = Camera.GetWorldRect()
}

type Cursor struct {
	Points []bmath.Vector2d
	Max int
	Position bmath.Vector2d
	LastPos bmath.Vector2d
	VaoPos bmath.Vector2d
	RendPos bmath.Vector2d
	time float64

	vertices []float32
	vaoDirty bool
	vao *glhf.VertexSlice
	subVao *glhf.VertexSlice
	lastLen int
	mutex *sync.Mutex
}

func NewCursor() *Cursor {
	if cursorShader == nil {
		initCursor()
	}

	vao := glhf.MakeVertexSlice(cursorShader, 1024*1024, 1024*1024) //creating approx. 4MB vao, just in case
	return &Cursor{Max:1000, LastPos: bmath.NewVec2d(100, 100), Position: bmath.NewVec2d(100, 100), vao: vao, subVao: vao.Slice(0,0), mutex: &sync.Mutex{}, RendPos: bmath.NewVec2d(100, 100)}
}


func (cr *Cursor) SetPos(pt bmath.Vector2d) {
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

func (cr *Cursor) Update(tim float64) {
	points := cr.Position.Dst(cr.LastPos)*2

	olp := len(cr.Points)


	if points > 0 {
		for i:=0.0; i <= 1.0; i+=1.0/points {
			cr.Points = append(cr.Points, cr.Position.Sub(cr.LastPos).Scl(i).Add(cr.LastPos))
		}
	}

	olp1 := len(cr.Points)

	mult := float64(1)
	if olp < olp1 {
		mult = 0
	}

	cr.LastPos = cr.Position
	times := int64(math.Max(mult, float64(len(cr.Points))/(6*(60.0/tim))))

	if len(cr.Points) > 0 {
		if int(times) < len(cr.Points) {
			cr.Points = cr.Points[times:]
		} else {
			cr.Points = cr.Points[len(cr.Points):]
		}

		arr := make([]float32, len(cr.Points)*6*9)

		for i, o := range cr.Points {
			 bI := i*6*9
			 fillArray(arr, bI, -1+o.X32(), -1+o.Y32(), 0, o.X32(), o.Y32(), 0, 0, 0, float32(i))
			 fillArray(arr, bI+9, 1+o.X32(), -1+o.Y32(), 0, o.X32(), o.Y32(), 0, 1, 0, float32(i))
			 fillArray(arr, bI+9*2, -1+o.X32(), 1+o.Y32(), 0, o.X32(), o.Y32(), 0, 0, 1, float32(i))
			 fillArray(arr, bI+9*3, 1+o.X32(), -1+o.Y32(), 0, o.X32(), o.Y32(), 0, 1, 0, float32(i))
			 fillArray(arr, bI+9*4, 1+o.X32(), 1+o.Y32(), 0, o.X32(), o.Y32(), 0, 1, 1, float32(i))
			 fillArray(arr, bI+9*5, -1+o.X32(), 1+o.Y32(), 0, o.X32(), o.Y32(), 0, 0, 1, float32(i))
		}

		cr.mutex.Lock()
		cr.vertices = arr
		cr.VaoPos = cr.Position
		cr.vaoDirty = true
		cr.mutex.Unlock()

		cr.vaoDirty = true
	} else {
		cr.mutex.Lock()
		cr.vertices = make([]float32, 0)
		cr.VaoPos = cr.Position
		cr.vaoDirty = true
		cr.mutex.Unlock()
	}
}

func fillArray(dst []float32, index int, values... float32) {
	for i, j := range values {
		dst[index+i] = j
	}
}

func (cursor *Cursor) Draw(scale float64, batch *SpriteBatch, color mgl32.Vec4) {
	cursor.DrawM(scale, batch, color, color)
}

func (cursor *Cursor) DrawM(scale float64, batch *SpriteBatch, color mgl32.Vec4, color2 mgl32.Vec4) {
	gl.Disable(gl.DEPTH_TEST)

	gl.ActiveTexture(gl.TEXTURE0)
	CursorTex.Begin()
	gl.ActiveTexture(gl.TEXTURE1)
	CursorTrail.Begin()
	gl.ActiveTexture(gl.TEXTURE2)
	CursorTop.Begin()

	cursorFbo.Begin()
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	cursorShader.Begin()

	siz := settings.Cursor.CursorSize

	if settings.Cursor.EnableCustomTrailGlowOffset {
		color2 = utils.GetColorShifted(color, settings.Cursor.TrailGlowOffset)
	}

	cursorShader.SetUniformAttr(0, color2)
	cursorShader.SetUniformAttr(1, int32(1))
	cursorShader.SetUniformAttr(2, batch.Projection)
	cursorShader.SetUniformAttr(3, float32(len(cursor.Points)))
	cursorShader.SetUniformAttr(4, float32(siz*(16.0/18)*scale))
	cursorShader.SetUniformAttr(5, float32(settings.Cursor.TrailEndScale))

	cursor.mutex.Lock()
	if cursor.vaoDirty {
		cursor.subVao = cursor.vao.Slice(0, len(cursor.vertices)/9)
		cursor.subVao.Begin()
		cursor.subVao.SetVertexData(cursor.vertices)
		cursor.subVao.End()
		cursor.RendPos = cursor.VaoPos
		cursor.vaoDirty = false
	}
	cursor.mutex.Unlock()
	cursor.subVao.Begin()

	cursor.subVao.Draw()

	cursorShader.SetUniformAttr(0, color)
	cursorShader.SetUniformAttr(4, float32(siz*(12.0/18)*scale))

	cursor.subVao.Draw()

	cursor.subVao.End()

	cursorFbo.End()

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	gl.ActiveTexture(gl.TEXTURE1)
	cursorFbo.Texture().Begin()
	fboShader.Begin()
	fboShader.SetUniformAttr(0, int32(1))
	fboSlice.Begin()
	fboSlice.Draw()
	fboSlice.End()
	fboShader.End()

	cursorShader.End()

	batch.Begin()

	batch.SetTranslation(bmath.NewVec2d(0, 0))
	batch.SetScale(siz*scale, siz*scale)

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), float64(color[3]))
	batch.DrawUnit(cursor.RendPos, 0)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color[3])))
	batch.DrawUnit(cursor.RendPos, 2)

	batch.End()

	CursorTrail.End()
	CursorTex.End()
	CursorTop.End()
}