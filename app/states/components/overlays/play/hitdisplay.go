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

func NewHitDisplay(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor, fnt *font.Font) *HitDisplay {
	aSprite := &HitDisplay{
		ruleset:          ruleset,
		cursor:           cursor,
		fnt:              fnt,
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
	if !settings.Gameplay.HitCounter.Show || settings.Gameplay.HitCounter.Opacity*alpha < 0.01 {
		return
	}

	batch.ResetTransform()

	alpha *= settings.Gameplay.HitCounter.Opacity
	scale := settings.Gameplay.HitCounter.Scale
	hSpacing := settings.Gameplay.HitCounter.Spacing * scale
	vSpacing := 0.0

	if settings.Gameplay.HitCounter.Vertical {
		vSpacing = hSpacing
		hSpacing = 0
	}

	fontScale := scale * settings.Gameplay.HitCounter.FontScale

	align := vector.ParseOrigin(settings.Gameplay.HitCounter.Align).AddS(1, 1).Scl(0.5)

	bC := 3.0

	if settings.Gameplay.HitCounter.Show300 {
		bC += 1.0
	}

	if settings.Gameplay.HitCounter.ShowSliderBreaks {
		bC += 1.0
	}

	valueAlign := vector.ParseOrigin(settings.Gameplay.HitCounter.ValueAlign)

	baseX := settings.Gameplay.HitCounter.XPosition - align.X*hSpacing*(bC-1)
	baseY := settings.Gameplay.HitCounter.YPosition - align.Y*vSpacing*(bC-1)

	offsetI := 0

	if settings.Gameplay.HitCounter.Show300 {
		sprite.drawShadowed(batch, baseX, baseY, valueAlign, fontScale, 0, float32(alpha), sprite.hit300Text)

		offsetI = 1
		baseX += hSpacing
		baseY += vSpacing
	}

	sprite.drawShadowed(batch, baseX, baseY, valueAlign, fontScale, offsetI, float32(alpha), sprite.hit100Text)
	sprite.drawShadowed(batch, baseX+hSpacing, baseY+vSpacing, valueAlign, fontScale, offsetI+1, float32(alpha), sprite.hit50Text)
	sprite.drawShadowed(batch, baseX+hSpacing*2, baseY+vSpacing*2, valueAlign, fontScale, offsetI+2, float32(alpha), sprite.hitMissText)

	if settings.Gameplay.HitCounter.ShowSliderBreaks {
		sprite.drawShadowed(batch, baseX+hSpacing*3, baseY+vSpacing*3, valueAlign, fontScale, offsetI+3, float32(alpha), sprite.sliderBreaksText)
	}

	batch.ResetTransform()
}

func (sprite *HitDisplay) drawShadowed(batch *batch.QuadBatch, x, y float64, origin vector.Vector2d, size float64, cI int, alpha float32, text string) {
	cS := settings.Gameplay.HitCounter.Color[cI%len(settings.Gameplay.HitCounter.Color)]
	color := color2.NewHSVA(float32(cS.Hue), float32(cS.Saturation), float32(cS.Value), alpha)

	batch.SetColor(0, 0, 0, float64(color.A)*0.8)
	sprite.fnt.DrawOrigin(batch, x+size, y+size, origin, 20*size, true, text)
	batch.SetColorM(color)
	sprite.fnt.DrawOrigin(batch, x, y, origin, 20*size, true, text)
}
