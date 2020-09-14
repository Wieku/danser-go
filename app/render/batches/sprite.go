package batches

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"io/ioutil"
	"math"
)

const defaultBatchSize = 2000

type SpriteBatch struct {
	shader     *shader.RShader
	additive   bool
	color      mgl32.Vec4
	Projection mgl32.Mat4
	position   bmath.Vector2d
	scale      bmath.Vector2d
	subscale   bmath.Vector2d
	rotation   float64

	transform mgl32.Mat4
	texture   texture.Texture

	vertexSize    int
	data          []float32
	vao           *buffer.VertexArrayObject
	ibo           *buffer.IndexBufferObject
	currentSize   int
	currentFloats int
	drawing       bool
	maxSprites    int
}

func NewSpriteBatch() *SpriteBatch {
	return NewSpriteBatchSize(defaultBatchSize)
}

func NewSpriteBatchSize(maxSprites int) *SpriteBatch {
	if maxSprites*6 > 0xFFFF {
		panic(fmt.Sprintf("SpriteBatch size is too big, maximum sprites allowed: 10922, given: %d", maxSprites))
	}

	vert, err := ioutil.ReadFile("assets/shaders/sprite.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := ioutil.ReadFile("assets/shaders/sprite.fsh")
	if err != nil {
		panic(err)
	}

	rShader := shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

	vao := buffer.NewVertexArrayObject()

	vao.AddVBO("default", maxSprites*4, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_tex_coord", Type: attribute.Vec3},
		{Name: "in_color", Type: attribute.Vec4},
		{Name: "in_additive", Type: attribute.Float},
	})

	vao.Bind()
	vao.Attach(rShader)
	vao.Unbind()

	ibo := buffer.NewIndexBufferObject(maxSprites * 6)

	ibo.Bind()

	indices := make([]uint16, maxSprites*6)

	index := uint16(0)

	for i := 0; i < maxSprites*6; i += 6 {
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

	vertexSize := vao.GetVBOFormat("default").Size() / 4
	data := make([]float32, defaultBatchSize*4*vertexSize)

	return &SpriteBatch{
		shader:     rShader,
		color:      mgl32.Vec4{1, 1, 1, 1},
		Projection: mgl32.Ident4(),
		scale:      bmath.NewVec2d(1, 1),
		subscale:   bmath.NewVec2d(1, 1),
		transform:  mgl32.Ident4(),
		vertexSize: vertexSize,
		data:       data,
		vao:        vao,
		ibo:        ibo,
		maxSprites: maxSprites,
	}
}

func (batch *SpriteBatch) Begin() {
	if batch.drawing {
		panic("Batching has already begun")
	}

	batch.drawing = true

	batch.shader.Bind()
	batch.shader.SetUniform("proj", batch.Projection)

	batch.vao.Bind()
	batch.ibo.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)
}

func (batch *SpriteBatch) bind(texture texture.Texture) {
	if batch.texture != nil {
		if batch.texture.GetID() == texture.GetID() {
			return
		}

		batch.Flush()
	}

	batch.texture = texture
}

