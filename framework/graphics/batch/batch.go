package batch

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/profiler"
	"math"
)

const defaultBatchSize = 2000

type QuadBatch struct {
	shader     *shader.RShader
	additive   bool
	color      color2.Color
	Projection mgl32.Mat4
	position   vector.Vector2d
	scale      vector.Vector2d
	subscale   vector.Vector2d
	rotation   float64

	transform mgl32.Mat4
	texture   texture.Texture

	vertexSize    int
	data          []float32
	vao           *buffer.VertexArrayObject
	currentSize   int
	currentFloats int
	drawing       bool
	maxSprites    int
	chunkOffset   int
}

func NewQuadBatch() *QuadBatch {
	return NewQuadBatchSize(defaultBatchSize)
}

func NewQuadBatchSize(maxSprites int) *QuadBatch {
	return newQuadBatchSize(maxSprites, false)
}

func NewQuadBatchPersistent() *QuadBatch {
	return NewQuadBatchSizePersistent(defaultBatchSize)
}

func NewQuadBatchSizePersistent(maxSprites int) *QuadBatch {
	return newQuadBatchSize(maxSprites, true)
}

func newQuadBatchSize(maxSprites int, persistent bool) *QuadBatch {
	if maxSprites*6 > 0xFFFF {
		panic(fmt.Sprintf("QuadBatch size is too big, maximum quads allowed: 10922, given: %d", maxSprites))
	}

	vert, err := assets.GetString("assets/shaders/sprite.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/sprite.fsh")
	if err != nil {
		panic(err)
	}

	rShader := shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	vao := buffer.NewVertexArrayObject()

	vao.AddVBO("default", 4, 0, attribute.Format{
		{Name: "base_pos", Type: attribute.Vec2},
		{Name: "base_uv", Type: attribute.Vec2},
	})

	format := attribute.Format{
		{Name: "in_origin", Type: attribute.Vec2Packed},
		{Name: "in_scale", Type: attribute.Vec2},
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_rotation", Type: attribute.Float},
		{Name: "in_u", Type: attribute.Vec2},
		{Name: "in_v", Type: attribute.Vec2},
		{Name: "in_layer", Type: attribute.Float},
		{Name: "in_color", Type: attribute.ColorPacked},
		{Name: "in_additive", Type: attribute.Float},
	}

	if persistent {
		vao.AddPersistentVBO("quads", maxSprites*100, 1, format)
	} else {
		vao.AddMappedVBO("quads", maxSprites, 1, format)
	}

	vao.SetData("default", 0, []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	})

	vao.Attach(rShader)

	ibo := buffer.NewIndexBufferObject(6)

	ibo.SetData(0, []uint16{
		0, 1, 2, 2, 3, 0,
	})

	vao.AttachIBO(ibo)

	vertexSize := vao.GetVBOFormat("quads").Size() / 4

	chunk := vao.MapVBO("quads", maxSprites*vertexSize)

	return &QuadBatch{
		shader:      rShader,
		color:       color2.NewL(1),
		Projection:  mgl32.Ident4(),
		scale:       vector.NewVec2d(1, 1),
		subscale:    vector.NewVec2d(1, 1),
		transform:   mgl32.Ident4(),
		vertexSize:  vertexSize,
		data:        chunk.Data,
		chunkOffset: chunk.Offset,
		vao:         vao,
		maxSprites:  maxSprites,
	}
}

func (batch *QuadBatch) Begin() {
	if batch.drawing {
		panic("Batching has already begun")
	}

	batch.drawing = true

	batch.shader.Bind()
	batch.shader.SetUniform("proj", batch.Projection)

	batch.vao.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)
}

func (batch *QuadBatch) bind(texture texture.Texture) {
	if batch.texture != nil {
		if batch.texture == texture {
			return
		}

		batch.Flush()
	}

	batch.texture = texture
}

func (batch *QuadBatch) Flush() {
	if batch.currentSize == 0 {
		return
	}

	if batch.texture.GetLocation() == 0 {
		batch.texture.Bind(0)
	}

	batch.shader.SetUniform("tex", int32(batch.texture.GetLocation()))

	batch.vao.UnmapVBO("quads", 0, batch.currentFloats)

	batch.vao.DrawInstanced(batch.chunkOffset/batch.vertexSize, batch.currentSize)

	profiler.AddStat(profiler.SpritesDrawn, int64(batch.currentSize))

	nextChunk := batch.vao.MapVBO("quads", batch.maxSprites*batch.vertexSize)

	batch.data = nextChunk.Data
	batch.chunkOffset = nextChunk.Offset

	batch.currentSize = 0
	batch.currentFloats = 0
}

func (batch *QuadBatch) End() {
	if !batch.drawing {
		panic("Batching has already ended")
	}

	batch.drawing = false

	batch.Flush()

	batch.vao.Unbind()

	batch.shader.Unbind()

	blend.Pop()
}

func (batch *QuadBatch) SetColor(r, g, b, a float64) {
	batch.color.R = float32(r)
	batch.color.G = float32(g)
	batch.color.B = float32(b)
	batch.color.A = float32(a)
}

func (batch *QuadBatch) SetColor32(r, g, b, a float32) {
	batch.color.R = r
	batch.color.G = g
	batch.color.B = b
	batch.color.A = a
}

func (batch *QuadBatch) SetColorM(color color2.Color) {
	batch.color = color
}

