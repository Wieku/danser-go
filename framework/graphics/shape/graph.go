package shape

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/profiler"
	"strconv"
)

const gVert = `
#version 330

in vec2 base_pos;
in float y_pos;
in float height;

uniform vec2 pos;
uniform float gHeight;
uniform float scale;
uniform int offset;
uniform int layers;

uniform mat4 proj;

flat out int col_id;

void main()
{
    gl_Position = proj * vec4(base_pos * vec2(1, -height*scale) + vec2((gl_InstanceID/layers)+offset, gHeight-y_pos*scale) + pos, 0.0, 1.0);
    col_id = (gl_InstanceID) % layers;
}
`

const gFrag = `
#version 330

flat in int col_id;

uniform vec4 colors[32];

out vec4 color;

void main() {
	color = colors[col_id];/*col_tint*/;
}
`

type label struct {
	name   string
	color2 color.Color
	ctr    *frame.Counter
}

type SteppingGraph struct {
	shader     *shader.RShader
	additive   bool
	projection mgl32.Mat4

	width  int
	height int
	idx    int
	layers int

	maxH float64

	labels []*label

	vao *buffer.VertexArrayObject

	buf []float32

	bgRenderer   *Renderer
	fontRenderer *batch.QuadBatch

	maxVal      float32
	unit        string
	labelsDirty bool
	pos         vector.Vector2f
}

func NewSteppingGraph(x, y, width, height, layers int, maxVal float32, unit string) *SteppingGraph {
	graph := &SteppingGraph{
		shader:       shader.NewRShader(shader.NewSource(gVert, shader.Vertex), shader.NewSource(gFrag, shader.Fragment)),
		projection:   mgl32.Ident4(),
		pos:          vector.NewVec2f(float32(x), float32(y)),
		width:        width,
		height:       height,
		maxVal:       maxVal,
		unit:         unit,
		idx:          0,
		layers:       layers,
		vao:          buffer.NewVertexArrayObject(),
		buf:          make([]float32, 2*layers),
		maxH:         2,
		labelsDirty:  true,
		bgRenderer:   NewRendererSize(18),
		fontRenderer: batch.NewQuadBatchSize(200),
	}

	graph.shader.SetUniform("pos", graph.pos)
	graph.shader.SetUniform("gHeight", float32(height))
	graph.shader.SetUniform("gHeight", float32(height))
	graph.shader.SetUniform("layers", layers)

	graph.vao.AddVBO("default", 4, 0, attribute.Format{
		{Name: "base_pos", Type: attribute.Vec2},
	})

	graph.vao.AddVBO("steps", width*graph.layers, 1, attribute.Format{
		{Name: "y_pos", Type: attribute.Float},
		{Name: "height", Type: attribute.Float},
	})

	graph.vao.Attach(graph.shader)

	graph.vao.SetData("default", 0, []float32{
		0, 0,
		1, 0,
		1, 1,
		0, 1,
	})

	ibo := buffer.NewIndexBufferObject(6)

	ibo.SetData(0, []uint16{
		0, 1, 2, 2, 3, 0,
	})

	graph.vao.AttachIBO(ibo)

	graph.labels = make([]*label, layers)

	for i := 0; i < layers; i++ {
		graph.labels[i] = &label{
			name:   "Unknown",
			color2: color.NewL(1),
			ctr:    frame.NewCounter(),
		}
	}

	return graph
}

func (graph *SteppingGraph) SetLabel(id int, name string, col color.Color) {
	if id >= graph.layers {
		panic("id out of range")
	}

	graph.labels[id].name = name
	graph.labels[id].color2 = col
	graph.labelsDirty = true
}

func (graph *SteppingGraph) SetMaxValue(maxVal float64) {
	graph.maxVal = float32(maxVal)
}

