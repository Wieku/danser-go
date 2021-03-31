package spinners

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type DanceSpinner struct {
	*objects.HitObject

	mover SpinnerMover
}

func NewSpinner(spinner *objects.Spinner, moverCtor func() SpinnerMover) *DanceSpinner {
	// data copy
	hO := *spinner.HitObject

	mover := moverCtor()

	mover.Init(hO.StartTime, hO.EndTime)

	danceSpinner := &DanceSpinner{
		HitObject: &hO,
		mover:     mover,
	}

	danceSpinner.PositionDelegate = mover.GetPositionAt
	danceSpinner.StartPosRaw = mover.GetPositionAt(danceSpinner.StartTime)
	danceSpinner.EndPosRaw = mover.GetPositionAt(danceSpinner.EndTime)

	return danceSpinner
}

func (spinner *DanceSpinner) GetStartAngle() float32 {
	return spinner.GetStackedStartPosition().AngleRV(spinner.GetStackedPositionAt(spinner.StartTime + math.Min(10, spinner.GetDuration()))) //temporary solution
}

func (spinner *DanceSpinner) GetStartAngleMod(modifier difficulty.Modifier) float32 {
	return spinner.GetStackedStartPositionMod(modifier).AngleRV(spinner.GetStackedPositionAtMod(spinner.StartTime+math.Min(10, spinner.GetDuration()), modifier)) //temporary solution
}

func (spinner *DanceSpinner) GetEndAngle() float32 {
	return spinner.GetStackedEndPosition().AngleRV(spinner.GetStackedPositionAt(spinner.EndTime - math.Min(10, spinner.GetDuration()))) //temporary solution
}

func (spinner *DanceSpinner) GetEndAngleMod(modifier difficulty.Modifier) float32 {
	return spinner.GetStackedEndPositionMod(modifier).AngleRV(spinner.GetStackedPositionAtMod(spinner.EndTime-math.Min(10, spinner.GetDuration()), modifier)) //temporary solution
}

func (spinner *DanceSpinner) GetPartLen() float32 {
	return float32(20.0) / float32(spinner.GetDuration()) * float32(settings.Dance.SpinnerRadius)
}

func (spinner *DanceSpinner) GetStackedPositionAtMod(time float64, _ difficulty.Modifier) vector.Vector2f {
	return spinner.GetStackedPositionAt(time)
}

func (spinner *DanceSpinner) GetStackedStartPositionMod(_ difficulty.Modifier) vector.Vector2f {
	return spinner.GetStackedStartPosition()
}

func (spinner *DanceSpinner) GetStackedEndPositionMod(_ difficulty.Modifier) vector.Vector2f {
	return spinner.GetStackedEndPosition()
}

func (spinner *DanceSpinner) GetType() objects.Type {
	return objects.SPINNER
}