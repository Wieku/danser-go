package play

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const errorBaseScale = 1.5

var colors = []color2.Color{{0.2, 0.8, 1, 1}, {0.44, 0.98, 0.18, 1}, {0.85, 0.68, 0.27, 1}}

type HitErrorMeter struct {
	diff             *difficulty.Difficulty
	errorDisplay     *sprite.SpriteManager
	errorCurrent     float64
	triangle         *sprite.Sprite
	errorDisplayFade *animation.Glider

	Width    float64
	Height   float64
	lastTime float64
}

func NewHitErrorMeter(width, height float64, diff *difficulty.Difficulty) *HitErrorMeter {
	meter := new(HitErrorMeter)
	meter.Width = width
	meter.Height = height

	meter.diff = diff
	meter.errorDisplay = sprite.NewSpriteManager()
	meter.errorDisplayFade = animation.NewGlider(0.0)

	sum := meter.diff.Hit50

	scale := errorBaseScale * settings.Gameplay.HitErrorMeter.Scale

	pixel := graphics.Pixel.GetRegion()
	bg := sprite.NewSpriteSingle(&pixel, 0.0, vector.NewVec2d(meter.Width/2, meter.Height-10*scale), bmath.Origin.Centre)
	bg.SetScaleV(vector.NewVec2d(float64(sum)*2*scale, 20*scale))
	bg.SetColor(color2.Color{0, 0, 0, 1})
	bg.SetAlpha(0.8)
	meter.errorDisplay.Add(bg)

	vals := []float64{float64(meter.diff.Hit300), float64(meter.diff.Hit100), float64(meter.diff.Hit50)}

	for i, v := range vals {
		pos := 0.0
		width := v

		if i > 0 {
			pos = vals[i-1]
			width -= vals[i-1]
		}

		left := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(meter.Width/2-pos*scale, meter.Height-10*scale), bmath.Origin.CentreRight)
		left.SetScaleV(vector.NewVec2d(width*scale, 4*scale))
		left.SetColor(colors[i])
		left.SetAlpha(0.8)

		meter.errorDisplay.Add(left)

		right := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(meter.Width/2+pos*scale, meter.Height-10*scale), bmath.Origin.CentreLeft)
		right.SetScaleV(vector.NewVec2d(width*scale, 4*scale))
		right.SetColor(colors[i])
		right.SetAlpha(0.8)

		meter.errorDisplay.Add(right)
	}

	middle := sprite.NewSpriteSingle(&pixel, 2.0, vector.NewVec2d(meter.Width/2, meter.Height-10*scale), bmath.Origin.Centre)
	middle.SetScaleV(vector.NewVec2d(2*scale, 20*scale))
	middle.SetAlpha(0.8)

	meter.errorDisplay.Add(middle)

	meter.triangle = sprite.NewSpriteSingle(graphics.TriangleSmall, 2.0, vector.NewVec2d(meter.Width/2, meter.Height-12*scale), bmath.Origin.BottomCentre)
	meter.triangle.SetScaleV(vector.NewVec2d(scale/8, scale/8))
	meter.triangle.SetAlpha(0.8)

	meter.errorDisplay.Add(meter.triangle)

	return meter
}

func (meter *HitErrorMeter) Add(time, error float64) {
	errorA := int64(math.Abs(error))

	scale := settings.Gameplay.HitErrorMeter.Scale * errorBaseScale

	pixel := graphics.Pixel.GetRegion()

	middle := sprite.NewSpriteSingle(&pixel, 3.0, vector.NewVec2d(meter.Width/2+error*scale, meter.Height-10*scale), bmath.Origin.Centre)
	middle.SetScaleV(vector.NewVec2d(1.5, 20).Scl(scale))
	middle.SetAdditive(true)

	var col color2.Color
	switch {
	case errorA < meter.diff.Hit300:
		col = colors[0]
	case errorA < meter.diff.Hit100:
		col = colors[1]
	case errorA < meter.diff.Hit50:
		col = colors[2]
	}

	middle.SetColor(col)

	middle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, time, time+10000, 0.4, 0.0))
	middle.AdjustTimesToTransformations()

	meter.errorDisplay.Add(middle)

	meter.errorCurrent = meter.errorCurrent*0.8 + error*0.2

	meter.triangle.ClearTransformations()
	meter.triangle.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.OutQuad, time, time+800, meter.triangle.GetPosition().X, meter.Width/2+meter.errorCurrent*scale))

	meter.errorDisplayFade.Reset()
	meter.errorDisplayFade.SetValue(1.0)
	meter.errorDisplayFade.AddEventSEase(time+4000, time+5000, 1.0, 0.0, easing.InQuad)
}

func (meter *HitErrorMeter) Update(time float64) {
	meter.errorDisplayFade.Update(time)
	meter.errorDisplay.Update(int64(time))

	meter.lastTime = time
}

func (meter *HitErrorMeter) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()
	meterAlpha := settings.Gameplay.HitErrorMeter.Opacity * meter.errorDisplayFade.GetValue() * alpha
	if meterAlpha > 0.001 && settings.Gameplay.HitErrorMeter.Show {
		batch.SetColor(1, 1, 1, meterAlpha)
		meter.errorDisplay.Draw(int64(meter.lastTime), batch)
	}
	batch.ResetTransform()
}