func (graph *SteppingGraph) Draw() {
	if graph.labelsDirty {
		for i := 0; i < graph.layers; i++ {
			graph.shader.SetUniformArr("colors", i, graph.labels[i].color2)
		}

		graph.labelsDirty = false
	}

	thickness := float32(2)

	pX := graph.pos.X
	pY := graph.pos.Y

	graph.bgRenderer.SetCamera(graph.projection)
	graph.bgRenderer.Begin()

	graph.bgRenderer.SetColorM(color.NewLA(0.0, 0.7))
	graph.bgRenderer.DrawQuad(pX, pY, float32(graph.width)+pX, pY, float32(graph.width)+pX, float32(graph.height)+pY, pX, float32(graph.height)+pY)

	graph.bgRenderer.SetColorM(color.NewL(0.8))

	half := thickness / 2

	graph.bgRenderer.DrawLine(-thickness+pX, -half+pY, float32(graph.width)+thickness+pX, -half+pY, thickness)
	graph.bgRenderer.DrawLine(-thickness+pX, float32(graph.height)+half+pY, float32(graph.width)+thickness+pX, float32(graph.height)+half+pY, thickness)

	graph.bgRenderer.DrawLine(-half+pX, pY, -half+pX, float32(graph.height)+pY, thickness)
	graph.bgRenderer.DrawLine(float32(graph.width)+half+pX, pY, float32(graph.width)+half+pX, float32(graph.height)+pY, thickness)

	graph.bgRenderer.End()

	graph.shader.Bind()
	graph.shader.SetUniform("proj", graph.projection)
	graph.shader.SetUniform("scale", float32(graph.height)/graph.maxVal)

	graph.vao.Bind()

	blend.Push()
	blend.Enable()
	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)

	graph.shader.SetUniform("offset", 0)
	if graph.idx == graph.width {
		graph.vao.DrawInstanced(0, graph.width*graph.layers)
	} else {
		graph.vao.DrawInstanced(graph.idx*graph.layers, (graph.width-graph.idx+1)*graph.layers)
		graph.shader.SetUniform("offset", graph.width-graph.idx)
		graph.vao.DrawInstanced(0, graph.idx*graph.layers)
	}

	profiler.AddStat(profiler.SpritesDrawn, 1)

	graph.vao.Unbind()
	graph.shader.Unbind()

	graph.fontRenderer.Begin()
	graph.fontRenderer.SetCamera(graph.projection)

	graph.fontRenderer.SetColor(1, 1, 1, 1)

	font.GetFont("Quicksand Bold").DrawOrigin(graph.fontRenderer, float64(graph.width)-5+float64(pX), 5+float64(pY), vector.TopRight, float64(graph.height)/20, false, mutils.FormatWOZeros(graph.maxVal, 4)+graph.unit)
	font.GetFont("Quicksand Bold").DrawOrigin(graph.fontRenderer, float64(graph.width)-5+float64(pX), float64(graph.height)-5+float64(pY), vector.BottomRight, float64(graph.height)/20, false, "0"+graph.unit)

	font.GetFont("Quicksand Bold").DrawBg(true)
	font.GetFont("Quicksand Bold").SetBgBorderSize(0)
	font.GetFont("Quicksand Bold").SetBgColor(color.NewLA(0, 0.8))

	for i := 0; i < graph.layers; i++ {
		lb := graph.labels[graph.layers-i-1]

		graph.fontRenderer.SetColorM(lb.color2)

		font.GetFont("Quicksand Bold").DrawOrigin(graph.fontRenderer, 5+float64(pX), 5+float64(graph.height)/20*float64(i)+float64(pY), vector.TopLeft, float64(graph.height)/20, true, lb.name+" "+strconv.FormatFloat(lb.ctr.GetAverage(), 'f', 5, 64)+"ms")
	}

	font.GetFont("Quicksand Bold").DrawBg(false)

	graph.fontRenderer.End()
}

func (graph *SteppingGraph) Advance(data ...float64) {
	if len(data) != graph.layers {
		panic("wrong number of layers given")
	}

	var startPos float32

	graph.idx = graph.idx % graph.width

	for i := 0; i < graph.layers; i++ {
		graph.labels[i].ctr.PutSample(data[i])

		graph.buf[i*2] = startPos
		graph.buf[i*2+1] = float32(data[i])
		startPos += float32(data[i])
	}

	graph.vao.SetData("steps", graph.idx*2*graph.layers, graph.buf)

	graph.idx++
}

func (graph *SteppingGraph) SetCamera(camera mgl32.Mat4) {
	graph.projection = camera
}
