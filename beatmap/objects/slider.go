package objects

import (
	"github.com/wieku/danser/bmath/sliders"
	"github.com/wieku/danser/bmath"
	"strconv"
	"strings"
	"github.com/wieku/danser/audio"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/settings"
	"github.com/wieku/glhf"
	"math"
	"sort"
	"github.com/faiface/mainthread"
	"runtime"
	"github.com/wieku/danser/render/batches"
	"github.com/wieku/danser/utils"
	"github.com/wieku/danser/animation"
)

type tickPoint struct {
	time int64
	Pos  bmath.Vector2d
}

type Slider struct {
	objData              *basicData
	multiCurve           sliders.SliderAlgo
	Timings              *Timings
	TPoint               TimingPoint
	pixelLength          float64
	partLen              float64
	repeat               int64
	clicked              bool
	sampleSets           []int
	additionSets         []int
	samples              []int
	lastT                int64
	Pos                  bmath.Vector2d
	divides              int
	TickPoints           []tickPoint
	lastTick             int
	End                  bool
	vao                  *glhf.VertexSlice
	created              bool
	discreteCurve        []bmath.Vector2d
	startAngle, endAngle float64
	sliderSnakeIn        *animation.Glider
	sliderSnakeOut       *animation.Glider
	fade                 *animation.Glider
	fadeApproach         *animation.Glider
	fadeCircle           *animation.Glider
}

func NewSlider(data []string) *Slider {
	slider := &Slider{clicked: false}
	slider.objData = commonParse(data)
	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.repeat, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []bmath.Vector2d{slider.objData.StartPos}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 64)
		y, _ := strconv.ParseFloat(list2[1], 64)
		points = append(points, bmath.NewVec2d(x, y))
	}

	slider.multiCurve = sliders.NewSliderAlgo(list[0], points, slider.pixelLength)

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
			extras := strings.Split(v, ":")
			sampleSet, _ := strconv.ParseInt(extras[0], 10, 64)
			additionSet, _ := strconv.ParseInt(extras[1], 10, 64)
			slider.sampleSets[i] = int(sampleSet)
			slider.additionSets[i] = int(additionSet)
		}
	}

	slider.objData.parseExtras(data, 10)

	slider.End = false
	slider.lastTick = -1
	slider.fade = animation.NewGlider(1)
	slider.fadeCircle = animation.NewGlider(1)
	slider.fadeApproach = animation.NewGlider(1)
	slider.sliderSnakeIn = animation.NewGlider(1)
	slider.sliderSnakeOut = animation.NewGlider(0)
	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() bmath.Vector2d {
	return self.multiCurve.PointAt(0.5).Add(self.objData.StackOffset)
}

func (self Slider) GetStartAngle() float64 {
	return self.GetBasicData().StartPos.AngleRV(self.GetPointAt(self.objData.StartTime + 10)) //temporary solution
}

func (self Slider) GetEndAngle() float64 {
	return self.GetBasicData().EndPos.AngleRV(self.GetPointAt(self.objData.EndTime - 10)) //temporary solution
}

func (self Slider) GetPartLen() float64 {
	return 20.0 / float64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)) * self.pixelLength
}

func (self Slider) GetPointAt(time int64) bmath.Vector2d {
	times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))

	ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1)*self.partLen

	var pos bmath.Vector2d
	if (times % 2) == 1 {
		pos = self.multiCurve.PointAt(ttime / self.partLen)
	} else {
		pos = self.multiCurve.PointAt(1.0 - ttime/self.partLen)
	}

	return pos.Add(self.objData.StackOffset)
}

func (self *Slider) GetAsDummyCircles() []BaseObject {
	partLen := self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)

	var circles []BaseObject

	for i := int64(0); i <= self.repeat; i++ {
		time := self.objData.StartTime + i*partLen
		circles = append(circles, DummyCircleInherit(self.GetPointAt(time), time, true))
	}

	for _, p := range self.TickPoints {
		circles = append(circles, DummyCircleInherit(p.Pos, p.time, true))
	}

	sort.Slice(circles, func(i, j int) bool { return circles[i].GetBasicData().StartTime < circles[j].GetBasicData().StartTime })

	return circles
}

func (self *Slider) SetTiming(timings *Timings) {
	self.Timings = timings
	self.TPoint = timings.GetPoint(self.objData.StartTime)

	sliderTime := self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)
	self.partLen = float64(sliderTime)
	self.objData.EndTime = self.objData.StartTime + sliderTime*self.repeat
	self.objData.EndPos = self.GetPointAt(self.objData.EndTime)

	self.calculateFollowPoints()
	self.discreteCurve = self.GetCurve()
	self.startAngle = self.GetStartAngle()
	self.endAngle = self.discreteCurve[len(self.discreteCurve)-1].AngleRV(self.discreteCurve[len(self.discreteCurve)-2])
}

