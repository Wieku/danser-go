package objects

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
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
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"sort"
	"strconv"
	"strings"
)

const (
	maxPathLength = 100_000_000 // Sanity limits, XNOR reaches 10M pixel length so 100M should be enough
	maxRepeats    = 10_000      // Same limit as osu!
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
	LastPoint bool
	EdgeIndex int
}

var easeBezier = curves.NewMultiCurve([]curves.CurveDef{{CurveType: curves.CBezier, Points: []vector.Vector2f{{X: 0, Y: 0}, {X: 0.1, Y: 1}, {X: 0.5, Y: 0.5}, {X: 1, Y: 1}}}})

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
	RepeatCount int

	sampleSets   []int
	additionSets []int
	samples      []int
	baseSample   int

	Pos         vector.Vector2f
	TickPoints  []TickPoint
	TickReverse []TickPoint
	ScorePoints []TickPoint

	startCircle *Circle

	sliderSnakeTail *animation.Glider
	sliderSnakeHead *animation.Glider
	fade            *animation.Glider
	bodyFade        *animation.Glider

	diff     *difficulty.Difficulty
	body     *sliderrenderer.Body
	lastTime float64

	ball     *sprite.Animation
	follower *sprite.Animation

	edges          []*Circle
	endCircles     []*Circle
	headEndCircles []*Circle
	tailEndCircles []*Circle

	isSliding bool

	EndTimeLazer     float64
	ScorePointsLazer []TickPoint
	spanDuration     float64

	updatedAtLeastOnce bool

	lastScorePoint int
}

func NewSlider(data []string) *Slider {
	slider := &Slider{
		HitObject: commonParse(data, 10),
	}

	slider.PositionDelegate = slider.PositionAt

	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.RepeatCount, _ = strconv.Atoi(data[6])

	if slider.pixelLength*float64(slider.RepeatCount) > maxPathLength*10 {
		return nil
	}

	slider.pixelLength = min(slider.pixelLength, maxPathLength)
	slider.RepeatCount = min(slider.RepeatCount, maxRepeats) // The same limit as in Lazer

	slider.multiCurve = slider.parseCurve(data[5])
	if slider.multiCurve == nil {
		return nil
	}

	if slider.pixelLength == 0 {
		slider.pixelLength = float64(slider.multiCurve.GetLength())
	}

	slider.EndTime = slider.StartTime
	slider.EndPosRaw = slider.multiCurve.PointAt(1.0)
	slider.Pos = slider.StartPosRaw

	slider.samples = make([]int, slider.RepeatCount+1)
	slider.sampleSets = make([]int, slider.RepeatCount+1)
	slider.additionSets = make([]int, slider.RepeatCount+1)

	f, _ := strconv.ParseInt(data[4], 10, 64)
	slider.baseSample = int(f)

	for i := range slider.samples {
		slider.samples[i] = slider.baseSample
		slider.sampleSets[i] = slider.BasicHitSound.SampleSet
		slider.additionSets[i] = slider.BasicHitSound.AdditionSet
	}

	if len(data) > 8 {
		subData := strings.Split(data[8], "|")

		n := min(len(subData), len(slider.samples))

		for i := 0; i < n; i++ {
			sample, _ := strconv.Atoi(subData[i])
			slider.samples[i] = sample
		}
	}

	if len(data) > 9 {
		subData := strings.Split(data[9], "|")

		n := min(len(subData), len(slider.sampleSets))

		for i := 0; i < n; i++ {
			extras := strings.Split(subData[i], ":")

			sampleSet, _ := strconv.Atoi(extras[0])
			additionSet, _ := strconv.Atoi(extras[1])

			slider.sampleSets[i] = sampleSet
			slider.additionSets[i] = additionSet
		}
	}

	slider.fade = animation.NewGlider(1)
	slider.bodyFade = animation.NewGlider(1)
	slider.sliderSnakeTail = animation.NewGlider(1)
	slider.sliderSnakeHead = animation.NewGlider(0)

	return slider
}

