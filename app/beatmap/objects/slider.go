package objects

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/curves"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/graphics/sliderrenderer"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

type PathLine struct {
	Time1 int64
	Time2 int64
	Line  curves.Linear
}

type TickPoint struct {
	Time      int64
	Pos       vector.Vector2f
	fade      *animation.Glider
	scale     *animation.Glider
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
	sampleSets   []int
	additionSets []int
	samples      []int
	Pos          vector.Vector2f
	TickPoints   []TickPoint
	TickReverse  []TickPoint
	ScorePoints  []TickPoint

	startCircle *Circle

	startAngle, endAngle float64
	sliderSnakeTail      *animation.Glider
	sliderSnakeHead      *animation.Glider
	fade                 *animation.Glider
	bodyFade             *animation.Glider
	//fadeFollow           *animation.Glider
	//scaleFollow          *animation.Glider
	reversePoints [2][]*reversePoint

	diff     *difficulty.Difficulty
	body     *sliderrenderer.Body
	lastTime int64

	ball     *sprite.Sprite
	follower *sprite.Sprite
}

func NewSlider(data []string) *Slider {
	slider := &Slider{}
	slider.objData = commonParse(data)
	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.repeat, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []vector.Vector2f{slider.objData.StartPos}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 32)
		y, _ := strconv.ParseFloat(list2[1], 32)
		points = append(points, vector.NewVec2f(float32(x), float32(y)))
	}

	slider.multiCurve = curves.NewMultiCurve(list[0], points, slider.pixelLength)

	slider.objData.EndTime = slider.objData.StartTime
	slider.objData.EndPos = slider.multiCurve.PointAt(1.0)
	slider.Pos = slider.objData.StartPos

	slider.samples = make([]int, slider.repeat+1)
	slider.sampleSets = make([]int, slider.repeat+1)
	slider.additionSets = make([]int, slider.repeat+1)

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

	slider.fade = animation.NewGlider(1)
	slider.bodyFade = animation.NewGlider(1)
	//slider.fadeCircle = animation.NewGlider(1)
	//slider.fadeApproach = animation.NewGlider(1)
	slider.sliderSnakeTail = animation.NewGlider(1)
	slider.sliderSnakeHead = animation.NewGlider(0)
	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() vector.Vector2f {
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

func (self Slider) GetPointAt(time int64) vector.Vector2f {
	if self.IsRetarded() {
		return self.objData.StartPos
	}

	index := sort.Search(len(self.scorePath), func(i int) bool {
		return self.scorePath[i].Time2 >= time
	})

	pLine := self.scorePath[bmath.ClampI(index, 0, len(self.scorePath)-1)]

	clamped := bmath.ClampI64(time, pLine.Time1, pLine.Time2)

	var pos vector.Vector2f
	if pLine.Time2 == pLine.Time1 {
		pos = pLine.Line.Point2
	} else {
		pos = pLine.Line.PointAt(float32(clamped-pLine.Time1) / float32(pLine.Time2-pLine.Time1))
	}

	return pos.Add(self.objData.StackOffset)
}

func (self *Slider) GetAsDummyCircles() []BaseObject {
	partLen := int64(self.Timings.GetSliderTimeP(self.TPoint, self.pixelLength))

	var circles []BaseObject

	for i := int64(0); i <= self.repeat; i++ {
		time := self.objData.StartTime + i*partLen

		if i == self.repeat && settings.KNOCKOUT {
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
			self.objData.EndTime = int64(math.Floor(startTime))
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

				point := TickPoint{scoreTime, self.GetPointAt(scoreTime), animation.NewGlider(0.0), animation.NewGlider(0.0), false}
				self.TickPoints = append(self.TickPoints, point)
				self.ScorePoints = append(self.ScorePoints, point)

			}
		}

		scoringLengthTotal += scoringDistance

		scoreTime := self.GetBasicData().StartTime + int64((float64(float32(scoringLengthTotal))/velocity)*1000)
		point := TickPoint{scoreTime, self.GetPointAt(scoreTime), nil, nil, true}

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

	self.partLen = float64(self.objData.EndTime-self.objData.StartTime) / float64(self.repeat)

	self.objData.EndPos = self.GetPointAt(self.objData.EndTime)

	if len(self.scorePath) == 0 || self.objData.StartTime == self.objData.EndTime {
		log.Println("Warning: slider", self.objData.Number, "at ", self.objData.StartTime, "is broken.")
	}

	self.calculateFollowPoints()

	self.startAngle = float64(self.GetStartAngle())
	if len(lines) > 0 {
		self.endAngle = float64(lines[len(lines)-1].GetEndAngle())
	} else {
		self.endAngle = self.startAngle + math.Pi
	}
}

func (self *Slider) calculateFollowPoints() {
	sort.Slice(self.TickPoints, func(i, j int) bool { return self.TickPoints[i].Time < self.TickPoints[j].Time })
	sort.Slice(self.ScorePoints, func(i, j int) bool { return self.ScorePoints[i].Time < self.ScorePoints[j].Time })
}

func (self *Slider) UpdateStacking() {
	for i, tp := range self.TickPoints {
		tp.Pos = tp.Pos.Add(self.objData.StackOffset)
		self.TickPoints[i] = tp
	}
}

func (self *Slider) SetDifficulty(diff *difficulty.Difficulty) {
	self.diff = diff
	self.sliderSnakeTail = animation.NewGlider(0)
	self.sliderSnakeHead = animation.NewGlider(0)

	slSnInS := float64(self.objData.StartTime) - diff.Preempt
	slSnInE := float64(self.objData.StartTime) - (diff.Preempt - difficulty.HitFadeIn) + self.partLen*(math.Max(0.0, math.Min(1.0, settings.Objects.SliderSnakeInMult)))

	if settings.Objects.SliderSnakeIn {
		self.sliderSnakeTail.AddEvent(slSnInS, slSnInE, 1)
	} else {
		self.sliderSnakeTail.SetValue(1)
	}

	if settings.Objects.SliderSnakeOut {
		if self.repeat%2 == 1 {
			self.sliderSnakeHead.AddEvent(float64(self.objData.EndTime)-self.partLen, float64(self.objData.EndTime), 1)
		} else {
			self.sliderSnakeTail.AddEvent(float64(self.objData.EndTime)-self.partLen, float64(self.objData.EndTime), 0)
		}
	}

	self.fade = animation.NewGlider(0)
	self.fade.AddEvent(float64(self.objData.StartTime)-diff.Preempt, float64(self.objData.StartTime)-(diff.Preempt-difficulty.HitFadeIn), 1)

	self.bodyFade = animation.NewGlider(0)
	self.bodyFade.AddEvent(float64(self.objData.StartTime)-diff.Preempt, float64(self.objData.StartTime)-(diff.Preempt-difficulty.HitFadeIn), 1)

	if diff.CheckModActive(difficulty.Hidden) {
		self.bodyFade.AddEvent(float64(self.objData.StartTime)-diff.Preempt+difficulty.HitFadeIn, float64(self.objData.EndTime), 0)
	} else {
		self.bodyFade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+difficulty.HitFadeOut, 0)
	}

	self.fade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+difficulty.HitFadeOut, 0)

	self.startCircle = DummyCircle(self.objData.StartPos, self.objData.StartTime)
	self.startCircle.objData.ComboNumber = self.objData.ComboNumber
	self.startCircle.objData.ComboSet = self.objData.ComboSet
	self.startCircle.objData.Number = self.objData.Number
	self.startCircle.SetDifficulty(diff)

	sixty := 1000.0 / 60
	frameDelay := math.Max(150/self.Timings.GetVelocity(self.TPoint)*sixty, sixty)

	self.ball = sprite.NewAnimation(skin.GetFrames("sliderb", false), frameDelay, true, 0.0, vector.NewVec2d(0, 0), bmath.Origin.Centre)

	if len(self.scorePath) > 0 {
		angle := self.scorePath[0].Line.GetStartAngle()
		self.ball.SetVFlip(angle > -math32.Pi/2 && angle < math32.Pi/2)
	}

	followerFrames := skin.GetFrames("sliderfollowcircle", true)

	self.follower = sprite.NewAnimation(followerFrames, 1000.0/float64(len(followerFrames)), true, 0.0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	self.follower.SetAlpha(0.0)

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

	for _, p := range self.TickPoints {
		a := float64((p.Time-self.objData.StartTime)/2+self.objData.StartTime) - diff.Preempt*2/3

		fs := float64(p.Time-self.objData.StartTime) / self.partLen

		if fs < 1.0 {
			a = math.Max(fs*(slSnInE-slSnInS)+slSnInS, a)
		}

		endTime := math.Min(a+150, float64(p.Time)-36)

		p.scale.AddEventS(a, endTime, 0.5, 1.2)
		p.scale.AddEventSEase(endTime, endTime+150, 1.2, 1.0, easing.OutQuad)
		p.fade.AddEventS(a, endTime, 0.0, 1.0)
		if diff.CheckModActive(difficulty.Hidden) {
			p.fade.AddEventS(math.Max(endTime, float64(p.Time)-1000), float64(p.Time), 1.0, 0.0)
		} else {
			p.fade.AddEventS(float64(p.Time), float64(p.Time), 1.0, 0.0)
		}

	}

	self.body = sliderrenderer.NewBody(self.multiCurve, float32(self.diff.CircleRadius))
}

