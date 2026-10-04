package evaluators

import (
	"math"

	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	aimWideAngleMultiplier      float64 = 1.5
	aimAcuteAngleMultiplier     float64 = 2.55
	aimSliderMultiplier         float64 = 1.35
	aimVelocityChangeMultiplier float64 = 0.75
	aimWiggleMultiplier         float64 = 1.02
)

func EvaluateAim(current *preprocessing.DifficultyObject, withSliders bool) float64 {
	if current.IsSpinner || current.Index <= 1 || current.Previous(0).IsSpinner {
		return 0
	}

	osuCurrObj := current
	osuLastObj := current.Previous(0)
	osuLastLastObj := current.Previous(1)
	osuLast2Obj := current.Previous(2)

	const (
		radius   = preprocessing.NormalizedRadius
		diameter = preprocessing.NormalizedRadius * 2
	)

	// Calculate the velocity to the current hitobject, which starts with a base distance / time assuming the last object is a hitcircle.
	currVelocity := osuCurrObj.LazyJumpDistance / osuCurrObj.AdjustedDeltaTime

	// But if the last object is a slider, then we extend the travel velocity through the slider into the current object.
	if osuLastObj.IsSlider && withSliders {
		travelVelocity := osuLastObj.TravelDistance / osuLastObj.TravelTime             // calculate the slider velocity from slider head to slider end.
		movementVelocity := osuCurrObj.MinimumJumpDistance / osuCurrObj.MinimumJumpTime // calculate the movement velocity from slider end to current object

		currVelocity = max(currVelocity, movementVelocity+travelVelocity) // take the larger total combined velocity.
	}

	// As above, do the same for the previous hitobject.
	prevVelocity := osuLastObj.LazyJumpDistance / osuLastObj.AdjustedDeltaTime

	if osuLastLastObj.IsSlider && withSliders {
		travelVelocity := osuLastLastObj.TravelDistance / osuLastLastObj.TravelTime
		movementVelocity := osuLastObj.MinimumJumpDistance / osuLastObj.MinimumJumpTime

		prevVelocity = max(prevVelocity, movementVelocity+travelVelocity)
	}

	wideAngleBonus := 0.0
	acuteAngleBonus := 0.0
	sliderBonus := 0.0
	velocityChangeBonus := 0.0
	wiggleBonus := 0.0

	aimStrain := currVelocity // Start strain with regular velocity.

	if !math.IsNaN(osuCurrObj.Angle) && !math.IsNaN(osuLastObj.Angle) {
		currAngle := osuCurrObj.Angle
		lastAngle := osuLastObj.Angle

		// Rewarding angles, take the smaller velocity as base.
		angleBonus := min(currVelocity, prevVelocity)

		if max(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime) < 1.25*min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime) { // If rhythms are the same.
			acuteAngleBonus = calcAcuteAngleBonus(currAngle)

			// Penalize angle repetition.
			acuteAngleBonus *= 0.08 + 0.92*(1-min(acuteAngleBonus, math.Pow(calcAcuteAngleBonus(lastAngle), 3)))

			// Apply acute angle bonus for BPM above 300 1/2 and distance more than one diameter
			acuteAngleBonus *= angleBonus *
				putils.Smootherstep(putils.MillisecondsToBPM(osuCurrObj.AdjustedDeltaTime, 2), 300, 400) *
				putils.Smootherstep(osuCurrObj.LazyJumpDistance, diameter, diameter*2)
		}

		wideAngleBonus = calcWideAngleBonus(currAngle)

		// Penalize angle repetition.
		wideAngleBonus *= 1 - min(wideAngleBonus, math.Pow(calcWideAngleBonus(lastAngle), 3))

		// Apply full wide angle bonus for distance more than one diameter
		wideAngleBonus *= angleBonus * putils.Smootherstep(osuCurrObj.LazyJumpDistance, 0, diameter)

		// Apply wiggle bonus for jumps that are [radius, 3*diameter] in distance, with < 110 angle
		// https://www.desmos.com/calculator/dp0v0nvowc
		wiggleBonus = angleBonus *
			putils.Smootherstep(osuCurrObj.LazyJumpDistance, radius, diameter) *
			math.Pow(putils.ReverseLerp(osuCurrObj.LazyJumpDistance, diameter*3, diameter), 1.8) *
			putils.Smootherstep(currAngle, putils.DegreesToRadians(110), putils.DegreesToRadians(60)) *
			putils.Smootherstep(osuLastObj.LazyJumpDistance, radius, diameter) *
			math.Pow(putils.ReverseLerp(osuLastObj.LazyJumpDistance, diameter*3, diameter), 1.8) *
			putils.Smootherstep(lastAngle, putils.DegreesToRadians(110), putils.DegreesToRadians(60))

		if osuLast2Obj != nil {
			// If objects just go back and forth through a middle point - don't give as much wide bonus
			// Use Previous(2) and Previous(0) because angles calculation is done prevprev-prev-curr, so any object's angle's center point is always the previous object
			lastBaseObject := osuLastObj.BaseObject
			last2BaseObject := osuLast2Obj.BaseObject

			distance := last2BaseObject.GetStackedStartPositionMod(current.Diff).Dst(lastBaseObject.GetStackedStartPositionMod(current.Diff))

			if distance < 1 {
				wideAngleBonus *= float64(1 - 0.35*(1-distance))
			}
		}
	}

	if max(prevVelocity, currVelocity) != 0 {
		// We want to use the average velocity over the whole object when awarding differences, not the individual jump and slider path velocities.
		prevVelocity = (osuLastObj.LazyJumpDistance + osuLastLastObj.TravelDistance) / osuLastObj.AdjustedDeltaTime
		currVelocity = (osuCurrObj.LazyJumpDistance + osuLastObj.TravelDistance) / osuCurrObj.AdjustedDeltaTime

		// Scale with ratio of difference compared to 0.5 * max dist.
		distRatio := putils.Smoothstep(mutils.Abs(prevVelocity-currVelocity)/max(prevVelocity, currVelocity), 0, 1)

		// Reward for % distance up to 125 / strainTime for overlaps where velocity is still changing.
		overlapVelocityBuff := min(diameter*1.25/min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime), math.Abs(prevVelocity-currVelocity))

		// Choose the largest bonus, multiplied by ratio.
		velocityChangeBonus = overlapVelocityBuff * distRatio

		// Penalize for rhythm changes.
		velocityChangeBonus *= math.Pow(min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime)/max(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime), 2)
	}

	if osuLastObj.IsSlider && withSliders {
		// Reward sliders based on velocity.
		sliderBonus = osuLastObj.TravelDistance / osuLastObj.TravelTime
	}

	aimStrain += wiggleBonus * aimWiggleMultiplier
	aimStrain += velocityChangeBonus * aimVelocityChangeMultiplier

	// Add in acute angle bonus or wide angle bonus + velocity change bonus, whichever is larger.
	aimStrain += max(acuteAngleBonus*aimAcuteAngleMultiplier, wideAngleBonus*aimWideAngleMultiplier)

	// Apply high circle size bonus
	aimStrain *= osuCurrObj.SmallCircleBonus

	if withSliders {
		// Add in additional slider velocity bonus.
		aimStrain += sliderBonus * aimSliderMultiplier
	}

	return aimStrain
}

func calcWideAngleBonus(angle float64) float64 {
	return putils.Smoothstep(angle, putils.DegreesToRadians(40), putils.DegreesToRadians(140))
}

func calcAcuteAngleBonus(angle float64) float64 {
	return putils.Smoothstep(angle, putils.DegreesToRadians(140), putils.DegreesToRadians(40))
}