func (slider *Slider) parseCurve(curveData string) *curves.MultiCurve {
	list := strings.Split(curveData, "|")

	var defs []curves.CurveDef

	cDef := curves.CurveDef{
		CurveType: curves.CType(-1),
		Points:    []vector.Vector2f{slider.StartPosRaw},
	}

	nextType := curves.CType(-1)

	for i, j := 0, 0; i < len(list); i++ {
		split := strings.Split(list[i], ":")

		if len(split) == 1 {
			if tType := tryGetType(split[0]); tType > -1 {
				if cDef.CurveType == -1 {
					cDef.CurveType = tType
				} else {
					nextType = tType
				}
			}
		} else {
			x, _ := strconv.ParseFloat(split[0], 32)
			y, _ := strconv.ParseFloat(split[1], 32)

			vec := vector.NewVec2f(float32(x), float32(y))

			if j > 0 || vec != slider.StartPosRaw { // skip the first point if it's the same as start position.
				cDef.Points = append(cDef.Points, vec)
			}

			j++

			if nextType > -1 {
				defs = append(defs, cDef)

				cDef = curves.CurveDef{
					CurveType: nextType,
					Points:    []vector.Vector2f{vec},
				}

				nextType = -1
			}
		}
	}

	if len(cDef.Points) > 1 || len(defs) == 0 { // Lazer's multi-type slider has 1 point line
		if cDef.CurveType == -1 { // osu! uses catmull if there's no curve type
			cDef.CurveType = curves.CCatmull
		}

		defs = append(defs, cDef)
	}

	// validation
	for _, def := range defs {
		if def.CurveType == curves.CBezier {
			var controlDistance float32

			for i := 1; i < len(def.Points); i++ {
				controlDistance += def.Points[i].Dst(def.Points[i-1])
			}

			if controlDistance >= 2*maxPathLength { // Skip sliders which are too computationally expensive
				return nil
			}
		}
	}

	return curves.NewMultiCurveT(defs, slider.pixelLength)
}

func tryGetType(str string) curves.CType {
	switch str {
	case "P":
		return curves.CCirArc
	case "L":
		return curves.CLine
	case "B":
		return curves.CBezier
	case "C":
		return curves.CCatmull
	default: // It's a point
		return -1
	}
}

func (slider *Slider) GetLength() float32 {
	return slider.multiCurve.GetLength()
}

func (slider *Slider) GetStartAngleMod(diff *difficulty.Difficulty) float32 {
	return slider.GetStackedStartPositionMod(diff).AngleRV(slider.GetStackedPositionAtMod(slider.StartTime+min(10, slider.partLen), diff)) //temporary solution
}

func (slider *Slider) GetEndAngleMod(diff *difficulty.Difficulty) float32 {
	return slider.GetStackedEndPositionMod(diff).AngleRV(slider.GetStackedPositionAtMod(slider.EndTime-min(10, slider.partLen), diff)) //temporary solution
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

	pLine := slider.scorePath[mutils.Clamp(index, 0, len(slider.scorePath)-1)]

	clamped := mutils.Clamp(time, float64(pLine.Time1), float64(pLine.Time2))

	var pos vector.Vector2f
	if pLine.Time2 == pLine.Time1 {
		pos = pLine.Line.Point2
	} else {
		pos = pLine.Line.PointAt(float32(clamped-float64(pLine.Time1)) / float32(pLine.Time2-pLine.Time1))
	}

	return pos
}

func (slider *Slider) PositionAtLazer(time float64) vector.Vector2f {
	if slider.IsRetarded() {
		return slider.StartPosRaw
	}

	t1 := mutils.Clamp(time, slider.StartTime, slider.EndTimeLazer)

	progress := (t1 - slider.StartTime) / slider.spanDuration

	progress = math.Mod(progress, 2)
	if progress >= 1 {
		progress = 2 - progress
	}

	return slider.multiCurve.PointAt(float32(progress))
}

func (slider *Slider) GetStackedPositionAtModLazer(time float64, diff *difficulty.Difficulty) vector.Vector2f {
	return ModifyPosition(slider.HitObject, slider.PositionAtLazer(time), diff)
}

