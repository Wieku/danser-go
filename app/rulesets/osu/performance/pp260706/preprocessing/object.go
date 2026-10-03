package preprocessing

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

const (
	NormalizedRadius   = 50.0
	NormalizedDiameter = NormalizedRadius * 2
	MinDeltaTime       = 25
)

type DifficultyObject struct {
	OriginalStartTime float64

	// That's stupid but oh well
	listOfDiffs *[]*DifficultyObject
	Index       int

	Diff *difficulty.Difficulty

	BaseObject objects.IHitObject

	IsSlider  bool
	IsSpinner bool

	LastObject objects.IHitObject

	lastLastObject objects.IHitObject

	DeltaTime float64

	StartTime float64

	EndTime float64

	Preempt float64

	JumpDistance float64

	LazyJumpDistance float64

	MinimumJumpDistance float64

	TravelDistance float64

	LazyTravelDistance float64

	Angle float64

	NormalisedVectorAngle float64

	MinimumJumpTime float64

	LazyTravelTime float64

	TravelTime float64

	AdjustedDeltaTime float64

	GreatWindow float64

	SmallCircleBonus float64

	LastObjectEndDeltaTime float64
}

func NewDifficultyObject(hitObject, lastLastObject, lastObject objects.IHitObject, d *difficulty.Difficulty, listOfDiffs *[]*DifficultyObject, index int) *DifficultyObject {
	endTime := hitObject.GetEndTime()
	if s, ok := hitObject.(*LazySlider); ok {
		endTime = s.EndTimeLazer
	}

	obj := &DifficultyObject{
		OriginalStartTime:     hitObject.GetStartTime(),
		listOfDiffs:           listOfDiffs,
		Index:                 index,
		Diff:                  d,
		BaseObject:            hitObject,
		LastObject:            lastObject,
		lastLastObject:        lastLastObject,
		DeltaTime:             (hitObject.GetStartTime() - lastObject.GetStartTime()) / d.Speed,
		StartTime:             hitObject.GetStartTime() / d.Speed,
		EndTime:               endTime / d.Speed,
		Preempt:               d.Preempt / d.Speed,
		Angle:                 math.NaN(),
		NormalisedVectorAngle: math.NaN(),
		GreatWindow:           2 * d.Hit300U / d.Speed,
		SmallCircleBonus:      max(1.0, 1.0+(30-d.CircleRadiusL)/70),
	}

	if _, ok := hitObject.(*objects.Spinner); ok {
		obj.IsSpinner = true
	}

	if _, ok := hitObject.(*LazySlider); ok {
		obj.IsSlider = true
	}

	obj.AdjustedDeltaTime = max(obj.DeltaTime, MinDeltaTime)
	obj.LastObjectEndDeltaTime = obj.AdjustedDeltaTime

	if last := obj.Previous(0); last != nil {
		obj.LastObjectEndDeltaTime = max(obj.StartTime-last.EndTime, MinDeltaTime)
	}

	obj.setDistances()

	return obj
}

func (o *DifficultyObject) CalculateDoubleTapFeasibility(nextObj *DifficultyObject) float64 {
	if nextObj == nil {
		return 0
	}

	currDeltaTime := max(1, o.DeltaTime)
	nextDeltaTime := max(1, nextObj.DeltaTime)

	deltaDifference := math.Abs(nextDeltaTime - currDeltaTime)

	speedRatio := currDeltaTime / max(currDeltaTime, deltaDifference)
	windowRatio := putils.Powi(min(1, currDeltaTime/o.GreatWindow), 5)

	// Can't doubletap if circles don't intersect
	distanceFactor := putils.Powi(putils.ReverseLerp(o.LazyJumpDistance, NormalizedDiameter, NormalizedRadius), 2)

	return 1 - math.Pow(speedRatio, distanceFactor*(1-windowRatio))
}

func (o *DifficultyObject) OpacityAt(time float64, hidden bool) float64 {
	if time > o.BaseObject.GetStartTime() {
		// Consider a hitobject as being invisible when its start time is passed.
		// In reality the hitobject will be visible beyond its start time up until its hittable window has passed,
		// but this is an approximation and such a case is unlikely to be hit where this function is used.
		return 0
	}

	fadeInStartTime := o.BaseObject.GetStartTime() - o.Diff.Preempt
	fadeInDuration := o.Diff.TimeFadeIn

	if hidden {
		fadeOutStartTime := o.BaseObject.GetStartTime() - o.Diff.Preempt + o.Diff.TimeFadeIn
		fadeOutDuration := o.Diff.Preempt * 0.3

		return min(
			mutils.Clamp((time-fadeInStartTime)/fadeInDuration, 0.0, 1.0),
			1.0-mutils.Clamp((time-fadeOutStartTime)/fadeOutDuration, 0.0, 1.0),
		)
	}

	return mutils.Clamp((time-fadeInStartTime)/fadeInDuration, 0.0, 1.0)
}

func (o *DifficultyObject) Previous(backwardsIndex int) *DifficultyObject {
	index := o.Index - (backwardsIndex + 1)

	if index < 0 {
		return nil
	}

	return (*o.listOfDiffs)[index]
}

func (o *DifficultyObject) Next(forwardsIndex int) *DifficultyObject {
	index := o.Index + (forwardsIndex + 1)

	if index >= len(*o.listOfDiffs) {
		return nil
	}

	return (*o.listOfDiffs)[index]
}

