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
	"github.com/wieku/danser-go/framework/statistic"
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
	ibo           *buffer.IndexBufferObject
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

	rShader := shader.NewRShader(shader.NewSource(string(vert), shader.Vertex), shader.NewSource(string(frag), shader.Fragment))

	vao := buffer.NewVertexArrayObject()

	vao.AddVBO("default", 4, 0, attribute.Format{
		{Name: "base_pos", Type: attribute.Vec2},
		{Name: "base_uv", Type: attribute.Vec2},
	})

	vao.SetData("default", 0, []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	})

	vao.AddMappedVBO("quads", maxSprites, 1, attribute.Format{
		{Name: "in_origin", Type: attribute.Vec2Packed},
		{Name: "in_scale", Type: attribute.Vec2},
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_rotation", Type: attribute.Float},
		{Name: "in_u", Type: attribute.Vec2Packed},
		{Name: "in_v", Type: attribute.Vec2Packed},
		{Name: "in_layer", Type: attribute.Float},
		{Name: "in_color", Type: attribute.ColorPacked},
		{Name: "in_additive", Type: attribute.Float},
		{Name: "in_msdf", Type: attribute.Float},
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
		ibo:         ibo,
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
	batch.ibo.Bind()

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

	//batch.vao.SetData("quads", 0, batch.data[:batch.currentFloats])
	batch.vao.UnmapVBO("quads", 0, batch.currentFloats)

	batch.ibo.DrawInstanced(batch.chunkOffset/batch.vertexSize, batch.currentSize)

	statistic.Add(statistic.SpritesDrawn, int64(batch.currentSize))

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

	batch.ibo.Unbind()
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
	batch.drawTextureBase(texture, false, false)
}

func (batch *QuadBatch) DrawUnitMSDF(texture texture.TextureRegion) {
	batch.drawTextureBase(texture, false, true)
}

func (batch *QuadBatch) DrawTexture(texture texture.TextureRegion) {
	batch.drawTextureBase(texture, true, false)
}

func (batch *QuadBatch) DrawTextureMSDF(texture texture.TextureRegion) {
	batch.drawTextureBase(texture, true, true)
}

func (batch *QuadBatch) drawTextureBase(texture texture.TextureRegion, useTextureSize, msdf bool) {
	if texture.Texture == nil || batch.color.A < 0.001 {
		return
	}

	batch.bind(texture.Texture)

	scaleX := float32(batch.scale.X * batch.subscale.X)
	scaleY := float32(batch.scale.Y * batch.subscale.Y)

	if useTextureSize {
		scaleX *= float32(texture.Width) / 2
		scaleY *= float32(texture.Height) / 2
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

	msdfI := float32(0)
	if msdf {
		msdfI = 1
	}

	idx := batch.currentFloats

	batch.data[idx] = packUV(0.5, 0.5)
	batch.data[idx+1] = scaleX
	batch.data[idx+2] = scaleY
	batch.data[idx+3] = posX
	batch.data[idx+4] = posY
	batch.data[idx+5] = rot
	batch.data[idx+6] = packUV(u1, u2)
	batch.data[idx+7] = packUV(v1, v2)
	batch.data[idx+8] = layer
	batch.data[idx+9] = batch.color.PackFloat()
	batch.data[idx+10] = add
	batch.data[idx+11] = msdfI

	batch.currentFloats += batch.vertexSize
	batch.currentSize++

	if batch.currentSize >= batch.maxSprites {
		batch.Flush()
	}
}

func (batch *QuadBatch) DrawStObject(position, origin, scale vector.Vector2d, flipX, flipY bool, rotation float64, color mgl32.Vec4, additive bool, texture texture.TextureRegion) {
	if texture.Texture == nil || color.W()*batch.color.A < 0.001 {
		return
	}

	batch.bind(texture.Texture)

	scaleX := float32(scale.X * float64(texture.Width) / 2 * batch.scale.X * batch.subscale.X)
	scaleY := float32(scale.Y * float64(texture.Height) / 2 * batch.scale.Y * batch.subscale.Y)

	posX := float32(position.X + batch.position.X)
	posY := float32(position.Y + batch.position.Y)

	rot := float32(math.Mod(rotation, math.Pi*2))

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

	r := color.X() * batch.color.R
	g := color.Y() * batch.color.G
	b := color.Z() * batch.color.B
	a := color.W() * batch.color.A

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
	batch.data[idx+6] = packUV(u1, u2)
	batch.data[idx+7] = packUV(v1, v2)
	batch.data[idx+8] = layer
	batch.data[idx+9] = pack(r, g, b, a)
	batch.data[idx+10] = add
	batch.data[idx+11] = 0

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

func pack(r, g, b, a float32) float32 {
	rI := uint32(r * 255)
	gI := uint32(g * 255)
	bI := uint32(b * 255)
	aI := uint32(a * 255)

	return math.Float32frombits(aI<<24 | bI<<16 | gI<<8 | rI)
}

func packUV(c1, c2 float32) float32 {
	c1I := uint32(c1 * 0xffff)
	c2I := uint32(c2 * 0xffff)

	return math.Float32frombits(c2I<<16 | c1I)
}
