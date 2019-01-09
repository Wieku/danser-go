package objects

import (
	"danser/animation"
	"danser/animation/easing"
	"danser/audio"
	"danser/bmath"
	m2 "danser/bmath"
	"danser/bmath/sliders"
	"danser/render"
	"danser/settings"
	"danser/utils"
	"github.com/faiface/mainthread"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/glhf"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type tickPoint struct {
	Time int64
	Pos  m2.Vector2d
}

type reversePoint struct {
	fade  *animation.Glider
	pulse *animation.Glider
}

func newReverse() (point *reversePoint) {
	point = &reversePoint{animation.NewGlider(0), animation.NewGlider(1)}
	point.fade.SetEasing(easing.OutQuad)
	point.pulse.SetEasing(easing.OutQuad)
	return
}

type Slider struct {
	objData       *basicData
	multiCurve    sliders.SliderAlgo
	Timings       *Timings
	TPoint        TimingPoint
	pixelLength   float64
	partLen       float64
	repeat        int64
	clicked       bool
	sampleSets    []int
	additionSets  []int
	samples       []int
	lastT         int64
	Pos           m2.Vector2d
	divides       int
	TickPoints    []tickPoint
	TickReverse   []tickPoint
	TickReverseTrue   []tickPoint
	ScorePoints   []tickPoint
	lastTick      int
	End           bool
	vao           *glhf.VertexSlice
	created       bool
	discreteCurve []bmath.Vector2d
	reversePoints [2][]*reversePoint
	startAngle, endAngle float64
	typ			  string
	// 曲线的真正终点
	curveEndPos     m2.Vector2d

	//加入tail真正的judge点
	TailJudgePoint  bmath.Vector2d
	TailJudgeOffset int64
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

	slider.multiCurve = sliders.NewSliderAlgo(list[0], points, slider.pixelLength)

	slider.typ = list[0]

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

	slider.curveEndPos = points[len(points) - 1]

	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() m2.Vector2d {
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

func (self Slider) GetPointAt(time int64) m2.Vector2d {
	times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))

	ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1)*self.partLen

	var pos m2.Vector2d
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
		circles = append(circles, DummyCircleInherit(p.Pos, p.Time, true))
	}

	sort.Slice(circles, func(i, j int) bool { return circles[i].GetBasicData().StartTime < circles[j].GetBasicData().StartTime })

	return circles
}

func (self Slider) endTime() int64 {
	return self.objData.StartTime + self.repeat*self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)
}

func (self *Slider) SetTiming(timings *Timings) {
	self.Timings = timings
	self.TPoint = timings.GetPoint(self.objData.StartTime)

	sliderTime := self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)
	self.partLen = float64(sliderTime)
	self.objData.EndTime = self.objData.StartTime + sliderTime*self.repeat
	self.objData.EndPos = self.GetPointAt(self.objData.EndTime)

	self.calculateFollowPoints()
	// 计算滑条尾判定点
	self.calculateTailJudgePoint()
	self.discreteCurve = self.GetCurve()
	self.startAngle = self.GetStartAngle()
	self.endAngle = self.curveEndPos.AngleRV(self.discreteCurve[len(self.discreteCurve)-1])
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

			point := tickPoint{time2 + self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)*int64(r), self.GetPointAt(time)}
			self.TickPoints = append(self.TickPoints, point)
			self.ScorePoints = append(self.ScorePoints, point)
		}

		time := self.objData.StartTime + int64(float64(r)*self.partLen)
		point := tickPoint{time, self.GetPointAt(time)}
		self.TickReverse = append(self.TickReverse, point)
		//// 去掉第一个点（滑条头）
		if r != 0 {
			self.TickReverseTrue = append(self.TickReverseTrue, point)
			self.ScorePoints = append(self.ScorePoints, point)
		}
	}
	self.TickReverse = append(self.TickReverse, tickPoint{self.objData.EndTime, self.GetPointAt(self.objData.EndTime)})
	self.TickReverseTrue = append(self.TickReverseTrue, tickPoint{self.objData.EndTime, self.GetPointAt(self.objData.EndTime)})

	sort.Slice(self.TickPoints, func(i, j int) bool { return self.TickPoints[i].Time < self.TickPoints[j].Time })
	sort.Slice(self.ScorePoints, func(i, j int) bool { return self.ScorePoints[i].Time < self.ScorePoints[j].Time })
}

