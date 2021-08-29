package preprocessing

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	NormalizedRadius        = 52.0
	CircleSizeBuffThreshold = 30.0
	OsuStableAllowance      = 1.00041
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
	radius := o.diff.CircleRadius / OsuStableAllowance // we need to undo that weird allowance mentioned in difficulty.Difficulty.calculate()
	scalingFactor := NormalizedRadius / float32(radius)

	if radius < CircleSizeBuffThreshold {
		scalingFactor *= 1.0 +
			math32.Min(CircleSizeBuffThreshold-float32(radius), 5.0)/50.0
	}

	if s, ok := o.LastObject.(*LazySlider); ok {
		o.TravelDistance = float64(s.LazyTravelDistance * scalingFactor)
	}

	lastCursorPosition := getEndCursorPosition(o.LastObject, o.diff)

	if _, ok := o.BaseObject.(*objects.Spinner); !ok {
		o.JumpDistance = float64((o.BaseObject.GetStackedStartPositionMod(o.diff.Mods).Scl(scalingFactor)).Dst(lastCursorPosition.Scl(scalingFactor)))
	}

	if o.lastLastObject != nil {
		lastLastCursorPosition := getEndCursorPosition(o.lastLastObject, o.diff)

		v1 := lastLastCursorPosition.Sub(o.LastObject.GetStackedStartPositionMod(o.diff.Mods))
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