func (slider *Slider) GetAsDummyCircles() []IHitObject {
	circles := []IHitObject{slider.createDummyCircle(slider.GetStartTime(), true, false)}

	if slider.IsRetarded() {
		return circles
	}

	for i, p := range slider.ScorePoints {
		time := p.Time
		if i == len(slider.ScorePoints)-1 && settings.KNOCKOUT && !slider.diff.CheckModActive(difficulty.Lazer) { // Lazer ends work differently so skip -36ms
			time = math.Floor(max(slider.StartTime+(slider.EndTime-slider.StartTime)/2, slider.EndTime-36))
		}

		circles = append(circles, slider.createDummyCircle(time, false, i == len(slider.ScorePoints)-1))
	}

	return circles
}

func (slider *Slider) createDummyCircle(time float64, inheritStart, inheritEnd bool) *Circle {
	circle := DummyCircleInherit(slider.GetPositionAt(time), time, true, inheritStart, inheritEnd)
	circle.StackLeniency = slider.StackLeniency
	circle.StackIndexMap = slider.StackIndexMap
	circle.ComboSet = slider.ComboSet

	return circle
}

func (slider *Slider) SetTiming(timings *Timings, beatmapVersion int, diffCalcOnly bool) {
	slider.Timings = timings
	slider.TPoint = timings.GetPointAt(slider.StartTime)

	nanTimingPoint := math.IsNaN(slider.TPoint.beatLength)

	lines := slider.multiCurve.GetLines()

	startTime := slider.StartTime

	velocity := slider.Timings.GetVelocity(slider.TPoint)

	cLength := float64(slider.multiCurve.GetLength())

	slider.spanDuration = cLength * 1000 / velocity

	slider.EndTimeLazer = slider.StartTime + cLength*1000*float64(slider.RepeatCount)/velocity

	minDistanceFromEnd := velocity * 0.01

	tickDistance := slider.Timings.GetTickDistance(slider.TPoint)
	if beatmapVersion < 8 {
		tickDistance = slider.Timings.GetScoringDistance()
	}

	if slider.multiCurve.GetLength() > 0 && tickDistance > slider.pixelLength {
		tickDistance = slider.pixelLength
	}

	// Sanity limit to 32768 ticks per repeat
	if cLength/tickDistance > 32768 {
		tickDistance = cLength / 32768
	}

	// Lazer like score point calculations. Clean AF, but not unreliable enough for stable's replay processing. Would need more testing.
	for span := 0; span < int(slider.RepeatCount); span++ {
		spanStartTime := slider.StartTime + float64(span)*slider.spanDuration
		reversed := span%2 == 1

		// Skip ticks if timingPoint has NaN beatLength
		for d := tickDistance; d <= cLength && !nanTimingPoint; d += tickDistance {
			if d >= cLength-minDistanceFromEnd {
				break
			}

			// Always generate ticks from the start of the path rather than the span to ensure that ticks in repeat spans are positioned identically to those in non-repeat spans
			timeProgress := d / cLength
			if reversed {
				timeProgress = 1 - timeProgress
			}

			slider.ScorePointsLazer = append(slider.ScorePointsLazer, TickPoint{
				Time: spanStartTime + timeProgress*slider.spanDuration,
			})
		}

		slider.ScorePointsLazer = append(slider.ScorePointsLazer, TickPoint{
			Time:      spanStartTime + slider.spanDuration,
			IsReverse: span < int(slider.RepeatCount)-1,
			LastPoint: span == int(slider.RepeatCount)-1,
		})
	}

	sort.Slice(slider.ScorePointsLazer, func(i, j int) bool {
		return slider.ScorePointsLazer[i].Time < slider.ScorePointsLazer[j].Time
	})

	if diffCalcOnly { // We're not interested in stable-like path in difficulty calculator mode
		return
	}

	scoringLengthTotal := 0.0
	scoringDistance := 0.0

	// Stable-like score point processing, ugly AF.
	for i := 0; i < slider.RepeatCount; i++ {
		distanceToEnd := float64(slider.multiCurve.GetLength())
		skipTick := nanTimingPoint // NaN SV acts like 1.0x SV, but doesn't spawn slider ticks

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

				point := TickPoint{scoreTime, slider.GetPositionAt(scoreTime), animation.NewGlider(0.0), animation.NewGlider(0.0), false, false, -1}
				slider.TickPoints = append(slider.TickPoints, point)
				slider.ScorePoints = append(slider.ScorePoints, point)
			}
		}

		scoringLengthTotal += scoringDistance

		scoreTime := slider.StartTime + math.Floor((float64(float32(scoringLengthTotal))/velocity)*1000)

		// Ensure last tick is not later than end time. Ruleset calculates the last tick regardless of this value
		if i == slider.RepeatCount-1 {
			scoreTime = slider.EndTime
		}

		point := TickPoint{scoreTime, slider.GetPositionAt(scoreTime), nil, nil, true, (i + 1) == slider.RepeatCount, i + 1}

		slider.TickReverse = append(slider.TickReverse, point)
		slider.ScorePoints = append(slider.ScorePoints, point)

		if skipTick {
			scoringDistance = 0
		} else {
			scoringLengthTotal -= tickDistance - scoringDistance
			scoringDistance = tickDistance - scoringDistance
		}
	}

	slider.partLen = (slider.EndTime - slider.StartTime) / float64(slider.RepeatCount)

	slider.EndPosRaw = slider.GetPositionAt(slider.EndTime)

	//if len(slider.scorePath) == 0 || slider.StartTime == slider.EndTime {
	//	log.Println("Warning: slider", slider.HitObjectID, "at ", slider.StartTime, "is broken.")
	//}

	slider.calculateFollowPoints()
}

