package pp211112

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

/* base pp value for stars, used internally by ppv2 */
func ppBase(stars float64) float64 {
	return math.Pow(5.0*max(1.0, stars/0.0675)-4.0, 3.0) /
		100000.0
}

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	attribs api.Attributes

	experimental bool

	score api.PerfScore

	diff *difficulty.Difficulty

	effectiveMissCount           float64
	totalHits                    int
	amountHitObjectsWithAccuracy int
}

func NewPPCalculator() api.IPerformanceCalculator {
	return &PPv2{}
}

func (pp *PPv2) Calculate(attribs api.Attributes, score api.PerfScore, diff *difficulty.Difficulty) api.PPv2Results {
	attribs.MaxCombo = max(1, attribs.MaxCombo)

	if score.MaxCombo < 0 {
		score.MaxCombo = attribs.MaxCombo
	}

	if score.CountGreat < 0 {
		score.CountGreat = attribs.ObjectCount - score.CountOk - score.CountMeh - score.CountMiss
	}

	pp.attribs = attribs
	pp.diff = diff
	pp.score = score

	pp.totalHits = score.CountGreat + score.CountOk + score.CountMeh + score.CountMiss
	pp.effectiveMissCount = pp.calculateEffectiveMissCount()
	pp.experimental = false

	if diff.CheckModActive(difficulty.ScoreV2) {
		pp.amountHitObjectsWithAccuracy = attribs.Circles + attribs.Sliders
	} else {
		pp.amountHitObjectsWithAccuracy = attribs.Circles
	}

	// total pp

	finalMultiplier := 1.12

	if diff.Mods.Active(difficulty.NoFail) {
		finalMultiplier *= max(0.90, 1.0-0.02*pp.effectiveMissCount)
	}

	if pp.totalHits > 0 && diff.Mods.Active(difficulty.SpunOut) {
		finalMultiplier *= 1.0 - math.Pow(float64(attribs.Spinners)/float64(pp.totalHits), 0.85)
	}

	if diff.Mods.Active(difficulty.Relax) {
		pp.effectiveMissCount = min(pp.effectiveMissCount+float64(pp.score.CountOk+pp.score.CountMeh), float64(pp.totalHits))
		finalMultiplier *= 0.6
	}

	results := api.PPv2Results{
		Aim:        pp.computeAimValue(),
		Speed:      pp.computeSpeedValue(),
		Acc:        pp.computeAccuracyValue(),
		Flashlight: pp.computeFlashlightValue(),
	}

	results.Total = math.Pow(
		math.Pow(results.Aim, 1.1)+
			math.Pow(results.Speed, 1.1)+
			math.Pow(results.Acc, 1.1)+
			math.Pow(results.Flashlight, 1.1),
		1.0/1.1) * finalMultiplier

	return results
}

func (pp *PPv2) computeAimValue() float64 {
	rawAim := pp.attribs.Aim

	if pp.diff.Mods.Active(difficulty.TouchDevice) {
		rawAim = math.Pow(rawAim, 0.8)
	}

	aimValue := ppBase(rawAim)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	aimValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		if pp.experimental {
			aimValue *= pp.calculateMissPenalty(pp.effectiveMissCount, pp.attribs.AimDifficultStrainCount)
		} else {
			aimValue *= 0.97 * math.Pow(1-math.Pow(pp.effectiveMissCount/float64(pp.totalHits), 0.775), pp.effectiveMissCount)
		}
	}

	// Combo scaling
	if pp.attribs.MaxCombo > 0 && !pp.experimental {
		aimValue *= pp.getComboScalingFactor()
	}

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor = 0.3 * (pp.diff.ARReal - 10.33)
	} else if pp.diff.ARReal < 8.0 {
		approachRateFactor = 0.1 * (8.0 - pp.diff.ARReal)
	}

	aimValue *= 1.0 + approachRateFactor*lengthBonus

	// We want to give more reward for lower AR when it comes to aim and HD. This nerfs high AR and buffs lower AR.
	if pp.diff.Mods.Active(difficulty.Hidden) {
		aimValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	// We assume 15% of sliders in a map are difficult since there's no way to tell from the performance calculator.
	estimateDifficultSliders := float64(pp.attribs.Sliders) * 0.15

	if pp.attribs.Sliders > 0 {
		estimateSliderEndsDropped := mutils.Clamp(float64(min(pp.score.CountOk+pp.score.CountMeh+pp.score.CountMiss, pp.attribs.MaxCombo-pp.score.MaxCombo)), 0, estimateDifficultSliders)
		sliderNerfFactor := (1-pp.attribs.SliderFactor)*math.Pow(1-estimateSliderEndsDropped/estimateDifficultSliders, 3) + pp.attribs.SliderFactor
		aimValue *= sliderNerfFactor
	}

	aimValue *= pp.score.Accuracy
	// It is important to also consider accuracy difficulty when doing that
	aimValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return aimValue
}

