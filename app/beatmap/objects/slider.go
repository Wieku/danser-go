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
	Time      float64
	Pos       vector.Vector2f
	fade      *animation.Glider
	scale     *animation.Glider
	IsReverse bool
}

var easeBezier = curves.NewMultiCurve("B", []vector.Vector2f{{X: 0, Y: 0}, {X: 0.1, Y: 1}, {X: 0.5, Y: 0.5}, {X: 1, Y: 1}})

var snakeEase = easing.Easing(func(f float64) float64 {
	return float64(easeBezier.PointAt(float32(f)).Y)
})

type Slider struct {
	*HitObject

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
	lastTime float64

	ball     *sprite.Sprite
	follower *sprite.Sprite

	edges          []*Circle
	endCircles     []*Circle
	headEndCircles []*Circle
	tailEndCircles []*Circle

	isSliding bool
}

func NewSlider(data []string) *Slider {
	slider := &Slider{
		HitObject: commonParse(data, 10),
	}

	slider.PositionDelegate = slider.PositionAt

	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.repeat, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []vector.Vector2f{slider.StartPosRaw}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 32)
		y, _ := strconv.ParseFloat(list2[1], 32)
		points = append(points, vector.NewVec2f(float32(x), float32(y)))
	}

	slider.multiCurve = curves.NewMultiCurveT(list[0], points, slider.pixelLength)

	slider.EndTime = slider.StartTime
	slider.EndPosRaw = slider.multiCurve.PointAt(1.0)
	slider.Pos = slider.StartPosRaw

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

	slider.fade = animation.NewGlider(1)
	slider.bodyFade = animation.NewGlider(1)
	slider.sliderSnakeTail = animation.NewGlider(1)
	slider.sliderSnakeHead = animation.NewGlider(0)

	return slider
}

func (slider *Slider) GetHalf() vector.Vector2f {
	return slider.multiCurve.PointAt(0.5).Add(slider.StackOffset)
}

func (slider *Slider) GetStartAngle() float32 {
	return slider.GetStackedStartPosition().AngleRV(slider.GetStackedPositionAt(slider.StartTime + math.Min(10, slider.partLen))) //temporary solution
}

func (slider *Slider) GetStartAngleMod(modifier difficulty.Modifier) float32 {
	return slider.GetStackedStartPositionMod(modifier).AngleRV(slider.GetStackedPositionAtMod(slider.StartTime+math.Min(10, slider.partLen), modifier)) //temporary solution
}

func (slider *Slider) GetEndAngle() float32 {
	return slider.GetStackedEndPosition().AngleRV(slider.GetStackedPositionAt(slider.EndTime - math.Min(10, slider.partLen))) //temporary solution
}

func (slider *Slider) GetEndAngleMod(modifier difficulty.Modifier) float32 {
	return slider.GetStackedEndPositionMod(modifier).AngleRV(slider.GetStackedPositionAtMod(slider.EndTime-math.Min(10, slider.partLen), modifier)) //temporary solution
}

func (slider *Slider) GetPartLen() float32 {
	return float32(20.0) / float32(slider.Timings.GetSliderTimeP(slider.TPoint, slider.pixelLength)) * float32(slider.pixelLength)
}

func (slider *Slider) PositionAt(time float64) vector.Vector2f {
	if slider.IsRetarded() {
		return slider.StartPosRaw
	}

	index := sort.Search(len(slider.scorePath), func(i int) bool {
		return float64(slider.scorePath[i].Time2) >= time
	})

	pLine := slider.scorePath[bmath.ClampI(index, 0, len(slider.scorePath)-1)]

	clamped := bmath.ClampF64(time, float64(pLine.Time1), float64(pLine.Time2))

	var pos vector.Vector2f
	if pLine.Time2 == pLine.Time1 {
		pos = pLine.Line.Point2
	} else {
		pos = pLine.Line.PointAt(float32(clamped-float64(pLine.Time1)) / float32(pLine.Time2-pLine.Time1))
	}

	return pos
}

