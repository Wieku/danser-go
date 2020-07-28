package objects

import (
	"github.com/faiface/mainthread"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/animation"
	"github.com/wieku/danser-go/animation/easing"
	"github.com/wieku/danser-go/audio"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/curves"
	"github.com/wieku/danser-go/bmath/difficulty"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/utils"
	"github.com/wieku/glhf"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const followBaseScale = 2.14

type PathLine struct {
	Time1 int64
	Time2 int64
	Line  curves.Linear
}

type TickPoint struct {
	Time      int64
	Pos       bmath.Vector2f
	fade      *animation.Glider
	IsReverse bool
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
	objData      *basicData
	multiCurve   *curves.MultiCurve
	scorePath    []PathLine
	Timings      *Timings
	TPoint       TimingPoint
	pixelLength  float64
	partLen      float64
	repeat       int64
	clicked      bool
	sampleSets   []int
	additionSets []int
	samples      []int
	lastT        int64
	Pos          bmath.Vector2f
	divides      int
	TickPoints   []TickPoint
	TickReverse  []TickPoint
	ScorePoints  []TickPoint
	lastTick     int
	End          bool

	startCircle *Circle

	vao                  *glhf.VertexSlice
	created              bool
	discreteCurve        []bmath.Vector2f
	startAngle, endAngle float64
	sliderSnakeIn        *animation.Glider
	sliderSnakeOut       *animation.Glider
	fade                 *animation.Glider
	fadeFollow           *animation.Glider
	scaleFollow          *animation.Glider
	reversePoints        [2][]*reversePoint
}

func NewSlider(data []string) *Slider {
	slider := &Slider{clicked: false}
	slider.objData = commonParse(data)
	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.repeat, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []bmath.Vector2f{slider.objData.StartPos}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 32)
		y, _ := strconv.ParseFloat(list2[1], 32)
		points = append(points, bmath.NewVec2f(float32(x), float32(y)))
	}

	slider.multiCurve = curves.NewMultiCurve(list[0], points, slider.pixelLength)

	slider.objData.EndTime = slider.objData.StartTime
	slider.objData.EndPos = slider.multiCurve.PointAt(1.0)
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
	//slider.fadeCircle = animation.NewGlider(1)
	//slider.fadeApproach = animation.NewGlider(1)
	slider.sliderSnakeIn = animation.NewGlider(1)
	slider.sliderSnakeOut = animation.NewGlider(0)
	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() bmath.Vector2f {
	return self.multiCurve.PointAt(0.5).Add(self.objData.StackOffset)
}

func (self Slider) GetStartAngle() float32 {
	return self.GetBasicData().StartPos.AngleRV(self.GetPointAt(self.objData.StartTime + int64(math.Min(10, self.partLen)))) //temporary solution
}

func (self Slider) GetEndAngle() float32 {
	return self.GetBasicData().EndPos.AngleRV(self.GetPointAt(self.objData.EndTime - int64(math.Min(10, self.partLen)))) //temporary solution
}

func (self Slider) GetPartLen() float32 {
	return float32(20.0) / float32(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)) * float32(self.pixelLength)
}

func (self Slider) GetPointAt(time int64) bmath.Vector2f {
	/*times := int64(math.Min(math.Floor(float64(time-self.objData.StartTime)/self.partLen)+1, float64(self.repeat)))

	ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1)*self.partLen

	var pos bmath.Vector2f
	if (times % 2) == 1 {
		pos = self.multiCurve.PointAt((ttime/*+0.6*/ /*) / self.partLen)
	} else {
		pos = self.multiCurve.PointAt(1.0 - ttime/self.partLen)
	}*/

	pos := self.GetBasicData().StartPos

	if time >= self.GetBasicData().StartTime {
		pLineI := len(self.scorePath)

		for i, p := range self.scorePath {
			if p.Time1 <= time && p.Time2 >= time {
				pLineI = i
				break
			}
		}

		if pLineI < len(self.scorePath) {
			pLine := self.scorePath[pLineI]
			if pLine.Time2 == pLine.Time1 {
				pos = pLine.Line.Point2
			} else {
				pos = pLine.Line.PointAt(float32(time-pLine.Time1) / float32(pLine.Time2-pLine.Time1))
			}
		} else {
			pos = self.scorePath[len(self.scorePath)-1].Line.Point2
		}
	}

	return pos.Add(self.objData.StackOffset)
}

