package objects

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics/sliderrenderer"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/curves"
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
	objData     *basicData
	multiCurve  *curves.MultiCurve
	scorePath   []PathLine
	Timings     *Timings
	TPoint      TimingPoint
	pixelLength float64
	partLen     float64
	repeat      int64

	sampleSets   []int
	additionSets []int
	samples      []int
	baseSample   int

	Pos         vector.Vector2f
	TickPoints  []TickPoint
	TickReverse []TickPoint
	ScorePoints []TickPoint

	startCircle *Circle

	startAngle, endAngle float64
	sliderSnakeTail      *animation.Glider
	sliderSnakeHead      *animation.Glider
	fade                 *animation.Glider
	bodyFade             *animation.Glider

	diff     *difficulty.Difficulty
	body     *sliderrenderer.Body
	lastTime int64

	ball     *sprite.Sprite
	follower *sprite.Sprite

	edges          []*Circle
	endCircles     []*Circle
	headEndCircles []*Circle
	tailEndCircles []*Circle

	isSliding bool
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

	f, _ := strconv.ParseInt(data[4], 10, 64)
	slider.baseSample = int(f)

	for i := range slider.samples {
		slider.samples[i] = slider.baseSample
	}

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

func (slider Slider) GetBasicData() *basicData {
	return slider.objData
}

func (slider Slider) GetHalf() vector.Vector2f {
	return slider.multiCurve.PointAt(0.5).Add(slider.objData.StackOffset)
}

func (slider Slider) GetStartAngle() float32 {
	return slider.GetBasicData().StartPos.AngleRV(slider.GetPointAt(slider.objData.StartTime + int64(math.Min(10, slider.partLen)))) //temporary solution
}

func (slider Slider) GetEndAngle() float32 {
	return slider.GetBasicData().EndPos.AngleRV(slider.GetPointAt(slider.objData.EndTime - int64(math.Min(10, slider.partLen)))) //temporary solution
}

func (slider Slider) GetPartLen() float32 {
	return float32(20.0) / float32(slider.Timings.GetSliderTimeP(slider.TPoint, slider.pixelLength)) * float32(slider.pixelLength)
}

func (slider Slider) GetPointAt(time int64) vector.Vector2f {
	if slider.IsRetarded() {
		return slider.objData.StartPos
	}

	index := sort.Search(len(slider.scorePath), func(i int) bool {
		return slider.scorePath[i].Time2 >= time
	})

	pLine := slider.scorePath[bmath.ClampI(index, 0, len(slider.scorePath)-1)]

	clamped := bmath.ClampI64(time, pLine.Time1, pLine.Time2)

	var pos vector.Vector2f
	if pLine.Time2 == pLine.Time1 {
		pos = pLine.Line.Point2
	} else {
		pos = pLine.Line.PointAt(float32(clamped-pLine.Time1) / float32(pLine.Time2-pLine.Time1))
	}

	return pos.Add(slider.objData.StackOffset)
}

func (slider *Slider) GetAsDummyCircles() []BaseObject {
	partLen := int64(slider.Timings.GetSliderTimeP(slider.TPoint, slider.pixelLength))

	var circles []BaseObject

	for i := int64(0); i <= slider.repeat; i++ {
		time := slider.objData.StartTime + i*partLen

		if i == slider.repeat && settings.KNOCKOUT {
			time = int64(math.Max(float64(slider.GetBasicData().StartTime)+float64((slider.GetBasicData().EndTime-slider.GetBasicData().StartTime)/2), float64(slider.GetBasicData().EndTime-36)))
		}

		circles = append(circles, DummyCircleInherit(slider.GetPointAt(time), time, true, i == 0, i == slider.repeat))
	}

	for _, p := range slider.TickPoints {
		circles = append(circles, DummyCircleInherit(p.Pos, p.Time, true, false, false))
	}

	sort.Slice(circles, func(i, j int) bool { return circles[i].GetBasicData().StartTime < circles[j].GetBasicData().StartTime })

	return circles
}

