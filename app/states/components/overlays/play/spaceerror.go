package play

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const baseSpaceSize = 64.0

type SpaceErrorMeter struct {
	diff             *difficulty.Difficulty
	errorDisplay     *sprite.SpriteManager
	errorCurrent     vector.Vector2d
	errorDot         *sprite.Sprite
	errorDisplayFade *animation.Glider

	lastTime float64

	hitCircle        *texture.TextureRegion
	hitCircleOverlay *texture.TextureRegion
}

func NewSpaceErrorMeter(diff *difficulty.Difficulty) *SpaceErrorMeter {
	meter := new(SpaceErrorMeter)

	meter.diff = diff
	meter.errorDisplay = sprite.NewSpriteManager()
	meter.errorDisplayFade = animation.NewGlider(0)

	pixel := graphics.Pixel.GetRegion()

	meter.errorDot = sprite.NewSpriteSingle(&pixel, 3.0, vector.NewVec2d(0, 0), vector.Centre)

	dotSize := settings.Gameplay.SpaceErrorMeter.DotScale / 8
	meter.errorDot.SetScaleV(vector.NewVec2d(dotSize, dotSize))

	meter.errorDot.SetAlpha(1)
	meter.errorDot.SetRotation(math.Pi / 4)

	meter.errorDisplay.Add(meter.errorDot)

	meter.hitCircle = skin.GetTexture("hitcircle")
	meter.hitCircleOverlay = skin.GetTexture("hitcircleoverlay")

	return meter
}

func (meter *SpaceErrorMeter) Add(time float64, error vector.Vector2f) {
	scl := baseSpaceSize * settings.Gameplay.SpaceErrorMeter.Scale

	error = error.Scl(float32(1 / meter.diff.CircleRadius))

	pixel := graphics.Pixel.GetRegion()

	middle := sprite.NewSpriteSingle(&pixel, 2.0, error.Copy64().Scl(scl), vector.Centre)

	middle.SetAdditive(true)

	errorA := error.Len()

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

	dotSize := settings.Gameplay.SpaceErrorMeter.DotScale / 16

	middle.SetScaleV(vector.NewVec2d(dotSize, dotSize))

	middle.SetColor(col)

	middle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, time, time+10000, 0.7, 0.0))
	middle.AdjustTimesToTransformations()

	meter.errorDisplay.Add(middle)

	meter.errorCurrent = meter.errorCurrent.Scl(0.8).Add(error.Copy64().Scl(0.2))

	meter.errorDot.ClearTransformations()
	meter.errorDot.AddTransform(animation.NewVectorTransformV(animation.Move, easing.OutQuad, time, time+800, meter.errorDot.GetPosition(), meter.errorCurrent.Scl(scl)))

	meter.errorDisplayFade.Reset()
	meter.errorDisplayFade.SetValue(1.0)
	meter.errorDisplayFade.AddEventSEase(time+4000, time+5000, 1.0, 0.0, easing.InQuad)
}

func (meter *SpaceErrorMeter) Update(time float64) {
	meter.errorDisplayFade.Update(time)
	meter.errorDisplay.Update(time)

	meter.lastTime = time
}

func (meter *SpaceErrorMeter) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()

	meterAlpha := settings.Gameplay.SpaceErrorMeter.Opacity * meter.errorDisplayFade.GetValue() * alpha
	if meterAlpha > 0.001 && settings.Gameplay.SpaceErrorMeter.Show {
		basePos := vector.NewVec2d(settings.Gameplay.SpaceErrorMeter.XPosition, settings.Gameplay.SpaceErrorMeter.YPosition)
		origin := vector.ParseOrigin(settings.Gameplay.SpaceErrorMeter.Align)

		scl := baseSpaceSize * settings.Gameplay.SpaceErrorMeter.Scale

		pos := basePos.Sub(origin.Scl(scl))

		batch.SetTranslation(pos)
		batch.SetScale(scl, scl)

		batch.SetColor(0.2, 0.2, 0.2, meterAlpha*0.8)
		batch.DrawUnit(*meter.hitCircle)

		batch.SetColor(1, 1, 1, meterAlpha*0.8)
		batch.DrawUnit(*meter.hitCircleOverlay)

		batch.SetColor(1, 1, 1, meterAlpha)

		meter.errorDisplay.Draw(meter.lastTime, batch)
	}

	batch.ResetTransform()
}
