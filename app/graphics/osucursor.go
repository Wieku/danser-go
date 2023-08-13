package graphics

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"sync"
)

const scaling = 0.625

var osuShader *shader.RShader = nil

func initOsuShader() {
	if osuShader != nil {
		return
	}

	vert, err := assets.GetString("assets/shaders/osutrail.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/osutrail.fsh")
	if err != nil {
		panic(err)
	}

	osuShader = shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))
}

type osuRenderer struct {
	Position vector.Vector2f
	LastPos  vector.Vector2f
	VaoPos   vector.Vector2f
	RendPos  vector.Vector2f

	Points  []vector.Vector2f
	PointsC []float64

	vertices  []float32
	vaoSize   int
	maxCap    int
	vaoDirty  bool
	vao       *buffer.VertexArrayObject
	mutex     *sync.Mutex
	vecSize   int
	instances int

	trail  *texture.TextureRegion
	cursor *sprite.Sprite
	middle *sprite.Sprite

	clock       float64
	manager     *sprite.Manager
	currentTime float64
	sixtyDelta  float64
	firstTime   bool
}

func newOsuRenderer() *osuRenderer {
	initOsuShader()

	points := int(settings.Cursor.TrailMaxLength)

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
			{Name: "fadeTime", Type: attribute.Float},
		},
	)

	vao.Attach(osuShader)

	cursor := &osuRenderer{LastPos: vector.NewVec2f(100, 100), Position: vector.NewVec2f(100, 100), vao: vao, mutex: &sync.Mutex{}, RendPos: vector.NewVec2f(100, 100), vertices: make([]float32, points*3), firstTime: true}
	cursor.vecSize = 3

	cursor.trail = skin.GetTexture("cursortrail")

	cursorTexture := skin.GetTexture("cursor")

	origin := vector.Centre
	if !skin.GetInfo().CursorCentre {
		origin = vector.TopLeft
	}

	cursor.cursor = sprite.NewSpriteSingle(cursorTexture, 0, vector.NewVec2d(0, 0), origin)
	cursor.middle = sprite.NewSpriteSingle(skin.GetTextureSource("cursormiddle", skin.GetSourceFromTexture(cursorTexture)), 0, vector.NewVec2d(0, 0), origin)

	cursor.manager = sprite.NewManager()

	return cursor
}

func (cursor *osuRenderer) SetPosition(position vector.Vector2f) {
	cursor.Position = position

	if cursor.firstTime {
		cursor.LastPos = position
		cursor.VaoPos = position

		cursor.firstTime = false
	}
}

