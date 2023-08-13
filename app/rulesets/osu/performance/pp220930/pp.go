package pp220930

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	PerformanceBaseMultiplier float64 = 1.14
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

/* base pp value for stars, used internally by ppv2 */
func ppBase(stars float64) float64 {
	return math.Pow(5.0*max(1.0, stars/0.0675)-4.0, 3.0) /
		100000.0
}

type PPv2Results struct {
	Aim, Speed, Acc, Flashlight, Total float64
}

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	Results PPv2Results

	attribs Attributes

	experimental bool

	scoreMaxCombo      int
	countGreat         int
	countOk            int
	countMeh           int
	countMiss          int
	effectiveMissCount float64

	diff *difficulty.Difficulty

	totalHits                    int
	accuracy                     float64
	amountHitObjectsWithAccuracy int
}

func (pp *PPv2) PPv2x(attribs Attributes, combo, n300, n100, n50, nmiss int, diff *difficulty.Difficulty) PPv2 {
	attribs.MaxCombo = max(1, attribs.MaxCombo)

	if combo < 0 {
		combo = attribs.MaxCombo
	}

	if n300 < 0 {
		n300 = attribs.ObjectCount - n100 - n50 - nmiss
	}

	pp.attribs = attribs
	pp.diff = diff
	pp.totalHits = n300 + n100 + n50 + nmiss
	pp.scoreMaxCombo = combo
	pp.countGreat = n300
	pp.countOk = n100
	pp.countMeh = n50
	pp.countMiss = nmiss
	pp.effectiveMissCount = pp.calculateEffectiveMissCount()

	// accuracy

	if pp.totalHits == 0 {
		pp.accuracy = 0.0
	} else {
		acc := (float64(n50)*50 +
			float64(n100)*100 +
			float64(n300)*300) /
			(float64(pp.totalHits) * 300)

		pp.accuracy = mutils.Clamp(acc, 0, 1)
	}

	if diff.CheckModActive(difficulty.ScoreV2) {
		pp.amountHitObjectsWithAccuracy = attribs.ObjectCount
	} else {
		pp.amountHitObjectsWithAccuracy = attribs.Circles
	}

	// total pp

	multiplier := PerformanceBaseMultiplier

	if diff.Mods.Active(difficulty.NoFail) {
		multiplier *= max(0.90, 1.0-0.02*pp.effectiveMissCount)
	}

	if diff.Mods.Active(difficulty.SpunOut) && pp.totalHits > 0 {
		multiplier *= 1.0 - math.Pow(float64(attribs.Spinners)/float64(pp.totalHits), 0.85)
	}

	if diff.Mods.Active(difficulty.Relax) {
		okMultiplier := 1.0
		mehMultiplier := 1.0

		if diff.ODReal > 0.0 {
			okMultiplier = max(0.0, 1-math.Pow(diff.ODReal/13.33, 1.8))
			mehMultiplier = max(0.0, 1-math.Pow(diff.ODReal/13.33, 5))
		}

		pp.effectiveMissCount = min(pp.effectiveMissCount+float64(pp.countOk)*okMultiplier+float64(pp.countMeh)*mehMultiplier, float64(pp.totalHits))
	}

	pp.Results.Aim = pp.computeAimValue()
	pp.Results.Speed = pp.computeSpeedValue()
	pp.Results.Acc = pp.computeAccuracyValue()
	pp.Results.Flashlight = pp.computeFlashlightValue()

	pp.Results.Total = math.Pow(
		math.Pow(pp.Results.Aim, 1.1)+
			math.Pow(pp.Results.Speed, 1.1)+
			math.Pow(pp.Results.Acc, 1.1)+
			math.Pow(pp.Results.Flashlight, 1.1),
		1.0/1.1) * multiplier

	return *pp
}

func (pp *PPv2) computeAimValue() float64 {
	aimValue := ppBase(pp.attribs.Aim)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	aimValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		aimValue *= 0.97 * math.Pow(1-math.Pow(pp.effectiveMissCount/float64(pp.totalHits), 0.775), pp.effectiveMissCount)
	}

	// Combo scaling
	aimValue *= pp.getComboScalingFactor()

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor = 0.3 * (pp.diff.ARReal - 10.33)
	} else if pp.diff.ARReal < 8.0 {
		approachRateFactor = 0.05 * (8.0 - pp.diff.ARReal)
	}

	if pp.diff.CheckModActive(difficulty.Relax) {
		approachRateFactor = 0.0
	}

	aimValue *= 1.0 + approachRateFactor*lengthBonus

	// We want to give more reward for lower AR when it comes to aim and HD. This nerfs high AR and buffs lower AR.
	if pp.diff.Mods.Active(difficulty.Hidden) {
		aimValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	// We assume 15% of sliders in a map are difficult since there's no way to tell from the performance calculator.
	estimateDifficultSliders := float64(pp.attribs.Sliders) * 0.15

	if pp.attribs.Sliders > 0 {
		estimateSliderEndsDropped := mutils.Clamp(float64(min(pp.countOk+pp.countMeh+pp.countMiss, pp.attribs.MaxCombo-pp.scoreMaxCombo)), 0, estimateDifficultSliders)
		sliderNerfFactor := (1-pp.attribs.SliderFactor)*math.Pow(1-estimateSliderEndsDropped/estimateDifficultSliders, 3) + pp.attribs.SliderFactor
		aimValue *= sliderNerfFactor
	}

	aimValue *= pp.accuracy
	// It is important to also consider accuracy difficulty when doing that
	aimValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return aimValue
}

