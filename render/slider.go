package render

import (
	"github.com/wieku/danser/bmath"
	"math"
	"github.com/wieku/glhf"
	"log"
	_ "image/png"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/settings"
	"io/ioutil"
)

var sliderShader *glhf.Shader = nil
var fboShader *glhf.Shader
var fboSlice *glhf.VertexSlice
var sliderVertexFormat glhf.AttrFormat
var cam mgl32.Mat4
var fbo *glhf.Frame
var fboUnit int32
var CS float64

func SetupSlider() {

	sliderVertexFormat = glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "center", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}
	var err error

	svert , _ := ioutil.ReadFile("assets/shaders/slider.vsh")
	sfrag , _ := ioutil.ReadFile("assets/shaders/slider.fsh")
	sliderShader, err = glhf.NewShader(sliderVertexFormat, glhf.AttrFormat{{Name: "col_border", Type: glhf.Vec4}, {Name: "tex", Type: glhf.Int}, {Name: "proj", Type: glhf.Mat4}, {Name: "trans", Type: glhf.Mat4}}, string(svert), string(sfrag))
	if err != nil {
		log.Println(err)
	}


	fvert , _ := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	ffrag , _ := ioutil.ReadFile("assets/shaders/fbopass.fsh")
	fboShader, err = glhf.NewShader(glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}, glhf.AttrFormat{{Name: "tex", Type: glhf.Int}}, string(fvert), string(ffrag))


	if err != nil {
		log.Println("FboPass: " + err.Error())
	}

	fbo = glhf.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, true)
	gl.ActiveTexture(gl.TEXTURE29)
	fbo.Texture().Begin()
	fboUnit = 29

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
}

func (self *SliderRenderer) SetCamera(camera mgl32.Mat4) {
	cam = camera
	sliderShader.SetUniformAttr(2, cam)
}

func (self *SliderRenderer) GetShape(curve []bmath.Vector2d) (*glhf.VertexSlice, int) {
	return createMesh(curve), int(settings.Objects.SliderLOD)
}

func createMesh(curve []bmath.Vector2d) *glhf.VertexSlice {
	var slice *glhf.VertexSlice

	vecr := make([]float32, 0)
	num := 0
	for _, v := range curve {
		tab := createCircle(v.X, v.Y, 64*CS, int(settings.Objects.SliderLOD))
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