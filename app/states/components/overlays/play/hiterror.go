package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
)

const errorBase = 4.8

var colors = []color2.Color{color2.NewRGBA(0.2, 0.8, 1, 1), color2.NewRGBA(0.44, 0.98, 0.18, 1), color2.NewRGBA(0.85, 0.68, 0.27, 1), color2.NewRGBA(0.98, 0.11, 0.011, 1)}

type HitErrorMeter struct {
	diff             *difficulty.Difficulty
	errorDisplay     *sprite.Manager
	errorCurrent     float64
	triangle         *sprite.Sprite
	errorDisplayFade *animation.Glider

	Width    float64
	Height   float64
	lastTime float64

	errors       []float64
	unstableRate float64
	avgPos       float64
	avgNeg       float64

	urText   string
	urGlider *animation.TargetGlider

	averageN float64
	averageP float64
	countN   int
	countP   int
}

func NewHitErrorMeter(width, height float64, diff *difficulty.Difficulty) *HitErrorMeter {
	meter := new(HitErrorMeter)
	meter.Width = width
	meter.Height = height

	meter.diff = diff
	meter.errorDisplay = sprite.NewManager()
	meter.errorDisplayFade = animation.NewGlider(0.0)
	meter.urText = "0UR"
	meter.urGlider = animation.NewTargetGlider(0, 0)

	baseScale := 0.8
	if settings.Gameplay.HitErrorMeter.ScaleWithSpeed {
		baseScale /= meter.diff.Speed
	}

	vals := []float64{float64(meter.diff.Hit300) * baseScale, float64(meter.diff.Hit100) * baseScale, float64(meter.diff.Hit50) * baseScale}

	scale := settings.Gameplay.HitErrorMeter.Scale

	pixel := graphics.Pixel.GetRegion()
	bg := sprite.NewSpriteSingle(&pixel, 0.0, vector.NewVec2d(meter.Width/2, meter.Height-errorBase*2*scale), vector.Centre)
	bg.SetScaleV(vector.NewVec2d(vals[2]*2, errorBase*4).Scl(scale))
	bg.SetColor(color2.NewL(0))
	bg.SetAlpha(0.6)
	meter.errorDisplay.Add(bg)

	for i, v := range vals {
		pos := 0.0
		width := v

		if i > 0 {
			pos = vals[i-1]
			width -= vals[i-1]
		}

		left := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(meter.Width/2-pos*scale, meter.Height-errorBase*2*scale), vector.CentreRight)
		left.SetScaleV(vector.NewVec2d(width, errorBase).Scl(scale))
		left.SetColor(colors[i])

		meter.errorDisplay.Add(left)

		right := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(meter.Width/2+pos*scale, meter.Height-errorBase*2*scale), vector.CentreLeft)
		right.SetScaleV(vector.NewVec2d(width, errorBase).Scl(scale))
		right.SetColor(colors[i])

		meter.errorDisplay.Add(right)
	}

	middle := sprite.NewSpriteSingle(&pixel, 2.0, vector.NewVec2d(meter.Width/2, meter.Height-errorBase*2*scale), vector.Centre)
	middle.SetScaleV(vector.NewVec2d(2, errorBase*4).Scl(scale))

	meter.errorDisplay.Add(middle)

	meter.triangle = sprite.NewSpriteSingle(graphics.TriangleSmall, 2.0, vector.NewVec2d(meter.Width/2, meter.Height-errorBase*2.5*scale), vector.BottomCentre)
	meter.triangle.SetScaleV(vector.NewVec2d(scale/6, scale/6))
	meter.triangle.SetAlpha(1)

	meter.errorDisplay.Add(meter.triangle)

	return meter
}

