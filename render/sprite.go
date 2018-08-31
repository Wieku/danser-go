package render

import (
	"github.com/wieku/glhf"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/bmath"
	"github.com/go-gl/gl/v3.3-core/gl"
	"io/ioutil"
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
	vert, _ := ioutil.ReadFile("assets/shaders/sprite.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/sprite.fsh")
	var err error
	shader, err = glhf.NewShader(circleVertexFormat, circleUniformFormat, string(vert), string(frag))

	if err != nil {
		panic("Sprite: " + err.Error())
	}

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
	color      mgl32.Vec4
	Projection mgl32.Mat4
	position   mgl32.Mat4
	scale      mgl32.Mat4
	transform  mgl32.Mat4
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
		mgl32.Ident4()}
}

func (batch *SpriteBatch) Begin() {
	shader.Begin()
	shader.SetUniformAttr(0, batch.color)
	shader.SetUniformAttr(3, batch.transform)
	shader.SetUniformAttr(2, batch.Projection)
	vao.BeginDraw()
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

func (batch *SpriteBatch) SetSubScale(scaleX, scaleY float64) {
	shader.SetUniformAttr(3, batch.position.Mul4(batch.scale.Mul4(mgl32.Scale3D(float32(scaleX), float32(scaleY), 1))))
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
	shader.SetUniformAttr(1, int32(0))

	vao.Draw()

	shader.SetUniformAttr(3, batch.transform)
	texture.End()
}

func (batch *SpriteBatch) DrawStObject(position, origin, scale bmath.Vector2d, rotation float64, color mgl32.Vec4, texture *glhf.Texture) {
	transf := mgl32.Translate3D(position.X32(), position.Y32(), 0).Mul4(mgl32.HomogRotate3DZ(float32(rotation))).Mul4(mgl32.Scale3D(scale.X32(), scale.Y32(), 1)).Mul4(mgl32.Translate3D(origin.X32()-float32(texture.Width()/2), origin.Y32()-float32(texture.Height()/2), 0))
	shader.SetUniformAttr(3, transf)
	shader.SetUniformAttr(0, color)
	shader.SetUniformAttr(1, int32(0))

	gl.ActiveTexture(gl.TEXTURE0)
	texture.Begin()
	vao.Draw()
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
	shader.SetUniformAttr(1, int32(unit))

	vao.Draw()

	shader.SetUniformAttr(3, batch.transform)
}

func (batch *SpriteBatch) SetCamera(camera mgl32.Mat4) {
	batch.Projection = camera
	shader.SetUniformAttr(2, batch.Projection)
}

func (batch *SpriteBatch) End() {
	vao.EndDraw()
	shader.End()
}