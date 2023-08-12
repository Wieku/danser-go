package graphics

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"sync"
)

var danserShader *shader.RShader = nil

func initDanserShader() {
	if danserShader != nil {
		return
	}

	vert, err := assets.GetString("assets/shaders/cursortrail.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/cursortrail.fsh")
	if err != nil {
		panic(err)
	}

	danserShader = shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))
}

type danserRenderer struct {
	Position vector.Vector2f
	LastPos  vector.Vector2f
	VaoPos   vector.Vector2f
	RendPos  vector.Vector2f

	Points        []vector.Vector2f
	PointsC       []float64
	removeCounter float64

	vertices  []float32
	vaoSize   int
	maxCap    int
	vaoDirty  bool
	vao       *buffer.VertexArrayObject
	mutex     *sync.Mutex
	hueBase   float64
	vecSize   int
	instances int

	firstTime bool
}

func newDanserRenderer() *danserRenderer {
	initDanserShader()

	points := int(math.Ceil(float64(settings.Cursor.TrailMaxLength) * settings.Cursor.TrailDensity))

	vao := buffer.NewVertexArrayObject()

	vao.AddVBO(
		"default",
		6,
		0,
		attribute.Format{
			{Name: "in_position", Type: attribute.Vec2},
			{Name: "in_tex_coord", Type: attribute.Vec2},
		},
	)

	vao.SetData("default", 0, []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		1, 1, 1, 1,
		-1, 1, 0, 1,
		-1, -1, 0, 0,
	})

	vao.AddVBO(
		"points",
		points,
		1,
		attribute.Format{
			{Name: "in_mid", Type: attribute.Vec2},
			{Name: "hue", Type: attribute.Float},
		},
	)

	vao.Attach(danserShader)

	cursor := &danserRenderer{LastPos: vector.NewVec2f(100, 100), Position: vector.NewVec2f(100, 100), vao: vao, mutex: &sync.Mutex{}, RendPos: vector.NewVec2f(100, 100), vertices: make([]float32, points*3), firstTime: true}
	cursor.vecSize = 3

	return cursor
}

func (cursor *danserRenderer) SetPosition(position vector.Vector2f) {
	cursor.Position = position

	if cursor.firstTime {
		cursor.LastPos = position
		cursor.VaoPos = position

		cursor.firstTime = false
	}
}

func (cursor *danserRenderer) Update(delta float64) {
	if settings.Cursor.TrailStyle == 3 {
		cursor.hueBase += settings.Cursor.Style23Speed / 360.0 * delta
		if cursor.hueBase > 1.0 {
			cursor.hueBase -= 1.0
		} else if cursor.hueBase < 0 {
			cursor.hueBase += 1.0
		}
	}

	lengthAdjusted := int(float64(settings.Cursor.TrailMaxLength) * settings.Cursor.TrailDensity)

	points := cursor.Position.Dst(cursor.LastPos)
	distance := float32(1.0 / settings.Cursor.TrailDensity)

	dirtyLocal := false

	if int(points/distance) > 0 {
		temp := cursor.LastPos
		for i := distance; i < points; i += distance {
			temp = cursor.Position.Sub(cursor.LastPos).Scl(i / points).Add(cursor.LastPos)
			cursor.Points = append(cursor.Points, temp)
			cursor.PointsC = append(cursor.PointsC, cursor.hueBase)

			if settings.Cursor.TrailStyle == 2 {
				cursor.hueBase += settings.Cursor.Style23Speed / 360.0 * float64(distance)
				if cursor.hueBase > 1.0 {
					cursor.hueBase -= 1.0
				} else if cursor.hueBase < 0 {
					cursor.hueBase += 1.0
				}
			}
		}
		dirtyLocal = true
		cursor.LastPos = temp
	}

	if len(cursor.Points) > 0 {
		cursor.removeCounter += float64(len(cursor.Points)+3) / (360.0 / delta) * settings.Cursor.TrailRemoveSpeed
		times := int(math.Floor(cursor.removeCounter))

		if len(cursor.Points) > lengthAdjusted {
			cursor.Points = cursor.Points[len(cursor.Points)-lengthAdjusted:]
			cursor.PointsC = cursor.PointsC[len(cursor.PointsC)-lengthAdjusted:]
			cursor.removeCounter = 0
			dirtyLocal = true
		} else if times > 0 {
			times = min(times, len(cursor.Points))

			cursor.Points = cursor.Points[times:]
			cursor.PointsC = cursor.PointsC[times:]
			cursor.removeCounter -= float64(times)

			dirtyLocal = true
		}
	}

	cursor.mutex.Lock()
	if dirtyLocal {
		if len(cursor.vertices) != lengthAdjusted*3 {
			cursor.vertices = make([]float32, lengthAdjusted*3)
		}

		for i, o := range cursor.Points {
			inv := float32(len(cursor.Points) - i - 1)

			hue := float32(cursor.PointsC[i])
			if settings.Cursor.TrailStyle == 4 {
				hue = float32(settings.Cursor.Style4Shift) * inv / float32(len(cursor.Points))
			}

			index := i * 3
			cursor.vertices[index] = o.X
			cursor.vertices[index+1] = o.Y
			cursor.vertices[index+2] = hue
		}

		cursor.maxCap = lengthAdjusted
		cursor.vaoSize = len(cursor.Points)

		cursor.vaoDirty = true
	}
	cursor.VaoPos = cursor.Position
	cursor.mutex.Unlock()
}

