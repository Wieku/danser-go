package objects

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
)

type Circle struct {
	objData *basicData
	sample  int
	Timings *Timings

	textFade *animation.Glider

	hitCircle        *sprite.Sprite
	hitCircleOverlay *sprite.Sprite
	approachCircle   *sprite.Sprite
	sprites          []*sprite.Sprite
	diff             *difficulty.Difficulty
	lastTime         int64
	silent           bool
}

func NewCircle(data []string) *Circle {
	circle := &Circle{}
	circle.objData = commonParse(data)
	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.parseExtras(data, 5)

	return circle
}

func DummyCircle(pos vector.Vector2f, time int64) *Circle {
	return DummyCircleInherit(pos, time, false, false, false)
}

func DummyCircleInherit(pos vector.Vector2f, time int64, inherit bool, inheritStart bool, inheritEnd bool) *Circle {
	circle := &Circle{objData: &basicData{}}
	circle.objData.StartPos = pos
	circle.objData.EndPos = pos
	circle.objData.StartTime = time
	circle.objData.EndTime = time
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.SliderPoint = inherit
	circle.objData.SliderPointStart = inheritStart
	circle.objData.SliderPointEnd = inheritEnd
	circle.silent = true
	return circle
}

func (self Circle) GetBasicData() *basicData {
	return self.objData
}

func (self *Circle) Update(time int64) bool {
	if !self.silent && ((!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1) && (self.lastTime < self.objData.StartTime && time >= self.objData.StartTime) {
		self.Arm(true, self.objData.StartTime)
		self.PlaySound()
	}

	for _, s := range self.sprites {
		s.Update(time)
	}

	if self.textFade != nil {
		self.textFade.Update(float64(time))
	}

	self.lastTime = time

	return true
}

func (self *Circle) PlaySound() {
	point := self.Timings.GetPoint(self.objData.StartTime)

	index := self.objData.customIndex
	sampleSet := self.objData.sampleSet

	if index == 0 {
		index = point.SampleIndex
	}

	if sampleSet == 0 {
		sampleSet = point.SampleSet
	}

	audio.PlaySample(sampleSet, self.objData.additionSet, self.sample, index, point.SampleVolume, self.objData.Number, self.GetBasicData().StartPos.X64())
}

func (self *Circle) SetTiming(timings *Timings) {
	self.Timings = timings
}

func (self *Circle) SetDifficulty(diff *difficulty.Difficulty) {
	self.diff = diff

	startTime := float64(self.objData.StartTime)

	self.textFade = animation.NewGlider(0)

	self.hitCircle = sprite.NewSpriteSingle(skin.GetTexture("hitcircle"), 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	self.hitCircleOverlay = sprite.NewSpriteSingle(skin.GetTexture("hitcircleoverlay"), 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	self.approachCircle = sprite.NewSpriteSingle(skin.GetTexture("approachcircle"), 0, vector.NewVec2d(0, 0), bmath.Origin.Centre)

	self.sprites = append(self.sprites, self.hitCircle)
	self.sprites = append(self.sprites, self.hitCircleOverlay)
	self.sprites = append(self.sprites, self.approachCircle)

	self.hitCircle.SetPosition(self.objData.StartPos.Copy64())
	self.hitCircle.SetAlpha(0)
	self.hitCircleOverlay.SetPosition(self.objData.StartPos.Copy64())
	self.hitCircleOverlay.SetAlpha(0)
	self.approachCircle.SetPosition(self.objData.StartPos.Copy64())
	self.approachCircle.SetAlpha(0)

	circles := []*sprite.Sprite{self.hitCircle, self.hitCircleOverlay}

	for _, t := range circles {
		if diff.CheckModActive(difficulty.Hidden) {
			t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime-diff.Preempt, startTime-diff.Preempt*0.6, 0.0, 1.0))
			t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime-diff.Preempt*0.6, startTime-diff.Preempt*0.3, 1.0, 0.0))
		} else {
			t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime-diff.Preempt, startTime-diff.Preempt+difficulty.HitFadeIn, 0.0, 1.0))
			t.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime+float64(diff.Hit100), startTime+float64(diff.Hit50), 1.0, 0.0))
		}
	}

	if diff.CheckModActive(difficulty.Hidden) {
		self.textFade.AddEventS(startTime-diff.Preempt, startTime-diff.Preempt*0.6, 0.0, 1.0)
		self.textFade.AddEventS(startTime-diff.Preempt*0.6, startTime-diff.Preempt*0.3, 1.0, 0.0)
	} else {
		self.textFade.AddEventS(startTime-diff.Preempt, startTime-diff.Preempt+difficulty.HitFadeIn, 0.0, 1.0)
		self.textFade.AddEventS(startTime+float64(diff.Hit100), startTime+float64(diff.Hit50), 1.0, 0.0)
	}

	if !diff.CheckModActive(difficulty.Hidden) || self.objData.Number == 0 {
		self.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime-diff.Preempt, math.Min(startTime, startTime-diff.Preempt+difficulty.HitFadeIn*2), 0.0, 0.9))
		if diff.CheckModActive(difficulty.Hidden) {
			self.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime-diff.Preempt*0.6, startTime-diff.Preempt*0.3, 0.9, 0.0))
		} else {
			self.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, startTime, 0.0, 0.0))
		}
		self.approachCircle.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, startTime-diff.Preempt, startTime, 4.0, 1.0))
	}
}

