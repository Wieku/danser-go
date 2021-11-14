package performance

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

/* base pp value for stars, used internally by ppv2 */
func ppBase(stars float64) float64 {
	return math.Pow(5.0*math.Max(1.0, stars/0.0675)-4.0, 3.0) /
		100000.0
}

type PPv2Results struct {
	Aim, Speed, Acc, Flashlight, Total float64
}

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	Results PPv2Results

	stars Stars

	experimental bool

	maxCombo, nsliders, ncircles, nobjects int

	scoreMaxCombo      int
	countGreat         int
	countOk            int
	countMeh           int
	countMiss          int
	effectiveMissCount int

	diff *difficulty.Difficulty

	totalHits                    int
	accuracy                     float64
	amountHitObjectsWithAccuracy int
}

func (pp *PPv2) PPv2x(stars Stars, experimental bool,
	maxCombo, nsliders, ncircles, nobjects,
	combo, n300, n100, n50, nmiss int, diff *difficulty.Difficulty) PPv2 {
	maxCombo = mutils.MaxI(1, maxCombo)

	pp.maxCombo, pp.nsliders, pp.ncircles, pp.nobjects = maxCombo, nsliders, ncircles, nobjects

	if combo < 0 {
		combo = maxCombo
	}

	if n300 < 0 {
		n300 = nobjects - n100 - n50 - nmiss
	}

	totalhits := n300 + n100 + n50 + nmiss

	pp.stars = stars
	pp.experimental = experimental
	pp.diff = diff
	pp.totalHits = totalhits
	pp.scoreMaxCombo = combo
	pp.countGreat = n300
	pp.countOk = n100
	pp.countMeh = n50
	pp.countMiss = nmiss
	pp.effectiveMissCount = pp.calculateEffectiveMissCount()

	// accuracy

	if totalhits == 0 {
		pp.accuracy = 0.0
	} else {
		acc := (float64(n50)*50 +
			float64(n100)*100 +
			float64(n300)*300) /
			(float64(totalhits) * 300)

		pp.accuracy = mutils.ClampF64(acc, 0, 1)
	}

	if diff.CheckModActive(difficulty.ScoreV2) {
		pp.amountHitObjectsWithAccuracy = nobjects
	} else {
		pp.amountHitObjectsWithAccuracy = ncircles
	}

	// total pp

	finalMultiplier := 1.12

	if diff.Mods.Active(difficulty.NoFail) {
		finalMultiplier *= math.Max(0.90, 1.0-0.02*float64(pp.effectiveMissCount))
	}

	if totalhits > 0 && diff.Mods.Active(difficulty.SpunOut) {
		nspinners := nobjects - nsliders - ncircles

		finalMultiplier *= 1.0 - math.Pow(float64(nspinners)/float64(totalhits), 0.85)
	}

	if diff.Mods.Active(difficulty.Relax) {
		pp.effectiveMissCount = mutils.MinI(pp.effectiveMissCount + pp.countOk + pp.countMeh, pp.totalHits)
		finalMultiplier *= 0.6
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
		1.0/1.1) * finalMultiplier

	return *pp
}

func (pp *PPv2) computeAimValue() float64 {
	rawAim := pp.stars.Aim

	if pp.diff.Mods.Active(difficulty.TouchDevice) {
		rawAim = math.Pow(rawAim, 0.8)
	}

	aimValue := ppBase(rawAim)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*math.Min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	aimValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		aimValue *= 0.97 * math.Pow(1-math.Pow(float64(pp.effectiveMissCount)/float64(pp.totalHits), 0.775), float64(pp.effectiveMissCount))
	}

	// Combo scaling
	if pp.maxCombo > 0 {
		aimValue *= math.Min(math.Pow(float64(pp.scoreMaxCombo), 0.8)/math.Pow(float64(pp.maxCombo), 0.8), 1.0)
	}

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor = 0.3 * (pp.diff.ARReal - 10.33)
	} else if pp.diff.ARReal < 8.0 {
		approachRateFactor = 0.1 * (8.0 - pp.diff.ARReal)
	}

	aimValue *= 1.0 + approachRateFactor * lengthBonus

	// We want to give more reward for lower AR when it comes to aim and HD. This nerfs high AR and buffs lower AR.
	if pp.diff.Mods.Active(difficulty.Hidden) {
		aimValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	// We assume 15% of sliders in a map are difficult since there's no way to tell from the performance calculator.
	estimateDifficultSliders := float64(pp.nsliders) * 0.15

	if pp.nsliders > 0 {
		estimateSliderEndsDropped := mutils.ClampF64(float64(mutils.MinI(pp.countOk + pp.countMeh + pp.countMiss, pp.maxCombo - pp.scoreMaxCombo)), 0, estimateDifficultSliders)
		sliderNerfFactor := (1 - pp.stars.SliderFactor) * math.Pow(1 - estimateSliderEndsDropped / estimateDifficultSliders, 3) + pp.stars.SliderFactor
		aimValue *= sliderNerfFactor
	}

	aimValue *= pp.accuracy
	// It is important to also consider accuracy difficulty when doing that
	aimValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return aimValue
}