func (self *Slider) calculateFollowPoints() {
	tickPixLen := (100.0 * self.Timings.SliderMult) / (self.Timings.TickRate * self.TPoint.GetRatio())
	tickpoints := int(math.Ceil(self.pixelLength/tickPixLen)) - 1

	for r := 0; r < int(self.repeat); r++ {
		lengthFromEnd := self.pixelLength
		for i := 1; i <= tickpoints; i++ {
			time := self.objData.StartTime + int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))
			time2 := self.objData.StartTime + int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))

			if r%2 == 1 {
				time2 = self.objData.StartTime + self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength) - int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))
			}

			lengthFromEnd -= tickPixLen

			if lengthFromEnd < 0.01*self.pixelLength {
				break
			}

			self.TickPoints = append(self.TickPoints, tickPoint{time2 + self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)*int64(r), self.GetPointAt(time)})
		}
	}

	sort.Slice(self.TickPoints, func(i, j int) bool { return self.TickPoints[i].time < self.TickPoints[j].time })
}

func (self *Slider) SetDifficulty(preempt, fadeIn float64) {
	self.sliderSnakeIn = animation.NewGlider(0)
	self.sliderSnakeOut = animation.NewGlider(0)
	if settings.Objects.SliderSnakeIn {
		self.sliderSnakeIn.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime)-(preempt-fadeIn), 1)
	}

	if settings.Objects.SliderSnakeOut {
		self.sliderSnakeOut.AddEvent(float64(self.objData.EndTime)-self.partLen, float64(self.objData.EndTime), 1)
	}
	self.fade = animation.NewGlider(0)
	self.fade.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime)-(preempt-fadeIn), 1)
	self.fade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+fadeIn, 0)

	self.fadeCircle = animation.NewGlider(0)
	self.fadeCircle.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime)-(preempt-fadeIn), 1)
	self.fadeCircle.AddEvent(float64(self.objData.StartTime), float64(self.objData.StartTime)+fadeIn, 0)

	self.fadeApproach = animation.NewGlider(1)
	self.fadeApproach.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime), 0)
}

func (self *Slider) GetCurve() []bmath.Vector2d {
	lod := math.Ceil(self.pixelLength * float64(settings.Objects.SliderPathLOD) / 100.0)
	t0 := 1.0 / lod
	points := make([]bmath.Vector2d, int(lod)+1)
	t := 0.0
	for i := 0; i <= int(lod); i += 1 {
		points[i] = self.multiCurve.PointAt(t)
		t += t0
	}
	return points
}

func (self *Slider) Update(time int64) bool {
	if time < self.objData.EndTime {
		times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))

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

		self.Pos = self.GetPointAt(time)

		if !self.clicked {
			self.playSample(self.sampleSets[0], self.additionSets[0], self.samples[0])
			self.clicked = true
		}

		return false
	}

	self.Pos = self.GetPointAt(self.objData.EndTime)

	self.playSample(self.sampleSets[self.repeat], self.additionSets[self.repeat], self.samples[self.repeat])
	self.End = true
	self.clicked = false

	return true
}

func (self *Slider) playSample(sampleSet, additionSet, sample int) {
	if sampleSet == 0 {
		sampleSet = self.objData.sampleSet
		if sampleSet == 0 {
			sampleSet = self.Timings.Current.SampleSet
		}
	}

	if additionSet == 0 {
		additionSet = self.objData.additionSet
	}

	audio.PlaySample(sampleSet, additionSet, sample, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume)
}

func (self *Slider) GetPosition() bmath.Vector2d {
	return self.Pos
}

func (self *Slider) InitCurve(renderer *render.SliderRenderer) {
	if !self.created {
		self.created = true
		go func() {
			var data []float32
			data, self.divides = renderer.GetShape(self.discreteCurve)
			mainthread.CallNonBlock(func() {
				self.vao = renderer.UploadMesh(data)
				runtime.KeepAlive(data)
			})
		}()
	}
}

func (self *Slider) DrawBody(time int64, color mgl32.Vec4, color1 mgl32.Vec4, renderer *render.SliderRenderer) {
	self.sliderSnakeIn.Update(float64(time))
	self.sliderSnakeOut.Update(float64(time))
	self.fade.Update(float64(time))

	in := 0
	out := int(self.sliderSnakeIn.GetValue() * float64(len(self.discreteCurve)))

	if int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat))) == self.repeat {
		if (self.repeat % 2) == 1 {
			in = int(self.sliderSnakeOut.GetValue() * (float64(len(self.discreteCurve)) - 1))
		} else {
			out = int((1.0 - self.sliderSnakeOut.GetValue()) * (float64(len(self.discreteCurve)) - 1))
		}
	}

	colorAlpha := self.fade.GetValue()

	renderer.SetColor(mgl32.Vec4{color[0], color[1], color[2], float32(colorAlpha)}, mgl32.Vec4{color1[0], color1[1], color1[2], float32(colorAlpha)})

	if self.vao != nil {
		subVao := self.vao.Slice(in*self.divides*3, out*self.divides*3)
		subVao.BeginDraw()
		subVao.Draw()
		subVao.EndDraw()
	}
}

