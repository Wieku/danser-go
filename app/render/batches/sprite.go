package batches

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"io/ioutil"
)

const batchSize = 2000

type SpriteBatch struct {
	shader     *shader.Shader
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
	vao         *buffer.VertexSlice
	ibo         *buffer.IndexBufferObject
	currentSize int
	drawing     bool
}

func NewSpriteBatch() *SpriteBatch {
	circleVertexFormat := shader.AttrFormat{
		{Name: "in_position", Type: shader.Vec2},
		{Name: "in_tex_coord", Type: shader.Vec3},
		{Name: "in_color", Type: shader.Vec4},
		{Name: "in_additive", Type: shader.Float},
	}

	circleUniformFormat := shader.AttrFormat{
		{Name: "proj", Type: shader.Mat4},
		{Name: "tex", Type: shader.Int},
	}
	vert, _ := ioutil.ReadFile("assets/shaders/sprite.vsh")
	frag, _ := ioutil.ReadFile("assets/shaders/sprite.fsh")

	var err error
	shader, err := shader.NewShader(circleVertexFormat, circleUniformFormat, string(vert), string(frag))

	if err != nil {
		panic("Sprite: " + err.Error())
	}

	ibo := buffer.NewIndexBufferObject(batchSize * 6)

	ibo.Bind()

	indices := make([]uint16, batchSize*6)

	index := uint16(0)
	for i := 0; i < batchSize*6; i += 6 {
		indices[i] = index
		indices[i+1] = index + 1
		indices[i+2] = index + 2

		indices[i+3] = index + 2
		indices[i+4] = index + 3
		indices[i+5] = index

		index += 4
	}
	ibo.SetData(0, indices)

	ibo.Unbind()

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
		make([]float32, batchSize*4*10),
		buffer.MakeVertexSlice(shader, batchSize*4, batchSize*4),
		ibo,
		0,
		false}
}

func (batch *SpriteBatch) Begin() {
	if batch.drawing {
		panic("Batching is already began")
	}
	batch.drawing = true
	batch.shader.Begin()
	batch.shader.SetUniformAttr(0, batch.Projection)

	batch.vao.Begin()
	batch.ibo.Bind()

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

	batch.addVertex(vec00, mgl32.Vec3{texture.U1, texture.V1, float32(texture.Layer)}, color)
	batch.addVertex(vec10, mgl32.Vec3{texture.U2, texture.V1, float32(texture.Layer)}, color)
	batch.addVertex(vec11, mgl32.Vec3{texture.U2, texture.V2, float32(texture.Layer)}, color)
	batch.addVertex(vec01, mgl32.Vec3{texture.U1, texture.V2, float32(texture.Layer)}, color)

	if batch.currentSize >= len(batch.data)-1 {
		batch.Flush()
	}
}

func (batch *SpriteBatch) Flush() {
	if batch.currentSize == 0 {
		return
	}

	subVao := batch.vao.Slice(0, batch.currentSize/10)
	//subVao.Begin()
	subVao.SetVertexData(batch.data[:batch.currentSize])

	batch.ibo.DrawPart(0, batch.currentSize/10/4*6)
	//batch.ibo.Unbind()

	//subVao.End()
	batch.currentSize = 0
}

func (batch *SpriteBatch) addVertex(vx bmath.Vector2d, texCoord mgl32.Vec3, color mgl32.Vec4) {
	add := 1
	if batch.additive {
		add = 0
	}
	fillArray(batch.data, batch.currentSize, vx.X32(), vx.Y32(), texCoord.X(), texCoord.Y(), texCoord.Z(), color.X(), color.Y(), color.Z(), color.W(), float32(add))
	batch.currentSize += 10
}

func fillArray(dst []float32, index int, values ...float32) {
	for i, j := range values {
		dst[index+i] = j
	}
}

func (batch *SpriteBatch) End() {
	if !batch.drawing {
		panic("Batching is already ended")
	}
	batch.drawing = false
	batch.Flush()

	batch.ibo.Unbind()
	batch.vao.End()

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

func (batch *SpriteBatch) GetRotation() float64 {
	return batch.rotation
}

func (batch *SpriteBatch) SetScale(scaleX, scaleY float64) {
	batch.scale = bmath.NewVec2d(scaleX, scaleY)
}

func (batch *SpriteBatch) GetScale() bmath.Vector2d {
	return batch.scale
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

func (batch *SpriteBatch) DrawStObject(position, origin, scale bmath.Vector2d, flip bmath.Vector2d, rotation float64, color mgl32.Vec4, additive bool, texture texture.TextureRegion, storyboard bool) {
	newScale := bmath.NewVec2d(scale.X*float64(texture.Width)/2, scale.Y*float64(texture.Height)/2)
	newPosition := position
	if storyboard {
		newPosition = bmath.NewVec2d(position.X-64, position.Y-48)
	}

	vec00 := bmath.NewVec2d(-1, -1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)
	vec10 := bmath.NewVec2d(1, -1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)
	vec11 := bmath.NewVec2d(1, 1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)
	vec01 := bmath.NewVec2d(-1, 1).Mult(flip).Sub(origin).Mult(newScale).Rotate(rotation).Add(newPosition)

	batch.SetAdditive(additive)
	batch.DrawUnitSep(vec00, vec10, vec11, vec01, mgl32.Vec4{color.X() * batch.color.X(), color.Y() * batch.color.Y(), color.Z() * batch.color.Z(), color.W() * batch.color.W()}, texture)
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
	if batch.drawing {
		batch.Flush()
	}

	batch.Projection = camera
	if batch.drawing {
		batch.shader.SetUniformAttr(0, batch.Projection)
	}
}
