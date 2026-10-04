package pp251020

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
)

func CalculateMissCount(score api.PerfScore, attributes api.Attributes, diff *difficulty.Difficulty) float64 {
	if attributes.MaxCombo == 0 || diff.CheckModActive(difficulty.Lazer) {
		return 0
	}

	scoreV1Multiplier := attributes.LegacyScoreBaseMultiplier * getLegacyScoreMultiplier(diff.Mods)

	relevantComboPerObject := calculateRelevantScoreComboPerObject(score, attributes)

	maximumMissCount := calculateMaximumComboBasedMissCount(score, attributes)

	scoreObtainedDuringMaxCombo := calculateScoreAtCombo(score, attributes, float64(score.MaxCombo), relevantComboPerObject, scoreV1Multiplier)

	remainingScore := float64(score.Score) - scoreObtainedDuringMaxCombo

	if remainingScore <= 0 {
		return maximumMissCount
	}

	remainingCombo := float64(attributes.MaxCombo - score.MaxCombo)
	expectedRemainingScore := calculateScoreAtCombo(score, attributes, remainingCombo, relevantComboPerObject, scoreV1Multiplier)

	scoreBasedMissCount := expectedRemainingScore / remainingScore

	// If there's less then one miss detected - let combo-based miss count decide if this is FC or not
	scoreBasedMissCount = max(scoreBasedMissCount, 1)

	// Cap result by very harsh version of combo-based miss count
	return min(scoreBasedMissCount, maximumMissCount)
}

// / <summary>
// / Calculates the amount of score that would be achieved at a given combo.
// / </summary>
func calculateScoreAtCombo(score api.PerfScore, attributes api.Attributes, combo, relevantComboPerObject, scoreV1Multiplier float64) float64 {
	countGreat := score.CountGreat
	countOk := score.CountOk
	countMeh := score.CountMeh
	countMiss := score.CountMiss

	totalHits := countGreat + countOk + countMeh + countMiss

	estimatedObjects := combo/relevantComboPerObject - 1

	// The combo portion of ScoreV1 follows arithmetic progression
	// Therefore, we calculate the combo portion of score using the combo per object and our current combo.
	comboScore := ternary(relevantComboPerObject > 0, (2*(relevantComboPerObject-1)+(estimatedObjects-1)*relevantComboPerObject)*estimatedObjects/2, 0)

	// We then apply the accuracy and ScoreV1 multipliers to the resulting score.
	comboScore *= score.Accuracy * 300 / 25 * scoreV1Multiplier

	objectsHit := float64(totalHits-countMiss) * combo / float64(attributes.MaxCombo)

	// Score also has a non-combo portion we need to create the final score value.
	nonComboScore := (300 + attributes.NestedScorePerObject) * score.Accuracy * objectsHit

	return comboScore + nonComboScore
}

// / <summary>
// / Calculates the relevant combo per object for legacy score.
// / This assumes a uniform distribution for circles and sliders.
// / This handles cases where objects (such as buzz sliders) do not fit a normal arithmetic progression model.
// / </summary>
func calculateRelevantScoreComboPerObject(score api.PerfScore, attributes api.Attributes) float64 {
	comboScore := float64(attributes.MaximumLegacyComboScore)

	// We then reverse apply the ScoreV1 multipliers to get the raw value.
	comboScore /= 300.0 / 25.0 * attributes.LegacyScoreBaseMultiplier

	// Reverse the arithmetic progression to work out the amount of combo per object based on the score.

	result := float64((attributes.MaxCombo - 2) * attributes.MaxCombo)
	result /= max(float64(attributes.MaxCombo)+2*(comboScore-1), 1)

	return result
}

// / <summary>
// / This function is a harsher version of current combo-based miss count, used to provide reasonable value for cases where score-based miss count can't do this.
// / </summary>
func calculateMaximumComboBasedMissCount(score api.PerfScore, attributes api.Attributes) float64 {
	countMiss := score.CountMiss

	if attributes.Sliders <= 0 {
		return float64(countMiss)
	}

	countOk := score.CountOk
	countMeh := score.CountMeh

	totalImperfectHits := countOk + countMeh + countMiss

	missCount := 0.0

	// Consider that full combo is maximum combo minus dropped slider tails since they don't contribute to combo but also don't break it
	// In classic scores we can't know the amount of dropped sliders so we estimate to 10% of all sliders on the map
	fullComboThreshold := float64(attributes.MaxCombo) - 0.1*float64(attributes.Sliders)

	if float64(score.MaxCombo) < fullComboThreshold {
		missCount = math.Pow(fullComboThreshold/max(1.0, float64(score.MaxCombo)), 2.5)
	}

	// In classic scores there can't be more misses than a sum of all non-perfect judgements
	missCount = min(missCount, float64(totalImperfectHits))

	// Every slider has *at least* 2 combo attributed in classic mechanics.
	// If they broke on a slider with a tick, then this still works since they would have lost at least 2 combo (the tick and the end)
	// Using this as a max means a score that loses 1 combo on a map can't possibly have been a slider break.
	// It must have been a slider end.
	maxPossibleSliderBreaks := float64(min(attributes.Sliders, (attributes.MaxCombo-score.MaxCombo)/2))

	scoreMissCount := float64(score.CountMiss)

	sliderBreaks := missCount - scoreMissCount

	if sliderBreaks > maxPossibleSliderBreaks {
		missCount = scoreMissCount + maxPossibleSliderBreaks
	}

	return missCount
}

// / <remarks>
// / Logic copied from <see cref="OsuLegacyScoreSimulator.GetLegacyScoreMultiplier"/>.
// / </remarks>
func getLegacyScoreMultiplier(mods difficulty.Modifier) float64 {
	scoreV2 := mods.Active(difficulty.ScoreV2)

	multiplier := 1.0

	if mods.Active(difficulty.NoFail) {
		multiplier *= ternary(scoreV2, 1.0, 0.5)
	}

	if mods.Active(difficulty.Easy) {
		multiplier *= 0.5
	}

	if mods.Active(difficulty.HalfTime | difficulty.Daycore) {
		multiplier *= 0.3
	}

	if mods.Active(difficulty.Hidden) {
		multiplier *= 1.06
	}

	if mods.Active(difficulty.HardRock) {
		multiplier *= ternary(scoreV2, 1.10, 1.06)
	}

	if mods.Active(difficulty.DoubleTime | difficulty.Nightcore) {
		multiplier *= ternary(scoreV2, 1.20, 1.12)
	}

	if mods.Active(difficulty.Flashlight) {
		multiplier *= 1.12
	}

	if mods.Active(difficulty.SpunOut) {
		multiplier *= 0.9
	}

	if mods.Active(difficulty.Relax | difficulty.Relax2) {
		return 0
	}

	return multiplier
}