func (pp *PPv2) computeSpeedValue() float64 {
	speedValue := ppBase(pp.stars.Speed)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*math.Min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	speedValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		speedValue *= 0.97 * math.Pow(1-math.Pow(float64(pp.effectiveMissCount)/float64(pp.totalHits), 0.775), math.Pow(float64(pp.effectiveMissCount), 0.875))
	}

	// Combo scaling
	if pp.maxCombo > 0 {
		speedValue *= math.Min(math.Pow(float64(pp.scoreMaxCombo), 0.8)/math.Pow(float64(pp.maxCombo), 0.8), 1.0)
	}

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor = 0.3 * (pp.diff.ARReal - 10.33)
	}

	speedValue *= 1.0 + approachRateFactor * lengthBonus

	if pp.diff.Mods.Active(difficulty.Hidden) {
		speedValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	// Scale the speed value with accuracy and OD
	speedValue *= (0.95 + math.Pow(pp.diff.ODReal, 2)/750) * math.Pow(pp.accuracy, (14.5-math.Max(pp.diff.ODReal, 8))/2)
	// Scale the speed value with # of 50s to punish doubletapping.

	mehMult := 0.0
	if float64(pp.countMeh) >= float64(pp.totalHits)/500 {
		mehMult = float64(pp.countMeh) - float64(pp.totalHits)/500.0
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
	accuracyValue *= math.Min(1.15, math.Pow(float64(pp.amountHitObjectsWithAccuracy)/1000.0, 0.3))

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

	rawFlashlight := pp.stars.Flashlight

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
		flashlightValue *= 0.97 * math.Pow(1-math.Pow(float64(pp.effectiveMissCount)/float64(pp.totalHits), 0.775), math.Pow(float64(pp.effectiveMissCount), 0.875))
	}

	// Combo scaling.
	if pp.maxCombo > 0 {
		flashlightValue *= math.Min(math.Pow(float64(pp.scoreMaxCombo), 0.8)/math.Pow(float64(pp.maxCombo), 0.8), 1.0)
	}

	// Account for shorter maps having a higher ratio of 0 combo/100 combo flashlight radius.
	scale := 0.7 + 0.1*math.Min(1.0, float64(pp.totalHits)/200.0)
	if pp.totalHits > 200 {
		scale += 0.2 * math.Min(1.0, float64(pp.totalHits-200)/200.0)
	}

	flashlightValue *= scale

	// Scale the flashlight value with accuracy _slightly_.
	flashlightValue *= 0.5 + pp.accuracy/2.0
	// It is important to also consider accuracy difficulty when doing that.
	flashlightValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return flashlightValue
}

func (pp *PPv2) calculateEffectiveMissCount() int {
	// guess the number of misses + slider breaks from combo
	comboBasedMissCount := 0.0

	if pp.nsliders > 0 {
		fullComboThreshold := float64(pp.maxCombo) - 0.1*float64(pp.nsliders)
		if float64(pp.scoreMaxCombo) < fullComboThreshold {
			comboBasedMissCount = fullComboThreshold / math.Max(1.0, float64(pp.scoreMaxCombo))
		}
	}

	// we're clamping misscount because since its derived from combo it can be higher than total hits and that breaks some calculations
	comboBasedMissCount = math.Min(comboBasedMissCount, float64(pp.totalHits))

	return mutils.MaxI(pp.countMiss, int(math.Floor(comboBasedMissCount)))
}
