package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
)

const baseSpaceSize = 64.0

type AimErrorMeter struct {
	diff             *difficulty.Difficulty
	errorDisplay     *sprite.Manager
	errorCurrent     vector.Vector2d
	errorDot         *sprite.Sprite
	errorDisplayFade *animation.Glider

	lastTime float64

	hitCircle        *texture.TextureRegion
	hitCircleOverlay *texture.TextureRegion

	errors []vector.Vector2d

	urText   string
	urGlider *animation.TargetGlider

	unstableRate float64
}

func NewAimErrorMeter(diff *difficulty.Difficulty) *AimErrorMeter {
	meter := new(AimErrorMeter)

	meter.diff = diff
	meter.errorDisplay = sprite.NewManager()
	meter.errorDisplayFade = animation.NewGlider(0)
	meter.urText = "0UR"
	meter.urGlider = animation.NewTargetGlider(0, 0)

	pixel := graphics.Pixel.GetRegion()

	meter.errorDot = sprite.NewSpriteSingle(&pixel, 3.0, vector.NewVec2d(0, 0), vector.Centre)

	dotSize := settings.Gameplay.AimErrorMeter.DotScale / 12
	meter.errorDot.SetScaleV(vector.NewVec2d(dotSize, dotSize))

	meter.errorDot.SetAlpha(1)
	meter.errorDot.SetRotation(math.Pi / 4)

	meter.errorDisplay.Add(meter.errorDot)

	meter.hitCircle = skin.GetTexture("hitcircle")
	meter.hitCircleOverlay = skin.GetTexture("hitcircleoverlay")

	return meter
}

func (meter *AimErrorMeter) Add(time float64, err vector.Vector2f) {
	scl := baseSpaceSize * settings.Gameplay.AimErrorMeter.Scale

	errorS := err.Scl(float32(1 / meter.diff.CircleRadius))

	pixel := graphics.Pixel.GetRegion()

	middle := sprite.NewSpriteSingle(&pixel, 2.0, errorS.Copy64().Scl(scl), vector.Centre)

	middle.SetAdditive(true)

	errorA := errorS.Len()

	var col color2.Color
	switch {
	case errorA < 0.33:
		col = colors[0]
	case errorA < 0.66:
		col = colors[1]
	case errorA <= 1:
		col = colors[2]
	case errorA > 1:
		col = colors[3]
	}

	dotSize := settings.Gameplay.AimErrorMeter.DotScale / 16

	middle.SetScaleV(vector.NewVec2d(dotSize, dotSize))

	middle.SetColor(col)

	middle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, time, time+10000, 0.7, 0.0))
	middle.AdjustTimesToTransformations()

	meter.errorDisplay.Add(middle)

	meter.errorCurrent = meter.errorCurrent.Scl(0.8).Add(errorS.Copy64().Scl(0.2))

	meter.errorDot.ClearTransformations()
	meter.errorDot.AddTransform(animation.NewVectorTransformV(animation.Move, easing.OutQuad, time, time+800, meter.errorDot.GetPosition(), meter.errorCurrent.Scl(scl)))

	meter.errorDisplayFade.Reset()
	meter.errorDisplayFade.SetValue(1.0)
	meter.errorDisplayFade.AddEventSEase(time+4000, time+5000, 1.0, 0.0, easing.InQuad)

	meter.errors = append(meter.errors, err.Copy64())

	var toAverage vector.Vector2d

	for _, e := range meter.errors {
		toAverage = toAverage.Add(e)
	}

	average := toAverage.Scl(1 / float64(len(meter.errors)))

	urBase := 0.0
	for _, e := range meter.errors {
		urBase += math.Pow(e.Dst(average), 2)
	}

	urBase /= float64(len(meter.errors))

	meter.unstableRate = math.Sqrt(urBase) * 10

	meter.urGlider.SetTarget(meter.unstableRate)
}

func (meter *AimErrorMeter) Update(time float64) {
	meter.errorDisplayFade.Update(time)
	meter.errorDisplay.Update(time)

	meter.lastTime = time

	meter.urGlider.SetDecimals(settings.Gameplay.AimErrorMeter.UnstableRateDecimals)
	meter.urGlider.Update(time)
	meter.urText = fmt.Sprintf("%."+strconv.Itoa(settings.Gameplay.AimErrorMeter.UnstableRateDecimals)+"fUR", meter.urGlider.GetValue())
}

func (meter *AimErrorMeter) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()

	meterAlpha := settings.Gameplay.AimErrorMeter.Opacity * meter.errorDisplayFade.GetValue() * alpha
	if meterAlpha > 0.001 && settings.Gameplay.AimErrorMeter.Show {
		basePos := vector.NewVec2d(settings.Gameplay.AimErrorMeter.XPosition, settings.Gameplay.AimErrorMeter.YPosition)
		origin := vector.ParseOrigin(settings.Gameplay.AimErrorMeter.Align)

		scl := baseSpaceSize * settings.Gameplay.AimErrorMeter.Scale

		pos := basePos.Sub(origin.Scl(scl))

		batch.SetTranslation(pos)
		batch.SetScale(scl, scl)

		batch.SetColor(0.2, 0.2, 0.2, meterAlpha*0.8)
		batch.DrawUnit(*meter.hitCircle)

		batch.SetColor(1, 1, 1, meterAlpha*0.8)
		batch.DrawUnit(*meter.hitCircleOverlay)

		batch.SetColor(1, 1, 1, meterAlpha)

		meter.errorDisplay.Draw(meter.lastTime, batch)

		batch.SetScale(1, 1)

		if settings.Gameplay.AimErrorMeter.ShowUnstableRate {
			scale := settings.Gameplay.AimErrorMeter.UnstableRateScale

			fnt := font.GetFont("Quicksand Bold")
			fnt.DrawOrigin(batch, 0, scl, vector.TopCentre, 15*scale, true, meter.urText)
		}
	}

	batch.ResetTransform()
}