func (slider *Slider) GetAsDummyCircles() []IHitObject {
	partLen := int64(slider.Timings.GetSliderTimeP(slider.TPoint, slider.pixelLength))

	var circles []IHitObject

	for i := int64(0); i <= slider.repeat; i++ {
		time := slider.StartTime + float64(i*partLen)

		if i == slider.repeat && settings.KNOCKOUT {
			time = math.Floor(math.Max(slider.StartTime+(slider.EndTime-slider.StartTime)/2, slider.EndTime-36))
		}

		circles = append(circles, DummyCircleInherit(slider.GetPositionAt(time), time, true, i == 0, i == slider.repeat))
	}

	for _, p := range slider.TickPoints {
		circles = append(circles, DummyCircleInherit(p.Pos, p.Time, true, false, false))
	}

	sort.Slice(circles, func(i, j int) bool { return circles[i].GetStartTime() < circles[j].GetStartTime() })

	return circles
}

func (slider *Slider) SetTiming(timings *Timings) {
	slider.Timings = timings
	slider.TPoint = timings.GetPoint(slider.StartTime)

	lines := slider.multiCurve.GetLines()

	startTime := slider.StartTime

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
			slider.EndTime = math.Floor(startTime)

			scoringDistance += float64(distance)

			for scoringDistance >= tickDistance && !skipTick {
				scoringLengthTotal += tickDistance
				scoringDistance -= tickDistance
				distanceToEnd -= tickDistance

				skipTick = distanceToEnd <= minDistanceFromEnd
				if skipTick {
					break
				}

				scoreTime := slider.StartTime + math.Floor(float64(float32(scoringLengthTotal)*1000)/velocity)

				point := TickPoint{scoreTime, slider.GetPositionAt(scoreTime), animation.NewGlider(0.0), animation.NewGlider(0.0), false}
				slider.TickPoints = append(slider.TickPoints, point)
				slider.ScorePoints = append(slider.ScorePoints, point)
			}
		}

		scoringLengthTotal += scoringDistance

		scoreTime := slider.StartTime + math.Floor((float64(float32(scoringLengthTotal))/velocity)*1000)
		point := TickPoint{scoreTime, slider.GetPositionAt(scoreTime), nil, nil, true}

		slider.TickReverse = append(slider.TickReverse, point)
		slider.ScorePoints = append(slider.ScorePoints, point)

		if skipTick {
			scoringDistance = 0
		} else {
			scoringLengthTotal -= tickDistance - scoringDistance
			scoringDistance = tickDistance - scoringDistance
		}
	}

	slider.partLen = (slider.EndTime - slider.StartTime) / float64(slider.repeat)

	slider.EndPosRaw = slider.GetPositionAt(slider.EndTime)

	if len(slider.scorePath) == 0 || slider.StartTime == slider.EndTime {
		log.Println("Warning: slider", slider.HitObjectID, "at ", slider.StartTime, "is broken.")
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
		tp.Pos = tp.Pos.Add(slider.StackOffset)
		slider.TickPoints[i] = tp
	}
}

