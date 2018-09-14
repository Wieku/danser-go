package render

import (
	"github.com/wieku/glhf"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/bmath"
	"github.com/go-gl/gl/v3.3-core/gl"
	"io/ioutil"
)

const batchSize = 2000

var shader *glhf.Shader = nil

func setup() {
	circleVertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
		{Name: "in_color", Type: glhf.Vec4},
		{Name: "in_additive", Type: glhf.Float},
	}

	circleUniformFormat := glhf.AttrFormat{
		{Name: "proj", Type: glhf.Mat4},
		{Name: "tex", Type: glhf.Int},
	}
	vert, _ := ioutil.ReadFile("assets/shaders/sprite.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/sprite.fsh")
	var err error
	shader, err = glhf.NewShader(circleVertexFormat, circleUniformFormat, string(vert), string(frag))

	if err != nil {
		panic("Sprite: " + err.Error())
	}

}

type SpriteBatch struct {
	additive   bool
	color      mgl32.Vec4
	Projection mgl32.Mat4
	position   bmath.Vector2d
	scale      bmath.Vector2d
	subscale   bmath.Vector2d

	transform mgl32.Mat4
	texture   *glhf.Texture
	unit      int

	data        []float32
	vao         *glhf.VertexSlice
	currentSize int
}

func NewSpriteBatch() *SpriteBatch {
	if shader == nil {
		setup()
	}
	return &SpriteBatch{
		false,
		mgl32.Vec4{1, 1, 1, 1},
		mgl32.Ident4(),
		bmath.NewVec2d(0, 0),
		bmath.NewVec2d(1, 1),
		bmath.NewVec2d(1, 1),
		mgl32.Ident4(),
		nil,
		-1,
		make([]float32, batchSize*6*10),
		glhf.MakeVertexSlice(shader, batchSize*6, batchSize*6),
		0}
}

func (batch *SpriteBatch) Begin() {
	shader.Begin()
	shader.SetUniformAttr(0, batch.Projection)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
}

func (batch *SpriteBatch) bind(texture *glhf.Texture) {
	if batch.texture != nil {
		if batch.texture.ID() == texture.ID() {
			return
		}
		batch.Flush()
		batch.texture.End()
	}

	gl.ActiveTexture(gl.TEXTURE0)
	texture.Begin()
	batch.texture = texture
	batch.unit = 0
	shader.SetUniformAttr(1, int32(0))
}

func (batch *SpriteBatch) DrawUnit(unit int) {
	transf := mgl32.Translate3D(batch.position.X32(), batch.position.Y32(), 0).Mul4(mgl32.Scale3D(batch.scale.X32()*batch.subscale.X32(), batch.scale.Y32()*batch.subscale.Y32(), 0))
	batch.DrawUnitSep(transf, batch.color, unit)
}

func (batch *SpriteBatch) DrawUnitSep(transform mgl32.Mat4, color mgl32.Vec4, unit int) {

	if batch.unit != unit {
		batch.Flush()
		batch.unit = unit
		shader.SetUniformAttr(1, int32(unit))
	}

	vec00 := transform.Mul4x1(mgl32.Vec4{-1, -1, 0, 1})
	vec10 := transform.Mul4x1(mgl32.Vec4{1, -1, 0, 1})
	vec11 := transform.Mul4x1(mgl32.Vec4{1, 1, 0, 1})
	vec01 := transform.Mul4x1(mgl32.Vec4{-1, 1, 0, 1})

	batch.addVertex(vec00.Vec3(), mgl32.Vec2{0, 0}, color)
	batch.addVertex(vec10.Vec3(), mgl32.Vec2{1, 0}, color)
	batch.addVertex(vec11.Vec3(), mgl32.Vec2{1, 1}, color)

	batch.addVertex(vec11.Vec3(), mgl32.Vec2{1, 1}, color)
	batch.addVertex(vec01.Vec3(), mgl32.Vec2{0, 1}, color)
	batch.addVertex(vec00.Vec3(), mgl32.Vec2{0, 0}, color)

	if batch.currentSize >= len(batch.data)-1 {
		batch.Flush()
	}

}

