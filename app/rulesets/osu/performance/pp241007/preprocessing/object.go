package preprocessing

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	NormalizedRadius        = 50.0
	CircleSizeBuffThreshold = 30.0
	MinDeltaTime            = 25
)

type DifficultyObject struct {
	// That's stupid but oh well
	listOfDiffs *[]*DifficultyObject
	Index       int

	Diff *difficulty.Difficulty

	BaseObject objects.IHitObject

	IsSlider  bool
	IsSpinner bool

	lastObject objects.IHitObject

	lastLastObject objects.IHitObject

	DeltaTime float64

	StartTime float64

	EndTime float64

	LazyJumpDistance float64

	MinimumJumpDistance float64

	TravelDistance float64

	Angle float64

	MinimumJumpTime float64

	TravelTime float64

	StrainTime float64

	GreatWindow float64
}

func NewDifficultyObject(hitObject, lastLastObject, lastObject objects.IHitObject, d *difficulty.Difficulty, listOfDiffs *[]*DifficultyObject, index int) *DifficultyObject {
	obj := &DifficultyObject{
		listOfDiffs:    listOfDiffs,
		Index:          index,
		Diff:           d,
		BaseObject:     hitObject,
		lastObject:     lastObject,
		lastLastObject: lastLastObject,
		DeltaTime:      (hitObject.GetStartTime() - lastObject.GetStartTime()) / d.Speed,
		StartTime:      hitObject.GetStartTime() / d.Speed,
		EndTime:        hitObject.GetEndTime() / d.Speed,
		Angle:          math.NaN(),
		GreatWindow:    2 * d.Hit300U / d.Speed,
	}

	if _, ok := hitObject.(*objects.Spinner); ok {
		obj.IsSpinner = true
	}

	if _, ok := hitObject.(*LazySlider); ok {
		obj.IsSlider = true
	}

	obj.StrainTime = max(obj.DeltaTime, MinDeltaTime)

	obj.setDistances()

	return obj
}

func (o *DifficultyObject) GetDoubletapness(osuNextObj *DifficultyObject) float64 {
	if osuNextObj != nil {
		currDeltaTime := max(1, o.DeltaTime)
		nextDeltaTime := max(1, osuNextObj.DeltaTime)
		deltaDifference := math.Abs(nextDeltaTime - currDeltaTime)
		speedRatio := currDeltaTime / max(currDeltaTime, deltaDifference)
		windowRatio := math.Pow(min(1, currDeltaTime/o.GreatWindow), 2)
		return 1 - math.Pow(speedRatio, 1-windowRatio)
	}

	return 0
}

func (o *DifficultyObject) OpacityAt(time float64) float64 {
	if time > o.BaseObject.GetStartTime() {
		return 0
	}

	fadeInStartTime := o.BaseObject.GetStartTime() - o.Diff.PreemptU
	fadeInDuration := o.Diff.TimeFadeIn

	if o.Diff.CheckModActive(difficulty.Hidden) {
		fadeOutStartTime := o.BaseObject.GetStartTime() - o.Diff.PreemptU + o.Diff.TimeFadeIn
		fadeOutDuration := o.Diff.PreemptU * 0.3

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
		o.TravelDistance = float64(currentSlider.LazyTravelDistance * float32(math.Pow(1+float64(currentSlider.RepeatCount-1)/2.5, 1.0/2.5)))
		o.TravelTime = max(currentSlider.LazyTravelTime/o.Diff.Speed, MinDeltaTime)
	}

	_, ok1 := o.BaseObject.(*objects.Spinner)
	_, ok2 := o.lastObject.(*objects.Spinner)

	if ok1 || ok2 {
		return
	}

	scalingFactor := NormalizedRadius / float32(o.Diff.CircleRadiusU)

	if o.Diff.CircleRadiusU < CircleSizeBuffThreshold {
		smallCircleBonus := min(CircleSizeBuffThreshold-float32(o.Diff.CircleRadiusU), 5.0) / 50.0
		scalingFactor *= 1.0 + smallCircleBonus
	}

	lastCursorPosition := getEndCursorPosition(o.lastObject, o.Diff)

	o.LazyJumpDistance = float64((o.BaseObject.GetStackedStartPositionMod(o.Diff.Mods).Scl(scalingFactor)).Dst(lastCursorPosition.Scl(scalingFactor)))
	o.MinimumJumpTime = o.StrainTime
	o.MinimumJumpDistance = o.LazyJumpDistance

	if lastSlider, ok := o.lastObject.(*LazySlider); ok {
		lastTravelTime := max(lastSlider.LazyTravelTime/o.Diff.Speed, MinDeltaTime)
		o.MinimumJumpTime = max(o.StrainTime-lastTravelTime, MinDeltaTime)

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

		tailJumpDistance := lastSlider.GetStackedPositionAtModLazer(lastSlider.EndTimeLazer, o.Diff.Mods).Dst(o.BaseObject.GetStackedStartPositionMod(o.Diff.Mods)) * scalingFactor
		o.MinimumJumpDistance = max(0, min(o.LazyJumpDistance-float64(maximumSliderRadius-assumedSliderRadius), float64(tailJumpDistance-maximumSliderRadius)))
	}

	if o.lastLastObject != nil {
		if _, ok := o.lastLastObject.(*objects.Spinner); ok {
			return
		}

		lastLastCursorPosition := getEndCursorPosition(o.lastLastObject, o.Diff)

		v1 := lastLastCursorPosition.Sub(o.lastObject.GetStackedStartPositionMod(o.Diff.Mods))
		v2 := o.BaseObject.GetStackedStartPositionMod(o.Diff.Mods).Sub(lastCursorPosition)
		dot := v1.Dot(v2)
		det := v1.X*v2.Y - v1.Y*v2.X
		o.Angle = float64(math32.Abs(math32.Atan2(det, dot)))
	}
}

func getEndCursorPosition(obj objects.IHitObject, d *difficulty.Difficulty) (pos vector.Vector2f) {
	pos = obj.GetStackedStartPositionMod(d.Mods)

	if s, ok := obj.(*LazySlider); ok {
		pos = s.LazyEndPosition
	}

	return
}