func (slider *Slider) SetDifficulty(diff *difficulty.Difficulty) {
	slider.diff = diff
	slider.sliderSnakeTail = animation.NewGlider(0)
	slider.sliderSnakeHead = animation.NewGlider(0)

	fadeMultiplier := 1.0 - bmath.ClampF64(settings.Objects.Sliders.Snaking.FadeMultiplier, 0.0, 1.0)
	durationMultiplier := bmath.ClampF64(settings.Objects.Sliders.Snaking.DurationMultiplier, 0.0, 1.0)

	slSnInS := slider.StartTime - diff.Preempt
	slSnInE := slider.StartTime - diff.Preempt*2/3*fadeMultiplier + slider.partLen*durationMultiplier

	if settings.Objects.Sliders.Snaking.In {
		slider.sliderSnakeTail.AddEvent(slSnInS, slSnInE, 1)
	} else {
		slider.sliderSnakeTail.SetValue(1)
	}

	slider.fade = animation.NewGlider(0)
	slider.fade.AddEvent(slider.StartTime-diff.Preempt, slider.StartTime-(diff.Preempt-difficulty.HitFadeIn), 1)

	slider.bodyFade = animation.NewGlider(0)
	slider.bodyFade.AddEvent(slider.StartTime-diff.Preempt, slider.StartTime-(diff.Preempt-difficulty.HitFadeIn), 1)

	if diff.CheckModActive(difficulty.Hidden) {
		slider.bodyFade.AddEventEase(slider.StartTime-diff.Preempt+difficulty.HitFadeIn, slider.EndTime, 0, easing.OutQuad)
	} else {
		slider.bodyFade.AddEvent(slider.EndTime, slider.EndTime+difficulty.HitFadeOut, 0)
	}

	slider.fade.AddEvent(slider.EndTime, slider.EndTime+difficulty.HitFadeOut, 0)

	slider.startCircle = DummyCircle(slider.StartPosRaw, slider.StartTime)
	slider.startCircle.ComboNumber = slider.ComboNumber
	slider.startCircle.ComboSet = slider.ComboSet
	slider.startCircle.HitObjectID = slider.HitObjectID
	slider.startCircle.StackOffset = slider.StackOffset
	slider.startCircle.StackOffsetHR = slider.StackOffsetHR
	slider.startCircle.StackOffsetEZ = slider.StackOffsetEZ
	slider.startCircle.SetDifficulty(diff)

	slider.edges = append(slider.edges, slider.startCircle)

	sixty := 1000.0 / 60
	frameDelay := math.Max(150/slider.Timings.GetVelocity(slider.TPoint)*sixty, sixty)

	slider.ball = sprite.NewAnimation(skin.GetFrames("sliderb", false), frameDelay, true, 0.0, vector.NewVec2d(0, 0), bmath.Origin.Centre)

	if settings.Objects.Sliders.Snaking.Out {
		slider.ball.SetAlpha(0)
	}

	if len(slider.scorePath) > 0 {
		angle := slider.scorePath[0].Line.GetStartAngle()
		slider.ball.SetVFlip(angle > -math32.Pi/2 && angle < math32.Pi/2)
	}

	followerFrames := skin.GetFrames("sliderfollowcircle", true)

	slider.follower = sprite.NewAnimation(followerFrames, 1000.0/float64(len(followerFrames)), true, 0.0, vector.NewVec2d(0, 0), bmath.Origin.Centre)
	slider.follower.SetAlpha(0.0)

	for i := int64(1); i <= slider.repeat; i++ {
		circleTime := slider.StartTime + math.Floor(slider.partLen*float64(i))

		appearTime := slider.StartTime - math.Floor(slider.diff.Preempt)
		if i > 1 {
			appearTime = circleTime - math.Floor(slider.partLen*2)
		}

		circle := NewSliderEndCircle(vector.NewVec2f(0, 0), appearTime, circleTime, i == 1, i == slider.repeat)
		circle.ComboNumber = slider.ComboNumber
		circle.ComboSet = slider.ComboSet
		circle.HitObjectID = slider.HitObjectID
		circle.StackOffset = slider.StackOffset
		circle.StackOffsetHR = slider.StackOffsetHR
		circle.StackOffsetEZ = slider.StackOffsetEZ
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

	for i, p := range slider.TickPoints {
		a := (p.Time-slider.StartTime)/2 + slider.StartTime - diff.Preempt*2/3

		fs := (p.Time - slider.StartTime) / slider.partLen

		if fs < 1.0 {
			a = math.Max(fs*(slSnInE-slSnInS)+slSnInS, a)
		}

		endTime := math.Min(a+150, p.Time-36)

		p.scale.AddEventS(a, endTime, 0.5, 1.2)
		p.scale.AddEventSEase(endTime, endTime+150, 1.2, 1.0, easing.OutQuad)
		p.fade.AddEventS(a, endTime, 0.0, 1.0)

		if diff.CheckModActive(difficulty.Hidden) {
			p.fade.AddEventS(math.Max(endTime, p.Time-1000), p.Time, 1.0, 0.0)
		} else {
			p.fade.AddEventS(p.Time, p.Time, 1.0, 0.0)
		}

		p.Pos = slider.GetStackedPositionAtMod(p.Time, slider.diff.Mods)

		slider.TickPoints[i] = p
	}

	for i, p := range slider.TickReverse {
		p.Pos = slider.GetStackedPositionAtMod(p.Time, slider.diff.Mods)

		slider.TickReverse[i] = p
	}

	slider.body = sliderrenderer.NewBody(slider.multiCurve, diff.Mods&difficulty.HardRock > 0, float32(slider.diff.CircleRadius))
}

func (slider *Slider) IsRetarded() bool {
	return len(slider.scorePath) == 0 || slider.StartTime == slider.EndTime
}

func (slider *Slider) Update(time float64) bool {
	if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {
		for i := int64(0); i <= slider.repeat; i++ {
			edgeTime := slider.StartTime + math.Floor(float64(i)*slider.partLen)

			if slider.lastTime < edgeTime && time >= edgeTime {
				slider.HitEdge(int(i), time, true)

				if i == 0 {
					slider.InitSlide(slider.StartTime)
				}
			}
		}

		for _, p := range slider.TickPoints {
			if slider.lastTime < p.Time && time >= p.Time {
				slider.PlayTick()
			}
		}
	} else if slider.isSliding {
		for i := int64(1); i < slider.repeat; i++ {
			edgeTime := slider.StartTime + math.Floor(float64(i)*slider.partLen)

			if slider.lastTime < edgeTime && time >= edgeTime {
				slider.HitEdge(int(i), time, true)
			}
		}

		for _, p := range slider.TickPoints {
			if slider.lastTime < p.Time && time >= p.Time {
				slider.PlayTick()
			}
		}
	}

	slider.sliderSnakeHead.Update(time)
	slider.sliderSnakeTail.Update(time)

	if slider.startCircle != nil {
		slider.startCircle.Update(time)
	}

	if slider.ball != nil {
		slider.ball.Update(time)
	}

	if slider.follower != nil {
		slider.follower.Update(time)
	}

	slider.fade.Update(time)
	slider.bodyFade.Update(time)

	headPos := slider.multiCurve.PointAt(float32(slider.sliderSnakeHead.GetValue()))
	tailPos := slider.multiCurve.PointAt(float32(slider.sliderSnakeTail.GetValue()))
	headAngle := slider.multiCurve.GetStartAngleAt(float32(slider.sliderSnakeHead.GetValue())) + math.Pi
	tailAngle := slider.multiCurve.GetEndAngleAt(float32(slider.sliderSnakeTail.GetValue())) + math.Pi

	if slider.diff.Mods&difficulty.HardRock > 0 {
		headAngle = -headAngle
		tailAngle = -tailAngle
	}

	for _, s := range slider.headEndCircles {
		s.ArrowRotation = float64(headAngle)
		s.StartPosRaw = headPos
		s.Update(time)
	}

	for _, s := range slider.tailEndCircles {
		s.ArrowRotation = float64(tailAngle)
		s.StartPosRaw = tailPos
		s.Update(time)
	}

	for _, p := range slider.TickPoints {
		p.fade.Update(time)
		p.scale.Update(time)
	}

	pos := slider.GetStackedPositionAtMod(time, slider.diff.Mods)

	if settings.Objects.Sliders.Snaking.Out && slider.repeat%2 == 1 && time >= math.Floor(slider.EndTime-slider.partLen) {
		snakeTime := slider.EndTime - slider.partLen*(1-slider.sliderSnakeHead.GetValue())
		p2 := slider.GetStackedPositionAtMod(snakeTime, slider.diff.Mods)
		slider.ball.SetPosition(p2.Copy64())
		slider.startCircle.StartPosRaw = slider.GetPositionAt(snakeTime)
	} else {
		slider.ball.SetPosition(pos.Copy64())
	}

	if time-slider.lastTime > 0 && time >= slider.StartTime {
		angle := pos.AngleRV(slider.Pos)

		reversed := int((time-slider.StartTime)/slider.partLen)%2 == 1

		if reversed {
			angle -= math32.Pi
		}

		if slider.ball != nil {
			slider.ball.SetHFlip(skin.GetInfo().SliderBallFlip && reversed)
			slider.ball.SetRotation(float64(angle))
		}
	}

	if slider.lastTime < slider.EndTime && time >= slider.EndTime && slider.isSliding {
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

func (slider *Slider) ArmStart(clicked bool, time float64) {
	slider.startCircle.Arm(clicked, time)

	if settings.Objects.Sliders.Snaking.Out {
		slider.ball.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, slider.StartTime, slider.StartTime, 0, 1))
		slider.ball.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, slider.EndTime, slider.EndTime, 1, 0))
		slider.ball.ResetValuesToTransforms()

		if time < math.Floor(slider.EndTime-slider.partLen) {
			if slider.repeat%2 == 1 {
				slider.sliderSnakeHead.AddEvent(slider.EndTime-slider.partLen, slider.EndTime, 1)
			} else {
				slider.sliderSnakeTail.AddEvent(slider.EndTime-slider.partLen, slider.EndTime, 0)
			}
		} else {
			endTime := slider.EndTime

			for _, p := range slider.ScorePoints {
				if p.Time > time {
					endTime = p.Time
					break
				}
			}

			partStart := slider.EndTime - slider.partLen
			remaining := endTime - time

			first := time - partStart

			dur := math.Min(first/2, remaining*0.66)
			eTime := time + dur

			if slider.repeat%2 == 1 {
				slider.sliderSnakeHead.AddEventEase(time, eTime, (first+dur)/slider.partLen, snakeEase)
				slider.sliderSnakeHead.AddEvent(eTime, slider.EndTime, 1)
			} else {
				slider.sliderSnakeTail.AddEventEase(time, eTime, 1-(first+dur)/slider.partLen, snakeEase)
				slider.sliderSnakeTail.AddEvent(eTime, slider.EndTime, 0)
			}
		}
	}
}

