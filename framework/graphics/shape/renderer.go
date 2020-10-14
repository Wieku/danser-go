package shape

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/statistic"
)

const defaultRendererSize = 2000

const vert = `
#version 330

in vec2 in_position;
in vec4 in_color;
in float in_additive;

uniform mat4 proj;

out vec4 col_tint;
out float additive;

void main()
{
    gl_Position = proj * vec4(in_position, 0.0, 1.0);
    col_tint = in_color;
    additive = in_additive;
}
`

const frag = `
#version 330

in vec4 col_tint;
in float additive;

out vec4 color;

void main() {
	color = col_tint;
	color.rgb *= color.a;
	color.a *= additive;
}
`

type Renderer struct {
	shader     *shader.RShader
	additive   bool
	color      color.Color
	Projection mgl32.Mat4

	vertexSize int
	vertices   []float32
	//indexes       []float32
	vao *buffer.VertexArrayObject
	//ibo           *buffer.IndexBufferObject
	currentSize   int
	currentFloats int
	drawing       bool
	maxSprites    int
	chunkOffset   int
}

func NewRenderer() *Renderer {
	return NewRendererSize(defaultRendererSize)
}

func NewRendererSize(maxSprites int) *Renderer {
	//if maxSprites*6 > 0xFFFF {
	//	panic(fmt.Sprintf("Renderer size is too big, maximum sprites allowed: 10922, given: %d", maxSprites))
	//}

	rShader := shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	vao := buffer.NewVertexArrayObject()

	vao.AddMappedVBO("default", maxSprites*3, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_color", Type: attribute.ColorPacked},
		{Name: "in_additive", Type: attribute.Float},
	})

	vao.Bind()
	vao.Attach(rShader)
	vao.Unbind()

	//ibo := buffer.NewIndexBufferObject(maxSprites*6)

	vertexSize := vao.GetVBOFormat("default").Size() / 4

	chunk := vao.MapVBO("default", maxSprites*3*vertexSize)

	return &Renderer{
		shader:      rShader,
		color:       color.NewL(1),
		Projection:  mgl32.Ident4(),
		vertexSize:  vertexSize,
		vertices:    chunk.Data,
		chunkOffset: chunk.Offset,
		vao:         vao,
		//ibo:         ibo,
		maxSprites: maxSprites,
	}
}

func (renderer *Renderer) Begin() {
	if renderer.drawing {
		panic("Batching has already begun")
	}

	renderer.drawing = true

	renderer.shader.Bind()
	renderer.shader.SetUniform("proj", renderer.Projection)

	renderer.vao.Bind()
	//renderer.ibo.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)
}

func (renderer *Renderer) Flush() {
	if renderer.currentSize == 0 {
		return
	}

	//renderer.vao.SetData("sprites", 0, renderer.data[:renderer.currentFloats])
	renderer.vao.UnmapVBO("default", 0, renderer.currentFloats)

	renderer.vao.DrawPart(0, renderer.currentSize)

	//renderer.ibo.DrawInstanced(renderer.chunkOffset/renderer.vertexSize, renderer.currentSize)

	statistic.Add(statistic.SpritesDrawn, int64(renderer.currentSize))

	nextChunk := renderer.vao.MapVBO("default", renderer.maxSprites*3*renderer.vertexSize)

	renderer.vertices = nextChunk.Data
	renderer.chunkOffset = nextChunk.Offset

	renderer.currentSize = 0
	renderer.currentFloats = 0
}

func (renderer *Renderer) End() {
	if !renderer.drawing {
		panic("Batching has already ended")
	}

	renderer.drawing = false

	renderer.Flush()

	//renderer.ibo.Unbind()
	renderer.vao.Unbind()

	renderer.shader.Unbind()

	blend.Pop()
}

func (renderer *Renderer) SetColor(r, g, b, a float64) {
	renderer.color = color.NewRGBA(float32(r), float32(g), float32(b), float32(a))
}

func (renderer *Renderer) SetColorM(color color.Color) {
	renderer.color = color
}

func (renderer *Renderer) SetAdditive(additive bool) {
	renderer.additive = additive
}

func (renderer *Renderer) DrawCircle(position vector.Vector2f, radius float32) {
	renderer.DrawCircleProgress(position, radius, 1.0)
}

func (renderer *Renderer) DrawCircleS(position vector.Vector2f, radius float32, sections int) {
	renderer.DrawCircleProgressS(position, radius, sections, 1.0)
}

func (renderer *Renderer) DrawCircleProgress(position vector.Vector2f, radius float32, progress float32) {
	renderer.DrawCircleProgressS(position, radius, 6*int(math32.Sqrt(radius)), progress)
}

func (renderer *Renderer) DrawCircleProgressS(position vector.Vector2f, radius float32, sections int, progress float32) {
	if progress < 0.001 {
		return
	}

	add := float32(1)
	if renderer.additive {
		add = 0
	}

	color := renderer.color.PackFloat()

	partRadius := 2 * math32.Pi / float32(sections)
	targetRadius := 2 * math32.Pi * progress

	floats := renderer.currentFloats

	cx := math32.Cos(-math32.Pi / 2)
	cy := math32.Sin(-math32.Pi / 2)

	for r := float32(0.0); r < targetRadius; r += partRadius {
		renderer.vertices[floats] = position.X
		renderer.vertices[floats+1] = position.Y
		renderer.vertices[floats+2] = color
		renderer.vertices[floats+3] = add

		renderer.vertices[floats+4] = cx*radius + position.X
		renderer.vertices[floats+5] = cy*radius + position.Y
		renderer.vertices[floats+6] = color
		renderer.vertices[floats+7] = add

		rads := math32.Min(targetRadius, r+partRadius) - math32.Pi/2

		cx = math32.Cos(rads)
		cy = math32.Sin(rads)

		renderer.vertices[floats+8] = cx*radius + position.X
		renderer.vertices[floats+9] = cy*radius + position.Y
		renderer.vertices[floats+10] = color
		renderer.vertices[floats+11] = add

		renderer.currentSize += 3
		floats += 12
	}

	renderer.currentFloats = floats

	if renderer.currentSize >= renderer.maxSprites*3 {
		renderer.Flush()
	}
}

func (renderer *Renderer) SetCamera(camera mgl32.Mat4) {
	if renderer.Projection == camera {
		return
	}

	if renderer.drawing {
		renderer.Flush()
	}

	renderer.Projection = camera
	if renderer.drawing {
		renderer.shader.SetUniform("proj", renderer.Projection)
	}
}