func (cursor *osuRenderer) Update(delta float64) {
	dirtyLocal := false

	cursor.clock += delta / 100

	if cursor.clock >= 1000000 {
		cursor.clock = 0

		for i := range cursor.PointsC {
			cursor.PointsC[i] -= 1000000
		}

		dirtyLocal = true
	}

	cursor.currentTime = cursor.clock * 100
	cursor.manager.Update(cursor.currentTime)

	if cursor.middle.Texture == nil && !settings.Skin.Cursor.ForceLongTrail {
		cursor.VaoPos = cursor.Position
		cursor.RendPos = cursor.Position

		cursor.sixtyDelta += delta
		if cursor.sixtyDelta >= 16.6667 {
			spr := sprite.NewSpriteSingle(cursor.trail, cursor.currentTime, cursor.Position.Copy64(), vector.Centre)

			spr.SetScale(settings.Skin.Cursor.TrailScale)
			spr.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, cursor.currentTime, cursor.currentTime+150, 1.0, 0.0))
			spr.ResetValuesToTransforms()
			spr.AdjustTimesToTransformations()
			spr.ShowForever(false)

			cursor.manager.Add(spr)

			cursor.sixtyDelta -= 16.6667
		}

		return
	}

	lengthAdjusted := int(settings.Skin.Cursor.LongTrailLength)

	fadeTime := 3.0 * (1.0 + float64(settings.Skin.Cursor.LongTrailLength-2048)/4096)
	distance := cursor.trail.Width / 2.5 * scaling / float32(settings.Skin.Cursor.LongTrailDensity)
	points := cursor.Position.Dst(cursor.LastPos)

	if int(points/distance) > 0 {
		temp := cursor.LastPos
		for i := distance; i < points; i += distance {
			temp = cursor.Position.Sub(cursor.LastPos).Scl(i / points).Add(cursor.LastPos)

			cursor.Points = append(cursor.Points, temp)
			cursor.PointsC = append(cursor.PointsC, cursor.clock+fadeTime)
		}

		dirtyLocal = true
		cursor.LastPos = temp
	}

	if len(cursor.Points) > 0 {
		times := 0
		for i, t := range cursor.PointsC {
			if t > cursor.clock {
				break
			}

			times = i + 1
		}

		if len(cursor.Points) > lengthAdjusted {
			cursor.Points = cursor.Points[len(cursor.Points)-lengthAdjusted:]
			cursor.PointsC = cursor.PointsC[len(cursor.PointsC)-lengthAdjusted:]

			dirtyLocal = true
		} else if times > 0 {
			times = min(times, len(cursor.Points))

			cursor.Points = cursor.Points[times:]
			cursor.PointsC = cursor.PointsC[times:]

			dirtyLocal = true
		}
	}

	cursor.mutex.Lock()
	if dirtyLocal {
		if len(cursor.vertices) != lengthAdjusted*3 {
			cursor.vertices = make([]float32, lengthAdjusted*3)
		}

		for i, o := range cursor.Points {
			index := i * 3
			cursor.vertices[index] = o.X
			cursor.vertices[index+1] = o.Y
			cursor.vertices[index+2] = float32(cursor.PointsC[i])
		}

		cursor.maxCap = lengthAdjusted
		cursor.vaoSize = len(cursor.Points)

		cursor.vaoDirty = true
	}

	cursor.VaoPos = cursor.Position
	cursor.mutex.Unlock()
}

func (cursor *osuRenderer) UpdateRenderer() {
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

func (cursor *osuRenderer) DrawM(scale, expand float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color) {
	scale *= settings.Skin.Cursor.Scale * scaling

	scaleExpanded := scale
	if skin.GetInfo().CursorExpand {
		scaleExpanded *= expand
	}

	if (settings.Skin.Cursor.ForceLongTrail || (cursor.middle.Texture != nil && cursor.middle.Texture.Texture != nil)) && cursor.trail.Texture != nil {
		osuShader.Bind()

		osuShader.SetUniform("tex", int32(cursor.trail.Texture.GetLocation()))
		osuShader.SetUniform("u", mgl32.Vec2{cursor.trail.U1, cursor.trail.U2})
		osuShader.SetUniform("v", mgl32.Vec2{cursor.trail.V1, cursor.trail.V2})
		osuShader.SetUniform("layer", float32(cursor.trail.Layer))
		osuShader.SetUniform("proj", batch.Projection)
		osuShader.SetUniform("clock", float32(cursor.clock))
		osuShader.SetUniform("alpha", color.A)

		cursor.vao.Bind()

		osuShader.SetUniform("scale", float32(scaleExpanded)*cursor.trail.Width/2*float32(settings.Skin.Cursor.TrailScale))

		blend.Push()
		blend.Enable()
		blend.SetFunction(blend.SrcAlpha, blend.One)

		cursor.vao.DrawInstanced(0, cursor.instances)

		blend.Pop()

		cursor.vao.Unbind()

		osuShader.Unbind()
	}

	batch.Begin()

	position := cursor.RendPos
	if settings.PLAY {
		position = cursor.Position
	}

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, float64(color.A))
	batch.SetScale(scaleExpanded, scaleExpanded)
	batch.SetSubScale(1, 1)

	cursor.manager.Draw(cursor.currentTime, batch)

	batch.SetTranslation(position.Copy64())

	if skin.GetInfo().CursorRotate {
		cursor.cursor.SetRotation(cursor.clock / 10 / 10 * 2 * math.Pi)
	} else {
		cursor.cursor.SetRotation(0)
	}

	cursor.cursor.Draw(cursor.currentTime, batch)

	batch.SetScale(scale, scale)

	cursor.middle.Draw(cursor.currentTime, batch)

	batch.End()
}