func (pp *PPv2) computeSpeedValue() float64 {
	if pp.diff.CheckModActive(difficulty.Relax) {
		return 0
	}

	speedValue := ppBase(pp.attribs.Speed)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	speedValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		speedValue *= 0.97 * math.Pow(1-math.Pow(pp.effectiveMissCount/float64(pp.totalHits), 0.775), math.Pow(pp.effectiveMissCount, 0.875))
	}

	// Combo scaling
	speedValue *= pp.getComboScalingFactor()

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor = 0.3 * (pp.diff.ARReal - 10.33)
	}

	speedValue *= 1.0 + approachRateFactor*lengthBonus

	if pp.diff.Mods.Active(difficulty.Hidden) {
		speedValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	relevantAccuracy := 0.0
	if pp.attribs.SpeedNoteCount != 0 {
		relevantTotalDiff := float64(pp.totalHits) - pp.attribs.SpeedNoteCount
		relevantCountGreat := max(0, float64(pp.countGreat)-relevantTotalDiff)
		relevantCountOk := max(0, float64(pp.countOk)-max(0, relevantTotalDiff-float64(pp.countGreat)))
		relevantCountMeh := max(0, float64(pp.countMeh)-max(0, relevantTotalDiff-float64(pp.countGreat)-float64(pp.countOk)))
		relevantAccuracy = (relevantCountGreat*6.0 + relevantCountOk*2.0 + relevantCountMeh) / (pp.attribs.SpeedNoteCount * 6.0)
	}

	// Scale the speed value with accuracy and OD
	speedValue *= (0.95 + math.Pow(pp.diff.ODReal, 2)/750) * math.Pow((pp.accuracy+relevantAccuracy)/2.0, (14.5-max(pp.diff.ODReal, 8))/2)

	// Scale the speed value with # of 50s to punish doubletapping.
	if float64(pp.countMeh) >= float64(pp.totalHits)/500 {
		speedValue *= math.Pow(0.99, float64(pp.countMeh)-float64(pp.totalHits)/500.0)
	}

	return speedValue
}

func (pp *PPv2) computeAccuracyValue() float64 {
	if pp.diff.Mods.Active(difficulty.Relax) {
		return 0.0
	}

	// This percentage only considers HitCircles of any value - in this part of the calculation we focus on hitting the timing hit window
	betterAccuracyPercentage := 0.0

	if pp.amountHitObjectsWithAccuracy > 0 {
		betterAccuracyPercentage = float64((pp.countGreat-(pp.totalHits-pp.amountHitObjectsWithAccuracy))*6+pp.countOk*2+pp.countMeh) / (float64(pp.amountHitObjectsWithAccuracy) * 6)
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

	flashlightValue := math.Pow(pp.attribs.Flashlight, 2.0) * 25.0

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		flashlightValue *= 0.97 * math.Pow(1-math.Pow(pp.effectiveMissCount/float64(pp.totalHits), 0.775), math.Pow(pp.effectiveMissCount, 0.875))
	}

	// Combo scaling.
	flashlightValue *= pp.getComboScalingFactor()

	// Account for shorter maps having a higher ratio of 0 combo/100 combo flashlight radius.
	scale := 0.7 + 0.1*min(1.0, float64(pp.totalHits)/200.0)
	if pp.totalHits > 200 {
		scale += 0.2 * min(1.0, float64(pp.totalHits-200)/200.0)
	}

	flashlightValue *= scale

	// Scale the flashlight value with accuracy _slightly_.
	flashlightValue *= 0.5 + pp.accuracy/2.0
	// It is important to also consider accuracy difficulty when doing that.
	flashlightValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return flashlightValue
}

func (pp *PPv2) calculateEffectiveMissCount() float64 {
	// guess the number of misses + slider breaks from combo
	comboBasedMissCount := 0.0

	if pp.attribs.Sliders > 0 {
		fullComboThreshold := float64(pp.attribs.MaxCombo) - 0.1*float64(pp.attribs.Sliders)
		if float64(pp.scoreMaxCombo) < fullComboThreshold {
			comboBasedMissCount = fullComboThreshold / max(1.0, float64(pp.scoreMaxCombo))
		}
	}

	// Clamp miss count to maximum amount of possible breaks
	comboBasedMissCount = min(comboBasedMissCount, float64(pp.countOk+pp.countMeh+pp.countMiss))

	return max(float64(pp.countMiss), comboBasedMissCount)
}

func (pp *PPv2) calculateMissPenalty(missCount, difficultStrainCount float64) float64 {
	return 0.95 / ((missCount / (3 * math.Sqrt(difficultStrainCount))) + 1)
}

func (pp *PPv2) getComboScalingFactor() float64 {
	if pp.attribs.MaxCombo <= 0 {
		return 1.0
	} else {
		return min(math.Pow(float64(pp.scoreMaxCombo), 0.8)/math.Pow(float64(pp.attribs.MaxCombo), 0.8), 1.0)
	}
}