func (slider *Slider) calculateFollowPoints() {
	sort.Slice(slider.TickPoints, func(i, j int) bool { return slider.TickPoints[i].Time < slider.TickPoints[j].Time })
	sort.Slice(slider.ScorePoints, func(i, j int) bool { return slider.ScorePoints[i].Time < slider.ScorePoints[j].Time })
}

func copySliderHOData(target, base *HitObject) {
	target.ComboNumber = base.ComboNumber
	target.ComboSet = base.ComboSet
	target.ComboSetHax = base.ComboSetHax
	target.HitObjectID = base.HitObjectID
	target.StackLeniency = base.StackLeniency
	target.StackIndexMap = base.StackIndexMap
}

func (slider *Slider) SetDifficulty(diff *difficulty.Difficulty) {
	slider.diff = diff
	slider.sliderSnakeTail = animation.NewGlider(0)
	slider.sliderSnakeHead = animation.NewGlider(0)

	slider.fade = animation.NewGlider(0)
	slider.fade.AddEvent(slider.StartTime-diff.Preempt, slider.StartTime-(diff.Preempt-diff.TimeFadeIn), 1)

	slider.bodyFade = animation.NewGlider(0)
	slider.bodyFade.AddEvent(slider.StartTime-diff.Preempt, slider.StartTime-(diff.Preempt-diff.TimeFadeIn), 1)

	if diff.CheckModActive(difficulty.Hidden) {
		slider.bodyFade.AddEventEase(slider.StartTime-diff.Preempt+diff.TimeFadeIn, slider.EndTime, 0, easing.OutQuad)
	}

	slider.fade.AddEvent(slider.EndTime, slider.EndTime+difficulty.HitFadeOut, 0)

	slider.startCircle = DummyCircle(slider.StartPosRaw, slider.StartTime)
	copySliderHOData(slider.startCircle.HitObject, slider.HitObject)
	slider.startCircle.SetDifficulty(diff)

	slider.edges = append(slider.edges, slider.startCircle)

	sixty := 1000.0 / 60
	frameDelay := max(150/slider.Timings.GetVelocity(slider.TPoint)*sixty, sixty)

	slider.ball = sprite.NewAnimation(skin.GetFrames("sliderb", false), frameDelay, true, 0.0, vector.NewVec2d(0, 0), vector.Centre)

	if len(slider.scorePath) > 0 {
		angle := slider.scorePath[0].Line.GetStartAngle()
		slider.ball.SetVFlip(angle > -math32.Pi/2 && angle < math32.Pi/2)
	}

	followerFrames := skin.GetFrames("sliderfollowcircle", true)

	slider.follower = sprite.NewAnimation(followerFrames, 1000.0/float64(len(followerFrames)), true, 0.0, vector.NewVec2d(0, 0), vector.Centre)
	slider.follower.SetAlpha(0.0)

	for i := 1; i <= slider.RepeatCount; i++ {
		circleTime := slider.StartTime + math.Floor(slider.partLen*float64(i))

		appearTime := slider.StartTime - math.Floor(slider.diff.Preempt)
		bounceStartTime := slider.StartTime - min(math.Floor(slider.diff.Preempt), 15000)

		if i > 1 {
			appearTime = circleTime - math.Floor(slider.partLen*2)
			bounceStartTime = appearTime
		}

		circle := NewSliderEndCircle(vector.NewVec2f(0, 0), appearTime, bounceStartTime, circleTime, i == 1, i == slider.RepeatCount)
		copySliderHOData(circle.HitObject, slider.HitObject)
		circle.SetTiming(slider.Timings, 14, false)
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
		p.Pos = slider.GetStackedPositionAtMod(p.Time, slider.diff)

		slider.TickPoints[i] = p
	}

	for i, p := range slider.TickReverse {
		p.Pos = slider.GetStackedPositionAtMod(p.Time, slider.diff)

		slider.TickReverse[i] = p
	}

	slider.body = sliderrenderer.NewBody(slider.multiCurve, diff.Mods&difficulty.HardRock > 0, float32(slider.diff.CircleRadius))
}

