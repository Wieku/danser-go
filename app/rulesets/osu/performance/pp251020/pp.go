package pp251020

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020/skills"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	PerformanceBaseMultiplier float64 = 1.14
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	attribs api.Attributes

	score api.PerfScore

	diff *difficulty.Difficulty

	effectiveMissCount     float64
	totalHits              int
	totalSuccessfulHits    int
	totalImperfectHits     int
	countSliderEndsDropped int

	usingClassicSliderAccuracy bool

	greatHitWindow float64
	okHitWindow    float64
	mehHitWindow   float64

	speedDeviation             *float64
	speedEstimatedSliderBreaks float64
	aimEstimatedSliderBreaks   float64
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

	if score.SliderEnd < 0 {
		score.SliderEnd = attribs.Sliders
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
	pp.totalSuccessfulHits = score.CountGreat + score.CountOk + score.CountMeh
	pp.totalImperfectHits = score.CountOk + score.CountMeh + score.CountMiss
	pp.effectiveMissCount = float64(score.CountMiss)

	pp.greatHitWindow = diff.Hit300U / diff.GetSpeed()
	pp.okHitWindow = diff.Hit100U / diff.GetSpeed()
	pp.mehHitWindow = diff.Hit50U / diff.GetSpeed()

	if pp.attribs.Sliders > 0 {
		if pp.usingClassicSliderAccuracy {
			pp.effectiveMissCount = CalculateMissCount(score, attribs, diff)
		} else {
			pp.effectiveMissCount = pp.calculateComboBasedEstimatedMissCount(pp.attribs)
		}
	}

	pp.effectiveMissCount = max(float64(pp.score.CountMiss), pp.effectiveMissCount)
	pp.effectiveMissCount = min(float64(pp.totalHits), pp.effectiveMissCount)

	// total pp

	multiplier := PerformanceBaseMultiplier

	if diff.Mods.Active(difficulty.NoFail) {
		multiplier *= max(0.90, 1.0-0.02*pp.effectiveMissCount)
	}

	if diff.Mods.Active(difficulty.SpunOut) && pp.totalHits > 0 {
		multiplier *= 1.0 - math.Pow(float64(attribs.Spinners)/float64(pp.totalHits), 0.85)
	}

	if diff.Mods.Active(difficulty.Relax) {
		okMultiplier := 0.75
		mehMultiplier := 1.0

		if diff.ODReal > 0.0 {
			okMultiplier *= max(0.0, 1-math.Pow(diff.ODReal/13.33, 1.8))
			mehMultiplier *= max(0.0, 1-math.Pow(diff.ODReal/13.33, 5))
		}

		pp.effectiveMissCount = min(pp.effectiveMissCount+float64(pp.score.CountOk)*okMultiplier+float64(pp.score.CountMeh)*mehMultiplier, float64(pp.totalHits))
	}

	pp.speedDeviation = pp.calculateSpeedDeviation(pp.attribs)

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
	if pp.diff.CheckModActive(difficulty.Relax2) {
		return 0
	}

	aimDifficulty := pp.attribs.Aim

	// We assume 15% of sliders in a map are difficult since there's no way to tell from the performance calculator.
	//estimateDifficultSliders := float64(pp.attribs.Sliders) * 0.15

	if pp.attribs.Sliders > 0 && pp.attribs.AimDifficultSliderCount > 0 {
		estimateImproperlyFollowedDifficultSliders := 0.0

		if pp.usingClassicSliderAccuracy {
			// When the score is considered classic (regardless if it was made on old client or not) we consider all missing combo to be dropped difficult sliders
			estimateImproperlyFollowedDifficultSliders = mutils.Clamp(min(float64(pp.totalImperfectHits), float64(pp.attribs.MaxCombo-pp.score.MaxCombo)), 0, pp.attribs.AimDifficultSliderCount)
		} else {
			// We add tick misses here since they too mean that the player didn't follow the slider properly
			// We however aren't adding misses here because missing slider heads has a harsh penalty by itself and doesn't mean that the rest of the slider wasn't followed properly
			estimateImproperlyFollowedDifficultSliders = mutils.Clamp(float64(pp.countSliderEndsDropped+pp.score.SliderBreaks), 0, pp.attribs.AimDifficultSliderCount)
		}

		sliderNerfFactor := (1-pp.attribs.SliderFactor)*math.Pow(1-estimateImproperlyFollowedDifficultSliders/pp.attribs.AimDifficultSliderCount, 3) + pp.attribs.SliderFactor
		aimDifficulty *= sliderNerfFactor
	}

	aimValue := skills.DefaultDifficultyToPerformance(aimDifficulty)

	// Longer maps are worth more
	lengthBonus := 0.95 + 0.4*min(1.0, float64(pp.totalHits)/2000.0)
	if pp.totalHits > 2000 {
		lengthBonus += math.Log10(float64(pp.totalHits)/2000.0) * 0.5
	}

	aimValue *= lengthBonus

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.effectiveMissCount > 0 {
		pp.aimEstimatedSliderBreaks = pp.calculateEstimatedSliderBreaks(pp.attribs.AimTopWeightedSliderFactor, pp.attribs)

		relevantMissCount := min(pp.effectiveMissCount+pp.aimEstimatedSliderBreaks, float64(pp.totalImperfectHits+pp.score.SliderBreaks))

		aimValue *= pp.calculateMissPenalty(relevantMissCount, pp.attribs.AimDifficultStrainCount)
	}

	if pp.diff.CheckModActive(difficulty.Traceable) {
		aimValue *= 1.0 + calculateVisibilityBonusSF(pp.diff, pp.diff.ARReal, pp.attribs.SliderFactor)
	}

	aimValue *= pp.score.Accuracy

	return aimValue
}