func (self *Slider) GetAsDummyCircles() []BaseObject {
	partLen := int64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength))

	var circles []BaseObject

	for i := int64(0); i <= self.repeat; i++ {
		time := self.objData.StartTime + i*partLen

		if i == self.repeat && settings.KNOCKOUT != "" {
			time = int64(math.Max(float64(self.GetBasicData().StartTime)+float64((self.GetBasicData().EndTime-self.GetBasicData().StartTime)/2), float64(self.GetBasicData().EndTime-36)))
		}

		circles = append(circles, DummyCircleInherit(self.GetPointAt(time), time, true, i == 0, i == self.repeat))
	}

	for _, p := range self.TickPoints {
		circles = append(circles, DummyCircleInherit(p.Pos, p.Time, true, false, false))
	}

	sort.Slice(circles, func(i, j int) bool { return circles[i].GetBasicData().StartTime < circles[j].GetBasicData().StartTime })

	return circles
}

func (self *Slider) SetTiming(timings *Timings) {
	self.Timings = timings
	self.TPoint = timings.GetPoint(self.objData.StartTime)

	sliderTime := self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)
	self.partLen = math.Round(sliderTime*float64(self.repeat)) / float64(self.repeat)

	lines := self.multiCurve.GetLines()

	startTime := float64(self.GetBasicData().StartTime)

	velocity := self.Timings.GetVelocity(self.TPoint)

	minDistanceFromEnd := velocity * 0.01

	scoringLengthTotal := 0.0
	scoringDistance := 0.0

	tickDistance := math.Min(self.Timings.GetTickDistance(self.TPoint), self.pixelLength)

	for i := int64(0); i < self.repeat; i++ {
		distanceToEnd := float64(self.multiCurve.GetLength())
		skipTick := false

		reverse := (i % 2) == 1

		start := 0
		end := len(lines)
		direction := 1
		if reverse {
			start = len(lines) - 1
			end = -1
			direction = -1
		}

		for j := start; j != end; j += direction {

			line := lines[j]

			p1, p2 := line.Point1, line.Point2

			if reverse {
				p1, p2 = p2, p1
			}

			distance := line.GetLength()

			progress := 1000.0 * float64(distance) / velocity

			self.scorePath = append(self.scorePath, PathLine{Time1: int64(startTime), Time2: int64(startTime + progress), Line: curves.NewLinear(p1, p2)})

			startTime += progress

			/*if self.objData.StartTime == 120273 {
				log.Printf("%.10f", distance)
				log.Printf("%.10f", progress)
				log.Printf("%.10f", velocity)
				log.Println(startTime)
			}*/

			scoringDistance += float64(distance)

			//sprites for scoring points (dots)
			for scoringDistance >= tickDistance && !skipTick {
				scoringLengthTotal += tickDistance
				scoringDistance -= tickDistance
				distanceToEnd -= tickDistance

				skipTick = distanceToEnd <= minDistanceFromEnd
				if skipTick {
					break
				}

				scoreTime := self.GetBasicData().StartTime + int64(float64(float32(scoringLengthTotal)*1000)/velocity)

				point := TickPoint{scoreTime, self.GetPointAt(scoreTime), animation.NewGlider(0.0), false}
				self.TickPoints = append(self.TickPoints, point)
				self.ScorePoints = append(self.ScorePoints, point)

			}
		}

		scoringLengthTotal += scoringDistance

		scoreTime := self.GetBasicData().StartTime + int64((float64(float32(scoringLengthTotal))/velocity)*1000)
		point := TickPoint{scoreTime, self.GetPointAt(scoreTime), nil, true}

		self.TickReverse = append(self.TickReverse, point)
		self.ScorePoints = append(self.ScorePoints, point)

		// If our scoring distance is small enough, then there was no "last" scoring point at the end. No need to mirror a non-existing point.
		if skipTick {
			scoringDistance = 0
		} else {
			scoringLengthTotal -= tickDistance - scoringDistance
			scoringDistance = tickDistance - scoringDistance
		}

	}

	//log.Println()

	self.objData.EndTime = int64(math.Floor(startTime))
	self.objData.EndPos = self.GetPointAt(self.objData.EndTime)

	self.calculateFollowPoints()
	self.discreteCurve = self.GetCurve()
	self.startAngle = float64(self.GetStartAngle())
	if len(self.discreteCurve) > 1 {
		self.endAngle = float64(self.discreteCurve[len(self.discreteCurve)-1].AngleRV(self.discreteCurve[len(self.discreteCurve)-2]))
	} else {
		self.endAngle = self.startAngle + math.Pi
	}
}

