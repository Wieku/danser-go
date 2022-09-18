package play

import (
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"strconv"
)

type HitDisplay struct {
	ruleset *osu.OsuRuleSet
	cursor  *graphics.Cursor
	fnt     *font.Font

	hit300     uint
	hit300Text string

	hit100     uint
	hit100Text string

	hit50     uint
	hit50Text string

	hitMiss     uint
	hitMissText string

	sliderBreaks     uint
	sliderBreaksText string
}

func NewHitDisplay(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor) *HitDisplay {
	aSprite := &HitDisplay{
		ruleset:          ruleset,
		cursor:           cursor,
		fnt:              font.GetFont("HUDFont"),
		hit300Text:       "0",
		hit100Text:       "0",
		hit50Text:        "0",
		hitMissText:      "0",
		sliderBreaksText: "0",
	}

	return aSprite
}

func (sprite *HitDisplay) Update(_ float64) {
	score := sprite.ruleset.GetScore(sprite.cursor)

	if sprite.hit300 != score.Count300 {
		sprite.hit300 = score.Count300
		sprite.hit300Text = strconv.Itoa(int(sprite.hit300))
	}

	if sprite.hit100 != score.Count100 {
		sprite.hit100 = score.Count100
		sprite.hit100Text = strconv.Itoa(int(sprite.hit100))
	}

	if sprite.hit50 != score.Count50 {
		sprite.hit50 = score.Count50
		sprite.hit50Text = strconv.Itoa(int(sprite.hit50))
	}

	if sprite.hitMiss != score.CountMiss {
		sprite.hitMiss = score.CountMiss
		sprite.hitMissText = strconv.Itoa(int(sprite.hitMiss))
	}

	if sprite.sliderBreaks != score.CountSB {
		sprite.sliderBreaks = score.CountSB
		sprite.sliderBreaksText = strconv.Itoa(int(sprite.sliderBreaks))
	}
}

func (sprite *HitDisplay) Draw(batch *batch.QuadBatch, alpha float64) {
	hCS := settings.Gameplay.HitCounter

	if !hCS.Show || hCS.Opacity*alpha < 0.01 {
		return
	}

	batch.ResetTransform()

	alpha *= hCS.Opacity
	scale := hCS.Scale
	hSpacing := hCS.Spacing * scale
	vSpacing := 0.0

	if hCS.Vertical {
		vSpacing = hSpacing
		hSpacing = 0
	}

	fontScale := scale * hCS.FontScale

	align := vector.ParseOrigin(hCS.Align).AddS(1, 1).Scl(0.5)

	bC := 3.0

	if hCS.Show300 {
		bC += 1.0
	}

	if hCS.ShowSliderBreaks {
		bC += 1.0
	}

	valueAlign := vector.ParseOrigin(hCS.ValueAlign)

	baseX := hCS.XPosition - align.X*hSpacing*(bC-1)
	baseY := hCS.YPosition - align.Y*vSpacing*(bC-1)

	if hCS.Show300 {
		sprite.drawShadowed(batch, baseX, baseY, valueAlign, fontScale, hCS.Color300, float32(alpha), sprite.hit300Text)

		baseX += hSpacing
		baseY += vSpacing
	}

	sprite.drawShadowed(batch, baseX, baseY, valueAlign, fontScale, hCS.Color100, float32(alpha), sprite.hit100Text)
	sprite.drawShadowed(batch, baseX+hSpacing, baseY+vSpacing, valueAlign, fontScale, hCS.Color50, float32(alpha), sprite.hit50Text)
	sprite.drawShadowed(batch, baseX+hSpacing*2, baseY+vSpacing*2, valueAlign, fontScale, hCS.ColorMiss, float32(alpha), sprite.hitMissText)

	if hCS.ShowSliderBreaks {
		sprite.drawShadowed(batch, baseX+hSpacing*3, baseY+vSpacing*3, valueAlign, fontScale, hCS.ColorSB, float32(alpha), sprite.sliderBreaksText)
	}

	batch.ResetTransform()
}

func (sprite *HitDisplay) drawShadowed(batch *batch.QuadBatch, x, y float64, origin vector.Vector2d, size float64, color *settings.HSV, alpha float32, text string) {
	rgba := color2.NewHSVA(float32(color.Hue), float32(color.Saturation), float32(color.Value), alpha)

	batch.SetColor(0, 0, 0, float64(rgba.A)*0.8)
	sprite.fnt.DrawOrigin(batch, x+size, y+size, origin, 20*size, true, text)
	batch.SetColorM(rgba)
	sprite.fnt.DrawOrigin(batch, x, y, origin, 20*size, true, text)
}