func (o *DifficultyObject) setDistances() {
	if currentSlider, ok := o.BaseObject.(*LazySlider); ok {
		// danser's RepeatCount considers first span, that's why we have to subtract 1 here
		o.LazyTravelDistance = currentSlider.LazyTravelDistance
		o.LazyTravelTime = currentSlider.LazyTravelTime

		o.TravelDistance = o.LazyTravelDistance * max(1, math.Pow(float64(currentSlider.RepeatCount-1), 0.3))
		o.TravelTime = max(o.LazyTravelTime/o.Diff.Speed, MinDeltaTime)
	}

	o.MinimumJumpTime = o.AdjustedDeltaTime

	_, ok1 := o.BaseObject.(*objects.Spinner)
	_, ok2 := o.LastObject.(*objects.Spinner)

	if ok1 || ok2 {
		return
	}

	// We will scale distances by this factor, so we can assume a uniform CircleSize among beatmaps.
	scalingFactor := NormalizedRadius / float32(o.Diff.CircleRadiusL)

	lastDifficultyObject := o.Previous(0)
	lastLastDifficultyObject := o.Previous(1)

	lastCursorPosition := o.LastObject.GetStackedStartPositionMod(o.Diff)
	if lastDifficultyObject != nil {
		lastCursorPosition = getEndCursorPosition(lastDifficultyObject)
	}

	o.JumpDistance = float64((o.LastObject.GetStackedStartPositionMod(o.Diff)).Dst(o.BaseObject.GetStackedStartPositionMod(o.Diff)) * scalingFactor)
	o.LazyJumpDistance = float64(o.BaseObject.GetStackedStartPositionMod(o.Diff).Dst(lastCursorPosition) * scalingFactor)
	o.MinimumJumpDistance = o.LazyJumpDistance

	if lastSlider, ok := o.LastObject.(*LazySlider); ok && lastDifficultyObject != nil {
		lastTravelTime := max(lastDifficultyObject.LazyTravelTime/o.Diff.Speed, MinDeltaTime)
		o.MinimumJumpTime = max(o.AdjustedDeltaTime-lastTravelTime, MinDeltaTime)

		//
		// There are two types of slider-to-object patterns to consider in order to better approximate the real movement a player will take to jump between the hitobjects.
		//
		// 1. The anti-flow pattern, where players cut the slider short in order to move to the next hitobject.
		//
		//      <======o==>  ← slider
		//             |     ← most natural jump path
		//             o     ← a follow-up hitcircle
		//
		// In this case the most natural jump path is approximated by LazyJumpDistance.
		//
		// 2. The flow pattern, where players follow through the slider to its visual extent into the next hitobject.
		//
		//      <======o==>---o
		//                  ↑
		//        most natural jump path
		//
		// In this case the most natural jump path is better approximated by a new distance called "tailJumpDistance" - the distance between the slider's tail and the next hitobject.
		//
		// Thus, the player is assumed to jump the minimum of these two distances in all cases.
		//

		tailJumpDistance := lastSlider.GetStackedPositionAtModLazer(lastSlider.EndTimeLazer, o.Diff).Dst(o.BaseObject.GetStackedStartPositionMod(o.Diff)) * scalingFactor
		o.MinimumJumpDistance = max(0, min(o.LazyJumpDistance-float64(maximumSliderRadius-assumedSliderRadius), float64(tailJumpDistance-maximumSliderRadius)))
	}

	if lastLastDifficultyObject != nil && !lastLastDifficultyObject.IsSpinner {
		if lastDifficultyObject.IsSlider && lastDifficultyObject.TravelDistance > 0 {
			lastCursorPosition = lastDifficultyObject.BaseObject.GetStackedStartPositionMod(lastDifficultyObject.Diff)
		}

		lastLastCursorPosition := getEndCursorPosition(lastLastDifficultyObject)

		angle := o.calculateAngle(o.BaseObject.GetStackedStartPositionMod(o.Diff), lastCursorPosition, lastLastCursorPosition)
		sliderAngle := o.calculateSliderAngle(lastDifficultyObject, lastLastCursorPosition)

		v := o.BaseObject.GetStackedStartPositionMod(o.Diff).Sub(lastCursorPosition)
		o.NormalisedVectorAngle = math.Atan2(math.Abs(float64(v.Y)), math.Abs(float64(v.X)))

		o.Angle = min(angle, sliderAngle)
	}
}

func (o *DifficultyObject) calculateSliderAngle(lastDifficultyObject *DifficultyObject, lastLastCursorPosition vector.Vector2f) float64 {
	lastCursorPosition := getEndCursorPosition(lastDifficultyObject)

	if prevSlider, ok := lastDifficultyObject.BaseObject.(*LazySlider); ok && lastDifficultyObject.TravelDistance > 0 {
		if len(prevSlider.ScorePointsLazer) < 2 {
			lastLastCursorPosition = prevSlider.GetStackedStartPositionMod(lastDifficultyObject.Diff)
		} else {
			p := prevSlider.ScorePointsLazer[len(prevSlider.ScorePointsLazer)-2]
			lastLastCursorPosition = prevSlider.GetStackedPositionAtModLazer(p.Time, lastDifficultyObject.Diff)
		}
	}

	return o.calculateAngle(o.BaseObject.GetStackedStartPositionMod(o.Diff), lastCursorPosition, lastLastCursorPosition)
}

func (o *DifficultyObject) calculateAngle(currentPosition, lastPosition, lastLastPosition vector.Vector2f) float64 {
	v1 := lastLastPosition.Sub(lastPosition)
	v2 := currentPosition.Sub(lastPosition)

	dot := v1.Dot(v2)
	det := v1.X*v2.Y - v1.Y*v2.X

	return math.Abs(math.Atan2(float64(det), float64(dot)))
}

func getEndCursorPosition(obj *DifficultyObject) (pos vector.Vector2f) {
	if s, ok := obj.BaseObject.(*LazySlider); ok {
		return s.LazyEndPosition
	}

	return obj.BaseObject.GetStackedStartPositionMod(obj.Diff)
}