func (meter *HitErrorMeter) Add(time, error float64, positionalMiss bool) {
	if positionalMiss && !settings.Gameplay.HitErrorMeter.ShowPositionalMisses {
		return
	}

	errorA := int64(math.Abs(error))

	scale := settings.Gameplay.HitErrorMeter.Scale

	pixel := graphics.Pixel.GetRegion()

	errorPos := error * 0.8
	if settings.Gameplay.HitErrorMeter.ScaleWithSpeed {
		errorPos /= meter.diff.Speed
	}

	middle := sprite.NewSpriteSingle(&pixel, 3.0, vector.NewVec2d(meter.Width/2+errorPos*scale, meter.Height-errorBase*2*scale), vector.Centre)
	middle.ShowForever(false)
	middle.SetScaleV(vector.NewVec2d(3, errorBase*4).Scl(scale))
	middle.SetAdditive(true)

	baseFade := 0.4

	if positionalMiss {
		baseFade = 0.8

		middle.SetColor(colors[3])

		middle.SetScaleV(vector.NewVec2d(3, errorBase*4*settings.Gameplay.HitErrorMeter.PositionalMissScale).Scl(scale))
		middle.SetAdditive(false)
	} else {
		switch {
		case errorA < meter.diff.Hit300:
			middle.SetColor(colors[0])
		case errorA < meter.diff.Hit100:
			middle.SetColor(colors[1])
		case errorA < meter.diff.Hit50:
			middle.SetColor(colors[2])
		}
	}

	middle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, time, time+max(0, settings.Gameplay.HitErrorMeter.PointFadeOutTime*1000), baseFade, 0.0))
	middle.AdjustTimesToTransformations()

	meter.errorDisplay.Add(middle)

	if positionalMiss {
		return
	}

	meter.errorCurrent = meter.errorCurrent*0.8 + errorPos*0.2

	meter.triangle.ClearTransformations()
	meter.triangle.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.OutQuad, time, time+800, meter.triangle.GetPosition().X, meter.Width/2+meter.errorCurrent*scale))

	meter.errorDisplayFade.Reset()
	meter.errorDisplayFade.SetValue(1.0)
	meter.errorDisplayFade.AddEventSEase(time+4000, time+5000, 1.0, 0.0, easing.InQuad)

	if error >= 0 {
		meter.averageP += error
		meter.countP++
	} else {
		meter.averageN += error
		meter.countN++
	}

	meter.errors = append(meter.errors, error)

	average := (meter.averageN + meter.averageP) / float64(meter.countN+meter.countP)

	urBase := 0.0
	for _, e := range meter.errors {
		urBase += math.Pow(e-average, 2)
	}

	urBase /= float64(len(meter.errors))

	meter.avgNeg = meter.averageN / max(float64(meter.countN), 1)
	meter.avgPos = meter.averageP / max(float64(meter.countP), 1)
	meter.unstableRate = math.Sqrt(urBase) * 10

	meter.urGlider.SetValue(meter.GetUnstableRateConverted(), settings.Gameplay.HitErrorMeter.StaticUnstableRate)
}

func (meter *HitErrorMeter) Update(time float64) {
	meter.errorDisplayFade.Update(time)
	meter.errorDisplay.Update(time)

	meter.lastTime = time

	meter.urGlider.SetDecimals(settings.Gameplay.HitErrorMeter.UnstableRateDecimals)
	meter.urGlider.Update(time)
	meter.urText = fmt.Sprintf("%."+strconv.Itoa(settings.Gameplay.HitErrorMeter.UnstableRateDecimals)+"fUR", meter.urGlider.GetValue())
}

func (meter *HitErrorMeter) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()

	meterAlpha := settings.Gameplay.HitErrorMeter.Opacity * meter.errorDisplayFade.GetValue() * alpha
	if meterAlpha > 0.001 && settings.Gameplay.HitErrorMeter.Show {
		batch.SetColor(1, 1, 1, meterAlpha)
		batch.SetTranslation(vector.NewVec2d(settings.Gameplay.HitErrorMeter.XOffset, settings.Gameplay.HitErrorMeter.YOffset))

		meter.errorDisplay.Draw(meter.lastTime, batch)

		if settings.Gameplay.HitErrorMeter.ShowUnstableRate {
			pY := meter.Height - (errorBase*4+3.75)*settings.Gameplay.HitErrorMeter.Scale
			scale := settings.Gameplay.HitErrorMeter.UnstableRateScale

			fnt := font.GetFont("HUDFont")
			fnt.DrawOrigin(batch, meter.Width/2, pY, vector.BottomCentre, 15*scale, true, meter.urText)
		}
	}

	batch.ResetTransform()
}

func (meter *HitErrorMeter) GetAvgNeg() float64 {
	return meter.avgNeg
}

func (meter *HitErrorMeter) GetAvgNegConverted() float64 {
	return meter.avgNeg / meter.diff.Speed
}

func (meter *HitErrorMeter) GetAvgPos() float64 {
	return meter.avgPos
}

func (meter *HitErrorMeter) GetAvgPosConverted() float64 {
	return meter.avgPos / meter.diff.Speed
}

func (meter *HitErrorMeter) GetUnstableRate() float64 {
	return meter.unstableRate
}

func (meter *HitErrorMeter) GetUnstableRateConverted() float64 {
	return meter.unstableRate / meter.diff.Speed
}