func (pp *PPv2) computeSpeedValue() float64 {
	speedValue := ppBase(pp.attribs.Speed)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	speedValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		if pp.experimental {
			speedValue *= pp.calculateMissPenalty(pp.effectiveMissCount, pp.attribs.SpeedDifficultStrainCount)
		} else {
			speedValue *= 0.97 * math.Pow(1-math.Pow(pp.effectiveMissCount/float64(pp.totalHits), 0.775), math.Pow(pp.effectiveMissCount, 0.875))
		}
	}

	// Combo scaling
	if pp.attribs.MaxCombo > 0 && !pp.experimental {
		speedValue *= pp.getComboScalingFactor()
	}

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor = 0.3 * (pp.diff.ARReal - 10.33)
	}

	speedValue *= 1.0 + approachRateFactor*lengthBonus

	if pp.diff.Mods.Active(difficulty.Hidden) {
		speedValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	// Scale the speed value with accuracy and OD
	speedValue *= (0.95 + math.Pow(pp.diff.ODReal, 2)/750) * math.Pow(pp.score.Accuracy, (14.5-max(pp.diff.ODReal, 8))/2)
	// Scale the speed value with # of 50s to punish doubletapping.

	mehMult := 0.0
	if float64(pp.score.CountMeh) >= float64(pp.totalHits)/500 {
		mehMult = float64(pp.score.CountMeh) - float64(pp.totalHits)/500.0
	}

	speedValue *= math.Pow(0.98, mehMult)

	return speedValue
}

func (pp *PPv2) computeAccuracyValue() float64 {
	if pp.diff.Mods.Active(difficulty.Relax) {
		return 0.0
	}

	// This percentage only considers HitCircles of any value - in this part of the calculation we focus on hitting the timing hit window
	betterAccuracyPercentage := 0.0

	if pp.amountHitObjectsWithAccuracy > 0 {
		betterAccuracyPercentage = float64((pp.score.CountGreat-(pp.totalHits-pp.amountHitObjectsWithAccuracy))*6+pp.score.CountOk*2+pp.score.CountMeh) / (float64(pp.amountHitObjectsWithAccuracy) * 6)
	}

	// It is possible to reach a negative accuracy with this formula. Cap it at zero - zero points
	if betterAccuracyPercentage < 0 {
		betterAccuracyPercentage = 0
	}

	// Lots of arbitrary values from testing.
	// Considering to use derivation from perfect accuracy in a probabilistic manner - assume normal distribution
	accuracyValue := math.Pow(1.52163, pp.diff.ODReal) * math.Pow(betterAccuracyPercentage, 24) * 2.83

	// Bonus for many hitcircles - it's harder to keep good accuracy up for longer
	accuracyValue *= min(1.15, math.Pow(float64(pp.amountHitObjectsWithAccuracy)/1000.0, 0.3))

	if pp.diff.Mods.Active(difficulty.Hidden) {
		accuracyValue *= 1.08
	}

	if pp.diff.Mods.Active(difficulty.Flashlight) {
		accuracyValue *= 1.02
	}

	return accuracyValue
}

func (pp *PPv2) computeFlashlightValue() float64 {
	if !pp.diff.CheckModActive(difficulty.Flashlight) {
		return 0
	}

	rawFlashlight := pp.attribs.Flashlight

	if pp.diff.CheckModActive(difficulty.TouchDevice) {
		rawFlashlight = math.Pow(rawFlashlight, 0.8)
	}

	flashlightValue := math.Pow(rawFlashlight, 2.0) * 25.0

	// Add an additional bonus for HDFL.
	if pp.diff.CheckModActive(difficulty.Hidden) {
		flashlightValue *= 1.3
	}

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		flashlightValue *= 0.97 * math.Pow(1-math.Pow(pp.effectiveMissCount/float64(pp.totalHits), 0.775), math.Pow(pp.effectiveMissCount, 0.875))
	}

	// Combo scaling.
	if pp.attribs.MaxCombo > 0 {
		flashlightValue *= pp.getComboScalingFactor()
	}

	// Account for shorter maps having a higher ratio of 0 combo/100 combo flashlight radius.
	scale := 0.7 + 0.1*min(1.0, float64(pp.totalHits)/200.0)
	if pp.totalHits > 200 {
		scale += 0.2 * min(1.0, float64(pp.totalHits-200)/200.0)
	}

	flashlightValue *= scale

	// Scale the flashlight value with accuracy _slightly_.
	flashlightValue *= 0.5 + pp.score.Accuracy/2.0
	// It is important to also consider accuracy difficulty when doing that.
	flashlightValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return flashlightValue
}

func (pp *PPv2) calculateEffectiveMissCount() float64 {
	// guess the number of misses + slider breaks from combo
	comboBasedMissCount := 0.0

	if pp.attribs.Sliders > 0 {
		fullComboThreshold := float64(pp.attribs.MaxCombo) - 0.1*float64(pp.attribs.Sliders)
		if float64(pp.score.MaxCombo) < fullComboThreshold {
			comboBasedMissCount = fullComboThreshold / max(1.0, float64(pp.score.MaxCombo))
		}
	}

	// we're clamping misscount because since its derived from combo it can be higher than total hits and that breaks some calculations
	comboBasedMissCount = min(comboBasedMissCount, float64(pp.totalHits))

	if pp.experimental {
		return max(float64(pp.score.CountMiss), comboBasedMissCount)
	} else {
		return max(float64(pp.score.CountMiss), math.Floor(comboBasedMissCount))
	}
}

func (pp *PPv2) calculateMissPenalty(missCount, difficultStrainCount float64) float64 {
	return 0.95 / ((missCount / (3 * math.Sqrt(difficultStrainCount))) + 1)
}

func (pp *PPv2) getComboScalingFactor() float64 {
	if pp.attribs.MaxCombo <= 0 {
		return 1.0
	} else {
		return min(math.Pow(float64(pp.score.MaxCombo), 0.8)/math.Pow(float64(pp.attribs.MaxCombo), 0.8), 1.0)
	}
}
