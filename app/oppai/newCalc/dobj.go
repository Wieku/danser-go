package newCalc

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	NormalizedRadius        = 52.0
	CirclesizeBuffThreshold = 30.0
)

type DifficultyObject struct {
	diff *difficulty.Difficulty

	BaseObject objects.IHitObject

	LastObject objects.IHitObject

	DeltaTime float64

	StartTime float64

	EndTime float64

	JumpDistance float64

	TravelDistance float64

	Angle float64

	StrainTime float64

	lastLastObject objects.IHitObject
}

func NewDifficultyObject(hitObject, lastLastObject, lastObject objects.IHitObject, d *difficulty.Difficulty) *DifficultyObject {
	obj := &DifficultyObject{
		diff:           d,
		BaseObject:     hitObject,
		LastObject:     lastObject,
		lastLastObject: lastLastObject,
		DeltaTime:      (hitObject.GetStartTime() - lastObject.GetStartTime()) / d.Speed,
		StartTime:      hitObject.GetStartTime() / d.Speed,
		EndTime:        hitObject.GetEndTime() / d.Speed,
	}

	obj.setDistances()

	obj.StrainTime = math.Max(50, obj.DeltaTime)

	return obj
}

func (o *DifficultyObject) setDistances() {
	radius := o.diff.CircleRadius
	scalingFactor := NormalizedRadius / radius

	if radius < CirclesizeBuffThreshold {
		scalingFactor *= 1.0 +
			math.Min(CirclesizeBuffThreshold-radius, 5.0)/50.0
	}

	if s, ok := o.LastObject.(*LazySlider); ok {
		o.TravelDistance = s.LazyTravelDistance * scalingFactor
	}

	lastCursorPosition := getEndCursorPosition(o.LastObject, o.diff)

	if _, ok := o.BaseObject.(*objects.Spinner); !ok {
		o.JumpDistance = (o.BaseObject.GetStackedStartPositionMod(o.diff.Mods).Copy64().Scl(scalingFactor)).Dst(lastCursorPosition.Copy64().Scl(scalingFactor))
	}

	if o.lastLastObject != nil {
		lastLastCursorPosition := getEndCursorPosition(o.lastLastObject, o.diff)

		v1 := lastLastCursorPosition.Sub(o.LastObject.GetStackedStartPositionMod(o.diff.Mods))
		v2 := o.BaseObject.GetStackedStartPositionMod(o.diff.Mods).Sub(lastCursorPosition)
		dot := v1.Dot(v2)
		det := v1.X*v2.Y - v1.Y*v2.X
		o.Angle = math.Abs(math.Atan2(float64(det), float64(dot)))
	}
}

func getEndCursorPosition(obj objects.IHitObject, d *difficulty.Difficulty) (pos vector.Vector2f) {
	pos = obj.GetStackedStartPositionMod(d.Mods)

	if s, ok := obj.(*LazySlider); ok {
		pos = s.LazyEndPosition
	}

	return
}
