package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/math/animation"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"strconv"
)

type PPDisplay struct {
	ppFont *font.Font

	aimGlider *animation.TargetGlider
	aimText   string

	tapGlider *animation.TargetGlider
	tapText   string

	accGlider *animation.TargetGlider
	accText   string

	flashlightGlider *animation.TargetGlider
	flashlightText   string

	ppGlider *animation.TargetGlider
	ppText   string

	mText string

	decimals       int
	format         string

	mods           difficulty.Modifier
	experimentalPP bool
}

func NewPPDisplay(mods difficulty.Modifier, experimentalPP bool) *PPDisplay {
	return &PPDisplay{
		ppFont:    font.GetFont("Quicksand Bold"),
		aimGlider: animation.NewTargetGlider(0, 0),
		tapGlider: animation.NewTargetGlider(0, 0),
		accGlider: animation.NewTargetGlider(0, 0),
		flashlightGlider: animation.NewTargetGlider(0, 0),
		ppGlider:  animation.NewTargetGlider(0, 0),
		aimText:   "0pp",
		tapText:   "0pp",
		accText:   "0pp",
		ppText:    "0pp",
		mText:     "0pp",
		decimals:  0,
		format:    "%.0fpp",
		mods: mods,
		experimentalPP: experimentalPP,
	}
}

func (ppDisplay *PPDisplay) Add(results performance.PPv2Results) {
	ppDisplay.aimGlider.SetTarget(results.Aim)
	ppDisplay.tapGlider.SetTarget(results.Speed)
	ppDisplay.accGlider.SetTarget(results.Acc)
	ppDisplay.flashlightGlider.SetTarget(results.Flashlight)
	ppDisplay.ppGlider.SetTarget(results.Total)
}

func (ppDisplay *PPDisplay) Update(time float64) {
	if settings.Gameplay.PPCounter.Decimals > ppDisplay.decimals {
		ppDisplay.decimals = settings.Gameplay.PPCounter.Decimals
		ppDisplay.format = "%." + strconv.Itoa(ppDisplay.decimals) + "fpp"
	}

	var mText string

	ppDisplay.updatePP(ppDisplay.ppGlider, &ppDisplay.ppText, time, &mText)

	if settings.Gameplay.PPCounter.ShowPPComponents {
		ppDisplay.updatePP(ppDisplay.aimGlider, &ppDisplay.aimText, time, &mText)
		ppDisplay.updatePP(ppDisplay.tapGlider, &ppDisplay.tapText, time, &mText)
		ppDisplay.updatePP(ppDisplay.accGlider, &ppDisplay.accText, time, &mText)
		ppDisplay.updatePP(ppDisplay.flashlightGlider, &ppDisplay.flashlightText, time, &mText)
	}

	ppDisplay.mText = mText
}

func (ppDisplay *PPDisplay) updatePP(glider *animation.TargetGlider, text *string, time float64, mText *string) {
	glider.SetDecimals(settings.Gameplay.PPCounter.Decimals)
	glider.Update(time)

	*text = fmt.Sprintf(ppDisplay.format, glider.GetValue())

	if len(*text) > len(*mText) {
		*mText = *text
	}
}

func (ppDisplay *PPDisplay) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()

	ppAlpha := settings.Gameplay.PPCounter.Opacity * alpha

	if ppAlpha < 0.001 || !settings.Gameplay.PPCounter.Show {
		return
	}

	ppScale := settings.Gameplay.PPCounter.Scale

	position := vector.NewVec2d(settings.Gameplay.PPCounter.XPosition, settings.Gameplay.PPCounter.YPosition)
	origin := vector.ParseOrigin(settings.Gameplay.PPCounter.Align)

	cS := settings.Gameplay.PPCounter.Color
	color := color2.NewHSVA(float32(cS.Hue), float32(cS.Saturation), float32(cS.Value), float32(ppAlpha))

	if settings.Gameplay.PPCounter.ShowPPComponents {
		length := ppDisplay.ppFont.GetWidthMonospaced(40*ppScale, "Total: ")
		pLength := ppDisplay.ppFont.GetWidthMonospaced(40*ppScale, ppDisplay.mText)

		position = position.Add(origin.AddS(1, 1).Mult(vector.NewVec2d(-(length+pLength)/2, -(160*ppScale)/2)))

		ppDisplay.drawPP(batch, "Aim:", ppDisplay.aimText, position, length, ppScale, color, vector.TopLeft)
		ppDisplay.drawPP(batch, "Tap:", ppDisplay.tapText, position.AddS(0, 40*ppScale), length, ppScale, color, vector.TopLeft)
		ppDisplay.drawPP(batch, "Acc:", ppDisplay.accText, position.AddS(0, 80*ppScale), length, ppScale, color, vector.TopLeft)

		offset := 0.0

		if ppDisplay.mods.Active(difficulty.Flashlight) {
			ppDisplay.drawPP(batch, "FL:", ppDisplay.flashlightText, position.AddS(0, 120*ppScale), length, ppScale, color, vector.TopLeft)

			offset = 40
		}

		ppDisplay.drawPP(batch, "Total:", ppDisplay.ppText, position.AddS(0, (120+offset)*ppScale), length, ppScale, color, vector.TopLeft)
	} else {
		ppDisplay.drawPP(batch, "", ppDisplay.ppText, position, 0, ppScale, color, origin)
	}

	batch.ResetTransform()
}

func (ppDisplay *PPDisplay) drawPP(batch *batch.QuadBatch, title, ppXText string, position vector.Vector2d, length float64, ppScale float64, color color2.Color, origin vector.Vector2d) {
	if title != "" {
		batch.SetColor(0, 0, 0, float64(color.A)*0.8)
		ppDisplay.ppFont.DrawOriginV(batch, position.AddS(ppScale, ppScale), origin, 40*ppScale, true, title)

		batch.SetColorM(color)
		ppDisplay.ppFont.DrawOriginV(batch, position, origin, 40*ppScale, true, title)
	}

	batch.SetColor(0, 0, 0, float64(color.A)*0.8)
	ppDisplay.ppFont.DrawOriginV(batch, position.AddS(ppScale+length, ppScale), origin, 40*ppScale, true, ppXText)

	batch.SetColorM(color)
	ppDisplay.ppFont.DrawOriginV(batch, position.AddS(length, 0), origin, 40*ppScale, true, ppXText)
}
