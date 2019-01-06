package render

import (
	"github.com/wieku/glhf"
	"github.com/go-gl/mathgl/mgl32"
	"danser/bmath"
	"github.com/go-gl/gl/v3.3-core/gl"
	"io/ioutil"
	"danser/render/texture"
)

const batchSize = 2000

type SpriteBatch struct {
	shader *glhf.Shader
	additive   bool
	color      mgl32.Vec4
	Projection mgl32.Mat4
	position   bmath.Vector2d
	scale      bmath.Vector2d
	subscale   bmath.Vector2d
	rotation   float64

	transform mgl32.Mat4
	texture   texture.Texture

	data        []float32
	vao         *glhf.VertexSlice
	currentSize int
}

func NewSpriteBatch() *SpriteBatch {
	circleVertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec3},
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
	shader, err := glhf.NewShader(circleVertexFormat, circleUniformFormat, string(vert), string(frag))

	if err != nil {
		panic("Sprite: " + err.Error())
	}

	return &SpriteBatch{
		shader,
		false,
		mgl32.Vec4{1, 1, 1, 1},
		mgl32.Ident4(),
		bmath.NewVec2d(0, 0),
		bmath.NewVec2d(1, 1),
		bmath.NewVec2d(1, 1),
		0,
		mgl32.Ident4(),
		nil,
		make([]float32, batchSize*6*11),
		glhf.MakeVertexSlice(shader, batchSize*6, batchSize*6),
		0}
}

func (batch *SpriteBatch) Begin() {
	batch.shader.Begin()
	batch.shader.SetUniformAttr(0, batch.Projection)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	if batch.texture != nil && batch.texture.GetLocation() == 0 {
		batch.texture.Bind(0)
	}
}

func (batch *SpriteBatch) bind(texture texture.Texture) {
	if batch.texture != nil {
		if batch.texture.GetID() == texture.GetID() {
			return
		}
		batch.Flush()
	}

	if texture.GetLocation() == 0 {
		texture.Bind(0)
	}
	batch.texture = texture
	batch.shader.SetUniformAttr(1, int32(texture.GetLocation()))
}

func (batch *SpriteBatch) DrawUnit(texture texture.TextureRegion) {
	newScale := batch.scale.Mult(batch.subscale)

	vec00 := bmath.NewVec2d(-1, -1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)
	vec10 := bmath.NewVec2d(1, -1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)
	vec11 := bmath.NewVec2d(1, 1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)
	vec01 := bmath.NewVec2d(-1, 1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)

	batch.DrawUnitSep(vec00, vec10, vec11, vec01, batch.color, texture)
}

func (batch *SpriteBatch) DrawUnitSep(vec00, vec10, vec11, vec01 bmath.Vector2d, color mgl32.Vec4, texture texture.TextureRegion) {

	batch.bind(texture.Texture)

	batch.addVertex(vec00.AsVec3(), mgl32.Vec3{texture.U1, texture.V1, float32(texture.Layer)}, color)
	batch.addVertex(vec10.AsVec3(), mgl32.Vec3{texture.U2, texture.V1, float32(texture.Layer)}, color)
	batch.addVertex(vec11.AsVec3(), mgl32.Vec3{texture.U2, texture.V2, float32(texture.Layer)}, color)

	batch.addVertex(vec11.AsVec3(), mgl32.Vec3{texture.U2, texture.V2, float32(texture.Layer)}, color)
	batch.addVertex(vec01.AsVec3(), mgl32.Vec3{texture.U1, texture.V2, float32(texture.Layer)}, color)
	batch.addVertex(vec00.AsVec3(), mgl32.Vec3{texture.U1, texture.V1, float32(texture.Layer)}, color)

	if batch.currentSize >= len(batch.data)-1 {
		batch.Flush()
	}

}

func (batch *SpriteBatch) Flush() {
	if batch.currentSize == 0 {
		return
	}

	subVao := batch.vao.Slice(0, batch.currentSize/11)
	subVao.Begin()
	subVao.SetVertexData(batch.data[:batch.currentSize])
	subVao.Draw()
	subVao.End()
	batch.currentSize = 0
}

func (batch *SpriteBatch) addVertex(vx mgl32.Vec3, texCoord mgl32.Vec3, color mgl32.Vec4) {
	add := 0
	if batch.additive {
		add = 1
	}
	fillArray(batch.data, batch.currentSize, vx.X(), vx.Y(), vx.Z(), texCoord.X(), texCoord.Y(), texCoord.Z(), color.X(), color.Y(), color.Z(), color.W(), float32(add))
	batch.currentSize += 11
}

func (batch *SpriteBatch) End() {
	batch.Flush()
	batch.shader.End()
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

func (batch *SpriteBatch) SetRotation(rad float64) {
	batch.rotation = rad
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
	batch.rotation = 0
}

func (batch *SpriteBatch) SetAdditive(additive bool) {
	batch.additive = additive
}

func (batch *SpriteBatch) DrawTexture(texture texture.TextureRegion) {
	newScale := bmath.NewVec2d(batch.scale.X*batch.subscale.X*float64(texture.Width)/2, batch.scale.Y*batch.subscale.Y*float64(texture.Height)/2)

	vec00 := bmath.NewVec2d(-1, -1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)
	vec10 := bmath.NewVec2d(1, -1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)
	vec11 := bmath.NewVec2d(1, 1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)
	vec01 := bmath.NewVec2d(-1, 1).Mult(newScale).Rotate(batch.rotation).Add(batch.position)

	batch.DrawUnitSep(vec00, vec10, vec11, vec01, batch.color, texture)
}

func (batch *SpriteBatch) DrawStObject(position, origin, scale bmath.Vector2d, flip bmath.Vector2d, rotation float64, color mgl32.Vec4, additive bool, texture texture.TextureRegion) {
	newScale := bmath.NewVec2d(scale.X*float64(texture.Width)/2, scale.Y*float64(texture.Height)/2)
	newPosition := bmath.NewVec2d(position.X-64, position.Y-48)

	vec00 := bmath.NewVec2d(-1, -1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)
	vec10 := bmath.NewVec2d(1, -1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)
	vec11 := bmath.NewVec2d(1, 1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)
	vec01 := bmath.NewVec2d(-1, 1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)

	batch.SetAdditive(additive)
	batch.DrawUnitSep(vec00, vec10, vec11, vec01, mgl32.Vec4{color.X()*batch.color.X(), color.Y()*batch.color.Y(), color.Z()*batch.color.Z(), color.W()*batch.color.W()}, texture)
	batch.SetAdditive(false)
}

func (batch *SpriteBatch) DrawUnscaled(texture texture.TextureRegion) {
	vec00 := bmath.NewVec2d(-1, -1).Add(batch.position)
	vec10 := bmath.NewVec2d(1, -1).Add(batch.position)
	vec11 := bmath.NewVec2d(1, 1).Add(batch.position)
	vec01 := bmath.NewVec2d(-1, 1).Add(batch.position)

	batch.DrawUnitSep(vec00, vec10, vec11, vec01, batch.color, texture)
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
	batch.shader.SetUniformAttr(0, batch.Projection)
}
