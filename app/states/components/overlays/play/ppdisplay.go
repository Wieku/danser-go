package play

import (
	"fmt"
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
	ppGlider *animation.TargetGlider
	ppText   string
}

func NewPPDisplay() *PPDisplay {
	return &PPDisplay{
		ppGlider: animation.NewTargetGlider(0, 0),
		ppText:   "0pp",
	}
}

func (ppDisplay *PPDisplay) Add(results performance.PPv2Results) {
	ppDisplay.ppGlider.SetTarget(results.Total)
}

func (ppDisplay *PPDisplay) Update(time float64) {
	ppDisplay.ppGlider.SetDecimals(settings.Gameplay.PPCounter.Decimals)
	ppDisplay.ppGlider.Update(time)
	ppDisplay.ppText = fmt.Sprintf("%."+strconv.Itoa(settings.Gameplay.PPCounter.Decimals)+"fpp", ppDisplay.ppGlider.GetValue())
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

	fnt := font.GetFont("Quicksand Bold")

	batch.SetColor(0, 0, 0, ppAlpha*0.8)
	fnt.DrawOriginV(batch, position.AddS(ppScale, ppScale), origin, 40*ppScale, true, ppDisplay.ppText)

	cS := settings.Gameplay.PPCounter.Color
	batch.SetColorM(color2.NewHSVA(float32(cS.Hue), float32(cS.Saturation), float32(cS.Value), float32(ppAlpha)))
	fnt.DrawOriginV(batch, position, origin, 40*ppScale, true, ppDisplay.ppText)
}
