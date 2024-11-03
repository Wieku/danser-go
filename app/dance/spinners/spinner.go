package spinners

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
)

type DanceSpinner struct {
	*objects.HitObject

	mover SpinnerMover
	id    int
}

func NewSpinner(spinner *objects.Spinner, moverCtor func() SpinnerMover, id int) *DanceSpinner {
	// data copy
	hO := *spinner.HitObject

	mover := moverCtor()

	mover.Init(hO.StartTime, hO.EndTime, id)

	danceSpinner := &DanceSpinner{
		HitObject: &hO,
		mover:     mover,
		id:        id,
	}

	danceSpinner.PositionDelegate = mover.GetPositionAt
	danceSpinner.StartPosRaw = mover.GetPositionAt(danceSpinner.StartTime)
	danceSpinner.EndPosRaw = mover.GetPositionAt(danceSpinner.EndTime)

	return danceSpinner
}

func (spinner *DanceSpinner) GetStartAngleMod(diff *difficulty.Difficulty) float32 {
	return spinner.GetStackedStartPositionMod(diff).AngleRV(spinner.GetStackedPositionAtMod(spinner.StartTime+min(10, spinner.GetDuration()), diff)) //temporary solution
}

func (spinner *DanceSpinner) GetEndAngleMod(diff *difficulty.Difficulty) float32 {
	return spinner.GetStackedEndPositionMod(diff).AngleRV(spinner.GetStackedPositionAtMod(spinner.EndTime-min(10, spinner.GetDuration()), diff)) //temporary solution
}

func (spinner *DanceSpinner) GetPartLen() float32 {
	radius := settings.CursorDance.Spinners[spinner.id%len(settings.CursorDance.Spinners)].Radius

	return float32(20.0) / float32(spinner.GetDuration()) * float32(radius)
}

func (spinner *DanceSpinner) GetStackedPositionAtMod(time float64, _ *difficulty.Difficulty) vector.Vector2f {
	return spinner.GetPositionAt(time)
}

func (spinner *DanceSpinner) GetStackedStartPositionMod(_ *difficulty.Difficulty) vector.Vector2f {
	return spinner.GetStartPosition()
}

func (spinner *DanceSpinner) GetStackedEndPositionMod(_ *difficulty.Difficulty) vector.Vector2f {
	return spinner.GetEndPosition()
}

func (spinner *DanceSpinner) GetType() objects.Type {
	return objects.SPINNER
}
