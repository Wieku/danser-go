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

	vao.AddVBO("default", 4, 0, attribute.Format{
		{Name: "base_pos", Type: attribute.Vec2},
		{Name: "base_uv", Type: attribute.Vec2},
	})

	vao.SetData("default", 0, []float32{
		-1, -1, 0, 2,
		1, -1, 1, 2,
		1, 1, 1, 3,
		-1, 1, 0, 3,
	})

	vao.AddVBO("sprites", maxSprites, 1, attribute.Format{
		{Name: "in_origin", Type: attribute.Vec2},
		{Name: "in_scale", Type: attribute.Vec2},
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_rotation", Type: attribute.Float},
		{Name: "in_uvs", Type: attribute.Vec4},
		{Name: "in_layer", Type: attribute.Float},
		{Name: "in_color", Type: attribute.Vec4},
		{Name: "in_additive", Type: attribute.Float},
	})

	vao.Bind()
	vao.Attach(rShader)
	vao.Unbind()

	ibo := buffer.NewIndexBufferObject(6)

	ibo.Bind()

	ibo.SetData(0, []uint16{
		0, 1, 2, 2, 3, 0,
	})

	ibo.Unbind()

	vertexSize := vao.GetVBOFormat("sprites").Size() / 4

	data := make([]float32, maxSprites*vertexSize)

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
	batch.bind(texture.Texture)

	scaleX := float32(batch.scale.X * batch.subscale.X)
	scaleY := float32(batch.scale.Y * batch.subscale.Y)

	posX := float32(batch.position.X)
	posY := float32(batch.position.Y)

	rot := float32(math.Mod(batch.rotation, math.Pi*2))

	u1 := texture.U1
	u2 := texture.U2
	v1 := texture.V1
	v2 := texture.V2

	layer := float32(texture.Layer)

	r := batch.color.X()
	g := batch.color.Y()
	b := batch.color.Z()
	a := batch.color.W()

	add := float32(1)
	if batch.additive {
		add = 0
	}

	idx := batch.currentFloats

	batch.data[idx] = 0
	batch.data[idx+1] = 0
	batch.data[idx+2] = scaleX
	batch.data[idx+3] = scaleY
	batch.data[idx+4] = posX
	batch.data[idx+5] = posY
	batch.data[idx+6] = rot
	batch.data[idx+7] = u1
	batch.data[idx+8] = u2
	batch.data[idx+9] = v1
	batch.data[idx+10] = v2
	batch.data[idx+11] = layer
	batch.data[idx+12] = r
	batch.data[idx+13] = g
	batch.data[idx+14] = b
	batch.data[idx+15] = a
	batch.data[idx+16] = add

	batch.currentFloats += batch.vertexSize
	batch.currentSize++

	if batch.currentSize >= batch.maxSprites {
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

	batch.vao.SetData("sprites", 0, batch.data[:batch.currentFloats])

	batch.ibo.DrawInstanced(0, batch.currentSize)

	batch.currentSize = 0
	batch.currentFloats = 0
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
	batch.bind(texture.Texture)

	scaleX := float32(float64(texture.Width) / 2 * batch.scale.X * batch.subscale.X)
	scaleY := float32(float64(texture.Height) / 2 * batch.scale.Y * batch.subscale.Y)

	posX := float32(batch.position.X)
	posY := float32(batch.position.Y)

	rot := float32(math.Mod(batch.rotation, math.Pi*2))

	u1 := texture.U1
	u2 := texture.U2
	v1 := texture.V1
	v2 := texture.V2

	layer := float32(texture.Layer)

	r := batch.color.X()
	g := batch.color.Y()
	b := batch.color.Z()
	a := batch.color.W()

	add := float32(1)
	if batch.additive {
		add = 0
	}

	idx := batch.currentFloats

	batch.data[idx] = 0
	batch.data[idx+1] = 0
	batch.data[idx+2] = scaleX
	batch.data[idx+3] = scaleY
	batch.data[idx+4] = posX
	batch.data[idx+5] = posY
	batch.data[idx+6] = rot
	batch.data[idx+7] = u1
	batch.data[idx+8] = u2
	batch.data[idx+9] = v1
	batch.data[idx+10] = v2
	batch.data[idx+11] = layer
	batch.data[idx+12] = r
	batch.data[idx+13] = g
	batch.data[idx+14] = b
	batch.data[idx+15] = a
	batch.data[idx+16] = add

	batch.currentFloats += batch.vertexSize
	batch.currentSize++

	if batch.currentSize >= batch.maxSprites {
		batch.Flush()
	}
}

func (batch *SpriteBatch) DrawStObject(position, origin, scale bmath.Vector2d, flip bmath.Vector2d, rotation float64, color mgl32.Vec4, additive bool, texture texture.TextureRegion, storyboard bool) {
	batch.bind(texture.Texture)

	scaleX := float32(scale.X * float64(texture.Width) / 2 * batch.scale.X * batch.subscale.X)
	scaleY := float32(scale.Y * float64(texture.Height) / 2 * batch.scale.Y * batch.subscale.Y)

	posX := float32(position.X + batch.position.X)
	posY := float32(position.Y + batch.position.Y)

	if storyboard {
		posX -= 64
		posY -= 48
	}

	rot := float32(math.Mod(rotation, math.Pi*2))

	u1 := texture.U1
	u2 := texture.U2
	v1 := texture.V1
	v2 := texture.V2

	layer := float32(texture.Layer)

	if flip.X < 0 {
		u1, u2 = u2, u1
	}

	if flip.Y < 0 {
		v1, v2 = v2, v1
	}

	r := color.X() * batch.color.X()
	g := color.Y() * batch.color.Y()
	b := color.Z() * batch.color.Z()
	a := color.W() * batch.color.W()

	add := float32(1)
	if additive {
		add = 0
	}

	idx := batch.currentFloats

	batch.data[idx] = origin.X32()
	batch.data[idx+1] = origin.Y32()
	batch.data[idx+2] = scaleX
	batch.data[idx+3] = scaleY
	batch.data[idx+4] = posX
	batch.data[idx+5] = posY
	batch.data[idx+6] = rot
	batch.data[idx+7] = u1
	batch.data[idx+8] = u2
	batch.data[idx+9] = v1
	batch.data[idx+10] = v2
	batch.data[idx+11] = layer
	batch.data[idx+12] = r
	batch.data[idx+13] = g
	batch.data[idx+14] = b
	batch.data[idx+15] = a
	batch.data[idx+16] = add

	batch.currentFloats += batch.vertexSize
	batch.currentSize++

	if batch.currentSize >= batch.maxSprites {
		batch.Flush()
	}
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