func (batch *SpriteBatch) DrawUnit(texture texture.TextureRegion) {
	newScale := batch.scale.Mult(batch.subscale)

	cos := math.Cos(batch.rotation)
	sin := math.Sin(batch.rotation)

	vec00 := bmath.NewVec2d(-1, -1)
	vec00.X = vec00.X * newScale.X
	vec00.Y = vec00.Y * newScale.Y
	vec00.X, vec00.Y = vec00.X*cos-vec00.Y*sin, vec00.X*sin+vec00.Y*cos
	vec00.X += batch.position.X
	vec00.Y += batch.position.Y

	vec10 := bmath.NewVec2d(1, -1)
	vec10.X = vec10.X * newScale.X
	vec10.Y = vec10.Y * newScale.Y
	vec10.X, vec10.Y = vec10.X*cos-vec10.Y*sin, vec10.X*sin+vec10.Y*cos
	vec10.X += batch.position.X
	vec10.Y += batch.position.Y

	vec11 := bmath.NewVec2d(1, 1)
	vec11.X = vec11.X * newScale.X
	vec11.Y = vec11.Y * newScale.Y
	vec11.X, vec11.Y = vec11.X*cos-vec11.Y*sin, vec11.X*sin+vec11.Y*cos
	vec11.X += batch.position.X
	vec11.Y += batch.position.Y

	vec01 := bmath.NewVec2d(-1, 1)
	vec01.X = vec01.X * newScale.X
	vec01.Y = vec01.Y * newScale.Y
	vec01.X, vec01.Y = vec01.X*cos-vec01.Y*sin, vec01.X*sin+vec01.Y*cos
	vec01.X += batch.position.X
	vec01.Y += batch.position.Y

	batch.DrawUnitSep(vec00, vec10, vec11, vec01, batch.color, texture)
}

func (batch *SpriteBatch) DrawUnitSep(vec00, vec10, vec11, vec01 bmath.Vector2d, color mgl32.Vec4, texture texture.TextureRegion) {
	batch.bind(texture.Texture)

	batch.addVertex(vec00, mgl32.Vec3{texture.U1, texture.V1, float32(texture.Layer)}, color)
	batch.addVertex(vec10, mgl32.Vec3{texture.U2, texture.V1, float32(texture.Layer)}, color)
	batch.addVertex(vec11, mgl32.Vec3{texture.U2, texture.V2, float32(texture.Layer)}, color)
	batch.addVertex(vec01, mgl32.Vec3{texture.U1, texture.V2, float32(texture.Layer)}, color)

	if batch.currentSize/4 >= batch.maxSprites {
		batch.Flush()
	}
}

func (batch *SpriteBatch) Flush() {
	if batch.currentSize == 0 {
		return
	}

	if batch.texture.GetLocation() == 0 {
		batch.texture.Bind(0)
	}

	batch.shader.SetUniform("tex", int32(batch.texture.GetLocation()))

	batch.vao.SetData("default", 0, batch.data[:batch.currentFloats])

	batch.ibo.DrawPart(0, batch.currentSize/4*6)

	batch.currentSize = 0
	batch.currentFloats = 0
}

func (batch *SpriteBatch) addVertex(vx bmath.Vector2d, texCoord mgl32.Vec3, color mgl32.Vec4) {
	add := 1
	if batch.additive {
		add = 0
	}

	batch.data[batch.currentFloats] = vx.X32()
	batch.data[batch.currentFloats+1] = vx.Y32()
	batch.data[batch.currentFloats+2] = texCoord.X()
	batch.data[batch.currentFloats+3] = texCoord.Y()
	batch.data[batch.currentFloats+4] = texCoord.Z()
	batch.data[batch.currentFloats+5] = color.X()
	batch.data[batch.currentFloats+6] = color.Y()
	batch.data[batch.currentFloats+7] = color.Z()
	batch.data[batch.currentFloats+8] = color.W()
	batch.data[batch.currentFloats+9] = float32(add)

	batch.currentFloats += batch.vertexSize

	batch.currentSize++
}

