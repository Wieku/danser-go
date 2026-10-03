package evaluators

import (
	"math"

	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	velocityChangeMultiplier = 0.52
)

func EvaluateFlowAim(current *preprocessing.DifficultyObject, withSliderTravelDistance bool) float64 {
	if current.IsSpinner || current.Index <= 1 || current.Previous(0).IsSpinner {
		return 0
	}

	osuCurrObj := current
	osuLastObj := current.Previous(0)
	osuLastLastObj := current.Previous(1)

	currDistance := osuCurrObj.JumpDistance
	prevDistance := osuLastObj.JumpDistance

	if withSliderTravelDistance {
		currDistance = osuCurrObj.LazyJumpDistance
		prevDistance = osuLastObj.LazyJumpDistance
	}

	currVelocity := currDistance / osuCurrObj.AdjustedDeltaTime

	if osuLastObj.IsSlider && withSliderTravelDistance {
		// If the last object is a slider, then we extend the travel velocity through the slider into the current object.
		sliderDistance := osuLastObj.LazyTravelDistance + osuCurrObj.LazyJumpDistance
		currVelocity = max(currVelocity, sliderDistance/osuCurrObj.AdjustedDeltaTime)
	}

	prevVelocity := prevDistance / osuLastObj.AdjustedDeltaTime

	flowDifficulty := currVelocity

	// Apply high circle size bonus to the base velocity.
	// We use reduced CS bonus here because the bonus was made for an evaluator with a different d/t scaling
	flowDifficulty *= math.Sqrt(osuCurrObj.SmallCircleBonus)

	// Rhythm changes are harder to flow
	flowDifficulty *= 1 + min(0.25,
		putils.Powi((max(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime)-min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime))/50, 4))

	if !math.IsNaN(osuCurrObj.Angle) && !math.IsNaN(osuLastObj.Angle) {
		angleDifference := math.Abs(osuCurrObj.Angle - osuLastObj.Angle)
		angleDifferenceAdjusted := math.Sin(angleDifference/2) * 180.0
		angularVelocity := angleDifferenceAdjusted / (osuCurrObj.AdjustedDeltaTime * 0.1)

		// Low angular velocity flow (angles are consistent) is easier to follow than erratic flow
		flowDifficulty *= 0.8 + math.Sqrt(angularVelocity/270.0)
	}

	// If all three notes are overlapping - don't reward bonuses as you don't have to do additional movement
	overlappedNotesWeight := 1.0

	if current.Index > 2 {
		o1 := calculateOverlapFactor(osuCurrObj, osuLastObj)
		o2 := calculateOverlapFactor(osuCurrObj, osuLastLastObj)
		o3 := calculateOverlapFactor(osuLastObj, osuLastLastObj)

		overlappedNotesWeight = 1 - o1*o2*o3
	}

	if !math.IsNaN(osuCurrObj.Angle) {
		// Acute angles are also hard to flow
		flowDifficulty += currVelocity *
			calcAngleAcuteness(osuCurrObj.Angle) *
			overlappedNotesWeight
	}

	if max(prevVelocity, currVelocity) != 0 {
		if withSliderTravelDistance {
			currVelocity = currDistance / osuCurrObj.AdjustedDeltaTime
		}

		// Scale with ratio of difference compared to 0.5 * max dist.
		distRatio := putils.Smoothstep(math.Abs(prevVelocity-currVelocity)/max(prevVelocity, currVelocity), 0, 1)

		// Reward for % distance up to 125 / strainTime for overlaps where velocity is still changing.
		overlapVelocityBuff := min(preprocessing.NormalizedDiameter*1.25/min(osuCurrObj.AdjustedDeltaTime, osuLastObj.AdjustedDeltaTime),
			math.Abs(prevVelocity-currVelocity))

		flowDifficulty += overlapVelocityBuff * distRatio * overlappedNotesWeight * velocityChangeMultiplier
	}

	if osuCurrObj.IsSlider && withSliderTravelDistance {
		// Include slider velocity to make velocity more consistent with snap
		flowDifficulty += osuCurrObj.TravelDistance / osuCurrObj.TravelTime
	}

	// Final velocity is being raised to a power because flow difficulty scales harder with both high distance and time, and we want to account for that

	flowDifficulty = math.Pow(flowDifficulty, 1.45)

	// Reduce difficulty for low spacing since spacing below radius is always to be flowed
	return flowDifficulty * putils.Smootherstep(currDistance, 0, preprocessing.NormalizedRadius)
}

func calculateOverlapFactor(first, second *preprocessing.DifficultyObject) float64 {
	var firstBase = first.BaseObject
	var secondBase = second.BaseObject
	objectRadius := first.Diff.CircleRadiusL

	distance := float64(firstBase.GetStackedStartPositionMod(first.Diff).Dst(secondBase.GetStackedStartPositionMod(second.Diff)))
	return mutils.Clamp(1-putils.Powi(max(distance-objectRadius, 0)/objectRadius, 2), 0, 1)
}
