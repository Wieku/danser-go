package pp260706

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
)

const difficultyMultiplier = 0.0675

func CalculateAimDifficultyRating(difficultyValue float64) float64 {
	return math.Pow(difficultyValue, 0.63) * 0.02275
}

func CalculateDifficultyRating(difficultyValue float64) float64 {
	return math.Sqrt(difficultyValue) * difficultyMultiplier
}

func calculateVisibilityBonus(diff *difficulty.Difficulty, approachRate float64) float64 {
	return calculateVisibilityBonusVFSF(diff, approachRate, 1, 1)
}

func calculateVisibilityBonusSF(diff *difficulty.Difficulty, approachRate, sliderFactor float64) float64 {
	return calculateVisibilityBonusVFSF(diff, approachRate, 1, sliderFactor)
}

func calculateVisibilityBonusVF(diff *difficulty.Difficulty, approachRate, visibilityFactor float64) float64 {
	return calculateVisibilityBonusVFSF(diff, approachRate, visibilityFactor, 1)
}

// Calculates a visibility bonus that is applicable to Hidden and Traceable.
func calculateVisibilityBonusVFSF(diff *difficulty.Difficulty, approachRate, visibilityFactor, sliderFactor float64) float64 {
	// NOTE: TC's effect is only noticeable in performance calculations until lazer mods are accounted for server-side.

	isAlwaysPartiallyVisible := false
	//if conf, ok := difficulty.GetModConfig[difficulty.HiddenSettings](diff); ok && conf.OnlyFadeApproachCircles {
	//	isAlwaysPartiallyVisible = true
	//}

	if diff.CheckModActive(difficulty.Traceable) {
		isAlwaysPartiallyVisible = true
	}

	// Start from normal curve, rewarding lower AR up to AR7
	// TC forcefully requires a lower reading bonus for now as it's post-applied in PP which makes it multiplicative with the regular AR bonuses
	// This means it has an advantage over HD, so we decrease the multiplier to compensate
	// This should be removed once we're able to apply TC bonuses in SR (depends on real-time difficulty calculations being possible)
	readingBonus := ternary(isAlwaysPartiallyVisible, 0.025, 0.04) * (12.0 - max(approachRate, 7))

	readingBonus *= visibilityFactor

	// We want to reward slideraim on low AR less
	sliderVisibilityFactor := math.Pow(sliderFactor, 3)

	// For AR up to 0 - reduce reward for very low ARs when object is visible
	if approachRate < 7 {
		readingBonus += ternary(isAlwaysPartiallyVisible, 0.02, 0.045) * (7.0 - max(approachRate, 0)) * sliderVisibilityFactor
	}

	// Starting from AR0 - cap values so they won't grow to infinity
	if approachRate < 0 {
		readingBonus += ternary(isAlwaysPartiallyVisible, 0.01, 0.1) * (1 - math.Pow(1.5, approachRate)) * sliderVisibilityFactor
	}

	return readingBonus
}

func ternary[T any](condition bool, a T, b T) T {
	if condition {
		return a
	}

	return b
}
