package objects

import (
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/audio"
	"strconv"
	"strings"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/settings"
	"github.com/go-gl/gl/v3.3-core/gl"
)

type Circle struct {
	objData *basicData
	sample int
	Timings *Timings
	ownSampleSet int
}

func NewCircle(data []string) *Circle {
	circle := &Circle{}
	circle.objData = commonParse(data)
	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	if len(data) > 5 {
		e, _ := strconv.ParseInt(strings.Split(data[5],":")[0], 10, 64)
		circle.ownSampleSet = int(e)
	} else {
		circle.ownSampleSet = 0
	}
	return circle
}

func DummyCircle(pos bmath.Vector2d, time int64) *Circle {
	circle := &Circle{objData:&basicData{}}
	circle.objData.StartPos = pos
	circle.objData.EndPos = pos
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	return circle
}

func (self Circle) GetBasicData() *basicData {
	return self.objData
}

func (self *Circle) Update(time int64, cursor *render.Cursor) bool {

	cursor.SetPos(self.objData.StartPos)
	if self.ownSampleSet == 0 {
		audio.PlaySample(self.Timings.Current.SampleSet, self.sample)
	} else {
		audio.PlaySample(self.ownSampleSet, self.sample)
	}

	return true
}

func (self *Circle) SetTiming(timings *Timings) {
	self.Timings = timings
}

func BeginCircleRender() {
	gl.ActiveTexture(gl.TEXTURE0)
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		render.CircleFull.Begin()
	} else {
		render.Circle.Begin()
	}

	gl.ActiveTexture(gl.TEXTURE1)
	render.CircleOverlay.Begin()

	gl.ActiveTexture(gl.TEXTURE2)
	render.ApproachCircle.Begin()
}

func EndCircleRender() {
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		render.CircleFull.End()
	} else {
		render.Circle.End()
	}

	render.CircleOverlay.End()
	render.ApproachCircle.End()
}

func (self *Circle) Render(time int64, preempt float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {

	alpha := 1.0
	arr := float64(self.objData.StartTime-time) / preempt

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha = float64(time - (self.objData.StartTime-int64(preempt)))/(preempt/2)
	} else if time >= self.objData.StartTime {
		alpha = 1.0-float64(time - self.objData.StartTime)/(preempt/2)
	} else {
		alpha = float64(color[3])
	}

	batch.SetTranslation(self.objData.StartPos)

	if time >= self.objData.StartTime {
		batch.SetSubScale(1+(1.0-alpha)*0.5, 1+(1.0-alpha)*0.5)
	}

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
	batch.DrawUnitR(0)

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {
		batch.SetColor(1, 1, 1, alpha)
		batch.DrawUnitR(1)

		if settings.Objects.DrawApproachCircles && time <= self.objData.StartTime {
			batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
			batch.SetSubScale(1.0+arr*2, 1.0+arr*2)
			batch.DrawUnitR(2)
		}

	}

	if time >= self.objData.StartTime+int64(preempt/2) {
		return true
	}
	return false
}