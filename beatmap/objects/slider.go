package objects

import (
	"github.com/wieku/danser/bmath/sliders"
	m2 "github.com/wieku/danser/bmath"
	"strconv"
	"strings"
	"github.com/wieku/danser/audio"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser/settings"
	"github.com/wieku/glhf"
	"math"
	"github.com/wieku/danser/utils"
	"sort"
)

type tickPoint struct {
	time int64
	Pos m2.Vector2d
}

type Slider struct {
	objData     *basicData
	multiCurve  sliders.SliderAlgo
	Timings     *Timings
	TPoint     TimingPoint
	pixelLength float64
	partLen 	float64
	repeat      int64
	clicked     bool
	sampleSets 	[]int
	additionSets 	[]int
	samples 	[]int
	lastT		int64
	Pos 		m2.Vector2d
	divides 	int
	TickPoints []tickPoint
	lastTick int
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

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 64)
		y, _ := strconv.ParseFloat(list2[1], 64)
		points = append(points, m2.NewVec2d(x, y))
	}

	slider.multiCurve = sliders.NewSliderAlgo(list[0], points)

	slider.objData.EndTime = slider.objData.StartTime
	slider.objData.EndPos = slider.objData.StartPos
	slider.Pos = slider.objData.StartPos

	slider.samples = make([]int, slider.repeat+1)
	slider.sampleSets = make([]int, slider.repeat+1)
	slider.additionSets = make([]int, slider.repeat+1)
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
			extras := strings.Split(v,":")
			sampleSet, _ := strconv.ParseInt(extras[0], 10, 64)
			additionSet, _ := strconv.ParseInt(extras[1], 10, 64)
			slider.sampleSets[i] = int(sampleSet)
			slider.additionSets[i] = int(additionSet)
		}
	}

	slider.objData.parseExtras(data, 10)

	slider.End = false
	slider.lastTick = -1
	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() m2.Vector2d {
	return self.multiCurve.PointAt(0.5).Add(self.objData.StackOffset)
}

func (self Slider) GetStartAngle() float64 {
	return self.GetBasicData().StartPos.AngleRV(self.GetPointAt(self.objData.StartTime+10)) //temporary solution
}

func (self Slider) GetEndAngle() float64 {
	return self.GetBasicData().EndPos.AngleRV(self.GetPointAt(self.objData.EndTime-10)) //temporary solution
}

func (self Slider) GetPartLen() float64 {
	return 20.0 / float64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)) * self.pixelLength
}

func (self Slider) GetPointAt(time int64) m2.Vector2d {
	partLen := float64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength))
	times := int64(math.Min(float64(time - self.objData.StartTime) / partLen + 1, float64(self.repeat)))

	ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * partLen

	rt := float64(self.pixelLength) / self.multiCurve.GetLength()

	var pos m2.Vector2d
	if (times%2) == 1 {
		pos = self.multiCurve.PointAt(rt*ttime/partLen)
	} else {
		pos = self.multiCurve.PointAt((1.0 - ttime/partLen)*rt)
	}

	return pos.Add(self.objData.StackOffset)
}

func (self *Slider) GetAsDummyCircles() []BaseObject {
	partLen := self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)

	var circles []BaseObject

	for i := int64(0); i <= self.repeat; i++ {
		time := self.objData.StartTime + i * partLen
		circles = append(circles, DummyCircleInherit(self.GetPointAt(time), time, true))
	}

	for _, p := range self.TickPoints {
		circles = append(circles, DummyCircleInherit(p.Pos, p.time, true))
	}

	sort.Slice(circles, func(i, j int) bool {return circles[i].GetBasicData().StartTime < circles[j].GetBasicData().StartTime})

	return circles
}

func (self Slider) endTime() int64 {
	return self.objData.StartTime + self.repeat * self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)
}

func (self *Slider) SetTiming(timings *Timings) {
	self.Timings = timings
	self.TPoint = timings.GetPoint(self.objData.StartTime)

	sliderTime := self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)
	self.partLen = float64(sliderTime)
	self.objData.EndTime = self.objData.StartTime + sliderTime * self.repeat
	self.objData.EndPos = self.GetPointAt(self.objData.EndTime)

	self.calculateFollowPoints()
}