func (self *Slider) IsRetarded() bool {
	return len(self.scorePath) == 0 || self.objData.StartTime == self.objData.EndTime
}

func (self *Slider) Update(time int64) bool {
	if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {

		for i := int64(0); i <= self.repeat; i++ {
			edgeTime := self.objData.StartTime + int64(float64(i)*self.partLen)

			if self.lastTime < edgeTime && time >= edgeTime {
				self.PlayEdgeSample(int(i))

				if i == 0 {
					self.InitSlide(self.objData.StartTime)
					self.ArmStart(true, self.objData.StartTime)
				}
			}
		}

		for _, p := range self.TickPoints {
			if self.lastTime < p.Time && time >= p.Time {
				audio.PlaySliderTick(self.Timings.Current.SampleSet, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume, self.objData.Number, p.Pos.X64())
			}
		}
	}

	self.sliderSnakeHead.Update(float64(time))
	self.sliderSnakeTail.Update(float64(time))

	if self.startCircle != nil {
		self.startCircle.Update(time)
	}

	if self.ball != nil {
		self.ball.Update(time)
	}

	if self.follower != nil {
		self.follower.Update(time)
	}

	self.fade.Update(float64(time))
	self.bodyFade.Update(float64(time))

	for i := 0; i < 2; i++ {
		for j := 0; j < len(self.reversePoints[i]); j++ {
			self.reversePoints[i][j].pulse.Update(float64(time))
			self.reversePoints[i][j].fade.Update(float64(time))
		}
	}

	for _, p := range self.TickPoints {
		p.fade.Update(float64(time))
		p.scale.Update(float64(time))
	}

	pos := self.GetPointAt(time)

	if time-self.lastTime > 0 && time >= self.objData.StartTime {
		angle := pos.AngleRV(self.Pos)

		reversed := int(float64(time-self.objData.StartTime)/self.partLen)%2 == 1

		if reversed {
			angle -= math32.Pi
		}

		if self.ball != nil {
			self.ball.SetHFlip(reversed)
			self.ball.SetRotation(float64(angle))
		}
	}

	self.Pos = pos

	self.lastTime = time

	return true
}

