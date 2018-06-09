package objects

import (
	"github.com/wieku/danser/bmath/sliders"
	m2 "github.com/wieku/danser/bmath"
	"strconv"
	"strings"
	"log"
	"github.com/wieku/danser/audio"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser/settings"
	"github.com/faiface/glhf"
)

type Slider struct {
	objData     *basicData
	multiCurve  sliders.SliderAlgo
	Timings     *Timings
	pixelLength float64
	partLen 	float64
	repeat      int64
	clicked     bool
	sampleSets 	[]int
	samples 	[]int
	lastT		int64
	Pos 		m2.Vector2d
	divides 	int
	LasTTI int64
	End bool
	vao *glhf.VertexSlice
}

func NewSlider(data []string) *Slider {
	slider := &Slider{clicked: false}
	slider.objData = commonParse(data)
	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.repeat, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []m2.Vector2d{slider.objData.StartPos}

	if list[0] == "C" {
		return nil
	}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 64)
		y, _ := strconv.ParseFloat(list2[1], 64)
		points = append(points, m2.NewVec2d(x, y))
	}

	slider.multiCurve = sliders.NewSliderAlgo(list[0], points)

	slider.objData.EndTime = slider.objData.StartTime
	slider.objData.EndPos = slider.multiCurve.PointAt(float64(slider.repeat%2))
	slider.Pos = slider.objData.StartPos

	slider.samples = make([]int, slider.repeat+1)
	slider.sampleSets = make([]int, slider.repeat+1)
	slider.lastT = 1
	if len(data) > 8 {
		subData := strings.Split(data[8], "|")
		for i, v := range subData {
			f, _ := strconv.ParseInt(v, 10, 64)
			slider.samples[i] = int(f)
		}
	}

	if len(data) > 9 {
		subData := strings.Split(data[9], "|")
		for i, v := range subData {
			f, _ := strconv.ParseInt(strings.Split(v,":")[0], 10, 64)
			slider.sampleSets[i] = int(f)
		}
	}
	slider.End = false
	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() m2.Vector2d {
	return self.multiCurve.PointAt(0.5).Add(self.objData.StackOffset)
}

func (self Slider) GetStartAngle() float64 {
	return self.GetBasicData().StartPos.AngleRV(self.multiCurve.PointAt(0.02).Add(self.objData.StackOffset))
}

func (self Slider) GetEndAngle() float64 {
	return self.GetBasicData().EndPos.AngleRV(self.multiCurve.PointAt(0.98 - float64(1-self.repeat%2)*0.96).Add(self.objData.StackOffset))
}

func (self Slider) GetPartLen() float64 {
	return 20.0 / float64(self.Timings.GetSliderTime(self.pixelLength)) * self.pixelLength
}

func (self Slider) GetPointAt(time int64) m2.Vector2d {
	partLen := float64(self.Timings.GetSliderTimeS(time, self.pixelLength))
	times := int64(float64(time - self.objData.StartTime) / partLen) + 1

	ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * partLen

	rt := float64(self.pixelLength) / self.multiCurve.Length

	var pos m2.Vector2d
	if (times%2) == 1 {
		pos = self.multiCurve.PointAt(rt*ttime/partLen)
	} else {
		pos = self.multiCurve.PointAt((1.0 - ttime/partLen)*rt)
	}

	return pos.Add(self.objData.StackOffset)
}

func (self Slider) endTime() int64 {
	return self.objData.StartTime + self.repeat * self.Timings.GetSliderTime(self.pixelLength)
}

func (self *Slider) SetTiming(timings *Timings) {
	self.Timings = timings
	if timings.GetSliderTimeS(self.objData.StartTime, self.pixelLength) < 0 {
		log.Println( self.objData.StartTime, self.pixelLength, "wuuuuuuuuuuuuuut")
	}
	self.objData.EndTime = self.objData.StartTime + timings.GetSliderTimeS(self.objData.StartTime, self.pixelLength) * self.repeat
}

func (self *Slider) GetCurve() []m2.Vector2d {
	t0 := 2 / self.pixelLength
	rt := float64(self.pixelLength) / self.multiCurve.Length
	points := make([]m2.Vector2d, int(self.pixelLength/2))
	t:= 0.0
	for i:=0; i < int(self.pixelLength/2); i+=1 {
		points[i] = self.multiCurve.PointAt(t*rt)
		t+=t0
	}
	return points
}

