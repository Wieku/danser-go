package preprocessing

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	NormalizedRadius        = 50.0
	CircleSizeBuffThreshold = 30.0
	MinDeltaTime            = 25
)

type DifficultyObject struct {
	diff *difficulty.Difficulty

	BaseObject objects.IHitObject

	lastObject objects.IHitObject

	lastLastObject objects.IHitObject

	DeltaTime float64

	StartTime float64

	EndTime float64

	JumpDistance float64

	MovementDistance float64

	TravelDistance float64

	Angle float64

	MovementTime float64

	TravelTime float64

	StrainTime float64
}

func NewDifficultyObject(hitObject, lastLastObject, lastObject objects.IHitObject, d *difficulty.Difficulty, experimental bool) *DifficultyObject {
	obj := &DifficultyObject{
		diff:           d,
		BaseObject:     hitObject,
		lastObject:     lastObject,
		lastLastObject: lastLastObject,
		DeltaTime:      (hitObject.GetStartTime() - lastObject.GetStartTime()) / d.Speed,
		StartTime:      hitObject.GetStartTime() / d.Speed,
		EndTime:        hitObject.GetEndTime() / d.Speed,
		Angle:          math.NaN(),
	}

	obj.StrainTime = math.Max(obj.DeltaTime, MinDeltaTime)

	obj.setDistances(experimental)

	return obj
}

func (o *DifficultyObject) setDistances(experimental bool) {
	_, ok1 := o.BaseObject.(*objects.Spinner)
	_, ok2 := o.lastObject.(*objects.Spinner)

	if ok1 || ok2 {
		return
	}

	scalingFactor := NormalizedRadius / float32(o.diff.CircleRadiusU)

	if o.diff.CircleRadiusU < CircleSizeBuffThreshold {
		scalingFactor *= 1.0 +
			math32.Min(CircleSizeBuffThreshold-float32(o.diff.CircleRadiusU), 5.0)/50.0
	}

	lastCursorPosition := getEndCursorPosition(o.lastObject, o.diff)
	o.JumpDistance = float64((o.BaseObject.GetStackedStartPositionMod(o.diff.Mods).Scl(scalingFactor)).Dst(lastCursorPosition.Scl(scalingFactor)))

	if lastSlider, ok := o.lastObject.(*LazySlider); ok {
		o.TravelDistance = float64(lastSlider.LazyTravelDistance)
		o.TravelTime = math.Max(lastSlider.LazyTravelTime/o.diff.Speed, MinDeltaTime)
		o.MovementTime = math.Max(o.StrainTime-o.TravelTime, MinDeltaTime)

		// Jump distance from the slider tail to the next object, as opposed to the lazy position of JumpDistance.
		tailJumpDistance := lastSlider.GetStackedPositionAtModLazer(lastSlider.EndTimeLazer, o.diff.Mods).Dst(o.BaseObject.GetStackedStartPositionMod(o.diff.Mods)) * scalingFactor

		// For hitobjects which continue in the direction of the slider, the player will normally follow through the slider,
		// such that they're not jumping from the lazy position but rather from very close to (or the end of) the slider.
		// In such cases, a leniency is applied by also considering the jump distance from the tail of the slider, and taking the minimum jump distance.
		// Additional distance is removed based on position of jump relative to slider follow circle radius.
		// JumpDistance is the leniency distance beyond the assumed_slider_radius. tailJumpDistance is maximum_slider_radius since the full distance of radial leniency is still possible.
		o.MovementDistance = math.Max(0, math.Min(o.JumpDistance-float64(maximumSliderRadius-assumedSliderRadius), float64(tailJumpDistance-maximumSliderRadius)))
	} else {
		o.MovementTime = o.StrainTime
		o.MovementDistance = o.JumpDistance
	}

	if o.lastLastObject != nil {
		if _, ok := o.lastLastObject.(*objects.Spinner); ok {
			return
		}

		lastLastCursorPosition := getEndCursorPosition(o.lastLastObject, o.diff)

		v1 := lastLastCursorPosition.Sub(o.lastObject.GetStackedStartPositionMod(o.diff.Mods))
		v2 := o.BaseObject.GetStackedStartPositionMod(o.diff.Mods).Sub(lastCursorPosition)
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