func (slider *Slider) IsRetarded() bool {
	return len(slider.scorePath) == 0 || slider.StartTime == slider.EndTime
}

func (slider *Slider) Update(time float64) bool {
	if !slider.updatedAtLeastOnce {
		slider.initSnake()

		slider.updatedAtLeastOnce = true
	}

	if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {
		if slider.lastTime < slider.StartTime && time >= slider.StartTime {
			slider.HitEdge(0, time, true)
			slider.InitSlide(slider.StartTime)
		}

		if slider.lastTime < slider.EndTime && time >= slider.EndTime {
			slider.HitEdge(slider.RepeatCount, time, true)
		}
	}

	if slider.isSliding {
		for i := slider.lastScorePoint; i < len(slider.ScorePoints)-1; i++ {
			p := slider.ScorePoints[i]

			if time < p.Time {
				break
			} else if slider.lastTime < p.Time {
				if p.IsReverse {
					slider.HitEdge(p.EdgeIndex, time, true)
				} else {
					slider.PlayTick()
				}
			}

			slider.lastScorePoint = i + 1
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

	pos := slider.GetStackedPositionAtMod(time, slider.diff)

	if settings.Objects.Sliders.Snaking.Out && slider.RepeatCount%2 == 1 && time >= math.Floor(slider.EndTime-slider.partLen) {
		snakeTime := slider.EndTime - slider.partLen*(1-slider.sliderSnakeHead.GetValue())
		p2 := slider.GetStackedPositionAtMod(snakeTime, slider.diff)
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

	if slider.isSliding && time >= slider.StartTime && time <= slider.EndTime {
		slider.PlaySlideSamples()
	}

	if slider.lastTime <= slider.EndTime && time > slider.EndTime && slider.isSliding {
		slider.StopSlideSamples()
		slider.isSliding = false
	}

	slider.Pos = pos

	slider.lastTime = time

	return true
}

func (slider *Slider) ArmStart(clicked bool, time float64) {
	slider.startCircle.Arm(clicked, time)

	slider.ball.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, slider.StartTime, slider.StartTime, 1, 1))

	if settings.Objects.Sliders.Snaking.Out {
		if time < math.Floor(slider.EndTime-slider.partLen) {
			if slider.RepeatCount%2 == 1 {
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

			dur := min(first/2, remaining*0.66)
			eTime := time + dur

			if slider.RepeatCount%2 == 1 {
				slider.sliderSnakeHead.AddEventEase(time, eTime, (first+dur)/slider.partLen, snakeEase)
				slider.sliderSnakeHead.AddEvent(eTime, slider.EndTime, 1)
			} else {
				slider.sliderSnakeTail.AddEventEase(time, eTime, 1-(first+dur)/slider.partLen, snakeEase)
				slider.sliderSnakeTail.AddEvent(eTime, slider.EndTime, 0)
			}
		}
	}

	if !slider.diff.CheckModActive(difficulty.Hidden) {
		if settings.Objects.Sliders.Snaking.Out && settings.Objects.Sliders.Snaking.OutFadeInstant {
			slider.bodyFade.AddEvent(slider.EndTime, slider.EndTime, 0)
		} else {
			slider.bodyFade.AddEvent(slider.EndTime, slider.EndTime+difficulty.HitFadeOut, 0)
		}
	}
}

func (slider *Slider) initSnake() {
	slSnInS := slider.StartTime - slider.diff.Preempt
	slSnInE := slider.StartTime - slider.diff.Preempt*2/3

	if settings.Objects.Sliders.Snaking.Out {
		slider.ball.SetAlpha(0)
	}

	if settings.Objects.Sliders.Snaking.In {
		fadeMultiplier := 1.0 - mutils.Clamp(settings.Objects.Sliders.Snaking.FadeMultiplier, 0.0, 1.0)
		durationMultiplier := mutils.Clamp(settings.Objects.Sliders.Snaking.DurationMultiplier, 0.0, 1.0)

		slSnInE = slider.StartTime - slider.diff.Preempt*2/3*fadeMultiplier + slider.partLen*durationMultiplier

		slider.sliderSnakeTail.AddEvent(slSnInS, slSnInE, 1)
	} else {
		slider.sliderSnakeTail.SetValue(1)
	}

	for i, p := range slider.TickPoints {
		var startTime, endTime float64

		repeatProgress := (p.Time - slider.StartTime) / slider.partLen

		if repeatProgress < 1.0 {
			normalStart := (p.Time-slider.StartTime)/2 + slider.StartTime - slider.diff.Preempt*2/3

			startTime = max(repeatProgress*(slSnInE-slSnInS)+slSnInS, normalStart)

			endTime = min(startTime+150, p.Time-36)
		} else {
			rStart := slider.StartTime + slider.partLen*math.Floor(repeatProgress)

			endTime = rStart + (p.Time-rStart)/2
			startTime = endTime - 200
		}

		p.scale.AddEventS(startTime, endTime, 0.5, 1.2)
		p.scale.AddEventSEase(endTime, endTime+150, 1.2, 1.0, easing.OutQuad)
		p.fade.AddEventS(startTime, endTime, 0.0, 1.0)

		if slider.diff.CheckModActive(difficulty.Hidden) {
			p.fade.AddEventS(max(endTime, p.Time-1000), p.Time, 1.0, 0.0)
		} else {
			p.fade.AddEventS(p.Time, p.Time, 1.0, 0.0)
		}

		p.Pos = slider.GetStackedPositionAtMod(p.Time, slider.diff)

		slider.TickPoints[i] = p
	}
}

func (slider *Slider) InitSlide(time float64) {
	if time > slider.EndTime {
		return
	}

	slider.follower.ClearTransformations()

	startTime := time

	fadeInEnd := min(startTime+180, slider.EndTime)

	slider.follower.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, startTime, min(startTime+60, slider.EndTime), 0, 1))
	slider.follower.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.OutQuad, startTime, fadeInEnd, 0.5, 1))

	slider.follower.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.InQuad, slider.EndTime, slider.EndTime+200, 1, 0))
	slider.follower.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.OutQuad, slider.EndTime, slider.EndTime+200, 1, 0.8))

	fadeBase := 200.0

	fadeTime := fadeBase
	if len(slider.ScorePoints) >= 2 {
		fadeTime = min(fadeTime, slider.ScorePoints[1].Time-slider.ScorePoints[0].Time)
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

		slider.follower.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, startTime, min(slider.EndTime, endTime), startValue, endValue))
	}

	slider.follower.SortTransformations()

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

	sampleSet := slider.sampleSets[index]
	if sampleSet == 0 && index == 0 {
		sampleSet = slider.BasicHitSound.SampleSet
	}

	slider.playSampleT(sampleSet, slider.additionSets[index], slider.samples[index], slider.Timings.GetPointAt(slider.StartTime+math.Floor(float64(index)*slider.partLen)+5), slider.GetStackedPositionAtMod(slider.StartTime+math.Floor(float64(index)*slider.partLen), slider.diff))
}