func (self *Slider) Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) bool {
	self.fade.Update(float64(time))
	self.fadeCircle.Update(float64(time))
	alpha := self.fade.GetValue()
	alphaF := self.fadeCircle.GetValue()

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {

		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos)

			batch.DrawUnit(*render.Circle)
			batch.SetColor(1, 1, 1, alpha)
			batch.DrawUnit(*render.CircleOverlay)
			render.Combo.DrawCentered(batch, self.objData.StartPos.X, self.objData.StartPos.Y, 0.65, strconv.Itoa(int(self.objData.ComboNumber)))
			batch.SetSubScale(1, 1)
			/*if self.repeat > 1 {
				batch.SetTranslation(self.discreteCurve[len(self.discreteCurve)-1])
				batch.SetRotation(self.endAngle + math.Pi)
				batch.DrawUnit(*render.SliderReverse)
				batch.SetRotation(0)
			}

			batch.SetTranslation(self.objData.StartPos)*/

		} else {
			if time < self.objData.EndTime {
				/*times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))

				batch.SetColor(1, 1, 1, alpha)
				acv := func(rp int64) {
					if rp%2 == 0 {
						batch.SetTranslation(self.discreteCurve[len(self.discreteCurve)-1])
						batch.SetRotation(self.endAngle + math.Pi)
					} else {
						batch.SetTranslation(self.discreteCurve[0])
						batch.SetRotation(self.startAngle + math.Pi)
					}
					batch.DrawUnit(*render.SliderReverse)
				}

				if times < self.repeat {
					acv(self.repeat)
				}

				if times < self.repeat-1 {
					acv(self.repeat - 1)
				}

				batch.SetRotation(0)*/

				if settings.Objects.DrawFollowPoints {
					shifted := utils.GetColorShifted(color, settings.Objects.FollowPointColorOffset)

					for _, p := range self.TickPoints {
						al := 0.0
						if p.time > time {
							al = math.Min(1.0, math.Max((float64(time)-(float64(p.time)-self.TPoint.Bpm*2))/(self.TPoint.Bpm), 0.0))
						}
						if al > 0.0 {
							batch.SetTranslation(p.Pos)
							batch.SetSubScale(1.0/5, 1.0/5)
							if settings.Objects.WhiteFollowPoints {
								batch.SetColor(1, 1, 1, alpha*al)
							} else {
								batch.SetColor(float64(shifted[0]), float64(shifted[1]), float64(shifted[2]), alpha*al)
							}

							batch.DrawUnit(*render.SliderTick)
						}
					}
				}
			}

			if time >= self.objData.StartTime && alphaF > 0.0 {
				batch.SetTranslation(self.objData.StartPos)
				batch.SetSubScale(1+(1.0-alphaF)*0.5, 1+(1.0-alphaF)*0.5)
				batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alphaF)
				batch.DrawUnit(*render.Circle)
				batch.SetColor(1, 1, 1, alphaF)
				batch.DrawUnit(*render.CircleOverlay)
			}

			batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
			batch.SetSubScale(1.0, 1.0)
			batch.SetTranslation(self.Pos)
			batch.DrawUnit(*render.SliderBall)
		}
	} else {
		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos)
			batch.DrawUnit(*render.CircleFull)
		} else if time < self.objData.EndTime {
			batch.SetTranslation(self.Pos)

			if settings.Objects.ForceSliderBallTexture {
				batch.DrawUnit(*render.SliderBall)
			} else {
				batch.DrawUnit(*render.CircleFull)
			}
		}
	}

	batch.SetSubScale(1, 1)

	if time >= self.objData.EndTime && self.fade.GetValue() <= 0.001 {
		if self.vao != nil {
			self.vao.Delete()
		}
		return true
	}
	return false
}

func (self *Slider) DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	self.fadeApproach.Update(float64(time))
	alpha := self.fade.GetValue()
	arr := self.fadeApproach.GetValue()

	batch.SetTranslation(self.objData.StartPos)

	if settings.Objects.DrawApproachCircles && time <= self.objData.StartTime {
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
		batch.SetSubScale(1.0+arr*2, 1.0+arr*2)
		batch.DrawUnit(*render.ApproachCircle)
	}

	batch.SetSubScale(1, 1)
}
