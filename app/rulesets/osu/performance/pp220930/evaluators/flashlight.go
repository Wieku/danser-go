package evaluators

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/preprocessing"
	"math"
)

const (
	flMaxOpacityBonus    = 0.4
	flHiddenBonus        = 0.2
	flMinVelocity        = 0.5
	flSliderMultiplier   = 1.3
	flMinAngleMultiplier = 0.2
)

func EvaluateFlashlight(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	scalingFactor := 52.0 / current.Diff.CircleRadiusU
	smallDistNerf := 1.0
	cumulativeStrainTime := 0.0

	result := 0.0

	lastObj := current

	angleRepeatCount := 0.0

	for i := 0; i < min(current.Index, 10); i++ {
		currentObj := current.Previous(i)

		if _, ok := currentObj.BaseObject.(*objects.Spinner); !ok {
			jumpDistance := float64(current.BaseObject.GetStackedStartPositionMod(current.Diff).Dst(currentObj.BaseObject.GetStackedEndPositionMod(currentObj.Diff)))

			cumulativeStrainTime += lastObj.StrainTime

			// We want to nerf objects that can be easily seen within the Flashlight circle radius.
			if i == 0 {
				smallDistNerf = min(1.0, jumpDistance/75.0)
			}

			// We also want to nerf stacks so that only the first object of the stack is accounted for.
			stackNerf := min(1.0, (currentObj.LazyJumpDistance/scalingFactor)/25.0)

			opacityBonus := 1.0 + flMaxOpacityBonus*(1.0-current.OpacityAt(currentObj.BaseObject.GetStartTime()))

			result += stackNerf * opacityBonus * scalingFactor * jumpDistance / cumulativeStrainTime

			if !math.IsNaN(currentObj.Angle) && !math.IsNaN(current.Angle) {
				// Objects further back in time should count less for the nerf.
				if math.Abs(currentObj.Angle-current.Angle) < 0.02 {
					angleRepeatCount += max(1.0-0.1*float64(i), 0)
				}
			}
		}

		lastObj = currentObj
	}

	result = math.Pow(smallDistNerf*result, 2.0)

	// Additional bonus for Hidden due to there being no approach circles.
	if current.Diff.CheckModActive(difficulty.Hidden) {
		result *= 1.0 + flHiddenBonus
	}

	// Nerf patterns with repeated angles.
	result *= flMinAngleMultiplier + (1.0-flMinAngleMultiplier)/(angleRepeatCount+1.0)

	sliderBonus := 0.0

	if osuSlider, ok := current.BaseObject.(*preprocessing.LazySlider); ok {
		// Invert the scaling factor to determine the true travel distance independent of circle size.
		pixelTravelDistance := float64(osuSlider.LazyTravelDistance) / scalingFactor

		// Reward sliders based on velocity.
		sliderBonus = math.Pow(max(0.0, pixelTravelDistance/current.TravelTime-flMinVelocity), 0.5)

		// Longer sliders require more memorisation.
		sliderBonus *= pixelTravelDistance

		// Nerf sliders with repeats, as less memorisation is required.
		// danser's RepeatCount considers first span
		// if osuSlider.RepeatCount > 0 {
		//     sliderBonus /= float64(osuSlider.RepeatCount + 1)
		// }
		if osuSlider.RepeatCount > 1 {
			sliderBonus /= float64(osuSlider.RepeatCount)
		}
	}

	result += sliderBonus * flSliderMultiplier

	return result
}
