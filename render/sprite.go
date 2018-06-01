package render

import (
	"github.com/faiface/glhf"
	"github.com/go-gl/mathgl/mgl32"
	"danser/bmath"
	"github.com/go-gl/gl/v3.3-core/gl"
)

var shader *glhf.Shader = nil
var vao *glhf.VertexSlice

func setup() {
	circleVertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}

	circleUniformFormat := glhf.AttrFormat{
		{Name: "col_tint", Type: glhf.Vec4},
		{Name: "tex", Type: glhf.Int},
		{Name: "proj", Type: glhf.Mat4},
		{Name: "model", Type: glhf.Mat4},
		{Name: "endTrans", Type: glhf.Mat4},
	}
	shader, _ = glhf.NewShader(circleVertexFormat, circleUniformFormat, vertex, fragment)

	vao = glhf.MakeVertexSlice(shader, 6, 6)
	vao.Begin()
	vao.SetVertexData([]float32{
		-1, -1, 0, 0, 0,
		1, -1, 0, 1, 0,
		-1, 1, 0, 0, 1,
		1, -1, 0, 1, 0,
		1, 1, 0, 1, 1,
		-1, 1, 0, 0, 1,
	})
	vao.End()
	
}

type SpriteBatch struct {
	color mgl32.Vec4
	projection mgl32.Mat4
	position mgl32.Mat4
	scale mgl32.Mat4
	transform mgl32.Mat4
	lastTrans mgl32.Mat4
}

func NewSpriteBatch() *SpriteBatch {
	if shader == nil {
		setup()
	}
	return &SpriteBatch{
		mgl32.Vec4{1, 1, 1, 1},
		mgl32.Ident4(),
		mgl32.Ident4(),
		mgl32.Ident4(),
		mgl32.Ident4(),
		mgl32.Ident4()}
}

func (batch *SpriteBatch) Begin() {
	shader.Begin()
	shader.SetUniformAttr(0, batch.color)
	shader.SetUniformAttr(3, batch.transform)
	shader.SetUniformAttr(2, batch.projection)
	shader.SetUniformAttr(4, batch.lastTrans)
	vao.Begin()
}

func (batch *SpriteBatch) SetColor(r, g, b, a float64) {
	batch.color = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
	shader.SetUniformAttr(0, batch.color)
}

func (batch *SpriteBatch) SetColorM(color mgl32.Vec4) {
	batch.color = color
	shader.SetUniformAttr(0, batch.color)
}

func (batch *SpriteBatch) SetTranslation(vec bmath.Vector2d) {
	batch.position = mgl32.Translate3D(float32(vec.X), float32(vec.Y), 0)
	batch.transform = batch.position.Mul4(batch.scale)
	shader.SetUniformAttr(3, batch.transform)
}

func (batch *SpriteBatch) SetScale(scaleX, scaleY float64) {
	batch.scale = mgl32.Scale3D(float32(scaleX), float32(scaleY), 1)
	batch.transform = batch.position.Mul4(batch.scale)
	shader.SetUniformAttr(3, batch.transform)
}

func (batch *SpriteBatch) ResetTransform() {
	batch.scale = mgl32.Ident4()
	batch.position = mgl32.Ident4()
	batch.transform = mgl32.Ident4()
	shader.SetUniformAttr(3, batch.transform)
}

func (batch *SpriteBatch) DrawTexture(vec bmath.Vector2d, texture *glhf.Texture) {
	gl.ActiveTexture(gl.TEXTURE0)
	texture.Begin()
	transf := (batch.position.Mul4(mgl32.Translate3D(float32(vec.X), float32(vec.Y), 0))).Mul4(batch.scale.Mul4(mgl32.Scale3D(float32(texture.Width())/2, float32(texture.Height())/2, 1)))
	shader.SetUniformAttr(3, transf)
	shader.SetUniformAttr(1, uint32(0))

	vao.Draw()

	shader.SetUniformAttr(3, batch.transform)
	texture.End()
}

func (batch *SpriteBatch) DrawUnscaled(vec bmath.Vector2d, texture *glhf.Texture) {
	gl.ActiveTexture(gl.TEXTURE0)
	texture.Begin()

	batch.DrawUnit(vec, 0)

	texture.End()
}

func (batch *SpriteBatch) DrawUnit(vec bmath.Vector2d, unit int) {
	transf := (batch.position.Mul4(mgl32.Translate3D(float32(vec.X), float32(vec.Y), 0))).Mul4(batch.scale)
	shader.SetUniformAttr(3, transf)
	shader.SetUniformAttr(1, int32(unit))

	vao.Draw()

	shader.SetUniformAttr(3, batch.transform)
}

func (batch *SpriteBatch) DrawUnitR(unit int) {
	shader.SetUniformAttr(1, int32(unit))
	vao.Draw()
}

func (batch *SpriteBatch) DrawSeparate(vec bmath.Vector2d, unit int) {
	transf := (batch.position.Mul4(mgl32.Translate3D(float32(vec.X), float32(vec.Y), 0))).Mul4(batch.scale)
	shader.SetUniformAttr(3, transf)
	shader.SetUniformAttr(1, unit)

	vao.Draw()

	shader.SetUniformAttr(3, batch.transform)
}

func (batch *SpriteBatch) SetCamera(camera mgl32.Mat4) {
	batch.projection = camera
	shader.SetUniformAttr(2, batch.projection)
}

func (batch *SpriteBatch) End() {
	vao.End()
	shader.End()
}
func (batch *SpriteBatch) SetEndTransform(dz mgl32.Mat4) {
	batch.lastTrans = dz
	shader.SetUniformAttr(4, dz)
}

const vertex = `
#version 330

in vec3 in_position;
in vec2 in_tex_coord;

uniform mat4 proj; 
uniform mat4 model; 
uniform mat4 endTrans; 

out vec2 tex_coord;
void main()
{
    gl_Position = endTrans * (proj * (model * vec4(in_position, 1)));
    tex_coord = in_tex_coord;
}
`

const fragment = `
#version 330

uniform sampler2D tex;
uniform vec4 col_tint;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture2D(tex, tex_coord);
	color = in_color*col_tint;
}
`