func (batch *SpriteBatch) Flush() {
	if batch.currentSize == 0 {
		return
	}

	subVao := batch.vao.Slice(0, batch.currentSize/10)
	subVao.Begin()
	subVao.SetVertexData(batch.data[:batch.currentSize])
	subVao.Draw()
	subVao.End()
	batch.currentSize = 0
}

func (batch *SpriteBatch) addVertex(vx mgl32.Vec3, texCoord mgl32.Vec2, color mgl32.Vec4) {
	add := 0
	if batch.additive {
		add = 1
	}
	fillArray(batch.data, batch.currentSize, vx.X(), vx.Y(), vx.Z(), texCoord.X(), texCoord.Y(), color.X(), color.Y(), color.Z(), color.W(), float32(add))
	batch.currentSize += 10
}

func (batch *SpriteBatch) End() {
	batch.Flush()
	shader.End()
	if batch.texture != nil {
		batch.texture.End()
		batch.texture = nil
	}
	batch.unit = -1

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (batch *SpriteBatch) SetColor(r, g, b, a float64) {
	batch.color = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
}

func (batch *SpriteBatch) SetColorM(color mgl32.Vec4) {
	batch.color = color
}

func (batch *SpriteBatch) SetTranslation(vec bmath.Vector2d) {
	batch.position = vec
}

func (batch *SpriteBatch) SetScale(scaleX, scaleY float64) {
	batch.scale = bmath.NewVec2d(scaleX, scaleY)
}

func (batch *SpriteBatch) SetSubScale(scaleX, scaleY float64) {
	batch.subscale = bmath.NewVec2d(scaleX, scaleY)
}

func (batch *SpriteBatch) ResetTransform() {
	batch.scale = bmath.NewVec2d(1, 1)
	batch.subscale = bmath.NewVec2d(1, 1)
	batch.position = bmath.NewVec2d(0, 0)
}

func (batch *SpriteBatch) SetAdditive(additive bool) {
	batch.additive = additive
}

func (batch *SpriteBatch) DrawTexture(texture *glhf.Texture) {
	batch.bind(texture)
	transf := mgl32.Translate3D(batch.position.X32(), batch.position.Y32(), 0).Mul4(mgl32.Scale3D(batch.scale.X32()*batch.subscale.X32()*float32(texture.Width())/2, batch.scale.Y32()*batch.subscale.Y32()*float32(texture.Width())/2, 0))
	batch.DrawUnitSep(transf, batch.color, 0)
}

func (batch *SpriteBatch) DrawStObject(position, origin, scale bmath.Vector2d, flip bmath.Vector2d, rotation float64, color mgl32.Vec4, additive bool, texture *glhf.Texture) {
	transf := mgl32.Translate3D(position.X32()-64, position.Y32()-48, 0).Mul4(mgl32.HomogRotate3DZ(float32(rotation))).Mul4(mgl32.Scale3D(scale.X32()*float32(texture.Width())/2, scale.Y32()*float32(texture.Height())/2, 1)).Mul4(mgl32.Translate3D(-origin.X32(), -origin.Y32(), 0)).Mul4(mgl32.Scale3D(flip.X32(), flip.Y32(), 1))

	batch.bind(texture)
	batch.SetAdditive(additive)
	batch.DrawUnitSep(transf, color, 0)
	batch.SetAdditive(false)
}

func (batch *SpriteBatch) DrawUnscaled(texture *glhf.Texture) {
	batch.bind(texture)
	transf := mgl32.Translate3D(batch.position.X32(), batch.position.Y32(), 0)
	batch.DrawUnitSep(transf, batch.color, 0)
}

/*func (batch *SpriteBatch) DrawUnitR(unit int) {
	shader.SetUniformAttr(1, int32(unit))
	vao.Draw()
}*/

/*func (batch *SpriteBatch) DrawSeparate(vec bmath.Vector2d, unit int) {
	transf := (batch.position.Mul4(mgl32.Translate3D(float32(vec.X), float32(vec.Y), 0))).Mul4(batch.scale)
	shader.SetUniformAttr(3, transf)
	shader.SetUniformAttr(1, int32(unit))

	vao.Draw()

	shader.SetUniformAttr(3, batch.transform)
}*/

func (batch *SpriteBatch) SetCamera(camera mgl32.Mat4) {
	batch.Flush()
	batch.Projection = camera
	shader.SetUniformAttr(0, batch.Projection)
}