func (self *Slider) Update(time int64, cursor *render.Cursor) bool {
	//TODO: CLEAN THIS
	if time < self.endTime() {
		sliderTime := self.Timings.GetSliderTime(self.pixelLength)
		pixLen := self.multiCurve.Length
		self.partLen = float64(sliderTime)
		self.objData.EndTime = self.objData.StartTime + sliderTime * self.repeat
		times := int64(float64(time - self.objData.StartTime) / self.partLen) + 1

		ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * self.partLen

		if self.lastT != times {
			ss := self.sampleSets[times-1]
			if ss == 0 {
				ss = self.Timings.Current.SampleSet
			}
			audio.PlaySample(ss, self.samples[times-1])
			self.lastT = times
		}


		rt := float64(self.pixelLength) / pixLen

		var pos m2.Vector2d
		if (times%2) == 1 {
			pos = self.multiCurve.PointAt(rt*ttime/self.partLen)
		} else {
			pos = self.multiCurve.PointAt((1.0 - ttime/self.partLen)*rt)
		}
		self.Pos = pos
		cursor.SetPos(pos.Add(self.objData.StackOffset))
		//io.MouseMoveVec(pos.Add(self.objData.StackOffset))

		if !self.clicked {
			//io.MouseClick(io.LEFT)
			ss := self.sampleSets[0]
			if ss == 0 {
				ss = self.Timings.Current.SampleSet
			}
			audio.PlaySample(ss, self.samples[0])
			self.clicked = true
		}

		return false
	}

	ss := self.sampleSets[self.repeat]
	if ss == 0 {
		ss = self.Timings.Current.SampleSet
	}
	audio.PlaySample(ss, self.samples[self.repeat])
	self.End = true
	self.clicked = false
	//io.MouseUnClick(io.LEFT)

	return true
}

func (self *Slider) InitCurve(renderer *render.SliderRenderer) {
	self.vao, self.divides = renderer.GetShape(self.GetCurve())
}

func (self *Slider) Render(time int64, preempt float64) {
	in := 0
	out := int(self.pixelLength/2)

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha := float64(time - (self.objData.StartTime-int64(preempt)))/(preempt/2)
		out = int(float64(out)*alpha)
	} else if time >= self.objData.StartTime && time <= self.objData.EndTime {
		times := int64(float64(time - self.objData.StartTime) / self.partLen) + 1
		if times == self.repeat {
			ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * self.partLen
			alpha := 0.0
			if (times%2) == 1 {
				alpha = ttime/self.partLen
				in = int(float64(out)*alpha)
			} else {
				alpha = 1.0-ttime/self.partLen
				out = int(float64(out)*alpha)
			}
		}
	} else if time > self.objData.EndTime {
		in = out
	}

	if self.LasTTI > time {
		log.Println("WHAAAAAAT")
	}

	self.LasTTI = time

	if time <= self.objData.EndTime && !self.End{
		subVao := self.vao.Slice(in*self.divides*3, out*self.divides*3)
		subVao.Begin()
		subVao.Draw()
		subVao.End()
	}
}

func (self *Slider) RenderOverlay(time int64, preempt float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {

	gl.ActiveTexture(gl.TEXTURE0)
	if settings.DIVIDES > 2 {
		render.CircleFull.Begin()
	} else {
		render.Circle.Begin()
	}
	gl.ActiveTexture(gl.TEXTURE1)
	render.CircleOverlay.Begin()
	gl.ActiveTexture(gl.TEXTURE2)
	render.SliderBall.Begin()

	gl.ActiveTexture(gl.TEXTURE3)
	render.ApproachCircle.Begin()

	alpha := 1.0
	arr := float64(self.objData.StartTime-time) / preempt

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha = float64(time - (self.objData.StartTime-int64(preempt)))/(preempt/2)
	} else if time >= self.objData.EndTime {
		alpha = 1.0-float64(time - self.objData.EndTime)/(preempt/4)
	} else {
		alpha = float64(color[3])
	}

	if settings.DIVIDES > 2 {
		alpha *= 0.2
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

	if settings.DIVIDES <= 2 {
		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos)

			batch.DrawUnitR(0)
			batch.SetColor(1, 1, 1, alpha)
			batch.DrawUnitR(1)

			if settings.Objects.DrawApproachCircles && time <= self.objData.StartTime {
				batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
				batch.SetSubScale(1.0+arr*2, 1.0+arr*2)
				batch.DrawUnitR(3)
			}

		} else if time < self.objData.EndTime {
			batch.SetTranslation(self.Pos)
			batch.DrawUnitR(2)
		}
	} else {
		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos)
			batch.DrawUnitR(0)
		} else if time < self.objData.EndTime {
			batch.SetTranslation(self.Pos)
			batch.DrawUnitR(0)
		}
	}

	if time >= self.objData.EndTime+int64(preempt/4) {
		return true
	}
	return false
}