func (self *Slider) calculateFollowPoints() {
	/*tickPixLen := (100.0 * self.Timings.SliderMult) / (self.Timings.TickRate * self.TPoint.GetRatio())
	tickpoints := int(math.Ceil(self.pixelLength/tickPixLen)) - 1

	for r := 0; r < int(self.repeat); r++ {
		lengthFromEnd := self.pixelLength
		for i := 1; i <= tickpoints; i++ {
			time := self.objData.StartTime + int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))
			time2 := self.objData.StartTime + int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))

			if r%2 == 1 {
				time2 = self.objData.StartTime + int64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength)) - int64(float64(i)*self.TPoint.Bpm/(self.Timings.TickRate*self.TPoint.GetRatio()))
			}

			lengthFromEnd -= tickPixLen

			if lengthFromEnd < 0.01*self.pixelLength {
				break
			}

			time2 += int64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength))*int64(r)

			glider := animation.NewGlider(0.0)

			point := TickPoint{time2, self.GetPointAt(time), glider}
			self.TickPoints = append(self.TickPoints, point)
			self.ScorePoints = append(self.ScorePoints, point)
		}

		time := self.objData.StartTime + int64(float64(r)*self.partLen)
		point := TickPoint{time, self.GetPointAt(time), nil}
		self.TickReverse = append(self.TickReverse, point)
		self.ScorePoints = append(self.ScorePoints, point)
	}
	self.TickReverse = append(self.TickReverse, TickPoint{self.objData.EndTime, self.GetPointAt(self.objData.EndTime), nil})
	*/
	sort.Slice(self.TickPoints, func(i, j int) bool { return self.TickPoints[i].Time < self.TickPoints[j].Time })
	sort.Slice(self.ScorePoints, func(i, j int) bool { return self.ScorePoints[i].Time < self.ScorePoints[j].Time })
}

func (self *Slider) UpdateStacking() {
	for _, tp := range self.TickPoints {
		tp.Pos = tp.Pos.Add(self.objData.StackOffset)
	}

	for i, p := range self.discreteCurve {
		self.discreteCurve[i] = p.Add(self.objData.StackOffset)
	}
}

func (self *Slider) SetDifficulty(diff *difficulty.Difficulty) {
	self.sliderSnakeIn = animation.NewGlider(0)
	self.sliderSnakeOut = animation.NewGlider(0)

	slSnInS := float64(self.objData.StartTime) - diff.Preempt
	slSnInE := float64(self.objData.StartTime) - (diff.Preempt - difficulty.HitFadeIn) + self.partLen*(math.Max(0.0, math.Min(1.0, settings.Objects.SliderSnakeInMult)))

	if settings.Objects.SliderSnakeIn {
		self.sliderSnakeIn.AddEvent(slSnInS, slSnInE, 1)
	} else {
		self.sliderSnakeIn.SetValue(1)
	}

	if settings.Objects.SliderSnakeOut {
		self.sliderSnakeOut.AddEvent(float64(self.objData.EndTime)-self.partLen, float64(self.objData.EndTime), 1)
	}

	self.fade = animation.NewGlider(0)
	self.fade.AddEvent(float64(self.objData.StartTime)-diff.Preempt, float64(self.objData.StartTime)-(diff.Preempt-difficulty.HitFadeIn), 1)
	self.fade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+difficulty.HitFadeIn/3, 0)

	self.startCircle = DummyCircle(self.objData.StartPos, self.objData.StartTime)
	self.startCircle.objData.ComboNumber = self.objData.ComboNumber
	self.startCircle.SetDifficulty(diff)

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
			start -= difficulty.HitFadeIn
		}

		arrow.fade.AddEvent(start, start+math.Min(300, end-start), 1)
		arrow.fade.AddEvent(end, end+300, 0)

		arrow.pulse.AddEventS(end, end+300, 1, 1.4)
		for subTime := start; subTime < end; subTime += 300 {
			arrow.pulse.AddEventS(subTime, subTime+math.Min(300, end-subTime), 1.3, 1)
		}

		self.reversePoints[1] = append(self.reversePoints[1], arrow)
	}

	self.fadeFollow = animation.NewGlider(0)
	self.scaleFollow = animation.NewGlider(0)

	/*self.fadeFollow.AddEventS(float64(self.objData.StartTime), math.Min(float64(self.objData.StartTime+60), float64(self.objData.EndTime)), 0, 1)
	self.fadeFollow.AddEventS(float64(self.objData.EndTime), float64(self.objData.EndTime+200), 1, 0)

	self.scaleFollow.AddEventS(float64(self.objData.StartTime), math.Min(float64(self.objData.StartTime+180), float64(self.objData.EndTime)), 0.5*followBaseScale, 1*followBaseScale)
	self.scaleFollow.AddEventS(float64(self.objData.EndTime), float64(self.objData.EndTime+200), 1*followBaseScale, 0.8*followBaseScale)

	for j, p := range self.ScorePoints {
		if j < 1 {
			continue
		}

		fade := 200.0
		delay := fade
		if len(self.ScorePoints) >= 2 {
			delay = math.Min(fade, float64(p.Time-self.ScorePoints[j-1].Time))
			ratio := delay / fade
			self.scaleFollow.AddEventS(float64(p.Time), float64(p.Time)+delay, 1.1*followBaseScale, (1.1-ratio*0.1)*followBaseScale)
		}

	}*/

	for _, p := range self.TickPoints {
		a := float64((p.Time-self.objData.StartTime)/2+self.objData.StartTime) - diff.Preempt*2/3

		fs := float64(p.Time-self.objData.StartTime) / self.partLen

		if fs < 1.0 {
			a = math.Max(fs*(slSnInE-slSnInS)+slSnInS, a)
		}

		p.fade.AddEventS(a, math.Min(a+150, float64(p.Time)-36), 0.0, 1.0)
		p.fade.AddEventS(float64(p.Time), float64(p.Time), 1.0, 0.0)
	}

}