func (self *Slider) ArmStart(clicked bool, time int64) {
	self.startCircle.Arm(clicked, time)
}

func (self *Slider) InitSlide(time int64) {
	self.follower.ClearTransformations()

	startTime := float64(time)

	self.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, math.Min(startTime+60, float64(self.objData.EndTime)), 0, 1))
	self.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, math.Min(startTime+180, float64(self.objData.EndTime)), 0.5, 1))

	for j, p := range self.ScorePoints {
		if j < 1 || p.Time < time {
			continue
		}

		fade := 200.0
		delay := fade
		if len(self.ScorePoints) >= 2 {
			delay = math.Min(fade, float64(p.Time-self.ScorePoints[j-1].Time))
			ratio := delay / fade

			self.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(p.Time), float64(p.Time)+delay, 1.1, 1.1-ratio*0.1))
		}
	}

	self.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, float64(self.objData.EndTime), float64(self.objData.EndTime+200), 1, 0))
	self.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, float64(self.objData.EndTime), float64(self.objData.EndTime+200), 1, 0.8))
}

func (self *Slider) KillSlide(time int64) {
	self.follower.ClearTransformations()

	nextPoint := self.objData.EndTime
	for _, p := range self.ScorePoints {
		if p.Time > time {
			nextPoint = p.Time
			break
		}
	}

	self.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(nextPoint-100), float64(nextPoint), 1, 0))
	self.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(nextPoint-100), float64(nextPoint), 1, 2))
}

