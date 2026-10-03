package evaluators

import (
	"math"

	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
)

const (
	aimWideAngleMultiplier      float64 = 9.67
	aimAcuteAngleMultiplier     float64 = 2.41
	aimSliderMultiplier         float64 = 1.5
	aimVelocityChangeMultiplier float64 = 0.9
	aimWiggleMultiplier         float64 = 1.02 // WARNING: Increasing this multiplier beyond 1.02 reduces difficulty as distance increases. Refer to the desmos link above the wiggle bonus calculation
	aimMaximumRepetitionNerf    float64 = 0.15
	aimMaximumVectorInfluence   float64 = 0.5
)

func EvaluateSnapAim(current *preprocessing.DifficultyObject, withSliderTravelDistance bool) float64 {
	if current.IsSpinner || current.Index <= 1 || current.Previous(0).IsSpinner {
		return 0
	}

	osuCurrObj := current
	osuLastObj := current.Previous(0)
	osuLast2Obj := current.Previous(2)

	const (
		radius   = preprocessing.NormalizedRadius
		diameter = preprocessing.NormalizedDiameter
	)

	// Calculate the velocity to the current hitobject, which starts with a base distance / time assuming the last object is a hitcircle.
	currDistance := osuCurrObj.JumpDistance
	if withSliderTravelDistance {
		currDistance = osuCurrObj.LazyJumpDistance
	}

	currVelocity := currDistance / osuCurrObj.AdjustedDeltaTime

	// But if the last object is a slider, then we extend the travel velocity through the slider into the current object.
	if osuLastObj.IsSlider && withSliderTravelDistance {
		sliderDistance := osuLastObj.LazyTravelDistance + osuCurrObj.LazyJumpDistance
		currVelocity = max(currVelocity, sliderDistance/osuCurrObj.AdjustedDeltaTime)
	}

	prevDistance := osuLastObj.JumpDistance
	if withSliderTravelDistance {
		prevDistance = osuLastObj.LazyJumpDistance
	}
	prevVelocity := prevDistance / osuLastObj.AdjustedDeltaTime

	snapDifficulty := currVelocity // Start difficulty with regular velocity.

	// Penalize angle repetition.
	snapDifficulty *= vectorAngleRepetition(osuCurrObj, osuLastObj)

	if !math.IsNaN(osuCurrObj.Angle) && !math.IsNaN(osuLastObj.Angle) {
		currAngle := osuCurrObj.Angle
		lastAngle := osuLastObj.Angle

		// Rewarding angles, take the smaller velocity as base.
		velocityInfluence := min(currVelocity, prevVelocity)

		acuteAngleBonus := 0.0

		if max(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime) < 1.25*min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime) { // If rhythms are the same.
			acuteAngleBonus = calcAngleAcuteness(currAngle)

			// Penalize angle repetition. It is important to do it _before_ multiplying by anything because we compare raw acuteness here
			acuteAngleBonus *= 0.08 + 0.92*(1-min(acuteAngleBonus, putils.Powi(calcAngleAcuteness(lastAngle), 3)))

			// Apply acute angle bonus for BPM above 300 1/2 and distance more than one diameter
			acuteAngleBonus *= velocityInfluence * putils.Smootherstep(putils.MillisecondsToBPM(osuCurrObj.AdjustedDeltaTime, 2), 300, 400) *
				putils.Smootherstep(currDistance, 0, diameter*2)
		}

		wideAngleBonus := calcAngleWideness(currAngle)

		// Penalize angle repetition. It is important to do it _before_ multiplying by velocity because we compare raw wideness here
		wideAngleBonus *= 0.25 + 0.75*(1-min(wideAngleBonus, putils.Powi(calcAngleWideness(lastAngle), 3)))

		// Rescaling velocity for the wide angle bonus
		const wideAngleTimeScale = 1.45
		wideAngleCurrVelocity := currDistance / math.Pow(osuCurrObj.AdjustedDeltaTime, wideAngleTimeScale)
		wideAnglePrevVelocity := prevDistance / math.Pow(osuLastObj.AdjustedDeltaTime, wideAngleTimeScale)

		if osuLastObj.IsSlider && withSliderTravelDistance {
			sliderDistance := osuLastObj.LazyTravelDistance + osuCurrObj.LazyJumpDistance
			wideAngleCurrVelocity = max(wideAngleCurrVelocity, sliderDistance/math.Pow(osuCurrObj.AdjustedDeltaTime, wideAngleTimeScale))
		}

		wideAngleBonus *= min(wideAngleCurrVelocity, wideAnglePrevVelocity)

		if osuLast2Obj != nil {
			// If objects just go back and forth through a middle point - don't give as much wide bonus
			// Use Previous(2) and Previous(0) because angles calculation is done prevprev-prev-curr, so any object's angle's center point is always the previous object
			var lastBaseObject = osuLastObj.BaseObject
			var last2BaseObject = osuLast2Obj.BaseObject

			distance := float64(last2BaseObject.GetStackedStartPositionMod(osuLast2Obj.Diff).Dst(lastBaseObject.GetStackedStartPositionMod(osuLastObj.Diff)))

			if distance < 1 {
				wideAngleBonus *= 1 - 0.55*(1-distance)
			}
		}

		// Add in acute angle bonus or wide angle bonus, whichever is larger.
		snapDifficulty += max(acuteAngleBonus*aimAcuteAngleMultiplier, wideAngleBonus*aimWideAngleMultiplier)

		// Apply wiggle bonus for jumps that are [radius, 3*diameter] in distance, with < 110 angle
		// https://www.desmos.com/calculator/dp0v0nvowc
		wiggleBonus := velocityInfluence *
			putils.Smootherstep(currDistance, radius, diameter) *
			math.Pow(putils.ReverseLerp(currDistance, diameter*3, diameter), 1.8) *
			putils.Smootherstep(currAngle, putils.DegreesToRadians(110), putils.DegreesToRadians(60)) *
			putils.Smootherstep(prevDistance, radius, diameter) *
			math.Pow(putils.ReverseLerp(prevDistance, diameter*3, diameter), 1.8) *
			putils.Smootherstep(lastAngle, putils.DegreesToRadians(110), putils.DegreesToRadians(60))

		snapDifficulty += wiggleBonus * aimWiggleMultiplier
	}

	if max(prevVelocity, currVelocity) != 0 {
		if withSliderTravelDistance {
			// We want to use just the object jump without slider velocity when awarding differences
			currVelocity = currDistance / osuCurrObj.AdjustedDeltaTime
		}

		// Scale with ratio of difference compared to 0.5 * max dist.
		distRatio := putils.Smoothstep(math.Abs(prevVelocity-currVelocity)/max(prevVelocity, currVelocity), 0, 1)

		// Reward for % distance up to 125 / strainTime for overlaps where velocity is still changing.
		overlapVelocityBuff := min(diameter*1.25/min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime), math.Abs(prevVelocity-currVelocity))

		velocityChangeBonus := overlapVelocityBuff * distRatio

		// Penalize for rhythm changes.
		velocityChangeBonus *= putils.Powi(min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime)/max(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime), 2)

		snapDifficulty += velocityChangeBonus * aimVelocityChangeMultiplier
	}

	if osuCurrObj.IsSlider && withSliderTravelDistance {
		// Reward sliders based on velocity.
		sliderBonus := osuCurrObj.TravelDistance / osuCurrObj.TravelTime

		if sliderBonus < 1 {
			snapDifficulty += sliderBonus * aimSliderMultiplier
		} else {
			snapDifficulty += math.Pow(sliderBonus, 0.75) * aimSliderMultiplier
		}
	}

	// Apply high circle size bonus
	snapDifficulty *= osuCurrObj.SmallCircleBonus

	snapDifficulty *= snapHighBpmBonus(osuCurrObj.AdjustedDeltaTime)

	return snapDifficulty
}

