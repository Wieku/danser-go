package common

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/attribute"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shader"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
)

const DefaultFlashlightSize = 168.0
const DefaultFlashlightDuration = 800.0

type Flashlight struct {
	flShader   *shader.RShader
	vao        *buffer.VertexArrayObject
	time       float64
	delta      float64
	position   vector.Vector2f
	size       *animation.Glider
	dim        *animation.Glider
	beatMap    *beatmap.BeatMap
	breakIndex int
	target     float64
	sliding    bool

	flSettings difficulty.FlashlightSettings
}

func NewFlashlight(beatMap *beatmap.BeatMap) *Flashlight {
	vert, err := assets.GetString("assets/shaders/flashlight.vsh")
	if err != nil {
		panic(err)
	}

	frag, err := assets.GetString("assets/shaders/flashlight.fsh")
	if err != nil {
		panic(err)
	}

	flShader := shader.NewRShader(shader.NewSource(vert, shader.Vertex), shader.NewSource(frag, shader.Fragment))

	vao := buffer.NewVertexArrayObject()
	vao.AddVBO("default", 6, 0, attribute.Format{
		attribute.VertexAttribute{Name: "in_position", Type: attribute.Vec2},
	})

	vao.SetData("default", 0, []float32{
		-1, -1,
		1, -1,
		1, 1,
		1, 1,
		-1, 1,
		-1, -1,
	})

	vao.Attach(flShader)

	size := animation.NewGlider(DefaultFlashlightSize * 8)

	startTime := beatMap.HitObjects[0].GetStartTime() / settings.SPEED
	endTime := (beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime() + float64(beatMap.Diff.Hit50+5)) / settings.SPEED / beatMap.Diff.GetSpeed()

	size.AddEvent(startTime-DefaultFlashlightDuration, startTime, DefaultFlashlightSize)
	size.AddEvent(endTime, endTime+DefaultFlashlightDuration, DefaultFlashlightSize*8)

	flSettings := difficulty.FlashlightSettings{
		FollowDelay:    120,
		SizeMultiplier: 1,
		ComboBasedSize: true,
	}

	if fl, ok := difficulty.GetModConfig[difficulty.FlashlightSettings](beatMap.Diff); ok {
		flSettings = fl
	}

	return &Flashlight{
		flShader:   flShader,
		vao:        vao,
		size:       size,
		beatMap:    beatMap,
		breakIndex: -1,
		dim:        animation.NewGlider(0.0),
		flSettings: flSettings,
	}
}

func (fl *Flashlight) UpdatePosition(cursorPosition vector.Vector2f) {
	oldPosition := fl.position
	fl.position = cursorPosition.Sub(oldPosition).Scl(float32(easing.OutQuad(min(fl.delta, fl.flSettings.FollowDelay) / fl.flSettings.FollowDelay))).Add(oldPosition)
}

func (fl *Flashlight) UpdateCombo(combo int64) {
	target := DefaultFlashlightSize

	if fl.flSettings.ComboBasedSize {
		switch {
		case combo > 200:
			target *= 0.625
		case combo > 100:
			target *= 0.8125
		}
	}

	fl.target = target

	fl.size.AddEvent(fl.time, fl.time+DefaultFlashlightDuration, target)
}

func (fl *Flashlight) SetSliding(value bool) {
	if fl.sliding != value {
		dim := 0.0
		if value {
			dim = 0.8
		}

		fl.dim.AddEvent(fl.time, fl.time+50, dim)

		fl.sliding = value
	}
}

func (fl *Flashlight) Update(time float64) {
	fl.delta = time - fl.time

	fl.time = time

	for i := fl.breakIndex + 1; i < len(fl.beatMap.Pauses); i++ {
		pause := fl.beatMap.Pauses[i]

		pauseStart := pause.GetStartTime() / (settings.SPEED * fl.beatMap.Diff.GetSpeed())
		pauseEnd := pause.GetEndTime() / (settings.SPEED * fl.beatMap.Diff.GetSpeed())

		if time < pauseStart {
			break
		}

		fl.breakIndex = i

		if pauseEnd-pauseStart > DefaultFlashlightDuration*2 {
			fl.size.AddEvent(pauseStart, pauseStart+DefaultFlashlightDuration, DefaultFlashlightSize*2.5)
			fl.size.AddEvent(pauseEnd-DefaultFlashlightDuration, pauseEnd, fl.target)
		}
	}

	fl.size.Update(time)
	fl.dim.Update(time)
}

func (fl *Flashlight) Draw(matrix mgl32.Mat4) {
	blend.Push()
	blend.Enable()
	blend.SetFunctionSeparate(blend.SrcAlpha, blend.OneMinusSrcAlpha, blend.One, blend.One)

	fl.flShader.Bind()
	fl.flShader.SetUniform("cursorPosition", mgl32.Vec2{fl.position.X, fl.position.Y})
	fl.flShader.SetUniform("radius", float32(fl.size.GetValue()*fl.flSettings.SizeMultiplier))
	fl.flShader.SetUniform("dim", float32(fl.dim.GetValue()))
	fl.flShader.SetUniform("maxDim", float32(settings.Gameplay.FlashlightDim))
	fl.flShader.SetUniform("invMatrix", matrix.Inv())

	fl.vao.Bind()
	fl.vao.Draw()
	fl.vao.Unbind()

	fl.flShader.Unbind()

	blend.Pop()
}
