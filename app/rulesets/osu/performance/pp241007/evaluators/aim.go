package evaluators

import (
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	aimWideAngleMultiplier      float64 = 1.5
	aimAcuteAngleMultiplier     float64 = 1.95
	aimSliderMultiplier         float64 = 1.35
	aimVelocityChangeMultiplier float64 = 0.75
)

func EvaluateAim(current *preprocessing.DifficultyObject, withSliders bool) float64 {
	if current.IsSpinner || current.Index <= 1 {
		return 0
	}
	if current.Previous(0).IsSpinner {
		return 0
	}

	osuCurrObj := current
	osuLastObj := current.Previous(0)
	osuLastLastObj := current.Previous(1)

	// Calculate the velocity to the current hitobject, which starts with a base distance / time assuming the last object is a hitcircle.
	currVelocity := osuCurrObj.LazyJumpDistance / osuCurrObj.StrainTime

	// But if the last object is a slider, then we extend the travel velocity through the slider into the current object.
	if osuLastObj.IsSlider && withSliders {
		travelVelocity := osuLastObj.TravelDistance / osuLastObj.TravelTime             // calculate the slider velocity from slider head to slider end.
		movementVelocity := osuCurrObj.MinimumJumpDistance / osuCurrObj.MinimumJumpTime // calculate the movement velocity from slider end to current object

		currVelocity = max(currVelocity, movementVelocity+travelVelocity) // take the larger total combined velocity.
	}

	// As above, do the same for the previous hitobject.
	prevVelocity := osuLastObj.LazyJumpDistance / osuLastObj.StrainTime

	if osuLastLastObj.IsSlider && withSliders {
		travelVelocity := osuLastLastObj.TravelDistance / osuLastLastObj.TravelTime
		movementVelocity := osuLastObj.MinimumJumpDistance / osuLastObj.MinimumJumpTime

		prevVelocity = max(prevVelocity, movementVelocity+travelVelocity)
	}

	wideAngleBonus := 0.0
	acuteAngleBonus := 0.0
	sliderBonus := 0.0
	velocityChangeBonus := 0.0

	aimStrain := currVelocity // Start strain with regular velocity.

	if max(osuCurrObj.StrainTime, osuLastObj.StrainTime) < 1.25*min(osuCurrObj.StrainTime, osuLastObj.StrainTime) { // If rhythms are the same.
		if !math.IsNaN(osuCurrObj.Angle) && !math.IsNaN(osuLastObj.Angle) && !math.IsNaN(osuLastLastObj.Angle) {
			currAngle := osuCurrObj.Angle
			lastAngle := osuLastObj.Angle
			lastLastAngle := osuLastLastObj.Angle

			// Rewarding angles, take the smaller velocity as base.
			angleBonus := min(currVelocity, prevVelocity)

			wideAngleBonus = calcWideAngleBonus(currAngle)
			acuteAngleBonus = calcAcuteAngleBonus(currAngle)

			if osuCurrObj.StrainTime > 100 { // Only buff deltaTime exceeding 300 bpm 1/2.
				acuteAngleBonus = 0
			} else {
				acuteAngleBonus *= calcAcuteAngleBonus(lastAngle) * // Multiply by previous angle, we don't want to buff unless this is a wiggle type pattern.
					min(angleBonus, 125/osuCurrObj.StrainTime) * // The maximum velocity we buff is equal to 125 / strainTime
					math.Pow(math.Sin(math.Pi/2*min(1, (100-osuCurrObj.StrainTime)/25)), 2) * // scale buff from 150 bpm 1/4 to 200 bpm 1/4
					math.Pow(math.Sin(math.Pi/2*(mutils.Clamp(osuCurrObj.LazyJumpDistance, 50, 100)-50)/50), 2) // Buff distance exceeding 50 (radius) up to 100 (diameter).
			}

			// Penalize wide angles if they're repeated, reducing the penalty as the lastAngle gets more acute.
			wideAngleBonus *= angleBonus * (1 - min(wideAngleBonus, math.Pow(calcWideAngleBonus(lastAngle), 3)))
			// Penalize acute angles if they're repeated, reducing the penalty as the lastLastAngle gets more obtuse.
			acuteAngleBonus *= 0.5 + 0.5*(1-min(acuteAngleBonus, math.Pow(calcAcuteAngleBonus(lastLastAngle), 3)))
		}
	}

	if max(prevVelocity, currVelocity) != 0 {
		// We want to use the average velocity over the whole object when awarding differences, not the individual jump and slider path velocities.
		prevVelocity = (osuLastObj.LazyJumpDistance + osuLastLastObj.TravelDistance) / osuLastObj.StrainTime
		currVelocity = (osuCurrObj.LazyJumpDistance + osuLastObj.TravelDistance) / osuCurrObj.StrainTime

		// Scale with ratio of difference compared to 0.5 * max dist.
		distRatio := math.Pow(math.Sin(math.Pi/2*math.Abs(prevVelocity-currVelocity)/max(prevVelocity, currVelocity)), 2)

		// Reward for % distance up to 125 / strainTime for overlaps where velocity is still changing.
		overlapVelocityBuff := min(125/min(osuCurrObj.StrainTime, osuLastObj.StrainTime), math.Abs(prevVelocity-currVelocity))

		// Choose the largest bonus, multiplied by ratio.
		velocityChangeBonus = overlapVelocityBuff * distRatio

		// Penalize for rhythm changes.
		velocityChangeBonus *= math.Pow(min(osuCurrObj.StrainTime, osuLastObj.StrainTime)/max(osuCurrObj.StrainTime, osuLastObj.StrainTime), 2)
	}

	if osuLastObj.IsSlider && withSliders {
		// Reward sliders based on velocity.
		sliderBonus = osuLastObj.TravelDistance / osuLastObj.TravelTime
	}

	// Add in acute angle bonus or wide angle bonus + velocity change bonus, whichever is larger.
	aimStrain += max(acuteAngleBonus*aimAcuteAngleMultiplier, wideAngleBonus*aimWideAngleMultiplier+velocityChangeBonus*aimVelocityChangeMultiplier)

	if withSliders {
		// Add in additional slider velocity bonus.
		aimStrain += sliderBonus * aimSliderMultiplier
	}

	return aimStrain
}

func calcWideAngleBonus(angle float64) float64 {
	return math.Pow(math.Sin(3.0/4*(min(5.0/6*math.Pi, max(math.Pi/6, angle))-math.Pi/6)), 2)
}

func calcAcuteAngleBonus(angle float64) float64 {
	return 1 - calcWideAngleBonus(angle)
}