func (self *Slider) GetCurve() []bmath.Vector2f {
	lod := math.Ceil(self.pixelLength * float64(settings.Objects.SliderPathLOD) / 100.0)
	t0 := float32(1.0 / lod)
	points := make([]bmath.Vector2f, int(lod)+1)
	t := float32(0.0)
	for i := 0; i <= int(lod); i += 1 {
		points[i] = self.multiCurve.PointAt(t).Add(self.objData.StackOffset)
		t += t0
	}
	return points
}

func (self *Slider) Update(time int64) bool {
	self.sliderSnakeOut.Update(float64(time))

	if time < self.objData.EndTime {
		times := int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat)))

		if self.lastT != times {
			if (!settings.PLAY && settings.KNOCKOUT == "") || settings.PLAYERS > 1 {
				self.PlayEdgeSample(int(times - 1))
			}
			self.lastT = times
		}

		for i, p := range self.TickPoints {
			if p.Time < time && self.lastTick < i {
				if (!settings.PLAY && settings.KNOCKOUT == "") || settings.PLAYERS > 1 {
					audio.PlaySliderTick(self.Timings.Current.SampleSet, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume, self.objData.Number, p.Pos.X64())
				}
				self.lastTick = i
			}
		}

		self.Pos = self.GetPointAt(time)

		if !self.clicked {
			//self.playSample(self.sampleSets[0], self.additionSets[0], self.samples[0])
			if (!settings.PLAY && settings.KNOCKOUT == "") || settings.PLAYERS > 1 {
				self.PlayEdgeSample(0)
				self.InitSlide(time)
			}
			self.clicked = true
		}

		return false
	}

	self.Pos = self.GetPointAt(self.objData.EndTime)

	//self.playSample(self.sampleSets[self.repeat], self.additionSets[self.repeat], self.samples[self.repeat])
	if (!settings.PLAY && settings.KNOCKOUT == "") || settings.PLAYERS > 1 {
		self.PlayEdgeSample(int(self.repeat))
	}
	self.End = true
	self.clicked = false

	return true
}

func (self *Slider) ArmStart(clicked bool, time int64) {
	self.startCircle.Arm(clicked, time)
}

