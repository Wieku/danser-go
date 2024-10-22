package pp241007

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/skills"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	PerformanceBaseMultiplier float64 = 1.15
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	attribs api.Attributes

	score api.PerfScore

	diff *difficulty.Difficulty

	effectiveMissCount           float64
	totalHits                    int
	totalImperfectHits           int
	countSliderEndsDropped       int
	amountHitObjectsWithAccuracy int

	usingClassicSliderAccuracy bool
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

	pp.usingClassicSliderAccuracy = !diff.CheckModActive(difficulty.Lazer)

	if diff.CheckModActive(difficulty.Lazer) && diff.CheckModActive(difficulty.Classic) {
		if conf, ok := difficulty.GetModConfig[difficulty.ClassicSettings](diff); ok {
			pp.usingClassicSliderAccuracy = conf.NoSliderHeadAccuracy
		}
	}

	pp.attribs = attribs
	pp.diff = diff
	pp.score = score

	pp.countSliderEndsDropped = attribs.Sliders - score.SliderEnd
	pp.totalHits = score.CountGreat + score.CountOk + score.CountMeh + score.CountMiss
	pp.totalImperfectHits = score.CountOk + score.CountMeh + score.CountMiss
	pp.effectiveMissCount = 0

	if pp.attribs.Sliders > 0 {
		if pp.usingClassicSliderAccuracy {
			// Consider that full combo is maximum combo minus dropped slider tails since they don't contribute to combo but also don't break it
			// In classic scores we can't know the amount of dropped sliders so we estimate to 10% of all sliders on the map
			fullComboThreshold := float64(pp.attribs.MaxCombo) - 0.1*float64(pp.attribs.Sliders)

			if float64(pp.score.MaxCombo) < fullComboThreshold {
				pp.effectiveMissCount = fullComboThreshold / max(1.0, float64(pp.score.MaxCombo))
			}

			pp.effectiveMissCount = min(pp.effectiveMissCount, float64(pp.totalImperfectHits))
		} else {
			fullComboThreshold := float64(pp.attribs.MaxCombo - pp.countSliderEndsDropped)

			if float64(pp.score.MaxCombo) < fullComboThreshold {
				pp.effectiveMissCount = fullComboThreshold / max(1.0, float64(pp.score.MaxCombo))
			}

			// Combine regular misses with tick misses since tick misses break combo as well
			pp.effectiveMissCount = min(pp.effectiveMissCount, float64(pp.score.SliderBreaks+pp.score.CountMiss))
		}

	}

	pp.effectiveMissCount = max(float64(pp.score.CountMiss), pp.effectiveMissCount)

	pp.amountHitObjectsWithAccuracy = attribs.Circles
	if !pp.usingClassicSliderAccuracy {
		pp.amountHitObjectsWithAccuracy += attribs.Sliders
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

		pp.effectiveMissCount = min(pp.effectiveMissCount+float64(pp.score.CountOk)*okMultiplier+float64(pp.score.CountMeh)*mehMultiplier, float64(pp.totalHits))
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
		1.0/1.1) * multiplier

	return results
}

func (pp *PPv2) computeAimValue() float64 {
	aimValue := skills.DefaultDifficultyToPerformance(pp.attribs.Aim)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	aimValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		aimValue *= pp.calculateMissPenalty(pp.effectiveMissCount, pp.attribs.AimDifficultStrainCount)
	}

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
		estimateImproperlyFollowedDifficultSliders := 0.0

		if pp.usingClassicSliderAccuracy {
			// When the score is considered classic (regardless if it was made on old client or not) we consider all missing combo to be dropped difficult sliders
			estimateImproperlyFollowedDifficultSliders = mutils.Clamp(min(float64(pp.totalImperfectHits), float64(pp.attribs.MaxCombo-pp.score.MaxCombo)), 0, estimateDifficultSliders)
		} else {
			// We add tick misses here since they too mean that the player didn't follow the slider properly
			// We however aren't adding misses here because missing slider heads has a harsh penalty by itself and doesn't mean that the rest of the slider wasn't followed properly
			estimateImproperlyFollowedDifficultSliders = min(float64(pp.countSliderEndsDropped+pp.score.SliderBreaks), estimateDifficultSliders)
		}

		sliderNerfFactor := (1-pp.attribs.SliderFactor)*math.Pow(1-estimateImproperlyFollowedDifficultSliders/estimateDifficultSliders, 3) + pp.attribs.SliderFactor
		aimValue *= sliderNerfFactor
	}

	aimValue *= pp.score.Accuracy
	// It is important to also consider accuracy difficulty when doing that
	aimValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return aimValue
}

func (pp *PPv2) computeSpeedValue() float64 {
	if pp.diff.CheckModActive(difficulty.Relax) {
		return 0
	}

	speedValue := skills.DefaultDifficultyToPerformance(pp.attribs.Speed)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	speedValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		speedValue *= pp.calculateMissPenalty(pp.effectiveMissCount, pp.attribs.SpeedDifficultStrainCount)
	}

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
		relevantCountGreat := max(0, float64(pp.score.CountGreat)-relevantTotalDiff)
		relevantCountOk := max(0, float64(pp.score.CountOk)-max(0, relevantTotalDiff-float64(pp.score.CountGreat)))
		relevantCountMeh := max(0, float64(pp.score.CountMeh)-max(0, relevantTotalDiff-float64(pp.score.CountGreat)-float64(pp.score.CountOk)))
		relevantAccuracy = (relevantCountGreat*6.0 + relevantCountOk*2.0 + relevantCountMeh) / (pp.attribs.SpeedNoteCount * 6.0)
	}

	// Scale the speed value with accuracy and OD
	speedValue *= (0.95 + math.Pow(pp.diff.ODReal, 2)/750) * math.Pow((pp.score.Accuracy+relevantAccuracy)/2.0, (14.5-pp.diff.ODReal)/2)

	// Scale the speed value with # of 50s to punish doubletapping.
	if float64(pp.score.CountMeh) >= float64(pp.totalHits)/500 {
		speedValue *= math.Pow(0.99, float64(pp.score.CountMeh)-float64(pp.totalHits)/500.0)
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

	flashlightValue := skills.FlashlightDifficultyToPerformance(pp.attribs.Flashlight)

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
	flashlightValue *= 0.5 + pp.score.Accuracy/2.0
	// It is important to also consider accuracy difficulty when doing that.
	flashlightValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return flashlightValue
}

func (pp *PPv2) calculateMissPenalty(missCount, difficultStrainCount float64) float64 {
	return 0.96 / ((missCount / (4 * math.Pow(math.Log(difficultStrainCount), 0.94))) + 1)
}

func (pp *PPv2) getComboScalingFactor() float64 {
	if pp.attribs.MaxCombo <= 0 {
		return 1.0
	} else {
		return min(math.Pow(float64(pp.score.MaxCombo), 0.8)/math.Pow(float64(pp.attribs.MaxCombo), 0.8), 1.0)
	}
}
