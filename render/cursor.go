package render

import (
	"github.com/wieku/danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v3.3-core/gl"
	"math"
	"github.com/faiface/glhf"
	"sync"
	"github.com/lucasb-eyer/go-colorful"
)

var cursorShader *glhf.Shader = nil
var cursorFbo *glhf.Frame = nil

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
	}

	var err error
	cursorShader, err = glhf.NewShader(vertexFormat, uniformFormat, cursorVertex, cursorFragment)

	if err != nil {
		panic(err)
	}

	cursorFbo = glhf.NewFrame(1920, 1080, true)
}

type Cursor struct {
	Points []bmath.Vector2d
	Max int
	Position bmath.Vector2d
	LastPos bmath.Vector2d
	time float64

	vertices []float32
	vaoDirty bool
	vaoChannel chan[]float32
	vao *glhf.VertexSlice
	lastLen int
	mutex *sync.Mutex
}

func NewCursor() *Cursor {
	if cursorShader == nil {
		initCursor()
	}

	vao := glhf.MakeVertexSlice(cursorShader, 1024*1024, 1024*1024) //creating approx. 4MB vao, just in case
	verts := make ([]float32, 1024*1024*9)
	return &Cursor{Max:1000, LastPos: bmath.NewVec2d(100, 100), Position: bmath.NewVec2d(100, 100), vao: vao, vertices: verts, mutex: &sync.Mutex{}, vaoChannel: make(chan[]float32, 1)}
}


func (cr *Cursor) SetPos(pt bmath.Vector2d) {
	cr.Position = pt
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

		select {
			case cr.vaoChannel <- arr:
			default:
		}

		cr.vaoDirty = true
	} else {
		select {
		case cr.vaoChannel <- make([]float32, 0):
		default:
		}
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

	siz := 18.0

	tohsv := colorful.Color{float64(color[0]), float64(color[1]), float64(color[2])}
	h, s, v := tohsv.Hsv()
	h -= 36.0
	if h < 0 {
		h += 360.0
	}

	col2 := colorful.Hsv(h, s, v)
	colf2 := mgl32.Vec4{float32(col2.R), float32(col2.G), float32(col2.B), color.W()}

	cursorShader.SetUniformAttr(0, colf2)
	cursorShader.SetUniformAttr(1, int32(1))
	cursorShader.SetUniformAttr(2, batch.Projection)
	cursorShader.SetUniformAttr(3, float32(len(cursor.Points)))
	cursorShader.SetUniformAttr(4, float32(siz*(16.0/18)*scale))

	select {
		case arr := <-cursor.vaoChannel :
			subVao := cursor.vao.Slice(0, len(arr)/9)
			subVao.Begin()

			subVao.SetVertexData(arr)

			subVao.End()
			cursor.vaoDirty = false
			cursor.lastLen = len(arr)/(9*6)
			break
	default:

	}

	subVao := cursor.vao.Slice(0, cursor.lastLen*6)
	subVao.Begin()


	subVao.Draw()

	cursorShader.SetUniformAttr(0, color)
	cursorShader.SetUniformAttr(4, float32(siz*(12.0/18)*scale))

	subVao.Draw()

	subVao.End()
	cursorFbo.End()
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	gl.ActiveTexture(gl.TEXTURE1)
	cursorFbo.Texture().Begin()
	fboShader.Begin()
	fboShader.SetUniformAttr(0, int32(1))
	fboShader.SetUniformAttr(1, float32(1))
	fboSlice.Begin()
	fboSlice.Draw()
	fboSlice.End()
	fboShader.End()

	cursorShader.End()

	batch.Begin()

	batch.SetTranslation(bmath.NewVec2d(0, 0))
	batch.SetScale(scale*siz, scale*siz)

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), float64(color[3]))
	batch.DrawUnit(cursor.Position, 0)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color[3])))
	batch.DrawUnit(cursor.Position, 2)

	batch.End()

	CursorTrail.End()
	CursorTex.End()
	CursorTop.End()

}

const cursorVertex = `
#version 330

in vec3 in_position;
in vec3 in_mid;
in vec2 in_tex_coord;
in float in_index;

uniform mat4 proj;
uniform float scale;
uniform float points;

out vec2 tex_coord;
out float index;

void main() {
    gl_Position = proj * vec4((in_position - in_mid) * scale * (0.4f + 0.6f * in_index / points) + in_mid, 1);
    tex_coord = in_tex_coord;
	index = in_index;
}
`

const cursorFragment = `
#version 330

uniform sampler2D tex;
uniform vec4 col_tint;
uniform float points;

in vec2 tex_coord;
in float index;

out vec4 color;

void main() {
    vec4 in_color = texture2D(tex, tex_coord);
	color = in_color * col_tint * vec4(1, 1, 1, smoothstep(0, points / 3, index));
}
`