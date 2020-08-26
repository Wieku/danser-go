package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/glhf"
)

var fxshader *glhf.Shader = nil

func setupFx() {
	fxVertexFormat := glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
	}

	fxUniformFormat := glhf.AttrFormat{
		{Name: "in_color", Type: glhf.Vec4},
		{Name: "transform", Type: glhf.Mat4},
	}
	var err error
	fxshader, err = glhf.NewShader(fxVertexFormat, fxUniformFormat, fxvertex, fxfragment)

	if err != nil {
		panic("Fx: " + err.Error())
	}

}

type FxBatch struct {
	color     mgl32.Vec4
	transform mgl32.Mat4
}

func NewFxBatch() *FxBatch {
	if fxshader == nil {
		setupFx()
	}
	return &FxBatch{mgl32.Vec4{1, 1, 1, 1}, mgl32.Ident4()}
}

func (batch *FxBatch) Begin() {
	fxshader.Begin()
	fxshader.SetUniformAttr(0, batch.color)
	fxshader.SetUniformAttr(1, batch.transform)
}

func (batch *FxBatch) CreateVao(length int) *glhf.VertexSlice {
	return glhf.MakeVertexSlice(fxshader, length, length)
}

func (batch *FxBatch) SetColor(r, g, b, a float64) {
	batch.color = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
	fxshader.SetUniformAttr(0, batch.color)
}

func (batch *FxBatch) SetColorM(color mgl32.Vec4) {
	batch.color = color
	fxshader.SetUniformAttr(0, batch.color)
}

func (batch *FxBatch) ResetTransform() {
	batch.transform = mgl32.Ident4()
	fxshader.SetUniformAttr(1, batch.transform)
}

func (batch *FxBatch) End() {
	fxshader.End()
}

func (batch *FxBatch) SetTransform(dz mgl32.Mat4) {
	batch.transform = dz
	fxshader.SetUniformAttr(1, dz)
}

const fxvertex = `
#version 330

in vec3 in_position;

uniform mat4 transform; 

void main()
{
    gl_Position = transform * vec4(in_position, 1);
}
`

const fxfragment = `
#version 330

uniform vec4 in_color;
out vec4 color;

void main()
{
	color = in_color;
}
`
