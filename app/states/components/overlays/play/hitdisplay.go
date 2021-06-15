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

	hit300     int
	hit300Text string

	hit100     int
	hit100Text string

	hit50     int
	hit50Text string

	hitMiss     int
	hitMissText string
}

func NewHitDisplay(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor, fnt *font.Font) *HitDisplay {
	aSprite := &HitDisplay{
		ruleset:     ruleset,
		cursor:      cursor,
		fnt:         fnt,
		hit100Text:  "0",
		hit50Text:   "0",
		hitMissText: "0",
	}

	return aSprite
}

func (sprite *HitDisplay) Update(_ float64) {
	h300, h100, h50, hMiss, _, _ := sprite.ruleset.GetHits(sprite.cursor)

	if sprite.hit300 != int(h300) {
		sprite.hit300 = int(h300)
		sprite.hit300Text = strconv.Itoa(sprite.hit300)
	}

	if sprite.hit100 != int(h100) {
		sprite.hit100 = int(h100)
		sprite.hit100Text = strconv.Itoa(sprite.hit100)
	}

	if sprite.hit50 != int(h50) {
		sprite.hit50 = int(h50)
		sprite.hit50Text = strconv.Itoa(sprite.hit50)
	}

	if sprite.hitMiss != int(hMiss) {
		sprite.hitMiss = int(hMiss)
		sprite.hitMissText = strconv.Itoa(sprite.hitMiss)
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
	fontScale := scale * settings.Gameplay.HitCounter.FontScale

	alignX := vector.ParseOrigin(settings.Gameplay.HitCounter.Align).AddS(1, 1).X

	if settings.Gameplay.HitCounter.Show300 {
		alignX *= 1.5
	}

	valueAlign := vector.ParseOrigin(settings.Gameplay.HitCounter.ValueAlign)

	baseX := settings.Gameplay.HitCounter.XPosition - alignX*hSpacing
	baseY := settings.Gameplay.HitCounter.YPosition

	offsetI := 0

	if settings.Gameplay.HitCounter.Show300 {
		sprite.drawShadowed(batch, baseX, baseY, valueAlign, fontScale, 0, float32(alpha), sprite.hit300Text)

		offsetI = 1
		baseX += hSpacing
	}

	sprite.drawShadowed(batch, baseX, baseY, valueAlign, fontScale, offsetI, float32(alpha), sprite.hit100Text)
	sprite.drawShadowed(batch, baseX+hSpacing, baseY, valueAlign, fontScale, offsetI+1, float32(alpha), sprite.hit50Text)
	sprite.drawShadowed(batch, baseX+hSpacing*2, baseY, valueAlign, fontScale, offsetI+2, float32(alpha), sprite.hitMissText)

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