func (batch *SpriteBatch) End() {
	if !batch.drawing {
		panic("Batching has already ended")
	}

	batch.drawing = false

	batch.Flush()

	batch.ibo.Unbind()
	batch.vao.Unbind()

	batch.shader.Unbind()

	blend.Pop()
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

	cos := math.Cos(batch.rotation)
	sin := math.Sin(batch.rotation)

	vec00 := bmath.NewVec2d(-1, -1)
	vec00.X = vec00.X * newScale.X
	vec00.Y = vec00.Y * newScale.Y
	vec00.X, vec00.Y = vec00.X*cos-vec00.Y*sin, vec00.X*sin+vec00.Y*cos
	vec00.X += batch.position.X
	vec00.Y += batch.position.Y

	vec10 := bmath.NewVec2d(1, -1)
	vec10.X = vec10.X * newScale.X
	vec10.Y = vec10.Y * newScale.Y
	vec10.X, vec10.Y = vec10.X*cos-vec10.Y*sin, vec10.X*sin+vec10.Y*cos
	vec10.X += batch.position.X
	vec10.Y += batch.position.Y

	vec11 := bmath.NewVec2d(1, 1)
	vec11.X = vec11.X * newScale.X
	vec11.Y = vec11.Y * newScale.Y
	vec11.X, vec11.Y = vec11.X*cos-vec11.Y*sin, vec11.X*sin+vec11.Y*cos
	vec11.X += batch.position.X
	vec11.Y += batch.position.Y

	vec01 := bmath.NewVec2d(-1, 1)
	vec01.X = vec01.X * newScale.X
	vec01.Y = vec01.Y * newScale.Y
	vec01.X, vec01.Y = vec01.X*cos-vec01.Y*sin, vec01.X*sin+vec01.Y*cos
	vec01.X += batch.position.X
	vec01.Y += batch.position.Y

	batch.DrawUnitSep(vec00, vec10, vec11, vec01, batch.color, texture)
}

func (batch *SpriteBatch) DrawStObject(position, origin, scale bmath.Vector2d, flip bmath.Vector2d, rotation float64, color mgl32.Vec4, additive bool, texture texture.TextureRegion, storyboard bool) {
	newScale := bmath.NewVec2d(scale.X*float64(texture.Width)/2*batch.scale.X*batch.subscale.X, scale.Y*float64(texture.Height)/2*batch.scale.Y*batch.subscale.Y)
	newPosition := position.Add(batch.position)

	if storyboard {
		newPosition = bmath.NewVec2d(position.X-64, position.Y-48)
	}

	cos := math.Cos(rotation)
	sin := math.Sin(rotation)

	vec00 := bmath.NewVec2d(-1, -1)
	vec00.X = ((vec00.X * flip.X) - origin.X) * newScale.X
	vec00.Y = ((vec00.Y * flip.Y) - origin.Y) * newScale.Y
	vec00.X, vec00.Y = vec00.X*cos-vec00.Y*sin, vec00.X*sin+vec00.Y*cos
	vec00.X += newPosition.X
	vec00.Y += newPosition.Y

	vec10 := bmath.NewVec2d(1, -1)
	vec10.X = ((vec10.X * flip.X) - origin.X) * newScale.X
	vec10.Y = ((vec10.Y * flip.Y) - origin.Y) * newScale.Y
	vec10.X, vec10.Y = vec10.X*cos-vec10.Y*sin, vec10.X*sin+vec10.Y*cos
	vec10.X += newPosition.X
	vec10.Y += newPosition.Y

	vec11 := bmath.NewVec2d(1, 1)
	vec11.X = ((vec11.X * flip.X) - origin.X) * newScale.X
	vec11.Y = ((vec11.Y * flip.Y) - origin.Y) * newScale.Y
	vec11.X, vec11.Y = vec11.X*cos-vec11.Y*sin, vec11.X*sin+vec11.Y*cos
	vec11.X += newPosition.X
	vec11.Y += newPosition.Y

	vec01 := bmath.NewVec2d(-1, 1)
	vec01.X = ((vec01.X * flip.X) - origin.X) * newScale.X
	vec01.Y = ((vec01.Y * flip.Y) - origin.Y) * newScale.Y
	vec01.X, vec01.Y = vec01.X*cos-vec01.Y*sin, vec01.X*sin+vec01.Y*cos
	vec01.X += newPosition.X
	vec01.Y += newPosition.Y

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

func (batch *SpriteBatch) SetCamera(camera mgl32.Mat4) {
	if batch.Projection == camera {
		return
	}

	if batch.drawing {
		batch.Flush()
	}

	batch.Projection = camera
	if batch.drawing {
		batch.shader.SetUniform("proj", batch.Projection)
	}
}