func (slider *Slider) HitEdge(index int, time float64, isHit bool) {
	if index == 0 {
		slider.ArmStart(isHit, time)
	} else {
		e := slider.edges[index]
		e.Arm(isHit, time)
	}

	if isHit && (index == 0 || index == slider.RepeatCount || !slider.IsRetarded()) {
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
		sampleSet = point.SampleSet
	}

	if additionSet == 0 {
		additionSet = sampleSet
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

	bodyOpacityInner := mutils.Clamp(float32(settings.Objects.Colors.Sliders.Body.InnerAlpha), 0.0, 1.0)
	bodyOpacityOuter := mutils.Clamp(float32(settings.Objects.Colors.Sliders.Body.OuterAlpha), 0.0, 1.0)

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
			baseTrack = skin.GetColor(int(slider.ComboSet), int(slider.ComboSetHax), baseTrack)
		}

		bodyOuter = baseTrack.Shade2(-0.1)
		bodyInner = baseTrack.Shade2(0.5)
	} else {
		if settings.Objects.Colors.Sliders.Border.UseHitCircleColor {
			borderInner = skin.GetColor(int(slider.ComboSet), int(slider.ComboSetHax), borderInner)
			borderOuter = skin.GetColor(int(slider.ComboSet), int(slider.ComboSetHax), borderOuter)
		}

		if settings.Objects.Colors.Sliders.Body.UseHitCircleColor {
			bodyColor = skin.GetColor(int(slider.ComboSet), int(slider.ComboSetHax), bodyColor)
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

	stackIndex := slider.GetStackIndexMod(slider.diff)
	stackOffset := -float32(stackIndex) * float32(slider.diff.CircleRadius) / 10

	slider.body.DrawNormal(projection, vector.NewVec2f(stackOffset, stackOffset), scale, bodyInner, bodyOuter, borderInner, borderOuter)
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

	return time >= slider.EndTime && slider.fade.GetValue() <= 0.001
}

func (slider *Slider) Finalize() {
	slider.body.Dispose()
}

func (slider *Slider) drawBall(time float64, batch *batch.QuadBatch, color color2.Color, alpha float64, useBallTexture bool) {
	batch.SetTranslation(slider.ball.GetPosition())

	source := skin.GetSourceFromTexture(slider.ball.Texture)

	if useBallTexture && skin.GetTextureSource("sliderb-nd", source) != nil {
		batch.SetColor(0.1, 0.1, 0.1, alpha*slider.ball.GetAlpha())
		batch.DrawTexture(*skin.GetTexture("sliderb-nd"))
	}

	if settings.Skin.UseColorsFromSkin {
		color := color2.NewL(1)

		if skin.GetInfo().SliderBallTint {
			color = skin.GetColor(int(slider.ComboSet), int(slider.ComboSetHax), color)
		} else if skin.GetInfo().SliderBall != nil {
			color = *skin.GetInfo().SliderBall
		}

		batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)
	} else if settings.Objects.Colors.Sliders.SliderBallTint {
		color = skin.GetColor(int(slider.ComboSet), int(slider.ComboSetHax), color)
		batch.SetColor(float64(color.R), float64(color.G), float64(color.B), alpha)
	} else {
		batch.SetColor(1, 1, 1, alpha)
	}

	if useBallTexture {
		batch.SetTranslation(vector.NewVec2d(0, 0))
		slider.ball.Draw(time, batch)
		batch.SetTranslation(slider.ball.GetPosition())
	} else {
		batch.DrawTexture(*skin.GetTexture("hitcircle-full"))
	}

	if useBallTexture && skin.GetTextureSource("sliderb-spec", source) != nil {
		batch.SetColor(1, 1, 1, alpha*slider.ball.GetAlpha())
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