func (self *Slider) InitSlide(time int64) {
	self.fadeFollow.Reset()
	self.scaleFollow.Reset()

	startTime := float64(time)
	self.fadeFollow.AddEventS(startTime, math.Min(startTime+60, float64(self.objData.EndTime)), 0, 1)
	self.scaleFollow.AddEventS(startTime, math.Min(startTime+180, float64(self.objData.EndTime)), 0.5*followBaseScale, 1*followBaseScale)

	for j, p := range self.ScorePoints {
		if j < 1 || p.Time < time {
			continue
		}

		fade := 200.0
		delay := fade
		if len(self.ScorePoints) >= 2 {
			delay = math.Min(fade, float64(p.Time-self.ScorePoints[j-1].Time))
			ratio := delay / fade
			self.scaleFollow.AddEventS(float64(p.Time), float64(p.Time)+delay, 1.1*followBaseScale, (1.1-ratio*0.1)*followBaseScale)
		}

	}

	self.scaleFollow.AddEventS(float64(self.objData.EndTime), float64(self.objData.EndTime+200), 1*followBaseScale, 0.8*followBaseScale)
	self.fadeFollow.AddEventS(float64(self.objData.EndTime), float64(self.objData.EndTime+200), 1, 0)
}

func (self *Slider) KillSlide(time int64) {
	self.fadeFollow.Reset()
	self.scaleFollow.Reset()

	nextPoint := self.objData.EndTime
	for _, p := range self.ScorePoints {
		if p.Time > time {
			nextPoint = p.Time
			break
		}
	}

	self.fadeFollow.AddEventS(float64(nextPoint-100), float64(nextPoint), 1, 0)
	self.scaleFollow.AddEventS(float64(nextPoint-100), float64(nextPoint), 1, 2)
}

func (self *Slider) PlayEdgeSample(index int) {
	self.playSampleT(self.sampleSets[index], self.additionSets[index], self.samples[index], self.Timings.GetPoint(self.objData.StartTime+int64(float64(index)*self.partLen)), self.GetPointAt(self.objData.StartTime+int64(float64(index)*self.partLen)))
}

func (self *Slider) PlayTick() {
	//audio.PlaySliderTick(self.Timings.Current.SampleSet, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume, self.objData.Number)
}

func (self *Slider) playSample(sampleSet, additionSet, sample int) {
	self.playSampleT(sampleSet, additionSet, sampleSet, self.Timings.Current, bmath.NewVec2f(0, 0))
}

func (self *Slider) playSampleT(sampleSet, additionSet, sample int, point TimingPoint, pos bmath.Vector2f) {
	if sampleSet == 0 {
		sampleSet = self.objData.sampleSet
		if sampleSet == 0 {
			sampleSet = point.SampleSet
		}
	}

	if additionSet == 0 {
		additionSet = self.objData.additionSet
	}

	audio.PlaySample(sampleSet, additionSet, sample, point.SampleIndex, point.SampleVolume, self.objData.Number, pos.X64())
}

func (self *Slider) GetPosition() bmath.Vector2f {
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
	self.fade.Update(float64(time))

	in := 0
	out := int(self.sliderSnakeIn.GetValue() * float64(len(self.discreteCurve)))

	if int64(math.Min(float64(time-self.objData.StartTime)/self.partLen+1, float64(self.repeat))) == self.repeat {
		if (self.repeat % 2) == 1 {
			in = int(math.Ceil(self.sliderSnakeOut.GetValue() * (float64(len(self.discreteCurve)) - 1)))
		} else {
			out = int(math.Floor((1.0 - self.sliderSnakeOut.GetValue()) * (float64(len(self.discreteCurve)) - 1)))
		}
	}

	if out < in {
		in, out = out, in
	}

	colorAlpha := self.fade.GetValue()

	renderer.SetColor(mgl32.Vec4{color[0], color[1], color[2], float32(colorAlpha /** 0.15*/)}, mgl32.Vec4{color1[0], color1[1], color1[2] /*1.0, 1.0, 1.0*/, float32(colorAlpha)})

	if self.vao != nil {
		subVao := self.vao.Slice(in*self.divides*3, out*self.divides*3)
		subVao.BeginDraw()
		subVao.Draw()
		subVao.EndDraw()
	}
}