func (slider *Slider) InitSlide(time float64) {
	slider.follower.ClearTransformations()

	startTime := time

	fadeInEnd := math.Min(startTime+180, slider.EndTime)

	slider.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, math.Min(startTime+60, slider.EndTime), 0, 1))
	slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, fadeInEnd, 0.5, 1))

	slider.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, slider.EndTime, slider.EndTime+200, 1, 0))
	slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, slider.EndTime, slider.EndTime+200, 1, 0.8))

	fadeBase := 200.0

	fadeTime := fadeBase
	if len(slider.ScorePoints) >= 2 {
		fadeTime = math.Min(fadeTime, slider.ScorePoints[1].Time-slider.ScorePoints[0].Time)
	}

	endValue := 1.1 - (fadeTime/fadeBase)*0.1

	for i := 0; i < len(slider.ScorePoints)-1; i++ {
		p := slider.ScorePoints[i]
		endTime := p.Time + fadeTime

		if endTime < fadeInEnd {
			continue
		}

		startTime := p.Time
		startValue := 1.1

		if startTime < fadeInEnd {
			startValue = (startValue-endValue)*(endTime-startTime)/fadeTime + endValue
			startTime = fadeInEnd
		}

		slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, startTime, math.Min(slider.EndTime, endTime), startValue, endValue))
	}

	slider.isSliding = true
}