func (pp *PPv2) computeSpeedValue() float64 {
	if pp.diff.CheckModActive(difficulty.Relax) || pp.speedDeviation == nil {
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
		pp.speedEstimatedSliderBreaks = pp.calculateEstimatedSliderBreaks(pp.attribs.SpeedTopWeightedSliderFactor, pp.attribs)

		relevantMissCount := min(pp.effectiveMissCount+pp.speedEstimatedSliderBreaks, float64(pp.totalImperfectHits+pp.score.SliderBreaks))

		speedValue *= pp.calculateMissPenalty(relevantMissCount, pp.attribs.SpeedDifficultStrainCount)
	}

	if pp.diff.Mods.Active(difficulty.Traceable) {
		speedValue *= 1.0 + calculateVisibilityBonus(pp.diff, pp.diff.ARReal)
	}

	speedHighDeviationMultiplier := pp.calculateSpeedHighDeviationNerf(pp.attribs)
	speedValue *= speedHighDeviationMultiplier

	relevantAccuracy := 0.0
	if pp.attribs.SpeedNoteCount != 0 {
		relevantTotalDiff := max(0, float64(pp.totalHits)-pp.attribs.SpeedNoteCount)
		relevantCountGreat := max(0, float64(pp.score.CountGreat)-relevantTotalDiff)
		relevantCountOk := max(0, float64(pp.score.CountOk)-max(0, relevantTotalDiff-float64(pp.score.CountGreat)))
		relevantCountMeh := max(0, float64(pp.score.CountMeh)-max(0, relevantTotalDiff-float64(pp.score.CountGreat)-float64(pp.score.CountOk)))
		relevantAccuracy = (relevantCountGreat*6.0 + relevantCountOk*2.0 + relevantCountMeh) / (pp.attribs.SpeedNoteCount * 6.0)
	}

	// Scale the speed value with accuracy and OD
	speedValue *= math.Pow((pp.score.Accuracy+relevantAccuracy)/2.0, (14.5-pp.diff.ODReal)/2)

	return speedValue
}