func (self *Slider) Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) bool {
	self.fade.Update(float64(time))
	//self.fadeCircle.Update(float64(time))
	self.fadeFollow.Update(float64(time))
	self.scaleFollow.Update(float64(time))

	for i := 0; i < 2; i++ {
		for j := 0; j < len(self.reversePoints[i]); j++ {
			self.reversePoints[i][j].pulse.Update(float64(time))
			self.reversePoints[i][j].fade.Update(float64(time))
		}
	}

	alpha := self.fade.GetValue()
	//alphaF := self.fadeCircle.GetValue()

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {

		if settings.Objects.DrawReverseArrows {
			for i := 0; i < 2; i++ {
				for _, p := range self.reversePoints[i] {
					if p.fade.GetValue() >= 0.001 {
						if i == 1 {

							out := int(self.sliderSnakeIn.GetValue() * float64(len(self.discreteCurve)-1))
							batch.SetTranslation(self.discreteCurve[out].Copy64())
							if out == 0 {
								batch.SetRotation(self.startAngle)
							} else if out == len(self.discreteCurve)-1 {
								batch.SetRotation(self.endAngle + math.Pi)
							} else {
								batch.SetRotation(float64(self.discreteCurve[out-1].AngleRV(self.discreteCurve[out])))
							}

						} else {
							batch.SetTranslation(self.discreteCurve[0].Copy64())
							batch.SetRotation(self.startAngle + math.Pi)
						}
						batch.SetSubScale(p.pulse.GetValue(), p.pulse.GetValue())
						batch.SetColor(1, 1, 1, alpha*self.sliderSnakeIn.GetValue()*p.fade.GetValue())
						batch.DrawUnit(*render.SliderReverse)
					}
				}
			}
		}

		batch.SetTranslation(self.objData.StartPos.Copy64())
		batch.SetSubScale(1, 1)
		batch.SetRotation(0)

		if time < self.objData.EndTime {

			if settings.Objects.DrawFollowPoints {
				shifted := utils.GetColorShifted(color, settings.Objects.FollowPointColorOffset)

				for _, p := range self.TickPoints {
					p.fade.Update(float64(time))
					al := p.fade.GetValue()
					/*if p.Time > time {
						al = math.Min(1.0, math.Max((float64(time)-(float64(p.Time)-self.TPoint.Bpm*2))/(self.TPoint.Bpm), 0.0))
					}*/
					if al > 0.0 {
						batch.SetTranslation(p.Pos.Copy64())
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

		batch.SetSubScale(1, 1)
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

		if time < self.objData.StartTime {
			/*batch.SetTranslation(self.objData.StartPos.Copy64())

			batch.DrawUnit(*render.Circle)
			batch.SetColor(1, 1, 1, alpha)
			batch.DrawUnit(*render.CircleOverlay)
			if settings.DIVIDES < 2 && settings.Objects.DrawComboNumbers {
				render.Combo.DrawCentered(batch, self.objData.StartPos.X64(), self.objData.StartPos.Y64(), 0.65, strconv.Itoa(int(self.objData.ComboNumber)))
			}*/

			batch.SetSubScale(1, 1)

		} else {

			/*if time >= self.objData.StartTime && alphaF > 0.0 {
				batch.SetTranslation(self.objData.StartPos.Copy64())
				batch.SetSubScale(1+(1.0-alphaF)*0.5, 1+(1.0-alphaF)*0.5)
				batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alphaF)
				batch.DrawUnit(*render.Circle)
				batch.SetColor(1, 1, 1, alphaF)
				batch.DrawUnit(*render.CircleOverlay)
			}*/

			batch.SetColor( /*float64(color[0]), float64(color[1]), float64(color[2])*/ 1, 1, 1, alpha)
			batch.SetSubScale(1.0, 1.0)
			batch.SetTranslation(self.Pos.Copy64())
			batch.DrawUnit(*render.SliderBall)
		}

		if settings.Objects.DrawSliderFollowCircle {
			batch.SetTranslation(self.Pos.Copy64())
			batch.SetSubScale(self.scaleFollow.GetValue(), self.scaleFollow.GetValue())
			batch.SetColor(1, 1, 1, self.fadeFollow.GetValue())
			batch.DrawUnit(*render.SliderFollow)
		}

	} else {
		if time < self.objData.StartTime {
			batch.SetTranslation(self.objData.StartPos.Copy64())
			batch.DrawUnit(*render.CircleFull)
		} else if time < self.objData.EndTime {
			batch.SetTranslation(self.Pos.Copy64())

			if settings.Objects.ForceSliderBallTexture {
				batch.DrawUnit(*render.SliderBall)
			} else {
				batch.DrawUnit(*render.CircleFull)
			}
		}
	}

	self.startCircle.Draw(time, color, batch)

	batch.SetSubScale(1, 1)

	if time >= self.objData.EndTime && self.fade.GetValue() <= 0.001 {
		if self.vao != nil {
			if settings.Objects.SliderDynamicUnload {
				self.vao.Delete()
			}
		}
		return true
	}
	return false
}

func (self *Slider) DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	self.startCircle.DrawApproach(time, color, batch)
}