func snapHighBpmBonus(ms float64) float64 {
	return 1.0 / (1.0 - math.Pow(0.03, math.Pow(ms/1000, 0.65)))
}

func vectorAngleRepetition(current, previous *preprocessing.DifficultyObject) float64 {
	if math.IsNaN(current.Angle) || math.IsNaN(previous.Angle) {
		return 1
	}

	const noteLimit = 6

	constantAngleCount := 0.0

	for index := range noteLimit {
		prevObj := current.Previous(index)

		if prevObj == nil {
			break
		}

		// Only consider vectors in the same jump section, stopping to change rhythm ruins momentum
		if max(current.AdjustedDeltaTime, prevObj.AdjustedDeltaTime) > 1.1*min(current.AdjustedDeltaTime, prevObj.AdjustedDeltaTime) {
			break
		}

		if !math.IsNaN(prevObj.NormalisedVectorAngle) && !math.IsNaN(current.NormalisedVectorAngle) {
			angleDifference := math.Abs(current.NormalisedVectorAngle - prevObj.NormalisedVectorAngle)
			// Refer to this desmos for tuning, constants need to be precise so that values stay within the range of 0 and 1.
			// https://www.desmos.com/calculator/a8jesv5sv2
			constantAngleCount += math.Cos(8 * min(putils.DegreesToRadians(11.25), angleDifference))
		}
	}

	vectorRepetition := putils.Powi(min(0.5/constantAngleCount, 1), 2)

	stackFactor := putils.Smootherstep(current.LazyJumpDistance, 0, preprocessing.NormalizedDiameter)

	currAngle := current.Angle
	lastAngle := previous.Angle

	angleDifferenceAdjusted := math.Cos(2 * min(putils.DegreesToRadians(45), math.Abs(currAngle-lastAngle)*stackFactor))

	baseNerf := 1 - aimMaximumRepetitionNerf*calcAngleAcuteness(lastAngle)*angleDifferenceAdjusted

	return putils.Powi(baseNerf+(1-baseNerf)*vectorRepetition*aimMaximumVectorInfluence*stackFactor, 2)
}

func calcAngleWideness(angle float64) float64 {
	return putils.Smoothstep(angle, putils.DegreesToRadians(40), putils.DegreesToRadians(140))
}

func calcAngleAcuteness(angle float64) float64 {
	return putils.Smoothstep(angle, putils.DegreesToRadians(140), putils.DegreesToRadians(40))
}
