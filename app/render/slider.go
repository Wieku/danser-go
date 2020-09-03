package render

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/math/math32"
	_ "image/png"
	"io/ioutil"
	"log"
)

var sliderShader *shader.RShader = nil
var fboShader *shader.Shader
var fboSlice *buffer.VertexSlice
var sliderVertexFormat shader.AttrFormat
var fbo *buffer.Framebuffer
var fboUnit int32
var CS float64
var unitCircle []float32

func SetupSlider() {

	sliderVertexFormat = shader.AttrFormat{
		{Name: "in_position", Type: shader.Vec3},
		{Name: "center", Type: shader.Vec3},
		//{Name: "in_tex_coord", Type: shader.Vec2},
	}
	var err error

	svert, _ := ioutil.ReadFile("assets/shaders/slider.vsh")
	sfrag, _ := ioutil.ReadFile("assets/shaders/slider.fsh")
	sliderShader = shader.NewRShader(shader.NewSource(string(svert), shader.Vertex), shader.NewSource(string(sfrag), shader.Fragment))
	//if err != nil {
	//	log.Println(err)
	//}

	//shader.AttrFormat{
	//	{Name: "col_border", Type: shader.Vec4},
	//	{Name: "proj", Type: shader.Mat4},
	//	{Name: "trans", Type: shader.Mat4},
	//	{Name: "col_border1", Type: shader.Vec4},
	//	{Name: "distort", Type: shader.Mat4}}

	fvert, _ := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	ffrag, _ := ioutil.ReadFile("assets/shaders/fbopass.fsh")
	fboShader, err = shader.NewShader(shader.AttrFormat{
		{Name: "in_position", Type: shader.Vec3},
		{Name: "in_tex_coord", Type: shader.Vec2},
	}, shader.AttrFormat{{Name: "tex", Type: shader.Int}}, string(fvert), string(ffrag))

	if err != nil {
		log.Println("FboPass: " + err.Error())
	}

	fbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, true)
	fbo.Texture().Bind(29)
	fboUnit = 29

	fboSlice = buffer.MakeVertexSlice(fboShader, 6, 6)
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

type SliderRenderer struct {
	camera mgl32.Mat4
	scale  float64
}

func NewSliderRenderer() *SliderRenderer {
	return &SliderRenderer{}
}

func (sr *SliderRenderer) Begin() {

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	fbo.Begin()

	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	sliderShader.Bind()

	sliderShader.SetUniform("proj", sr.camera)
}

func (sr *SliderRenderer) BeginShader() {
	sliderShader.Bind()

	sliderShader.SetUniform("proj", sr.camera)
	sliderShader.SetUniform("scale", float32(sr.scale*CS))
}

func (sr *SliderRenderer) EndShader() {
	sliderShader.Unbind()
}

func (sr *SliderRenderer) SetColor(color mgl32.Vec4, prev mgl32.Vec4) {
	sliderShader.SetUniform("col_border", color)
	if settings.Objects.EnableCustomSliderBorderGradientOffset {
		sliderShader.SetUniform("col_border1", utils.GetColorShifted(color, settings.Objects.SliderBorderGradientOffset))
	} else {
		sliderShader.SetUniform("col_border1", prev)
	}
}

func (sr *SliderRenderer) SetScale(scale float64) {
	sr.scale = scale
	sliderShader.SetUniform("scale", float32(scale*CS))
}

func (sr *SliderRenderer) SetDistort(scaleX, scaleY float64) {
	sliderShader.SetUniform("distort", mgl32.Translate3D(-1, 1, 0).Mul4(mgl32.Scale3D(float32(scaleX), float32(scaleY), 1)).Mul4(mgl32.Translate3D(1, -1, 0)))
}

func (sr *SliderRenderer) GetCamera() mgl32.Mat4 {
	return sr.camera
}

func (sr *SliderRenderer) EndAndRender() {

	sliderShader.Unbind()
	fbo.End()
	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Enable(gl.BLEND)

	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	fboShader.Begin()
	fboShader.SetUniformAttr(0, int32(fboUnit))
	fboSlice.BeginDraw()
	fboSlice.Draw()
	fboSlice.EndDraw()
	fboShader.End()
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
}

func (self *SliderRenderer) SetCamera(camera mgl32.Mat4) {
	self.camera = camera
	sliderShader.SetUniform("proj", self.camera)
}

func (self *SliderRenderer) GetShape(curve []bmath.Vector2f) ([]float32, int) {
	return createMesh(curve), int(settings.Objects.SliderLOD)
}

func (self *SliderRenderer) UploadMesh(curve []bmath.Vector2f) *buffer.VertexArrayObject {
	if len(unitCircle) == 0 {
		createCircle(int(settings.Objects.SliderLOD))
	}

	vao := buffer.NewVertexArrayObject()

	vao.AddVBO("default", len(unitCircle)/3, 0, attribute.Format{
		{Name: "in_position", Type: attribute.Vec3},
	})

	vao.SetData("default", 0, unitCircle)

	points := make([]float32, len(curve)*2)

	vao.AddVBO("points", len(points)/2, 1, attribute.Format{
		{Name: "center", Type: attribute.Vec2},
	})

	for i := 0; i < len(curve); i++ {
		points[i*2] = curve[i].X
		points[i*2+1] = curve[i].Y
	}

	vao.SetData("points", 0, points)

	vao.Bind()
	vao.Attach(sliderShader)
	vao.Unbind()

	//slice := buffer.MakeVertexSlice(sliderShader, len(mesh)/6, len(mesh)/6)
	//slice.Begin()
	//slice.SetVertexData(mesh)
	//slice.End()
	return vao
}

func createMesh(curve []bmath.Vector2f) []float32 {
	if len(unitCircle) == 0 {
		createCircle(int(settings.Objects.SliderLOD))
	}

	vertices := make([]float32, 6*3*int(settings.Objects.SliderLOD)*len(curve))

	for i, v := range curve {
		for j, s := range unitCircle {
			vertices[i*len(unitCircle)+j] = s
			if j%3 == 0 {
				vertices[i*len(unitCircle)+j] *= float32(CS)
				vertices[i*len(unitCircle)+j] += v.X
			} else if j%3 == 1 {
				vertices[i*len(unitCircle)+j] *= float32(CS)
				vertices[i*len(unitCircle)+j] += v.Y
			}
		}
	}

	return vertices
}

func set(array []float32, index int, data ...float32) {
	copy(array[index:index+len(data)], data)
}

func createCircle(segments int) {
	points := make([]bmath.Vector2f, segments+2)
	points[0] = bmath.NewVec2f(0, 0)

	for i := 0; i < segments; i++ {
		points[i+1] = bmath.NewVec2fRad(float32(i)/float32(segments)*2*math32.Pi, 1)
	}

	points[segments+1] = points[1]

	unitCircle = make([]float32, 9*segments)

	for j := range points {
		if j >= 2 {
			p1, p2, p3 := points[j-1], points[j], points[0]
			set(unitCircle, (j-2)*9, p1.X, p1.Y, 1.0, p2.X, p2.Y, 1.0, p3.X, p3.Y, 0.0)
		}
	}
}