func (self *Slider) SetDifficulty(preempt, fadeIn float64) {
	for i := int64(2); i < self.repeat; i += 2 {
		arrow := newReverse()

		start := float64(self.objData.StartTime) + float64(i-2)*self.partLen
		end := float64(self.objData.StartTime) + float64(i)*self.partLen

		arrow.fade.AddEvent(start, start+math.Min(300, end-start), 1)
		arrow.fade.AddEvent(end, end+300, 0)

		arrow.pulse.AddEventS(end, end+300, 1, 1.4)
		for j := start; j < end; j += 300 {
			arrow.pulse.AddEvent(j-0.1, j-0.1, 1.3)
			arrow.pulse.AddEvent(j, j+math.Min(300, end-j), 1)
		}

		self.reversePoints[0] = append(self.reversePoints[0], arrow)
	}

	for i := int64(1); i < self.repeat; i += 2 {
		arrow := newReverse()

		start := float64(self.objData.StartTime) + float64(i-2)*self.partLen
		end := float64(self.objData.StartTime) + float64(i)*self.partLen
		if i == 1 {
			start -= fadeIn
		}

		arrow.fade.AddEvent(start, start+math.Min(300, end-start), 1)
		arrow.fade.AddEvent(end, end+300, 0)

		arrow.pulse.AddEventS(end, end+300, 1, 1.4)
		for subTime := start; subTime < end; subTime += 300 {
			arrow.pulse.AddEventS(subTime, subTime+math.Min(300, end-subTime), 1.3, 1)
		}

		self.reversePoints[1] = append(self.reversePoints[1], arrow)
	}

}

// 计算真正的TailJudge参数
func (self *Slider) calculateTailJudgePoint() {
	legacytailoffset := int64(36)
	// 计算滑条持续时间
	slidersuration := self.GetBasicData().EndTime - self.GetBasicData().StartTime
	if slidersuration < legacytailoffset * 2  {
		self.TailJudgeOffset = int64((slidersuration+1)/2)
	}else {
		self.TailJudgeOffset = legacytailoffset
	}
	// 计算实际判定点
	time := self.objData.EndTime - self.TailJudgeOffset
	self.TailJudgePoint = self.GetPointAt(time)
}

func (self *Slider) GetCurve() []m2.Vector2d {
	lod := math.Ceil(self.pixelLength * float64(settings.Objects.SliderPathLOD) / 100.0)
	t0 := 1.0 / lod
	points := make([]m2.Vector2d, int(lod)+1)
	t := 0.0
	for i := 0; i <= int(lod); i += 1 {
		points[i] = self.multiCurve.PointAt(t)
		t += t0
	}
	return points
}

func (self *Slider) Update(time int64) bool {
	if time < self.endTime() {
		times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))
		ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1)*self.partLen

		if self.lastT != times {
			//self.playSample(self.sampleSets[times-1], self.additionSets[times-1], self.samples[times-1])
			self.lastT = times
		}

		for i, p := range self.TickPoints {
			if p.Time < time && self.lastTick < i {
				audio.PlaySliderTick(self.Timings.Current.SampleSet, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume)
				self.lastTick = i
			}
		}

		var pos m2.Vector2d
		if (times % 2) == 1 {
			pos = self.multiCurve.PointAt(ttime / self.partLen)
		} else {
			pos = self.multiCurve.PointAt(1.0 - ttime/self.partLen)
		}
		self.Pos = pos.Add(self.objData.StackOffset)

		if !self.clicked {
			self.playSample(self.sampleSets[0], self.additionSets[0], self.samples[0])
			self.clicked = true
		}

		return false
	}

	var pos m2.Vector2d
	if (self.repeat % 2) == 1 {
		pos = self.multiCurve.PointAt(1.0)
	} else {
		pos = self.multiCurve.PointAt(0.0)
	}
	self.Pos = pos.Add(self.objData.StackOffset)

	//self.playSample(self.sampleSets[self.repeat], self.additionSets[self.repeat], self.samples[self.repeat])
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

func (self *Slider) GetPosition() m2.Vector2d {
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

func (self *Slider) DrawBody(time int64, preempt float64, color mgl32.Vec4, color1 mgl32.Vec4, renderer *render.SliderRenderer) {
	in := 0
	out := len(self.discreteCurve)

	if time < self.objData.StartTime-int64(preempt)/2 {
		if settings.Objects.SliderSnakeIn {
			alpha := math.Abs(float64(time-(self.objData.StartTime-int64(preempt)))) / (preempt / 2)
			out = int(float64(out) * alpha)
		}
	} else if settings.Objects.SliderSnakeOut {
		if time >= self.objData.StartTime && time <= self.objData.EndTime {
			times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))
			if times >= self.repeat {
				ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1)*self.partLen
				alpha := 0.0
				if (times % 2) == 1 {
					alpha = ttime / self.partLen
					in = int(float64(out) * alpha)
				} else {
					alpha = 1.0 - ttime/self.partLen
					out = int(float64(out) * alpha)
				}
			}
		} else if time > self.objData.EndTime {
			if (self.repeat % 2) == 1 {
				in = out - 1
			} else {
				out = 1
			}
		}
	}

	colorAlpha := 1.0

	if time < self.objData.StartTime-int64(preempt)/2 {
		colorAlpha = float64(time-(self.objData.StartTime-int64(preempt))) / (preempt / 2)
	} else if time >= self.objData.EndTime {
		colorAlpha = 1.0 - float64(time-self.objData.EndTime)/(preempt/4)
	} else {
		colorAlpha = float64(color[3])
	}

	renderer.SetColor(mgl32.Vec4{color[0], color[1], color[2], float32(colorAlpha)}, mgl32.Vec4{color1[0], color1[1], color1[2], float32(colorAlpha)})

	if self.vao != nil {
		subVao := self.vao.Slice(in*self.divides*3, out*self.divides*3)
		subVao.BeginDraw()
		subVao.Draw()
		subVao.EndDraw()
	}
}