func (slider *Slider) SetTiming(timings *Timings) {
	slider.Timings = timings
	slider.TPoint = timings.GetPoint(slider.objData.StartTime)

	lines := slider.multiCurve.GetLines()

	startTime := float64(slider.GetBasicData().StartTime)

	velocity := slider.Timings.GetVelocity(slider.TPoint)

	minDistanceFromEnd := velocity * 0.01

	scoringLengthTotal := 0.0
	scoringDistance := 0.0

	tickDistance := math.Min(slider.Timings.GetTickDistance(slider.TPoint), slider.pixelLength)

	for i := int64(0); i < slider.repeat; i++ {
		distanceToEnd := float64(slider.multiCurve.GetLength())
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

			slider.scorePath = append(slider.scorePath, PathLine{Time1: int64(startTime), Time2: int64(startTime + progress), Line: curves.NewLinear(p1, p2)})

			startTime += progress
			slider.objData.EndTime = int64(math.Floor(startTime))
			/*if slider.objData.StartTime == 120273 {
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

				scoreTime := slider.GetBasicData().StartTime + int64(float64(float32(scoringLengthTotal)*1000)/velocity)

				point := TickPoint{scoreTime, slider.GetPointAt(scoreTime), animation.NewGlider(0.0), animation.NewGlider(0.0), false}
				slider.TickPoints = append(slider.TickPoints, point)
				slider.ScorePoints = append(slider.ScorePoints, point)

			}
		}

		scoringLengthTotal += scoringDistance

		scoreTime := slider.GetBasicData().StartTime + int64((float64(float32(scoringLengthTotal))/velocity)*1000)
		point := TickPoint{scoreTime, slider.GetPointAt(scoreTime), nil, nil, true}

		slider.TickReverse = append(slider.TickReverse, point)
		slider.ScorePoints = append(slider.ScorePoints, point)

		// If our scoring distance is small enough, then there was no "last" scoring point at the end. No need to mirror a non-existing point.
		if skipTick {
			scoringDistance = 0
		} else {
			scoringLengthTotal -= tickDistance - scoringDistance
			scoringDistance = tickDistance - scoringDistance
		}
	}

	slider.partLen = float64(slider.objData.EndTime-slider.objData.StartTime) / float64(slider.repeat)

	slider.objData.EndPos = slider.GetPointAt(slider.objData.EndTime)

	if len(slider.scorePath) == 0 || slider.objData.StartTime == slider.objData.EndTime {
		log.Println("Warning: slider", slider.objData.Number, "at ", slider.objData.StartTime, "is broken.")
	}

	slider.calculateFollowPoints()

	slider.startAngle = float64(slider.GetStartAngle())
	if len(lines) > 0 {
		slider.endAngle = float64(lines[len(lines)-1].GetEndAngle())
	} else {
		slider.endAngle = slider.startAngle + math.Pi
	}
}

func (slider *Slider) calculateFollowPoints() {
	sort.Slice(slider.TickPoints, func(i, j int) bool { return slider.TickPoints[i].Time < slider.TickPoints[j].Time })
	sort.Slice(slider.ScorePoints, func(i, j int) bool { return slider.ScorePoints[i].Time < slider.ScorePoints[j].Time })
}

func (slider *Slider) UpdateStacking() {
	for i, tp := range slider.TickPoints {
		tp.Pos = tp.Pos.Add(slider.objData.StackOffset)
		slider.TickPoints[i] = tp
	}
}

func (slider *Slider) SetDifficulty(diff *difficulty.Difficulty) {
	slider.diff = diff
	slider.sliderSnakeTail = animation.NewGlider(0)
	slider.sliderSnakeHead = animation.NewGlider(0)

	fadeMultiplier := 1.0 - bmath.ClampF64(settings.Objects.Sliders.Snaking.FadeMultiplier, 0.0, 1.0)
	durationMultiplier := bmath.ClampF64(settings.Objects.Sliders.Snaking.DurationMultiplier, 0.0, 1.0)

	slSnInS := float64(slider.objData.StartTime) - diff.Preempt
	slSnInE := float64(slider.objData.StartTime) - diff.Preempt*2/3*fadeMultiplier + slider.partLen*durationMultiplier

	if settings.Objects.Sliders.Snaking.In {
		slider.sliderSnakeTail.AddEvent(slSnInS, slSnInE, 1)
	} else {
		slider.sliderSnakeTail.SetValue(1)
	}

	if settings.Objects.Sliders.Snaking.Out {
		if slider.repeat%2 == 1 {
			slider.sliderSnakeHead.AddEvent(float64(slider.objData.EndTime)-slider.partLen, float64(slider.objData.EndTime), 1)
		} else {
			slider.sliderSnakeTail.AddEvent(float64(slider.objData.EndTime)-slider.partLen, float64(slider.objData.EndTime), 0)
		}
	}

	slider.fade = animation.NewGlider(0)
	slider.fade.AddEvent(float64(slider.objData.StartTime)-diff.Preempt, float64(slider.objData.StartTime)-(diff.Preempt-difficulty.HitFadeIn), 1)

	slider.bodyFade = animation.NewGlider(0)
	slider.bodyFade.AddEvent(float64(slider.objData.StartTime)-diff.Preempt, float64(slider.objData.StartTime)-(diff.Preempt-difficulty.HitFadeIn), 1)

	if diff.CheckModActive(difficulty.Hidden) {
		slider.bodyFade.AddEvent(float64(slider.objData.StartTime)-diff.Preempt+difficulty.HitFadeIn, float64(slider.objData.EndTime), 0)
	} else {
		slider.bodyFade.AddEvent(float64(slider.objData.EndTime), float64(slider.objData.EndTime)+difficulty.HitFadeOut, 0)
	}

	slider.fade.AddEvent(float64(slider.objData.EndTime), float64(slider.objData.EndTime)+difficulty.HitFadeOut, 0)

	slider.startCircle = DummyCircle(slider.objData.StartPos, slider.objData.StartTime)
	slider.startCircle.objData.ComboNumber = slider.objData.ComboNumber
	slider.startCircle.objData.ComboSet = slider.objData.ComboSet
	slider.startCircle.objData.Number = slider.objData.Number
	slider.startCircle.SetDifficulty(diff)

	slider.edges = append(slider.edges, slider.startCircle)

	sixty := 1000.0 / 60
	frameDelay := math.Max(150/slider.Timings.GetVelocity(slider.TPoint)*sixty, sixty)

	slider.ball = sprite.NewAnimation(skin.GetFrames("sliderb", false), frameDelay, true, 0.0, vector.NewVec2d(0, 0), bmath.Origin.Centre)

	if len(slider.scorePath) > 0 {
		angle := slider.scorePath[0].Line.GetStartAngle()
		slider.ball.SetVFlip(angle > -math32.Pi/2 && angle < math32.Pi/2)
	}

	followerFrames := skin.GetFrames("sliderfollowcircle", true)

	slider.follower = sprite.NewAnimation(followerFrames, 1000.0/float64(len(followerFrames)), true, 0.0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	slider.follower.SetAlpha(0.0)

	for i := int64(1); i <= slider.repeat; i++ {
		appearTime := slider.objData.StartTime - int64(slider.diff.Preempt)
		circleTime := slider.objData.StartTime + int64(slider.partLen*float64(i))
		if i > 1 {
			appearTime = circleTime - int64(slider.partLen*2)
		}

		circle := NewSliderEndCircle(vector.NewVec2f(0, 0), appearTime, circleTime, i == 1, i == slider.repeat)
		circle.objData.ComboNumber = slider.objData.ComboNumber
		circle.objData.ComboSet = slider.objData.ComboSet
		circle.objData.Number = slider.objData.Number
		circle.SetTiming(slider.Timings)
		circle.SetDifficulty(diff)

		slider.endCircles = append(slider.endCircles, circle)
		slider.edges = append(slider.edges, circle)

		if i%2 == 0 {
			slider.headEndCircles = append(slider.headEndCircles, circle)
		} else {
			slider.tailEndCircles = append(slider.tailEndCircles, circle)
		}
	}

	for _, p := range slider.TickPoints {
		a := float64((p.Time-slider.objData.StartTime)/2+slider.objData.StartTime) - diff.Preempt*2/3

		fs := float64(p.Time-slider.objData.StartTime) / slider.partLen

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

	slider.body = sliderrenderer.NewBody(slider.multiCurve, float32(slider.diff.CircleRadius))
}

func (slider *Slider) IsRetarded() bool {
	return len(slider.scorePath) == 0 || slider.objData.StartTime == slider.objData.EndTime
}

func (slider *Slider) Update(time int64) bool {
	if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {

		for i := int64(0); i <= slider.repeat; i++ {
			edgeTime := slider.objData.StartTime + int64(float64(i)*slider.partLen)

			if slider.lastTime < edgeTime && time >= edgeTime {
				slider.HitEdge(int(i), time, true)

				if i == 0 {
					slider.InitSlide(slider.objData.StartTime)
				}
			}
		}

		for _, p := range slider.TickPoints {
			if slider.lastTime < p.Time && time >= p.Time {
				audio.PlaySliderTick(slider.Timings.Current.SampleSet, slider.Timings.Current.SampleIndex, slider.Timings.Current.SampleVolume, slider.objData.Number, p.Pos.X64())
			}
		}
	} else if slider.isSliding {
		for i := int64(1); i < slider.repeat; i++ {
			edgeTime := slider.objData.StartTime + int64(float64(i)*slider.partLen)

			if slider.lastTime < edgeTime && time >= edgeTime {
				slider.HitEdge(int(i), time, true)
			}
		}

		for _, p := range slider.TickPoints {
			if slider.lastTime < p.Time && time >= p.Time {
				audio.PlaySliderTick(slider.Timings.Current.SampleSet, slider.Timings.Current.SampleIndex, slider.Timings.Current.SampleVolume, slider.objData.Number, p.Pos.X64())
			}
		}
	}

	slider.sliderSnakeHead.Update(float64(time))
	slider.sliderSnakeTail.Update(float64(time))

	if slider.startCircle != nil {
		slider.startCircle.Update(time)
	}

	if slider.ball != nil {
		slider.ball.Update(time)
	}

	if slider.follower != nil {
		slider.follower.Update(time)
	}

	slider.fade.Update(float64(time))
	slider.bodyFade.Update(float64(time))

	headPos := slider.multiCurve.PointAt(float32(slider.sliderSnakeHead.GetValue())).Add(slider.objData.StackOffset)
	headAngle := slider.multiCurve.GetStartAngleAt(float32(slider.sliderSnakeHead.GetValue())) + math.Pi

	for _, s := range slider.headEndCircles {
		s.ArrowRotation = float64(headAngle)
		s.objData.StartPos = headPos
		s.Update(time)
	}

	tailPos := slider.multiCurve.PointAt(float32(slider.sliderSnakeTail.GetValue())).Add(slider.objData.StackOffset)
	tailAngle := slider.multiCurve.GetEndAngleAt(float32(slider.sliderSnakeTail.GetValue())) + math.Pi

	for _, s := range slider.tailEndCircles {
		s.ArrowRotation = float64(tailAngle)
		s.objData.StartPos = tailPos
		s.Update(time)
	}

	for _, p := range slider.TickPoints {
		p.fade.Update(float64(time))
		p.scale.Update(float64(time))
	}

	pos := slider.GetPointAt(time)

	if time-slider.lastTime > 0 && time >= slider.objData.StartTime {
		angle := pos.AngleRV(slider.Pos)

		reversed := int(float64(time-slider.objData.StartTime)/slider.partLen)%2 == 1

		if reversed {
			angle -= math32.Pi
		}

		if slider.ball != nil {
			slider.ball.SetHFlip(skin.GetInfo().SliderBallFlip && reversed)
			slider.ball.SetRotation(float64(angle))
		}
	}

	if slider.lastTime < slider.objData.EndTime && time >= slider.objData.EndTime && slider.isSliding {
		slider.StopSlideSamples()
		slider.isSliding = false
	}

	if slider.isSliding {
		slider.PlaySlideSamples()
	}

	slider.Pos = pos

	slider.lastTime = time

	return true
}

func (slider *Slider) ArmStart(clicked bool, time int64) {
	slider.startCircle.Arm(clicked, time)
}

func (slider *Slider) InitSlide(time int64) {
	slider.follower.ClearTransformations()

	startTime := float64(time)

	slider.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, math.Min(startTime+60, float64(slider.objData.EndTime)), 0, 1))
	slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, math.Min(startTime+180, float64(slider.objData.EndTime)), 0.5, 1))

	for j, p := range slider.ScorePoints {
		if j < 1 || p.Time < time {
			continue
		}

		fade := 200.0
		delay := fade
		if len(slider.ScorePoints) >= 2 {
			delay = math.Min(fade, float64(p.Time-slider.ScorePoints[j-1].Time))
			ratio := delay / fade

			slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(p.Time), float64(p.Time)+delay, 1.1, 1.1-ratio*0.1))
		}
	}

	slider.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, float64(slider.objData.EndTime), float64(slider.objData.EndTime+200), 1, 0))
	slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, float64(slider.objData.EndTime), float64(slider.objData.EndTime+200), 1, 0.8))

	slider.isSliding = true
}

func (slider *Slider) KillSlide(time int64) {
	slider.follower.ClearTransformations()

	nextPoint := slider.objData.EndTime
	for _, p := range slider.ScorePoints {
		if p.Time > time {
			nextPoint = p.Time
			break
		}
	}

	slider.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(nextPoint-100), float64(nextPoint), 1, 0))
	slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(nextPoint-100), float64(nextPoint), 1, 2))

	slider.isSliding = false
	slider.StopSlideSamples()
}

func (slider *Slider) PlaySlideSamples() {
	point := slider.Timings.Current

	sampleSet := slider.objData.sampleSet
	if sampleSet == 0 {
		sampleSet = point.SampleSet
	}

	audio.PlaySliderLoops(sampleSet, slider.objData.additionSet, slider.baseSample, point.SampleIndex, point.SampleVolume, slider.objData.Number, slider.Pos.X64())
}

func (slider *Slider) StopSlideSamples() {
	audio.StopSliderLoops()
}

func (slider *Slider) PlayEdgeSample(index int) {
	slider.playSampleT(slider.sampleSets[index], slider.additionSets[index], slider.samples[index], slider.Timings.GetPoint(slider.objData.StartTime+int64(float64(index)*slider.partLen)+5), slider.GetPointAt(slider.objData.StartTime+int64(float64(index)*slider.partLen)))
}

func (slider *Slider) HitEdge(index int, time int64, isHit bool) {
	e := slider.edges[index]
	e.Arm(isHit, time)

	if isHit {
		slider.PlayEdgeSample(index)
	}
}

func (slider *Slider) PlayTick() {
	audio.PlaySliderTick(slider.Timings.Current.SampleSet, slider.Timings.Current.SampleIndex, slider.Timings.Current.SampleVolume, slider.objData.Number, slider.Pos.X64())
}

func (slider *Slider) playSample(sampleSet, additionSet, sample int) {
	slider.playSampleT(sampleSet, additionSet, sampleSet, slider.Timings.Current, vector.NewVec2f(0, 0))
}

func (slider *Slider) playSampleT(sampleSet, additionSet, sample int, point TimingPoint, pos vector.Vector2f) {
	if sampleSet == 0 {
		sampleSet = slider.objData.sampleSet
		if sampleSet == 0 {
			sampleSet = point.SampleSet
		}
	}

	if additionSet == 0 {
		additionSet = slider.objData.additionSet
	}

	audio.PlaySample(sampleSet, additionSet, sample, point.SampleIndex, point.SampleVolume, slider.objData.Number, pos.X64())
}

func (slider *Slider) GetPosition() vector.Vector2f {
	return slider.Pos
}

func (slider *Slider) DrawBodyBase(time int64, projection mgl32.Mat4) {
	slider.body.DrawBase(slider.sliderSnakeHead.GetValue(), slider.sliderSnakeTail.GetValue(), projection)
}

func (slider *Slider) DrawBody(time int64, bodyColor, innerBorder, outerBorder color2.Color, projection mgl32.Mat4, scale float32) {
	colorAlpha := slider.bodyFade.GetValue()

	bodyOpacityInner := bmath.ClampF32(float32(settings.Objects.Colors.Sliders.Body.InnerAlpha), 0.0, 1.0)
	bodyOpacityOuter := bmath.ClampF32(float32(settings.Objects.Colors.Sliders.Body.OuterAlpha), 0.0, 1.0)

	borderInner := color2.NewRGBA(innerBorder.R, innerBorder.G, innerBorder.B, float32(colorAlpha))
	borderOuter := color2.NewRGBA(outerBorder.R, outerBorder.G, outerBorder.B, float32(colorAlpha))
	bodyInner := color2.NewL(0)
	bodyOuter := color2.NewL(0)

	if settings.Skin.UseColorsFromSkin {
		borderOuter = skin.GetInfo().SliderBorder
		borderInner = borderOuter

		borderOuter.A = float32(colorAlpha)
		borderInner.A = float32(colorAlpha)

		var baseTrack color2.Color

		if skin.GetInfo().SliderTrackOverride != nil {
			baseTrack = *skin.GetInfo().SliderTrackOverride
		} else {
			baseTrack = skin.GetInfo().ComboColors[int(slider.objData.ComboSet)%len(skin.GetInfo().ComboColors)]
		}

		bodyOuter = baseTrack.Shade2(-0.1)
		bodyInner = baseTrack.Shade2(0.5)
	} else {
		if settings.Objects.Colors.UseComboColors {
			cHSV := settings.Objects.Colors.ComboColors[int(slider.objData.ComboSet)%len(settings.Objects.Colors.ComboColors)]
			comnboColor := color2.NewHSV(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))

			if settings.Objects.Colors.Sliders.Border.UseHitCircleColor {
				borderInner = comnboColor
				borderOuter = comnboColor
			}

			if settings.Objects.Colors.Sliders.Body.UseHitCircleColor {
				bodyColor = comnboColor
			}
		}

		if settings.Objects.Colors.Sliders.Border.EnableCustomGradientOffset {
			borderOuter = borderInner.Shift(float32(settings.Objects.Colors.Sliders.Border.CustomGradientOffset), 0, 0)
		}

		bodyInner = bodyColor.Shade2(float32(settings.Objects.Colors.Sliders.Body.InnerOffset))
		bodyOuter = bodyColor.Shade2(float32(settings.Objects.Colors.Sliders.Body.OuterOffset))
	}

	borderInner.A = float32(colorAlpha)
	borderOuter.A = float32(colorAlpha)
	bodyInner.A = float32(colorAlpha) * bodyOpacityInner
	bodyOuter.A = float32(colorAlpha) * bodyOpacityOuter

	slider.body.DrawNormal(projection, slider.objData.StackOffset, scale, bodyInner, bodyOuter, borderInner, borderOuter)
}

func (slider *Slider) Draw(time int64, color color2.Color, batch *batch.QuadBatch) bool {
	if len(slider.scorePath) == 0 {
		return true
	}

	alpha := slider.fade.GetValue()

	if settings.DIVIDES >= settings.Objects.Colors.MandalaTexturesTrigger {
		alpha *= settings.Objects.Colors.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)

	if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger {
		if time < slider.objData.EndTime {
			if settings.Objects.Sliders.DrawScorePoints {
				shifted := color.Shift(float32(settings.Objects.Colors.Sliders.ScorePointColorOffset), 0, 0)

				scorePoint := skin.GetTexture("sliderscorepoint")

				for _, p := range slider.TickPoints {
					al := p.fade.GetValue()

					if al > 0.001 {
						batch.SetTranslation(p.Pos.Copy64())
						batch.SetSubScale(p.scale.GetValue(), p.scale.GetValue())

						if settings.Objects.Colors.Sliders.WhiteScorePoints || settings.Skin.UseColorsFromSkin {
							batch.SetColor(1, 1, 1, alpha*al)
						} else {
							batch.SetColor(float64(shifted.R), float64(shifted.G), float64(shifted.B), alpha*al)
						}

						batch.DrawTexture(*scorePoint)
					}
				}
			}
		}

		batch.SetSubScale(1, 1)

		if settings.Objects.Sliders.DrawEndCircles {
			for i := len(slider.endCircles) - 1; i >= 0; i-- {
				slider.endCircles[i].Draw(time, color, batch)
			}
		}
	}

	batch.SetColor(1, 1, 1, 1)
	slider.startCircle.Draw(time, color, batch)

	if time >= slider.objData.StartTime && time <= slider.objData.EndTime {
		slider.drawBall(time, batch, color, alpha, settings.Objects.Sliders.ForceSliderBallTexture || settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger)
	}

	if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.Sliders.DrawSliderFollowCircle && slider.follower != nil {
		batch.SetTranslation(slider.Pos.Copy64())
		batch.SetColor(1, 1, 1, alpha)
		slider.follower.Draw(time, batch)
	}

	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))

	if time >= slider.objData.EndTime && slider.fade.GetValue() <= 0.001 {
		if slider.body != nil {
			if !settings.Graphics.VSync {
				slider.body.Dispose()
			}
		}

		return true
	}

	return false
}

func (slider *Slider) drawBall(time int64, batch *batch.QuadBatch, color color2.Color, alpha float64, useBallTexture bool) {
	batch.SetColor(1, 1, 1, alpha)
	batch.SetTranslation(slider.Pos.Copy64())

	isB := skin.GetSource("sliderb") != skin.SKIN && useBallTexture

	if isB && skin.GetTexture("sliderb-nd") != nil {
		batch.SetColor(0.1, 0.1, 0.1, alpha)
		batch.DrawTexture(*skin.GetTexture("sliderb-nd"))
	}

	if settings.Skin.UseColorsFromSkin {
		color := color2.NewL(1)

		if skin.GetInfo().SliderBallTint {
			color = skin.GetInfo().ComboColors[int(slider.objData.ComboSet)%len(skin.GetInfo().ComboColors)]
		} else if skin.GetInfo().SliderBall != nil {
			color = *skin.GetInfo().SliderBall
		}

		batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)
	} else if settings.Objects.Colors.Sliders.SliderBallTint {
		if settings.Objects.Colors.UseComboColors {
			cHSV := settings.Objects.Colors.ComboColors[int(slider.objData.ComboSet)%len(settings.Objects.Colors.ComboColors)]
			r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
			batch.SetColor(float64(r), float64(g), float64(b), alpha)
		} else {
			batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)
		}
	} else {
		batch.SetColor(1, 1, 1, alpha)
	}

	//cHSV := settings.Objects.Colors.ComboColors[int(slider.objData.ComboSet)%len(settings.Objects.Colors.ComboColors)]
	//r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
	//batch.SetColor(float64(r), float64(g), float64(b), alpha)

	if useBallTexture {
		slider.ball.Draw(time, batch)
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

func (slider *Slider) DrawApproach(time int64, color color2.Color, batch *batch.QuadBatch) {
	if len(slider.scorePath) == 0 {
		return
	}

	slider.startCircle.DrawApproach(time, color, batch)
}