func (slider *Slider) KillSlide(time float64) {
	slider.follower.ClearTransformations()

	nextPoint := slider.EndTime
	for _, p := range slider.ScorePoints {
		if p.Time > time {
			nextPoint = p.Time
			break
		}
	}

	slider.follower.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, nextPoint-100, nextPoint, 1, 0))
	slider.follower.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, nextPoint-100, nextPoint, 1, 2))

	slider.isSliding = false
	slider.StopSlideSamples()
}

func (slider *Slider) PlaySlideSamples() {
	if slider.audioSubmissionDisabled {
		return
	}

	point := slider.Timings.Current

	sampleSet := slider.BasicHitSound.SampleSet
	if sampleSet == 0 {
		sampleSet = point.SampleSet
	}

	audio.PlaySliderLoops(sampleSet, slider.BasicHitSound.AdditionSet, slider.baseSample, point.SampleIndex, point.SampleVolume, slider.HitObjectID, slider.Pos.X64())
}

func (slider *Slider) StopSlideSamples() {
	if slider.audioSubmissionDisabled {
		return
	}

	audio.StopSliderLoops()
}

func (slider *Slider) PlayEdgeSample(index int) {
	if slider.audioSubmissionDisabled {
		return
	}

	slider.playSampleT(slider.sampleSets[index], slider.additionSets[index], slider.samples[index], slider.Timings.GetPoint(slider.StartTime+math.Floor(float64(index)*slider.partLen)+5), slider.GetStackedPositionAt(slider.StartTime+math.Floor(float64(index)*slider.partLen)))
}