func (cursor *danserRenderer) UpdateRenderer() {
	cursor.mutex.Lock()
	if cursor.vaoDirty {
		cursor.vao.Resize("points", cursor.maxCap)
		cursor.vao.SetData("points", 0, cursor.vertices[0:cursor.vaoSize*3])
		cursor.instances = cursor.vaoSize
		cursor.vaoDirty = false
	}
	cursor.RendPos = cursor.VaoPos
	cursor.mutex.Unlock()
}

func (cursor *danserRenderer) DrawM(scale, expand float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color) {
	if settings.Cursor.CursorExpand {
		scale *= expand
	}

	hueShift := color.GetHue()

	siz := settings.Cursor.CursorSize * scale

	if settings.Cursor.EnableCustomTrailGlowOffset {
		colorGlow = color.Shift(float32(settings.Cursor.TrailGlowOffset), 0, 0)
	}

	glowShift := colorGlow.GetHue()

	if settings.Cursor.TrailStyle > 1 {
		color = color.Shift(float32(cursor.hueBase*360), 0, 0)
		colorGlow = colorGlow.Shift(float32(cursor.hueBase*360), 0, 0)
	}

	colorD := color
	colorD2 := colorGlow

	if settings.Cursor.TrailStyle > 1 {
		colorD = color2.NewLA(1.0, color.A)
		colorD2 = color2.NewLA(1.0, colorGlow.A)
	}

	danserShader.Bind()

	CursorTrail.Bind(1)
	danserShader.SetUniform("tex", int32(1))

	danserShader.SetUniform("proj", batch.Projection)
	danserShader.SetUniform("points", float32(cursor.instances))
	danserShader.SetUniform("instances", float32(cursor.instances))

	if settings.Cursor.TrailStyle == 1 {
		danserShader.SetUniform("saturation", float32(0.0))
	} else {
		danserShader.SetUniform("saturation", float32(1.0))
	}

	cursor.vao.Bind()

	innerLengthMult := float32(1.0)

	cursorScl := float32(siz * (16.0 / 18) * scale)

	if settings.Cursor.EnableTrailGlow {
		cursorScl = float32(siz * (12.0 / 18) * scale)
		innerLengthMult = float32(settings.Cursor.InnerLengthMult)

		danserShader.SetUniform("col_tint", colorD2)
		danserShader.SetUniform("scale", float32(siz*(16.0/18)*scale*settings.Cursor.TrailScale))
		danserShader.SetUniform("endScale", float32(settings.Cursor.GlowEndScale))
		if settings.Cursor.TrailStyle > 1 {
			danserShader.SetUniform("hueshift", glowShift/360)
		}

		cursor.vao.DrawInstanced(0, cursor.instances)
	}

	danserShader.SetUniform("col_tint", colorD)
	danserShader.SetUniform("scale", cursorScl*float32(settings.Cursor.TrailScale))
	danserShader.SetUniform("points", float32(len(cursor.Points))*innerLengthMult)
	danserShader.SetUniform("endScale", float32(settings.Cursor.TrailEndScale))
	if settings.Cursor.TrailStyle > 1 {
		danserShader.SetUniform("hueshift", hueShift/360)
	}

	cursor.vao.DrawInstanced(0, cursor.instances)

	cursor.vao.Unbind()

	danserShader.Unbind()

	batch.Begin()

	position := cursor.RendPos
	if settings.PLAY {
		position = cursor.Position
	}

	batch.ResetTransform()

	batch.SetTranslation(position.Copy64())
	batch.SetScale(siz*scale, siz*scale)
	batch.SetSubScale(1, 1)

	batch.SetColorM(color)
	batch.DrawUnit(*CursorTex)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color.A)))
	batch.DrawUnit(*CursorTop)

	batch.End()
}
