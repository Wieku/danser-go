package objects

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/animation"
	"github.com/wieku/danser-go/audio"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/settings"
	"strconv"
)

type Circle struct {
	objData      *basicData
	sample       int
	Timings      *Timings
	fadeApproach *animation.Glider
	fadeCircle   *animation.Glider
}

func NewCircle(data []string) *Circle {
	circle := &Circle{}
	circle.objData = commonParse(data)
	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.parseExtras(data, 5)
	circle.fadeCircle = animation.NewGlider(1)
	circle.fadeApproach = animation.NewGlider(1)
	return circle
}

func DummyCircle(pos bmath.Vector2d, time int64) *Circle {
	return DummyCircleInherit(pos, time, false, false, false)
}

func DummyCircleInherit(pos bmath.Vector2d, time int64, inherit bool, inheritStart bool, inheritEnd bool) *Circle {
	circle := &Circle{objData: &basicData{}}
	circle.objData.StartPos = pos
	circle.objData.EndPos = pos
	circle.objData.StartTime = time
	circle.objData.EndTime = time
	circle.objData.EndPos = circle.objData.StartPos
	circle.objData.SliderPoint = inherit
	circle.objData.SliderPointStart = inheritStart
	circle.objData.SliderPointEnd = inheritEnd
	return circle
}

func (self Circle) GetBasicData() *basicData {
	return self.objData
}

func (self *Circle) Update(time int64) bool {

	if (!settings.PLAY && settings.KNOCKOUT == "") || settings.PLAYERS > 1 {
		self.PlaySound()
	}
	/*index := self.objData.customIndex

	if index == 0 {
		index = self.Timings.Current.SampleIndex
	}

	if time < 2000+self.objData.EndTime {
		if self.objData.sampleSet == 0 {
			audio.PlaySample(self.Timings.Current.SampleSet, self.objData.additionSet, self.sample, index, self.Timings.Current.SampleVolume)
		} else {
			audio.PlaySample(self.objData.sampleSet, self.objData.additionSet, self.sample, index, self.Timings.Current.SampleVolume)
		}
	}*/

	return true
}

func (self *Circle) PlaySound() {
	index := self.objData.customIndex

	point := self.Timings.GetPoint(self.objData.StartTime)

	if index == 0 {
		index = point.SampleIndex
	}

	if self.objData.sampleSet == 0 {
		audio.PlaySample(point.SampleSet, self.objData.additionSet, self.sample, index, point.SampleVolume)
	} else {
		audio.PlaySample(self.objData.sampleSet, self.objData.additionSet, self.sample, index, point.SampleVolume)
	}
}

func (self *Circle) SetTiming(timings *Timings) {
	self.Timings = timings
}

func (self *Circle) SetDifficulty(preempt, fadeIn float64) {
	self.fadeCircle = animation.NewGlider(0)
	self.fadeCircle.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime)-(preempt-fadeIn), 1)
	self.fadeCircle.AddEvent(float64(self.objData.StartTime), float64(self.objData.StartTime)+fadeIn, 0)
	//self.fadeCircle.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime)-preempt*0.6, 1)
	//self.fadeCircle.AddEvent(float64(self.objData.StartTime)-preempt*0.6, float64(self.objData.StartTime)-preempt*0.3, 0) HIDDEN

	self.fadeApproach = animation.NewGlider(1)
	self.fadeApproach.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime), 0)
}

func (self *Circle) GetPosition() bmath.Vector2d {
	return self.objData.StartPos
}

func (self *Circle) Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) bool {
	self.fadeCircle.Update(float64(time))
	alpha := self.fadeCircle.GetValue()

	batch.SetTranslation(self.objData.StartPos)

	if time >= self.objData.StartTime {
		subScale := 1 + (1.0-alpha)*0.5
		batch.SetSubScale(subScale, subScale)
	} else {
		batch.SetSubScale(1, 1)
	}

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
		if time < self.objData.StartTime {
			if settings.DIVIDES < 2 && settings.Objects.DrawComboNumbers {
				render.Combo.DrawCentered(batch, self.objData.StartPos.X, self.objData.StartPos.Y, 0.65, strconv.Itoa(int(self.objData.ComboNumber)))
			}
			batch.SetTranslation(self.objData.StartPos)
		}
	}

	batch.SetSubScale(1, 1)

	if time >= self.objData.StartTime && self.fadeCircle.GetValue() <= 0.001 {
		return true
	}
	return false
}

func (self *Circle) DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	self.fadeApproach.Update(float64(time))
	arr := self.fadeApproach.GetValue()
	alpha := self.fadeCircle.GetValue()

	batch.SetTranslation(self.objData.StartPos)

	if settings.Objects.DrawApproachCircles && time <= self.objData.StartTime {
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
		batch.SetSubScale(1.0+arr*4, 1.0+arr*4)
		batch.DrawUnit(*render.ApproachCircle)
	}

	batch.SetSubScale(1, 1)
}