func (slider *Slider) HitEdge(index int, time float64, isHit bool) {
	if index == 0 {
		slider.ArmStart(isHit, time)
	} else {
		e := slider.edges[index]
		e.Arm(isHit, time)
	}

	if isHit {
		slider.PlayEdgeSample(index)
	}
}

func (slider *Slider) PlayTick() {
	if slider.audioSubmissionDisabled {
		return
	}

	audio.PlaySliderTick(slider.Timings.Current.SampleSet, slider.Timings.Current.SampleIndex, slider.Timings.Current.SampleVolume, slider.HitObjectID, slider.Pos.X64())
}

func (slider *Slider) playSampleT(sampleSet, additionSet, sample int, point TimingPoint, pos vector.Vector2f) {
	if sampleSet == 0 {
		sampleSet = slider.BasicHitSound.SampleSet
		if sampleSet == 0 {
			sampleSet = point.SampleSet
		}
	}

	if additionSet == 0 {
		additionSet = slider.BasicHitSound.AdditionSet
	}

	audio.PlaySample(sampleSet, additionSet, sample, point.SampleIndex, point.SampleVolume, slider.HitObjectID, pos.X64())
}

func (slider *Slider) GetPosition() vector.Vector2f {
	return slider.Pos
}

func (slider *Slider) DrawBodyBase(_ float64, projection mgl32.Mat4) {
	slider.body.DrawBase(slider.sliderSnakeHead.GetValue(), slider.sliderSnakeTail.GetValue(), projection)
}

