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
	"github.com/wieku/danser-go/framework/profiler"
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

	vertexSize    int
	vertices      []float32
	vao           *buffer.VertexArrayObject
	currentSize   int
	currentFloats int
	drawing       bool
	maxSprites    int
	chunkOffset   int
	colorPacked   float32
}

func NewRenderer() *Renderer {
	return NewRendererSize(defaultRendererSize)
}

func NewRendererSize(maxTriangles int) *Renderer {
	rShader := shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	vao := buffer.NewVertexArrayObject()

	vao.AddMappedVBO("default", maxTriangles*3, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec2},
		{Name: "in_color", Type: attribute.ColorPacked},
		{Name: "in_additive", Type: attribute.Float},
	})

	vao.Bind()
	vao.Attach(rShader)
	vao.Unbind()

	vertexSize := vao.GetVBOFormat("default").Size() / 4

	chunk := vao.MapVBO("default", maxTriangles*3*vertexSize)

	return &Renderer{
		shader:      rShader,
		color:       color.NewL(1),
		Projection:  mgl32.Ident4(),
		vertexSize:  vertexSize,
		vertices:    chunk.Data,
		chunkOffset: chunk.Offset,
		vao:         vao,
		maxSprites:  maxTriangles,
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

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)
}

func (renderer *Renderer) Flush() {
	if renderer.currentSize == 0 {
		return
	}

	renderer.vao.UnmapVBO("default", 0, renderer.currentFloats)

	renderer.vao.DrawPart(0, renderer.currentSize)

	profiler.AddStat(profiler.SpritesDrawn, int64(renderer.currentSize))

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

	renderer.vao.Unbind()

	renderer.shader.Unbind()

	blend.Pop()
}

func (renderer *Renderer) SetColor(r, g, b, a float64) {
	renderer.SetColorM(color.NewRGBA(float32(r), float32(g), float32(b), float32(a)))
}

func (renderer *Renderer) SetColorM(color color.Color) {
	renderer.color = color
	renderer.colorPacked = color.PackFloat()
}

func (renderer *Renderer) SetAdditive(additive bool) {
	renderer.additive = additive
}

func (renderer *Renderer) DrawPixelV(position vector.Vector2f, size float32) {
	renderer.DrawPixel(position.X, position.Y, size)
}

func (renderer *Renderer) DrawPixel(x, y, size float32) {
	if size < 0.001 {
		return
	}

	r := size / 2

	renderer.DrawQuad(x-r, y-r, x-r, y+r, x+r, y+r, x+r, y-r)
}

func (renderer *Renderer) DrawLineV(position1, position2 vector.Vector2f, thickness float32) {
	renderer.DrawLine(position1.X, position1.Y, position2.X, position2.Y, thickness)
}

func (renderer *Renderer) DrawLine(x1, y1, x2, y2, thickness float32) {
	if thickness < 0.001 {
		return
	}

	thickHalf := thickness / 2

	dx := x2 - x1
	dy := y2 - y1

	length := math32.Sqrt(dx*dx + dy*dy)

	tx := -dy / length * thickHalf
	ty := dx / length * thickHalf

	renderer.DrawQuad(x1-tx, y1-ty, x1+tx, y1+ty, x2+tx, y2+ty, x2-tx, y2-ty)
}

func (renderer *Renderer) DrawQuadV(p1, p2, p3, p4 vector.Vector2f) {
	renderer.DrawQuad(p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y, p4.X, p4.Y)
}

func (renderer *Renderer) DrawQuad(x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	if renderer.color.A < 0.001 {
		return
	}

	renderer.DrawTriangle(x1, y1, x2, y2, x3, y3)
	renderer.DrawTriangle(x3, y3, x4, y4, x1, y1)
}

func (renderer *Renderer) DrawTriangleV(p1, p2, p3 vector.Vector2f) {
	renderer.DrawTriangle(p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y)
}

func (renderer *Renderer) DrawTriangle(x1, y1, x2, y2, x3, y3 float32) {
	if renderer.color.A < 0.001 {
		return
	}

	add := float32(1)
	if renderer.additive {
		add = 0
	}

	floats := renderer.currentFloats

	renderer.vertices[floats] = x1
	renderer.vertices[floats+1] = y1
	renderer.vertices[floats+2] = renderer.colorPacked
	renderer.vertices[floats+3] = add

	renderer.vertices[floats+4] = x2
	renderer.vertices[floats+5] = y2
	renderer.vertices[floats+6] = renderer.colorPacked
	renderer.vertices[floats+7] = add

	renderer.vertices[floats+8] = x3
	renderer.vertices[floats+9] = y3
	renderer.vertices[floats+10] = renderer.colorPacked
	renderer.vertices[floats+11] = add

	renderer.currentFloats += 12
	renderer.currentSize += 3

	if renderer.currentSize >= renderer.maxSprites*3 {
		renderer.Flush()
	}
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
	if math32.Abs(progress) < 0.001 || renderer.color.A < 0.001 {
		return
	}

	direction := float32(1.0)
	if progress < 0 {
		direction = -1
	}

	progress = math32.Abs(progress)

	partRadians := 2 * math32.Pi / float32(sections)
	targetRadians := 2 * math32.Pi * progress

	x := position.X
	y := position.Y

	cx := math32.Cos(-math32.Pi/2)*radius + x
	cy := math32.Sin(-math32.Pi/2)*radius + y

	for r := float32(0.0); r < targetRadians; r += partRadians {
		rads := min(targetRadians, r+partRadians)*direction - math32.Pi/2

		cx1 := math32.Cos(rads)*radius + x
		cy1 := math32.Sin(rads)*radius + y

		renderer.DrawTriangle(x, y, cx, cy, cx1, cy1)

		cx, cy = cx1, cy1
	}

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
