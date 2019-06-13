package osu

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/animation"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/difficulty"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/settings"
)

type HitCircleSprite struct {
	fadeApproach  *animation.Glider
	scaleApproach *animation.Glider
	fadeCircle    *animation.Glider
	scaleCircle   *animation.Glider
	comboFade     *animation.Glider
	shake         *animation.Glider
	number        string
	startTime     int64
	Position      bmath.Vector2d
	diff          difficulty.Difficulty
}

func NewHitCircleSprite(diff difficulty.Difficulty, position bmath.Vector2d, startTime int64) *HitCircleSprite {
	sprite := new(HitCircleSprite)
	sprite.fadeApproach = animation.NewGlider(0.0)
	sprite.scaleApproach = animation.NewGlider(4)
	sprite.fadeCircle = animation.NewGlider(0.0)
	sprite.scaleCircle = animation.NewGlider(1.0)
	sprite.comboFade = animation.NewGlider(1.0)
	sprite.shake = animation.NewGlider(0.0)
	sprite.diff = diff
	sprite.startTime = startTime
	sprite.Position = position

	if (diff.Mods & difficulty.Hidden) > 0 {
		sprite.fadeCircle.AddEvent(float64(startTime)-diff.Preempt, float64(startTime)-diff.Preempt*0.6, 1)
		sprite.fadeCircle.AddEvent(float64(startTime)-diff.Preempt*0.6, float64(startTime)-diff.Preempt*0.3, 0)
	} else {
		sprite.fadeCircle.AddEvent(float64(startTime)-diff.Preempt, float64(startTime)-(diff.Preempt-diff.FadeIn), 1)
		sprite.fadeApproach = animation.NewGlider(0)
		sprite.fadeApproach.AddEvent(float64(startTime)-diff.Preempt, float64(startTime)-(diff.Preempt-diff.FadeIn), 1)
		sprite.fadeApproach.AddEvent(float64(startTime), float64(startTime), 0)
		sprite.scaleApproach.AddEvent(float64(startTime)-diff.Preempt, float64(startTime), 1)
	}

	return sprite
}

func (self *HitCircleSprite) Update(time int64) {
	self.fadeApproach.Update(float64(time))
	self.scaleApproach.Update(float64(time))
	self.fadeCircle.Update(float64(time))
	self.scaleCircle.Update(float64(time))
	self.comboFade.Update(float64(time))
	self.shake.Update(float64(time))
}

func (self *HitCircleSprite) Hit(time int64) {
	self.comboFade.AddEventS(float64(time), float64(time+60), 1.0, 0.0)
	self.fadeCircle.AddEventS(float64(time), float64(time+PostEmpt), 1.0, 0.0)
	self.scaleCircle.AddEventS(float64(time), float64(time+PostEmpt), 1.0, 1.5)
}

func (self *HitCircleSprite) Shake(time int64) {
	self.shake.AddEventS(float64(time), float64(time+25), 0, 10)
	self.shake.AddEventS(float64(time+25), float64(time+50), 10, -10)
	self.shake.AddEventS(float64(time+50), float64(time+75), -10, 0)
}

func (self *HitCircleSprite) Miss(time int64) {
	self.comboFade.AddEventS(float64(time), float64(time+60), 1.0, 0.0)
	self.fadeCircle.AddEventS(float64(time), float64(time+60), 1.0, 0.0)
}

func (self *HitCircleSprite) Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	alpha := self.fadeCircle.GetValue()

	batch.SetTranslation(self.Position.AddS(self.shake.GetValue(), 0))

	/*if time >= self.objData.StartTime {
		subScale := 1 + (1.0-alpha)*0.5
		batch.SetSubScale(subScale, subScale)
	} else {
		batch.SetSubScale(1, 1)
	}*/

	batch.SetSubScale(self.scaleCircle.GetValue(), self.scaleCircle.GetValue())

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		batch.DrawUnit(*render.CircleFull)
	} else {
		batch.DrawUnit(*render.Circle)
	}

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {
		batch.SetColor(1, 1, 1, alpha)
		batch.DrawUnit(*render.CircleOverlay)
		if /*time < self.startTime && */ len(self.number) > 0 {
			render.Combo.DrawCentered(batch, self.Position.X, self.Position.Y, 0.65, self.number)
			batch.SetTranslation(self.Position)
		}
	}

	batch.SetSubScale(1, 1)
}

func (self *HitCircleSprite) DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	batch.SetTranslation(self.Position.AddS(self.shake.GetValue(), 0))

	if settings.Objects.DrawApproachCircles {
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), self.fadeApproach.GetValue())
		batch.SetSubScale(self.scaleApproach.GetValue(), self.scaleApproach.GetValue())
		batch.DrawUnit(*render.ApproachCircle)
	}

	batch.SetSubScale(1, 1)
}