func (self *Slider) PlayEdgeSample(index int) {
	self.playSampleT(self.sampleSets[index], self.additionSets[index], self.samples[index], self.Timings.GetPoint(self.objData.StartTime+int64(float64(index)*self.partLen)), self.GetPointAt(self.objData.StartTime+int64(float64(index)*self.partLen)))
}

func (self *Slider) PlayTick() {
	audio.PlaySliderTick(self.Timings.Current.SampleSet, self.Timings.Current.SampleIndex, self.Timings.Current.SampleVolume, self.objData.Number, self.Pos.X64())
}

func (self *Slider) playSample(sampleSet, additionSet, sample int) {
	self.playSampleT(sampleSet, additionSet, sampleSet, self.Timings.Current, vector.NewVec2f(0, 0))
}

func (self *Slider) playSampleT(sampleSet, additionSet, sample int, point TimingPoint, pos vector.Vector2f) {
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

func (self *Slider) GetPosition() vector.Vector2f {
	return self.Pos
}

func (self *Slider) DrawBodyBase(time int64, projection mgl32.Mat4) {
	self.body.DrawBase(self.sliderSnakeHead.GetValue(), self.sliderSnakeTail.GetValue(), projection)
}

func (self *Slider) DrawBody(time int64, color mgl32.Vec4, color1 mgl32.Vec4, projection mgl32.Mat4, scale float32) {
	colorAlpha := self.bodyFade.GetValue()

	bodyInner := mgl32.Vec4{color.X(), color.Y(), color.Z(), float32(colorAlpha)}
	bodyOuter := mgl32.Vec4{color1.X(), color1.Y(), color1.Z(), float32(colorAlpha)}
	borderInner := mgl32.Vec4{0.1, 0.1, 0.1, float32(colorAlpha) * 0.2}
	borderOuter := mgl32.Vec4{0.1, 0.1, 0.1, float32(colorAlpha) * 0.2}

	if settings.Skin.UseColorsFromSkin {
		borderOuter = skin.GetInfo().SliderBorder.ToVec4()
		borderInner = borderOuter

		borderOuter[3] = float32(colorAlpha)
		borderInner[3] = float32(colorAlpha)

		var baseTrack color2.Color

		if skin.GetInfo().SliderTrackOverride != nil {
			baseTrack = *skin.GetInfo().SliderTrackOverride
		} else {
			baseTrack = skin.GetInfo().ComboColors[int(self.objData.ComboSet)%len(skin.GetInfo().ComboColors)]
		}

		bodyOuter[0] = baseTrack.R / 1.1
		bodyOuter[1] = baseTrack.G / 1.1
		bodyOuter[2] = baseTrack.B / 1.1

		bodyInner[0] = baseTrack.R + (1-baseTrack.R)*0.25
		bodyInner[1] = baseTrack.G + (1-baseTrack.G)*0.25
		bodyInner[2] = baseTrack.B + (1-baseTrack.B)*0.25
	} else if settings.Objects.EnableCustomSliderBorderGradientOffset {
		borderOuter = utils.GetColorShifted(color, settings.Objects.SliderBorderGradientOffset)
		borderOuter[3] = float32(colorAlpha)
	}

	self.body.DrawNormal(projection, self.objData.StackOffset, scale, bodyInner, bodyOuter, borderInner, borderOuter)
}

func (self *Slider) Draw(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) bool {
	if len(self.scorePath) == 0 {
		return true
	}

	alpha := self.fade.GetValue()
	//alphaF := self.fadeCircle.GetValue()

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

	if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger {

		if settings.Objects.DrawReverseArrows {
			headPos := self.multiCurve.PointAt(float32(self.sliderSnakeHead.GetValue()))
			headAngle := self.multiCurve.GetStartAngleAt(float32(self.sliderSnakeHead.GetValue())) + math.Pi

			tailPos := self.multiCurve.PointAt(float32(self.sliderSnakeTail.GetValue()))
			tailAngle := self.multiCurve.GetEndAngleAt(float32(self.sliderSnakeTail.GetValue())) + math.Pi

			arrow := skin.GetTexture("reversearrow")

			for i := 0; i < 2; i++ {
				for _, p := range self.reversePoints[i] {
					if p.fade.GetValue() >= 0.001 {
						if i == 1 {
							batch.SetTranslation(tailPos.Copy64())
							batch.SetRotation(float64(tailAngle))
						} else {
							batch.SetTranslation(headPos.Copy64())
							batch.SetRotation(float64(headAngle))
						}

						batch.SetSubScale(p.pulse.GetValue(), p.pulse.GetValue())
						batch.SetColor(1, 1, 1, alpha*p.fade.GetValue())
						batch.DrawTexture(*arrow)
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

				scorePoint := skin.GetTexture("sliderscorepoint")

				for _, p := range self.TickPoints {
					al := p.fade.GetValue()

					if al > 0.001 {
						batch.SetTranslation(p.Pos.Copy64())
						batch.SetSubScale(p.scale.GetValue(), p.scale.GetValue())

						if settings.Objects.WhiteFollowPoints {
							batch.SetColor(1, 1, 1, alpha*al)
						} else {
							batch.SetColor(float64(shifted[0]), float64(shifted[1]), float64(shifted[2]), alpha*al)
						}

						batch.DrawTexture(*scorePoint)
					}
				}
			}
		}

		batch.SetSubScale(1, 1)
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), alpha)

		if time >= self.objData.StartTime && time <= self.objData.EndTime {
			self.drawBall(time, batch, alpha, true)
		}

		if settings.Objects.DrawSliderFollowCircle && self.follower != nil {
			batch.SetTranslation(self.Pos.Copy64())
			batch.SetColor(1, 1, 1, alpha)
			self.follower.Draw(time, batch)
		}

	} else {
		if time >= self.objData.StartTime && time <= self.objData.EndTime {
			self.drawBall(time, batch, alpha, settings.Objects.ForceSliderBallTexture)
		}
	}

	self.startCircle.Draw(time, color, batch)

	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))

	if time >= self.objData.EndTime && self.fade.GetValue() <= 0.001 {
		if self.body != nil {
			if !settings.Graphics.VSync {
				self.body.Dispose()
			}
		}

		return true
	}

	return false
}