func (slider *Slider) DrawBody(_ float64, bodyColor, innerBorder, outerBorder color2.Color, projection mgl32.Mat4, scale float32) {
	colorAlpha := slider.bodyFade.GetValue() * float64(bodyColor.A)

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
			baseTrack = skin.GetInfo().ComboColors[int(slider.ComboSet)%len(skin.GetInfo().ComboColors)]
		}

		bodyOuter = baseTrack.Shade2(-0.1)
		bodyInner = baseTrack.Shade2(0.5)
	} else {
		if settings.Objects.Colors.UseComboColors {
			cHSV := settings.Objects.Colors.ComboColors[int(slider.ComboSet)%len(settings.Objects.Colors.ComboColors)]
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

	stackOffset := slider.StackOffset
	if slider.diff.Mods&difficulty.HardRock > 0 {
		stackOffset = slider.StackOffsetHR
	} else if slider.diff.Mods&difficulty.Easy > 0 {
		stackOffset = slider.StackOffsetEZ
	}

	slider.body.DrawNormal(projection, stackOffset, scale, bodyInner, bodyOuter, borderInner, borderOuter)
}

func (slider *Slider) Draw(time float64, color color2.Color, batch *batch.QuadBatch) bool {
	if len(slider.scorePath) == 0 {
		return true
	}

	alpha := slider.fade.GetValue() * float64(color.A)

	if settings.DIVIDES >= settings.Objects.Colors.MandalaTexturesTrigger {
		alpha *= settings.Objects.Colors.MandalaTexturesAlpha
	}

	batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)

	if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger {
		if time < slider.EndTime {
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

	if time >= slider.StartTime && time <= slider.EndTime {
		slider.drawBall(time, batch, color, alpha, settings.Objects.Sliders.ForceSliderBallTexture || settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger)
	}

	if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.Sliders.DrawSliderFollowCircle && slider.follower != nil {
		batch.SetTranslation(slider.Pos.Copy64())
		batch.SetColor(1, 1, 1, alpha)
		slider.follower.Draw(time, batch)
	}

	batch.SetSubScale(1, 1)
	batch.SetTranslation(vector.NewVec2d(0, 0))

	if time >= slider.EndTime && slider.fade.GetValue() <= 0.001 {
		if slider.body != nil {
			//HACKHACKHACK: for some reason disposing FBOs with VSync causes 30ms frame skips...
			if !settings.Graphics.VSync {
				slider.body.Dispose()
			}
		}

		return true
	}

	return false
}

func (slider *Slider) drawBall(time float64, batch *batch.QuadBatch, color color2.Color, alpha float64, useBallTexture bool) {
	batch.SetColor(1, 1, 1, alpha)
	batch.SetTranslation(slider.ball.GetPosition())

	isB := skin.GetSource("sliderb") != skin.SKIN && useBallTexture

	if isB && skin.GetTexture("sliderb-nd") != nil {
		batch.SetColor(0.1, 0.1, 0.1, alpha)
		batch.DrawTexture(*skin.GetTexture("sliderb-nd"))
	}

	if settings.Skin.UseColorsFromSkin {
		color := color2.NewL(1)

		if skin.GetInfo().SliderBallTint {
			color = skin.GetInfo().ComboColors[int(slider.ComboSet)%len(skin.GetInfo().ComboColors)]
		} else if skin.GetInfo().SliderBall != nil {
			color = *skin.GetInfo().SliderBall
		}

		batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)
	} else if settings.Objects.Colors.Sliders.SliderBallTint {
		if settings.Objects.Colors.UseComboColors {
			cHSV := settings.Objects.Colors.ComboColors[int(slider.ComboSet)%len(settings.Objects.Colors.ComboColors)]
			r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
			batch.SetColor(float64(r), float64(g), float64(b), alpha)
		} else {
			batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)
		}
	} else {
		batch.SetColor(1, 1, 1, alpha)
	}

	//cHSV := settings.Objects.Colors.ComboColors[int(slider.ComboSet)%len(settings.Objects.Colors.ComboColors)]
	//r, g, b := color2.HSVToRGB(float32(cHSV.Hue), float32(cHSV.Saturation), float32(cHSV.Value))
	//batch.SetColor(float64(r), float64(g), float64(b), alpha)

	if useBallTexture {
		batch.SetTranslation(vector.NewVec2d(0, 0))
		slider.ball.Draw(time, batch)
		batch.SetTranslation(slider.ball.GetPosition())
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

func (slider *Slider) DrawApproach(time float64, color color2.Color, batch *batch.QuadBatch) {
	if len(slider.scorePath) == 0 {
		return
	}

	slider.startCircle.DrawApproach(time, color, batch)
}

func (slider *Slider) GetType() Type {
	return SLIDER
}