func (self *Slider) Draw(time int64, preempt float64, color mgl32.Vec4, batch *render.SpriteBatch) bool {
	alpha := 1.0
	alphaF := 1.0

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha = float64(time-(self.objData.StartTime-int64(preempt))) / (preempt / 2)
	} else if time >= self.objData.EndTime {
		alpha = 1.0 - float64(time-self.objData.EndTime)/(preempt/4)
		alphaF = 1.0 - float64(time-self.objData.StartTime)/(preempt/2)
	} else {
		alpha = float64(color[3])
		alphaF = 1.0 - float64(time-self.objData.StartTime)/(preempt/2)
	}

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {

		for i := 0; i < 2; i++ {
			for k, p := range self.reversePoints[i] {
				if p.fade.GetValue() >= 0 {
					if i == 1 {
						out := len(self.discreteCurve)-1
						batch.SetTranslation(self.discreteCurve[out])
						if out == 0 {
							batch.SetRotation(self.startAngle)
						} else if out == len(self.discreteCurve)-1 {
							batch.SetRotation(self.endAngle + math.Pi)
						} else {
							batch.SetRotation(self.discreteCurve[out-1].AngleRV(self.discreteCurve[out]))
						}
					} else {
						batch.SetTranslation(self.discreteCurve[0])
						batch.SetRotation(self.startAngle + math.Pi)
					}
					batch.SetSubScale(p.pulse.GetValue(), p.pulse.GetValue())
					mult := 1.0
					num := k*2
					if i == 0 {
						num += 1
					}
					fadetime := int64(float64(self.TickReverseTrue[num].Time - self.TickReverse[num].Time) / 4.5)
					if time >= self.TickReverseTrue[num].Time {
						mult = 0.0
					} else if time >= self.TickReverse[num].Time + fadetime{
						mult = 1.0
					} else {
						mult = float64((time - self.TickReverse[num].Time)) / float64(fadetime)
					}
					batch.SetColor(1, 1, 1, alpha * mult)
					batch.DrawUnit(*render.SliderReverse)
				}
			}
		}

		batch.SetTranslation(self.objData.StartPos)
		batch.SetSubScale(1, 1)
		batch.SetRotation(0)
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos)

			batch.DrawUnit(*render.Circle)
			batch.SetColor(1, 1, 1, alpha)
			batch.DrawUnit(*render.CircleOverlay)

		} else {
			if settings.Objects.DrawFollowPoints && time < self.objData.EndTime {
				shifted := utils.GetColorShifted(color, settings.Objects.FollowPointColorOffset)

				for _, p := range self.TickPoints {
					al := 0.0
					if p.Time > time {
						al = math.Min(1.0, math.Max((float64(time)-(float64(p.Time)-self.TPoint.Bpm*2))/(self.TPoint.Bpm), 0.0))
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

	if time >= self.objData.EndTime+int64(preempt/4) {
		if self.vao != nil {
			self.vao.Delete()
		}
		return true
	}
	return false
}

func (self *Slider) DrawApproach(time int64, preempt float64, color mgl32.Vec4, batch *render.SpriteBatch) {

	alpha := 1.0
	arr := float64(self.objData.StartTime-time) / preempt

	if time < self.objData.StartTime-int64(preempt)/2 {
		alpha = float64(time-(self.objData.StartTime-int64(preempt))) / (preempt / 2)
	} else if time >= self.objData.StartTime {
		alpha = 1.0 - float64(time-self.objData.StartTime)/(preempt/2)
	} else {
		alpha = float64(color[3])
	}

	batch.SetTranslation(self.objData.StartPos)

	if settings.Objects.DrawApproachCircles && time <= self.objData.StartTime {
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)
		batch.SetSubScale(1.0+arr*2, 1.0+arr*2)
		batch.DrawUnit(*render.ApproachCircle)
	}

	batch.SetSubScale(1, 1)
}