func (self *Slider) drawBall(time int64, batch *sprite.SpriteBatch, alpha float64, useBallTexture bool) {
	batch.SetColor(1, 1, 1, alpha)
	batch.SetTranslation(self.Pos.Copy64())

	isB := skin.GetTexture("sliderb") == nil && useBallTexture

	if isB && skin.GetTexture("sliderb-nd") != nil {
		batch.SetColor(0.1, 0.1, 0.1, alpha)
		batch.DrawTexture(*skin.GetTexture("sliderb-nd"))
	}

	batch.SetColor(1, 1, 1, alpha)
	//cHSV := settings.Objects.ComboColors[int(self.objData.ComboSet)%len(settings.Objects.ComboColors)]
	//r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
	//batch.SetColor(float64(r), float64(g), float64(b), alpha)

	if useBallTexture {
		self.ball.Draw(time, batch)
	} else {
		batch.DrawTexture(*skin.GetTexture("hitcircle-full"))
	}

	batch.SetColor(1, 1, 1, alpha)

	if isB && skin.GetTexture("sliderb-spec") != nil {
		batch.SetAdditive(true)
		batch.DrawTexture(*skin.GetTexture("sliderb-spec"))
		batch.SetAdditive(false)
	}
}

func (self *Slider) DrawApproach(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) {
	if len(self.scorePath) == 0 {
		return
	}

	self.startCircle.DrawApproach(time, color, batch)
}