func (self *Slider) calculateFollowPoints() {
	tickPixLen := (100.0*self.Timings.SliderMult)/(self.Timings.TickRate*self.TPoint.GetRatio())
	tickpoints := int(math.Ceil(self.pixelLength/tickPixLen))-1

	for r:=0; r < int(self.repeat); r++ {
		lengthFromEnd := self.pixelLength
		for i:=1; i <= tickpoints; i++ {
			time := self.objData.StartTime+int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))
			time2 := self.objData.StartTime+int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))

			if r%2 == 1 {
				time2 = self.objData.StartTime+self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)-int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))
			}

			lengthFromEnd -= tickPixLen

			if lengthFromEnd < 0.01*self.pixelLength {
				break
			}

			self.TickPoints = append(self.TickPoints, tickPoint{time2 + self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)*int64(r), self.GetPointAt(time)})
		}
	}

	sort.Slice(self.TickPoints, func(i, j int) bool {return self.TickPoints[i].time < self.TickPoints[j].time})
}

func (self *Slider) GetCurve() []m2.Vector2d {
	lod := float64(settings.Objects.SliderPathLOD) / 100.0
	t0 := (1.0 / lod) / self.pixelLength
	rt := float64(self.pixelLength) / self.multiCurve.GetLength()
	points := make([]m2.Vector2d, int(self.pixelLength*lod)+1)
	t:= 0.0
	for i:=0; i <= int(self.pixelLength*lod); i+=1 {
		points[i] = self.multiCurve.PointAt(t*rt)
		t+=t0
	}
	return points
}

func (self *Slider) Update(time int64) bool {
	if time < self.endTime() {
		times := int64(math.Min(float64(time - self.objData.StartTime) / self.partLen + 1, float64(self.repeat)))
		ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * self.partLen

		if self.lastT != times {
			self.playSample(self.sampleSets[times-1], self.additionSets[times-1], self.samples[times-1])
			self.lastT = times
		}

		for i, p := range self.TickPoints {
			if p.time < time && self.lastTick < i {
				audio.PlaySliderTick(self.Timings.Current.SampleSet, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume)
				self.lastTick = i
			}
		}

		rt := float64(self.pixelLength) / self.multiCurve.GetLength()

		var pos m2.Vector2d
		if (times%2) == 1 {
			pos = self.multiCurve.PointAt(rt*ttime/self.partLen)
		} else {
			pos = self.multiCurve.PointAt((1.0 - ttime/self.partLen)*rt)
		}
		self.Pos = pos.Add(self.objData.StackOffset)

		if !self.clicked {
			self.playSample(self.sampleSets[0], self.additionSets[0], self.samples[0])
			self.clicked = true
		}

		return false
	}

	self.playSample(self.sampleSets[self.repeat], self.additionSets[self.repeat], self.samples[self.repeat])
	self.End = true
	self.clicked = false

	return true
}

func (self *Slider) playSample(sampleSet, additionSet, sample int) {
	if sampleSet == 0 {
		sampleSet = self.Timings.Current.SampleSet
	}
	audio.PlaySample(sampleSet, additionSet, sample, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume)
}

func (self *Slider) GetPosition() m2.Vector2d {
	return self.Pos
}

func (self *Slider) InitCurve(renderer *render.SliderRenderer) {
	self.vao, self.divides = renderer.GetShape(self.GetCurve())
}