func (batch *QuadBatch) SetTranslation(vec vector.Vector2d) {
	batch.position = vec
}

func (batch *QuadBatch) SetRotation(rad float64) {
	batch.rotation = rad
}

func (batch *QuadBatch) GetRotation() float64 {
	return batch.rotation
}

func (batch *QuadBatch) SetScale(scaleX, scaleY float64) {
	batch.scale = vector.NewVec2d(scaleX, scaleY)
}

func (batch *QuadBatch) GetScale() vector.Vector2d {
	return batch.scale
}

func (batch *QuadBatch) SetSubScale(scaleX, scaleY float64) {
	batch.subscale = vector.NewVec2d(scaleX, scaleY)
}

func (batch *QuadBatch) GetSubScale() vector.Vector2d {
	return batch.subscale
}

func (batch *QuadBatch) ResetTransform() {
	batch.scale = vector.NewVec2d(1, 1)
	batch.subscale = vector.NewVec2d(1, 1)
	batch.position = vector.NewVec2d(0, 0)
	batch.rotation = 0
}

func (batch *QuadBatch) SetAdditive(additive bool) {
	batch.additive = additive
}

func (batch *QuadBatch) DrawUnit(texture texture.TextureRegion) {
	batch.drawTextureBase(texture, false)
}

func (batch *QuadBatch) DrawTexture(texture texture.TextureRegion) {
	batch.drawTextureBase(texture, true)
}

func (batch *QuadBatch) drawTextureBase(texture texture.TextureRegion, useTextureSize bool) {
	if texture.Texture == nil || batch.color.A < 0.001 {
		return
	}

	batch.bind(texture.Texture)

	scaleX := float32(batch.scale.X * batch.subscale.X)
	scaleY := float32(batch.scale.Y * batch.subscale.Y)

	if useTextureSize {
		scaleX *= texture.Width / 2
		scaleY *= texture.Height / 2
	}

	posX := float32(batch.position.X)
	posY := float32(batch.position.Y)

	rot := float32(math.Mod(batch.rotation, math.Pi*2))

	u1 := texture.U1
	u2 := texture.U2
	v1 := texture.V1
	v2 := texture.V2

	layer := float32(texture.Layer)

	add := float32(1)
	if batch.additive {
		add = 0
	}

	idx := batch.currentFloats

	batch.data[idx] = packUV(0.5, 0.5)
	batch.data[idx+1] = scaleX
	batch.data[idx+2] = scaleY
	batch.data[idx+3] = posX
	batch.data[idx+4] = posY
	batch.data[idx+5] = rot
	batch.data[idx+6] = u1
	batch.data[idx+7] = u2
	batch.data[idx+8] = v1
	batch.data[idx+9] = v2
	batch.data[idx+10] = layer
	batch.data[idx+11] = batch.color.PackFloat()
	batch.data[idx+12] = add

	batch.currentFloats += batch.vertexSize
	batch.currentSize++

	if batch.currentSize >= batch.maxSprites {
		batch.Flush()
	}
}

func (batch *QuadBatch) DrawStObject(position, origin, scale vector.Vector2d, flipX, flipY bool, rotation float64, color color2.Color, additive bool, texture texture.TextureRegion) {
	if texture.Texture == nil || color.A*batch.color.A < 0.001 {
		return
	}

	batch.bind(texture.Texture)

	scaleX := float32(scale.X * float64(texture.Width) / 2 * batch.scale.X * batch.subscale.X)
	scaleY := float32(scale.Y * float64(texture.Height) / 2 * batch.scale.Y * batch.subscale.Y)

	posX := float32(position.X + batch.position.X)
	posY := float32(position.Y + batch.position.Y)

	rot := float32(math.Mod(rotation+batch.rotation, math.Pi*2))

	u1 := texture.U1
	u2 := texture.U2
	v1 := texture.V1
	v2 := texture.V2

	layer := float32(texture.Layer)

	if flipX {
		u1, u2 = u2, u1
	}

	if flipY {
		v1, v2 = v2, v1
	}

	r := color.R * batch.color.R
	g := color.G * batch.color.G
	b := color.B * batch.color.B
	a := color.A * batch.color.A

	add := float32(1)
	if additive || batch.additive {
		add = 0
	}

	idx := batch.currentFloats

	batch.data[idx] = packUV(origin.X32()*0.5+0.5, origin.Y32()*0.5+0.5)
	batch.data[idx+1] = scaleX
	batch.data[idx+2] = scaleY
	batch.data[idx+3] = posX
	batch.data[idx+4] = posY
	batch.data[idx+5] = rot
	batch.data[idx+6] = u1
	batch.data[idx+7] = u2
	batch.data[idx+8] = v1
	batch.data[idx+9] = v2
	batch.data[idx+10] = layer
	batch.data[idx+11] = color2.PackFloat(r, g, b, a)
	batch.data[idx+12] = add

	batch.currentFloats += batch.vertexSize
	batch.currentSize++

	if batch.currentSize >= batch.maxSprites {
		batch.Flush()
	}
}

func (batch *QuadBatch) SetCamera(camera mgl32.Mat4) {
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

func packUV(c1, c2 float32) float32 {
	c1I := uint32(c1 * 0xffff)
	c2I := uint32(c2 * 0xffff)

	return math.Float32frombits(c2I<<16 | c1I)
}