func (self *Circle) Arm(clicked bool, time int64) {
	self.hitCircle.ClearTransformations()
	self.hitCircleOverlay.ClearTransformations()
	self.textFade.Reset()

	startTime := float64(time)

	self.approachCircle.ClearTransformations()
	self.approachCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, startTime, 0.0, 0.0))

	if clicked && !self.diff.CheckModActive(difficulty.Hidden) {
		endTime := startTime + difficulty.HitFadeOut
		self.hitCircle.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, endTime, 1.0, 1.4))
		self.hitCircleOverlay.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, endTime, 1.0, 1.4))

		self.hitCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, 1.0, 0.0))
		self.hitCircleOverlay.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, 1.0, 0.0))
		self.textFade.AddEventS(startTime, startTime+60, 1.0, 0.0)
	} else {
		endTime := startTime + 60
		self.hitCircle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, self.hitCircle.GetAlpha(), 0.0))
		self.hitCircleOverlay.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, startTime, endTime, self.hitCircleOverlay.GetAlpha(), 0.0))
		self.textFade.AddEventS(startTime, endTime, self.textFade.GetValue(), 0.0)
	}
}

func (self *Circle) Shake(time int64) {
	startTime := float64(time)
	for _, s := range self.sprites {
		s.ClearTransformationsOfType(animation.MoveX)
		startPosX := float64(self.objData.StartPos.X)
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, startTime, startTime+20, startPosX, startPosX+8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, startTime+20, startTime+40, startPosX+8, startPosX-8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, startTime+40, startTime+60, startPosX-8, startPosX+8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, startTime+60, startTime+80, startPosX+8, startPosX-8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, startTime+80, startTime+100, startPosX-8, startPosX+8))
		s.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.Linear, startTime+100, startTime+120, startPosX+8, startPosX))
	}
}

func (self *Circle) UpdateStacking() {

}

func (self *Circle) GetPosition() vector.Vector2f {
	return self.objData.StartPos
}

func (self *Circle) Draw(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) bool {
	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))

	alpha := 1.0
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(1, 1, 1, alpha)

	//if settings.Objects.UseComboColors && len(settings.Objects.ComboColors) > 0 {
	//	cHSV := settings.Objects.ComboColors[int(self.objData.ComboSet)%len(settings.Objects.ComboColors)]
	//	r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
	//	self.hitCircle.SetColor(bmath.Color{R: float64(r), G: float64(g), B: float64(b), A: 1.0})
	//} else {
	//	self.hitCircle.SetColor(bmath.Color{R: float64(color.X()), G: float64(color.Y()), B: float64(color.Z()), A: 1.0})
	//}

	self.hitCircle.SetColor(bmath.Color{R: float64(color.X()), G: float64(color.Y()), B: float64(color.Z()), A: 1.0})

	self.hitCircle.Draw(time, batch)

	/*batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		batch.DrawUnit(*render.CircleFull)
	} else {
		batch.DrawUnit(*render.Circle)
	}*/

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {
		if !skin.GetInfo().HitCircleOverlayAboveNumber {
			self.hitCircleOverlay.Draw(time, batch)
		}

		if settings.DIVIDES < 2 && settings.Objects.DrawComboNumbers {
			fnt := skin.GetFont("default")
			batch.SetColor(1, 1, 1, alpha*self.textFade.GetValue())
			fnt.DrawCentered(batch, self.objData.StartPos.X64(), self.objData.StartPos.Y64(), 0.8*fnt.GetSize(), strconv.Itoa(int(self.objData.ComboNumber)))
		}

		batch.SetSubScale(1, 1)
		batch.SetTranslation(vector.NewVec2d(0, 0))

		if skin.GetInfo().HitCircleOverlayAboveNumber {
			self.hitCircleOverlay.Draw(time, batch)
		}
	}

	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))

	if time >= self.objData.StartTime && self.hitCircle.GetAlpha() <= 0.001 {
		return true
	}
	return false
}

func (self *Circle) DrawApproach(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) {
	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))
	batch.SetColor(1, 1, 1, 1)

	//if settings.Objects.UseComboColors && len(settings.Objects.ComboColors) > 0 {
	//	cHSV := settings.Objects.ComboColors[int(self.objData.ComboSet)%len(settings.Objects.ComboColors)]
	//	r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
	//	self.approachCircle.SetColor(bmath.Color{R: float64(r), G: float64(g), B: float64(b), A: 1.0})
	//} else {
	//	self.approachCircle.SetColor(bmath.Color{R: float64(color.X()), G: float64(color.Y()), B: float64(color.Z()), A: 1.0})
	//}
	self.approachCircle.SetColor(bmath.Color{R: float64(color.X()), G: float64(color.Y()), B: float64(color.Z()), A: 1.0})

	self.approachCircle.Draw(time, batch)
}