func (self *Slider) Render(time int64, preempt float64, color mgl32.Vec4, color1 mgl32.Vec4, renderer *render.SliderRenderer) {
	in := 0
	out := int(math.Ceil(self.pixelLength * float64(settings.Objects.SliderPathLOD) / 100.0))+1

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha := math.Abs(float64(time - (self.objData.StartTime-int64(preempt))))/(preempt/2)
		out = int(float64(out)*alpha)
	} else if time >= self.objData.StartTime && time <= self.objData.EndTime {
		times := int64(math.Min(float64(time - self.objData.StartTime) / self.partLen + 1, float64(self.repeat)))
		if times >= self.repeat {
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
		if (self.repeat%2) == 1 {
			in = out - 1
		} else {
			out = 1
		}
	}

	colorAlpha := 1.0

	if time < self.objData.StartTime-int64(preempt)/2 {
		colorAlpha = float64(time - (self.objData.StartTime-int64(preempt)))/(preempt/2)
	} else if time >= self.objData.EndTime {
		colorAlpha = 1.0-float64(time - self.objData.EndTime)/(preempt/4)
	} else {
		colorAlpha = float64(color[3])
	}

	renderer.SetColor(mgl32.Vec4{color[0], color[1], color[2], float32(colorAlpha)}, mgl32.Vec4{color1[0], color1[1], color1[2], float32(colorAlpha)})

	subVao := self.vao.Slice(in*self.divides*3, out*self.divides*3)
	subVao.BeginDraw()
	subVao.Draw()
	subVao.EndDraw()
}

func BeginSliderOverlay() {
	gl.ActiveTexture(gl.TEXTURE0)
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
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

	gl.ActiveTexture(gl.TEXTURE4)
	render.SliderTick.Begin()
}

func EndSliderOverlay() {
	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		render.CircleFull.End()
	} else {
		render.Circle.End()
	}

	render.CircleOverlay.End()
	render.SliderBall.End()
	render.ApproachCircle.End()
	render.SliderTick.End()
}

func (self *Slider) RenderOverlay(time int64, preempt float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {
	alpha := 1.0
	alphaF := 1.0
	arr := float64(self.objData.StartTime-time) / preempt

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha = float64(time - (self.objData.StartTime-int64(preempt)))/(preempt/2)
	} else if time >= self.objData.EndTime {
		alpha = 1.0-float64(time - self.objData.EndTime)/(preempt/4)
		alphaF = 1.0-float64(time - self.objData.StartTime)/(preempt/2)
	} else {
		alpha = float64(color[3])
		alphaF = 1.0-float64(time - self.objData.StartTime)/(preempt/2)
	}

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {

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

		} else {
			if settings.Objects.DrawFollowPoints && time < self.objData.EndTime {
				shifted := utils.GetColorShifted(color, settings.Objects.FollowPointColorOffset)

				for _, p := range self.TickPoints {
					al := 0.0
					if p.time > time {
						al = math.Min(1.0, math.Max( (float64(time)-(float64(p.time)-self.TPoint.Bpm*2))/(self.TPoint.Bpm), 0.0))
					}
					if al > 0.0 {
						batch.SetTranslation(p.Pos)
						batch.SetSubScale(1.0/5, 1.0/5)
						if settings.Objects.WhiteFollowPoints {
							batch.SetColor(1, 1, 1, alpha*al)
						} else {
							batch.SetColor(float64(shifted[0]), float64(shifted[1]), float64(shifted[2]), alpha*al)
						}

						batch.DrawUnitR(4)
					}
				}
			}

			if time >= self.objData.StartTime && alphaF > 0.0 {
				batch.SetTranslation(self.objData.StartPos)
				batch.SetSubScale(1+(1.0-alphaF)*0.5, 1+(1.0-alphaF)*0.5)
				batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alphaF)
				batch.DrawUnitR(0)
				batch.SetColor(1, 1, 1, alphaF)
				batch.DrawUnitR(1)
			}

			batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
			batch.SetSubScale(1.0, 1.0)
			batch.SetTranslation(self.Pos)
			batch.DrawUnitR(2)
		}
	} else {
		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos)
			batch.DrawUnitR(0)
		} else if time < self.objData.EndTime {
			batch.SetTranslation(self.Pos)

			if settings.Objects.ForceSliderBallTexture {
				batch.DrawUnitR(2)
			} else {
				batch.DrawUnitR(0)
			}
		}
	}

	if time >= self.objData.EndTime+int64(preempt/4) {
		return true
	}
	return false
}