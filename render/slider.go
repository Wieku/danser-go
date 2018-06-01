package render

import (
	"danser/bmath"
	"math"
	"github.com/faiface/glhf"
	"log"
	_ "image/png"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var sliderShader *glhf.Shader = nil
var fboShader *glhf.Shader
var fboSlice *glhf.VertexSlice
var sliderVertexFormat glhf.AttrFormat
var cam mgl32.Mat4
var fbo *glhf.Frame

var CS float64

func SetupSlider() {

	sliderVertexFormat = glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "center", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}
	var err error

	sliderShader, err = glhf.NewShader(sliderVertexFormat, glhf.AttrFormat{{Name: "col_border", Type: glhf.Vec4}, {Name: "tex", Type: glhf.Int}, {Name: "proj", Type: glhf.Mat4}, {Name: "trans", Type: glhf.Mat4}}, sliderVec, sliderFrag)
	if err != nil {
		log.Println(err)
	}
	fboShader, err = glhf.NewShader(glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}, glhf.AttrFormat{{Name: "tex", Type: glhf.Int}, {Name: "alpha", Type: glhf.Float}}, fboVec, fboFrag)


	if err != nil {
		log.Println(err)
	}

	fbo = glhf.NewFrame(1920, 1080, true, false)

	fbo.Begin()
	var depthRenderBuffer uint32
	gl.GenRenderbuffers(1, &depthRenderBuffer)
	gl.BindRenderbuffer( gl.RENDERBUFFER, depthRenderBuffer)
	gl.RenderbufferStorage( gl.RENDERBUFFER, gl.DEPTH_COMPONENT, 1920, 1080)
	gl.FramebufferRenderbuffer( gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthRenderBuffer)
	fbo.End()

	fboSlice = glhf.MakeVertexSlice(fboShader, 6, 6)
	fboSlice.Begin()
	fboSlice.SetVertexData([]float32{
		-1, -1, 0, 0, 0,
		1, -1, 0, 1, 0,
		-1, 1, 0, 0, 1,
		1, -1, 0, 1, 0,
		1, 1, 0, 1, 1,
		-1, 1, 0, 0, 1,
	})
	fboSlice.End()

}

type SliderRenderer struct {}

func NewSliderRenderer() *SliderRenderer {
	return &SliderRenderer{}
}

func (sr *SliderRenderer) Begin() {

	gl.Disable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	fbo.Begin()

	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	sliderShader.Begin()
	gl.ActiveTexture(gl.TEXTURE0)
	SliderGradient.Begin()
	sliderShader.SetUniformAttr(1, int32(0))
	sliderShader.SetUniformAttr(2, cam)
}

func (sr *SliderRenderer) SetColor(color mgl32.Vec4) {
	sliderShader.SetUniformAttr(0, color)
}

func (sr *SliderRenderer) SetScale(scale float64) {
	sliderShader.SetUniformAttr(3, mgl32.Scale3D(float32(scale), float32(scale), 1))
}

func (sr *SliderRenderer) EndAndRender() {

	SliderGradient.End()
	sliderShader.End()
	fbo.End()
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	//gl.BlendFunc(gl.ONE_MINUS_DST_ALPHA, gl.DST_ALPHA)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.ActiveTexture(gl.TEXTURE0)
	fbo.Texture().Begin()

	fboShader.Begin()
	fboShader.SetUniformAttr(0, int32(0))
	fboShader.SetUniformAttr(1, float32(1))
	fboSlice.Begin()
	fboSlice.Draw()
	fboSlice.End()
	fboShader.End()

	fbo.Texture().End()
	/*gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	 Blit to default framebuffer *//*
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo.ID())
	gl.BlitFramebuffer(0, 0, 1920, 1080, 0, 0, 1920, 1080, gl.COLOR_BUFFER_BIT, gl.NEAREST)
	*/
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	//gl.DepthFunc(gl.LESS)
	//gl.DepthMask(false)
	//gl.Disable(gl.DEPTH_TEST)
}

func (self *SliderRenderer) SetCamera(camera mgl32.Mat4) {
	cam = camera
	sliderShader.SetUniformAttr(2, cam)
}

func (self *SliderRenderer) GetShape(curve []bmath.Vector2d) (*glhf.VertexSlice, int) {
	return createMesh(curve), 30
}

func createMesh(curve []bmath.Vector2d) *glhf.VertexSlice {
	var slice *glhf.VertexSlice

	vecr := make([]float32, 0)
	num := 0
	for _, v := range curve {
		tab := createCircle(v.X, v.Y, 64*CS, 30)
		for j := range tab {
			if j >= 2 {
				p1, p2, p3 := tab[j-1], tab[j], tab[0]
				vecr = append(vecr, float32(p1.X), float32(p1.Y), 1.0, float32(p3.X), float32(p3.Y), 0.0, 0.0, 0.0, float32(p2.X), float32(p2.Y), 1.0, float32(p3.X), float32(p3.Y), 0.0, 0.0, 0.0, float32(p3.X), float32(p3.Y), 0.0, float32(p3.X), float32(p3.Y), 0.0, 1.0, 0.0)
			}

		}
		num += len(tab)
	}

	slice = glhf.MakeVertexSlice(sliderShader, len(vecr)/8, len(vecr))
	slice.Begin()
	slice.SetVertexData(vecr)
	slice.End()

	return slice
}

func createCircle(x, y, radius float64, segments int) ([]bmath.Vector2d) {

	points := []bmath.Vector2d{bmath.NewVec2d(x, y)}

	for i:=0; i < segments; i++ {
		points = append(points, bmath.NewVec2dRad(float64(i)/float64(segments)*2*math.Pi, radius).AddS(x, y))
	}

	points = append(points, points[1])

	return points
}

const sliderFrag = `
#version 330

uniform sampler2D tex;
//uniform vec2 tex_size;
//uniform vec3 col_tint;
uniform vec4 col_border;

in vec2 tex_coord;
out vec4 color;
void main()
{
    vec4 in_color = texture2D(tex, tex_coord);
    //float blend_factor = in_color.r-in_color.b;
	//vec4 col_tint = vec4(1, 1, 1, 1f);
   // vec4 new_color = vec4(mix(in_color.xyz*col_border.xyz,col_tint.xyz, blend_factor),in_color.a);
	color = in_color*col_border;
}
`

const sliderVec = `
#version 330

in vec3 in_position;
in vec3 center;
in vec2 in_tex_coord;

uniform mat4 proj;
uniform mat4 trans;

out vec2 tex_coord;
void main()
{
    gl_Position = proj * ((trans * vec4(in_position-center, 1))+vec4(center, 0));
    tex_coord = in_tex_coord;
}
`

const fboFrag = `
#version 330

uniform sampler2D tex;
uniform float alpha;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture2D(tex, tex_coord);
	//if (in_color.xyz == vec3(0,0,0)) {
	//	discard;
	//}
	color = vec4(in_color.xyz, in_color.a*alpha);
}
`

const fboVec = `
#version 330

in vec3 in_position;
in vec2 in_tex_coord;

out vec2 tex_coord;
void main()
{
    gl_Position = vec4(in_position, 1);
    tex_coord = in_tex_coord;
}
`