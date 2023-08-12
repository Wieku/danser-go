package play

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type HpBar struct {
	healthBackground *sprite.Sprite
	healthBar        *sprite.Animation

	kiIcon *sprite.Sprite

	kiNormal  *texture.TextureRegion
	kiDanger  *texture.TextureRegion
	kiDanger2 *texture.TextureRegion

	currentHp      float64
	displayHp      float64
	lastTime       float64
	hpSlide        *animation.Glider
	hpFade         *animation.Glider
	hpBasePosition vector.Vector2d
	newStyle       bool
	explodes       *sprite.Manager
	kiPosY         float64
}

func NewHpBar() *HpBar {
	hpBar := new(HpBar)

	hpBar.healthBackground = sprite.NewSpriteSingle(skin.GetTexture("scorebar-bg"), 0, vector.NewVec2d(0, 0), vector.TopLeft)
	barTextures := skin.GetFrames("scorebar-colour", true) //nolint:misspell

	pos := vector.NewVec2d(4.8, 16)

	marker := skin.GetTextureSource("scorebar-marker", skin.GetSourceFromTexture(barTextures[0]))

	if marker != nil {
		hpBar.newStyle = true
		pos = vector.NewVec2d(12, 12.5)

		hpBar.kiNormal = marker
		hpBar.kiDanger = marker
		hpBar.kiDanger2 = marker

		hpBar.kiPosY = 17
	} else {
		hpBar.kiNormal = skin.GetTextureSource("scorebar-ki", skin.GetSourceFromTexture(barTextures[0]))
		hpBar.kiDanger = skin.GetTextureSource("scorebar-kidanger", skin.GetSourceFromTexture(barTextures[0]))
		hpBar.kiDanger2 = skin.GetTextureSource("scorebar-kidanger2", skin.GetSourceFromTexture(barTextures[0]))

		hpBar.kiPosY = 16
	}

	hpBar.hpBasePosition = pos

	hpBar.healthBar = sprite.NewAnimation(barTextures, skin.GetInfo().GetFrameTime(len(barTextures)), true, 0.0, pos, vector.TopLeft)
	hpBar.healthBar.SetCutOrigin(vector.CentreLeft)

	hpBar.kiIcon = sprite.NewSpriteSingle(nil, 0.0, vector.NewVec2d(0, 0), vector.Centre)

	hpBar.hpSlide = animation.NewGlider(0)
	hpBar.hpFade = animation.NewGlider(1)

	hpBar.explodes = sprite.NewManager()

	return hpBar
}

func (hpBar *HpBar) Update(time float64) {
	if hpBar.newStyle {
		additive := false

		var col color.Color
		if hpBar.displayHp < 0.2 {
			col = color.NewRGB(1, 0, 0).Mix(color.NewL(0), float32(hpBar.displayHp)/0.2)
		} else if hpBar.displayHp < 0.5 {
			col = color.NewL(0).Mix(color.NewL(1), float32(hpBar.displayHp)/0.5)
		} else {
			col = color.NewL(1)
			additive = true
		}

		hpBar.healthBar.SetColor(col)
		hpBar.kiIcon.SetColor(col)
		hpBar.kiIcon.SetAdditive(additive)
	}

	if hpBar.displayHp < 0.2 {
		hpBar.kiIcon.Texture = hpBar.kiDanger2
	} else if hpBar.displayHp < 0.5 {
		hpBar.kiIcon.Texture = hpBar.kiDanger
	} else {
		hpBar.kiIcon.Texture = hpBar.kiNormal
	}

	delta60 := (time - hpBar.lastTime) / 16.667

	if hpBar.displayHp < hpBar.currentHp {
		hpBar.displayHp = min(1.0, hpBar.displayHp+math.Abs(hpBar.currentHp-hpBar.displayHp)/4*delta60)
	} else if hpBar.displayHp > hpBar.currentHp {
		hpBar.displayHp = max(0.0, hpBar.displayHp-math.Abs(hpBar.displayHp-hpBar.currentHp)/6*delta60)
	}

	hpBar.kiIcon.SetPosition(vector.NewVec2d(hpBar.hpBasePosition.X, hpBar.kiPosY).AddS(float64(hpBar.healthBar.Texture.Width)*hpBar.displayHp, hpBar.hpSlide.GetValue()).Scl(settings.Gameplay.HpBar.Scale))

	hpBar.healthBar.SetCutX(0, 1.0-hpBar.displayHp)

	hpBar.healthBar.Update(time)
	hpBar.kiIcon.Update(time)

	hpBar.explodes.Update(time)

	hpBar.hpSlide.Update(time)
	hpBar.hpFade.Update(time)

	hpBar.lastTime = time
}

func (hpBar *HpBar) Draw(batch *batch.QuadBatch, alpha float64) {
	hpAlpha := settings.Gameplay.HpBar.Opacity * hpBar.hpFade.GetValue() * alpha

	if hpAlpha < 0.001 || !settings.Gameplay.HpBar.Show {
		return
	}

	hpScale := settings.Gameplay.HpBar.Scale

	batch.ResetTransform()

	batch.SetScale(hpScale, hpScale)
	batch.SetTranslation(vector.NewVec2d(settings.Gameplay.HpBar.XOffset, settings.Gameplay.HpBar.YOffset+hpBar.hpSlide.GetValue()))
	batch.SetColor(1, 1, 1, hpAlpha)

	hpBar.healthBackground.Draw(hpBar.lastTime, batch)

	hpBar.healthBar.SetPosition(hpBar.hpBasePosition.Scl(hpScale))
	hpBar.healthBar.Draw(hpBar.lastTime, batch)

	hpBar.kiIcon.Draw(hpBar.lastTime, batch)

	hpBar.explodes.Draw(hpBar.lastTime, batch)
}

func (hpBar *HpBar) SlideOut() {
	if settings.Gameplay.HpBar.YOffset < 0.01 {
		hpBar.hpSlide.AddEvent(hpBar.lastTime, hpBar.lastTime+500, -20)
	}

	hpBar.hpFade.AddEvent(hpBar.lastTime, hpBar.lastTime+500, 0)
}

func (hpBar *HpBar) SlideIn() {
	hpBar.hpSlide.AddEvent(hpBar.lastTime, hpBar.lastTime+500, 0)
	hpBar.hpFade.AddEvent(hpBar.lastTime, hpBar.lastTime+500, 1)
}

func (hpBar *HpBar) SetHp(hp float64) {
	if hp > hpBar.currentHp {
		hpBar.kiIcon.ClearTransformationsOfType(animation.Scale)
		hpBar.kiIcon.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, hpBar.lastTime, hpBar.lastTime+150, 1.2, 0.8))

		if hpBar.currentHp > 0.9 {
			eIcon := sprite.NewSpriteSingle(hpBar.kiNormal, 0.0, hpBar.kiIcon.GetPosition(), vector.Centre)
			eIcon.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, hpBar.lastTime, hpBar.lastTime+120, 1, 2))
			eIcon.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, hpBar.lastTime, hpBar.lastTime+120, 0.5, 0))
			eIcon.SetAdditive(true)
			eIcon.ShowForever(false)
			eIcon.AdjustTimesToTransformations()

			hpBar.explodes.Add(eIcon)
		}
	}

	hpBar.currentHp = hp
}