func (pp *PPv2) computeAccuracyValue() float64 {
	if pp.diff.Mods.Active(difficulty.Relax) {
		return 0.0
	}

	amountHitObjectsWithAccuracy := pp.attribs.Circles
	if !pp.usingClassicSliderAccuracy || pp.diff.CheckModActive(difficulty.ScoreV2) {
		amountHitObjectsWithAccuracy += pp.attribs.Sliders
	}

	// This percentage only considers HitCircles of any value - in this part of the calculation we focus on hitting the timing hit window
	betterAccuracyPercentage := 0.0

	if amountHitObjectsWithAccuracy > 0 {
		betterAccuracyPercentage = float64((pp.score.CountGreat-max(pp.totalHits-amountHitObjectsWithAccuracy, 0))*6+pp.score.CountOk*2+pp.score.CountMeh) / (float64(amountHitObjectsWithAccuracy) * 6)
	}

	// It is possible to reach a negative accuracy with this formula. Cap it at zero - zero points
	if betterAccuracyPercentage < 0 {
		betterAccuracyPercentage = 0
	}

	// Lots of arbitrary values from testing.
	// Considering to use derivation from perfect accuracy in a probabilistic manner - assume normal distribution
	accuracyValue := math.Pow(1.52163, pp.diff.ODReal) * math.Pow(betterAccuracyPercentage, 24) * 2.83

	// Bonus for many hitcircles - it's harder to keep good accuracy up for longer
	accuracyValue *= min(1.15, math.Pow(float64(amountHitObjectsWithAccuracy)/1000.0, 0.3))

	if pp.diff.Mods.Active(difficulty.Hidden) || pp.diff.Mods.Active(difficulty.Traceable) {
		accuracyValue *= 1 + 0.08*putils.ReverseLerp(pp.diff.ARReal, 11.5, 10)
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

	flashlightValue *= pp.getComboScalingFactor()

	// Scale the flashlight value with accuracy _slightly_.
	flashlightValue *= 0.5 + pp.score.Accuracy/2.0

	return flashlightValue
}

func (pp *PPv2) calculateComboBasedEstimatedMissCount(attributes api.Attributes) float64 {
	if attributes.Sliders <= 0 {
		return float64(pp.score.CountMiss)
	}

	missCount := float64(pp.score.CountMiss)

	if pp.usingClassicSliderAccuracy {
		// Consider that full combo is maximum combo minus dropped slider tails since they don't contribute to combo but also don't break it
		// In classic scores we can't know the amount of dropped sliders so we estimate to 10% of all sliders on the map
		fullComboThreshold := float64(attributes.MaxCombo) - 0.1*float64(attributes.Sliders)

		if float64(pp.score.MaxCombo) < fullComboThreshold {
			missCount = fullComboThreshold / max(1.0, float64(pp.score.MaxCombo))
		}
		// In classic scores there can't be more misses than a sum of all non-perfect judgements
		missCount = min(missCount, float64(pp.totalImperfectHits))

		// Every slider has *at least* 2 combo attributed in classic mechanics.
		// If they broke on a slider with a tick, then this still works since they would have lost at least 2 combo (the tick and the end)
		// Using this as a max means a score that loses 1 combo on a map can't possibly have been a slider break.
		// It must have been a slider end.
		maxPossibleSliderBreaks := min(attributes.Sliders, (attributes.MaxCombo-pp.score.MaxCombo)/2)

		sliderBreaks := missCount - float64(pp.score.CountMiss)

		if sliderBreaks > float64(maxPossibleSliderBreaks) {
			missCount = float64(pp.score.CountMiss + maxPossibleSliderBreaks)
		}
	} else {
		fullComboThreshold := float64(attributes.MaxCombo - pp.countSliderEndsDropped)

		if float64(pp.score.MaxCombo) < fullComboThreshold {
			missCount = fullComboThreshold / max(1.0, float64(pp.score.MaxCombo))
		}

		// Combine regular misses with tick misses since tick misses break combo as well
		missCount = min(missCount, float64(pp.score.SliderBreaks+pp.score.CountMiss))
	}

	return missCount
}

func (pp *PPv2) calculateEstimatedSliderBreaks(topWeightedSliderFactor float64, attributes api.Attributes) float64 {
	if !pp.usingClassicSliderAccuracy || pp.score.CountOk == 0 {
		return 0
	}

	missedComboPercent := 1.0 - float64(pp.score.MaxCombo)/float64(attributes.MaxCombo)
	estimatedSliderBreaks := min(float64(pp.score.CountOk), pp.effectiveMissCount*topWeightedSliderFactor)

	// Scores with more Oks are more likely to have slider breaks.
	okAdjustment := ((float64(pp.score.CountOk) - estimatedSliderBreaks) + 0.5) / float64(pp.score.CountOk)

	// There is a low probability of extra slider breaks on effective miss counts close to 1, as score based calculations are good at indicating if only a single break occurred.
	estimatedSliderBreaks *= putils.Smoothstep(pp.effectiveMissCount, 1, 2)

	return estimatedSliderBreaks * okAdjustment * putils.Logistic(missedComboPercent, 0.33, 15, 1)
}

// Estimates player's deviation on speed notes using <see cref="calculateDeviation"/>, assuming worst-case.
// Treats all speed notes as hit circles.
func (pp *PPv2) calculateSpeedDeviation(attributes api.Attributes) *float64 {
	if pp.totalSuccessfulHits == 0 {
		return nil
	}

	// Calculate accuracy assuming the worst case scenario
	speedNoteCount := attributes.SpeedNoteCount
	speedNoteCount += (float64(pp.totalHits) - attributes.SpeedNoteCount) * 0.1

	// Assume worst case: all mistakes were on speed notes
	relevantCountMiss := min(float64(pp.score.CountMiss), speedNoteCount)
	relevantCountMeh := min(float64(pp.score.CountMeh), speedNoteCount-relevantCountMiss)
	relevantCountOk := min(float64(pp.score.CountOk), speedNoteCount-relevantCountMiss-relevantCountMeh)
	relevantCountGreat := max(0, speedNoteCount-relevantCountMiss-relevantCountMeh-relevantCountOk)

	return pp.calculateDeviation(relevantCountGreat, relevantCountOk, relevantCountMeh)
}

// Estimates the player's tap deviation based on the OD, given number of greats, oks, mehs and misses,
// assuming the player's mean hit error is 0. The estimation is consistent in that two SS scores on the same map with the same settings
// will always return the same deviation. Misses are ignored because they are usually due to misaiming.
// Greats and oks are assumed to follow a normal distribution, whereas mehs are assumed to follow a uniform distribution.
func (pp *PPv2) calculateDeviation(relevantCountGreat, relevantCountOk, relevantCountMeh float64) *float64 {
	if relevantCountGreat+relevantCountOk+relevantCountMeh <= 0 {
		return nil
	}

	// The sample proportion of successful hits.
	n := max(1, relevantCountGreat+relevantCountOk)
	p := relevantCountGreat / n

	const z = 2.32634787404 // 99% critical value for the normal distribution (one-tailed).

	// We can be 99% confident that p is at least this value.
	pLowerBound := min(p, (n*p+z*z/2)/(n+z*z)-z/(n+z*z)*math.Sqrt(n*p*(1-p)+z*z/4))

	var deviation float64

	if pLowerBound > 0.01 {
		// Compute deviation assuming greats and oks are normally distributed.
		deviation = pp.greatHitWindow / (math.Sqrt(2) * math.Erfinv(pLowerBound))

		// Subtract the deviation provided by tails that land outside the ok hit window from the deviation computed above.
		// This is equivalent to calculating the deviation of a normal distribution truncated at +-okHitWindow.
		okHitWindowTailAmount := math.Sqrt(2/math.Pi) * pp.okHitWindow * math.Exp(-0.5*math.Pow(pp.okHitWindow/deviation, 2)) /
			(deviation * math.Erf(pp.okHitWindow/(math.Sqrt(2)*deviation)))

		deviation *= math.Sqrt(1 - okHitWindowTailAmount)
	} else {
		// A tested limit value for the case of a score only containing oks.
		deviation = pp.okHitWindow / math.Sqrt(3)
	}

	// Compute and add the variance for mehs, assuming that they are uniformly distributed.
	mehVariance := (pp.mehHitWindow*pp.mehHitWindow + pp.okHitWindow*pp.mehHitWindow + pp.okHitWindow*pp.okHitWindow) / 3

	deviation = math.Sqrt(((relevantCountGreat+relevantCountOk)*math.Pow(deviation, 2) + relevantCountMeh*mehVariance) / (relevantCountGreat + relevantCountOk + relevantCountMeh))

	return &deviation
}

// Calculates multiplier for speed to account for improper tapping based on the deviation and speed difficulty
// https://www.desmos.com/calculator/dmogdhzofn
func (pp *PPv2) calculateSpeedHighDeviationNerf(attributes api.Attributes) float64 {
	if pp.speedDeviation == nil {
		return 0
	}
	speedValue := skills.DefaultDifficultyToPerformance(attributes.Speed)

	// Decides a point where the PP value achieved compared to the speed deviation is assumed to be tapped improperly. Any PP above this point is considered "excess" speed difficulty.
	// This is used to cause PP above the cutoff to scale logarithmically towards the original speed value thus nerfing the value.
	excessSpeedDifficultyCutoff := 100 + 220*math.Pow(22 / *pp.speedDeviation, 6.5)

	if speedValue <= excessSpeedDifficultyCutoff {
		return 1
	}

	const scale = 50.0
	adjustedSpeedValue := scale * (math.Log((speedValue-excessSpeedDifficultyCutoff)/scale+1) + excessSpeedDifficultyCutoff/scale)

	// 220 UR and less are considered tapped correctly to ensure that normal scores will be punished as little as possible
	lerp := 1 - putils.ReverseLerp(*pp.speedDeviation, 22.0, 27.0)
	adjustedSpeedValue = mutils.Lerp(adjustedSpeedValue, speedValue, lerp)

	return adjustedSpeedValue / speedValue
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
