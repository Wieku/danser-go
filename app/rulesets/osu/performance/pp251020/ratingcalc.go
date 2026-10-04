package pp251020

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const difficultyMultiplier = 0.0675

type osuRatingCalculator struct {
	diff                       *difficulty.Difficulty
	totalHits                  int
	approachRate               float64
	overallDifficulty          float64
	mechanicalDifficultyRating float64
	sliderFactor               float64
}

func newOsuRatingCalculator(diff *difficulty.Difficulty, totalHits int, approachRate, overallDifficulty, mechanicalDifficultyRating, sliderFactor float64) *osuRatingCalculator {
	return &osuRatingCalculator{
		diff:                       diff,
		totalHits:                  totalHits,
		approachRate:               approachRate,
		overallDifficulty:          overallDifficulty,
		mechanicalDifficultyRating: mechanicalDifficultyRating,
		sliderFactor:               sliderFactor,
	}
}

func (c *osuRatingCalculator) ComputeAimRating(aimDifficultyValue float64) float64 {
	if c.diff.CheckModActive(difficulty.Relax2) {
		return 0
	}

	aimRating := CalculateDifficultyRating(aimDifficultyValue)

	if c.diff.CheckModActive(difficulty.TouchDevice) {
		aimRating = math.Pow(aimRating, 0.8)
	}

	if c.diff.CheckModActive(difficulty.Relax) {
		aimRating *= 0.9
	}

	ratingMultiplier := 1.0

	approachRateLengthBonus := 0.95 + 0.4*min(1.0, float64(c.totalHits)/2000.0) +
		ternary(c.totalHits > 2000, math.Log10(float64(c.totalHits)/2000.0)*0.5, 0.0)

	approachRateFactor := 0.0
	if c.approachRate > 10.33 {
		approachRateFactor = 0.3 * (c.approachRate - 10.33)
	} else if c.approachRate < 8.0 {
		approachRateFactor = 0.05 * (8.0 - c.approachRate)
	}

	if c.diff.CheckModActive(difficulty.Relax) {
		approachRateFactor = 0.0
	}

	ratingMultiplier += approachRateFactor * approachRateLengthBonus // Buff for longer maps with high AR.

	if c.diff.CheckModActive(difficulty.Hidden) {
		visibilityFactor := c.calculateAimVisibilityFactor(c.approachRate)
		ratingMultiplier += calculateVisibilityBonusVFSF(c.diff, c.approachRate, visibilityFactor, c.sliderFactor)
	}

	// It is important to consider accuracy difficulty when scaling with accuracy.
	ratingMultiplier *= 0.98 + math.Pow(max(0, c.overallDifficulty), 2)/2500

	return aimRating * math.Cbrt(ratingMultiplier)
}

func (c *osuRatingCalculator) ComputeSpeedRating(speedDifficultyValue float64) float64 {
	if c.diff.CheckModActive(difficulty.Relax) {
		return 0
	}

	speedRating := CalculateDifficultyRating(speedDifficultyValue)

	if c.diff.CheckModActive(difficulty.Relax2) {
		speedRating *= 0.5
	}

	ratingMultiplier := 1.0

	approachRateLengthBonus := 0.95 + 0.4*min(1.0, float64(c.totalHits)/2000.0) +
		ternary(c.totalHits > 2000, math.Log10(float64(c.totalHits)/2000.0)*0.5, 0.0)

	approachRateFactor := 0.0
	if c.approachRate > 10.33 {
		approachRateFactor = 0.3 * (c.approachRate - 10.33)
	}

	if c.diff.CheckModActive(difficulty.Relax2) {
		approachRateFactor = 0.0
	}

	ratingMultiplier += approachRateFactor * approachRateLengthBonus // Buff for longer maps with high AR.

	if c.diff.CheckModActive(difficulty.Hidden) {
		visibilityFactor := c.calculateSpeedVisibilityFactor(c.approachRate)
		ratingMultiplier += calculateVisibilityBonusVF(c.diff, c.approachRate, visibilityFactor)
	}

	ratingMultiplier *= 0.95 + math.Pow(max(0, c.overallDifficulty), 2)/750

	return speedRating * math.Cbrt(ratingMultiplier)
}

func (c *osuRatingCalculator) ComputeFlashlightRating(flashlightDifficultyValue float64) float64 {
	if !c.diff.CheckModActive(difficulty.Flashlight) {
		return 0
	}

	flashlightRating := CalculateDifficultyRating(flashlightDifficultyValue)

	if c.diff.CheckModActive(difficulty.TouchDevice) {
		flashlightRating = math.Pow(flashlightRating, 0.8)
	}

	if c.diff.CheckModActive(difficulty.Relax) {
		flashlightRating *= 0.7
	} else if c.diff.CheckModActive(difficulty.Relax2) {
		flashlightRating *= 0.4
	}

	// Account for shorter maps having a higher ratio of 0 combo/100 combo flashlight radius.
	ratingMultiplier := 0.7 + 0.1*min(1.0, float64(c.totalHits)/200.0)
	if c.totalHits > 200 {
		ratingMultiplier += 0.2 * min(1.0, (float64(c.totalHits)-200)/200.0)
	}

	// It is important to consider accuracy difficulty when scaling with accuracy.
	ratingMultiplier *= 0.98 + math.Pow(max(0, c.overallDifficulty), 2)/2500

	return flashlightRating * math.Sqrt(ratingMultiplier)
}

func CalculateDifficultyRating(difficultyValue float64) float64 {
	return math.Sqrt(difficultyValue) * difficultyMultiplier
}

func (c *osuRatingCalculator) calculateAimVisibilityFactor(approachRate float64) float64 {
	const arFactorEndPoint = 11.5

	mechanicalDifficultyFactor := putils.ReverseLerp(c.mechanicalDifficultyRating, 5, 10)
	arFactorStartingPoint := mutils.Lerp(9, 10.33, mechanicalDifficultyFactor)

	return putils.ReverseLerp(approachRate, arFactorEndPoint, arFactorStartingPoint)
}

func (c *osuRatingCalculator) calculateSpeedVisibilityFactor(approachRate float64) float64 {
	const arFactorEndPoint = 11.5

	mechanicalDifficultyFactor := putils.ReverseLerp(c.mechanicalDifficultyRating, 5, 10)
	arFactorStartingPoint := mutils.Lerp(10, 10.33, mechanicalDifficultyFactor)

	return putils.ReverseLerp(approachRate, arFactorEndPoint, arFactorStartingPoint)
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
