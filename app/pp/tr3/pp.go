package tr3

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"math"
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

/* base pp value for stars, used internally by ppv2 */
func ppBase(stars float64) float64 {
	return math.Pow(5.0*math.Max(1.0, stars/0.0675)-4.0, 3.0) /
		100000.0
}

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	Total, Aim, Speed, Acc float64

	aimStrain, speedStrain float64

	maxCombo, nsliders, ncircles, nobjects int

	scoreMaxCombo int
	countGreat    int
	countOk       int
	countMeh      int
	countMiss     int

	diff *difficulty.Difficulty

	totalHits                    int
	accuracy                     float64
	amountHitObjectsWithAccuracy int
}

func (pp *PPv2) PPv2x(aimStars, speedStars float64,
	maxCombo, nsliders, ncircles, nobjects,
	combo, n300, n100, n50, nmiss int, diff *difficulty.Difficulty,
	scoreVersion int) PPv2 {
	maxCombo = bmath.MaxI(1, maxCombo)

	if combo > maxCombo {
		maxCombo = combo
	}

	pp.maxCombo, pp.nsliders, pp.ncircles, pp.nobjects = maxCombo, nsliders, ncircles, nobjects

	if combo < 0 {
		combo = maxCombo
	}

	if n300 < 0 {
		n300 = nobjects - n100 - n50 - nmiss
	}

	totalhits := n300 + n100 + n50 + nmiss

	pp.aimStrain = aimStars
	pp.speedStrain = speedStars
	pp.diff = diff
	pp.totalHits = totalhits
	pp.scoreMaxCombo = combo
	pp.countGreat = n300
	pp.countOk = n100
	pp.countMeh = n50
	pp.countMiss = nmiss

	// accuracy

	if totalhits == 0 {
		pp.accuracy = 0.0
	} else {
		acc := (float64(n50)*50 +
			float64(n100)*100 +
			float64(n300)*300) /
			(float64(totalhits) * 300)

		pp.accuracy = bmath.ClampF64(acc, 0, 1)
	}

	switch scoreVersion {
	case 1:
		pp.amountHitObjectsWithAccuracy = ncircles
	case 2:
		pp.amountHitObjectsWithAccuracy = nobjects
	default:
		panic("unsupported score")
	}

	// total pp

	finalMultiplier := 1.12

	if diff.Mods.Active(difficulty.NoFail) {
		finalMultiplier *= math.Max(0.90, 1.0-0.02*float64(nmiss))
	}

	if totalhits > 0 && diff.Mods.Active(difficulty.SpunOut) {
		nspinners := nobjects - nsliders - ncircles

		finalMultiplier *= 1.0 - math.Pow(float64(nspinners)/float64(totalhits), 0.85)
	}

	aim := pp.computeAimValue()
	speed := pp.computeSpeedValue()
	accuracy := pp.computeAccuracyValue()

	pp.Total = math.Pow(
		math.Pow(aim, 1.1)+math.Pow(speed, 1.1)+
			math.Pow(accuracy, 1.1),
		1.0/1.1) * finalMultiplier

	return *pp
}

func (pp *PPv2) computeAimValue() float64 {
	rawAim := pp.aimStrain

	if pp.diff.Mods.Active(difficulty.TouchDevice) {
		rawAim = math.Pow(rawAim, 0.8)
	}

	aimValue := ppBase(rawAim)

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.countMiss > 0 {
		aimValue *= 0.97 * math.Pow(1-math.Pow(float64(pp.countMiss)/float64(pp.totalHits), 0.775), float64(pp.countMiss))
	}

	// Combo scaling
	if pp.maxCombo > 0 {
		aimValue *= math.Pow((math.Tan(math.Pi / 4 * (2 * (float64(pp.scoreMaxCombo) / float64(pp.maxCombo)) - 1)) + 1) / 2, 0.8)
	}

	//approachRateFactor := 0.0
	//if pp.diff.ARReal > 10.33 {
	//	approachRateFactor += 0.25 * (pp.diff.ARReal - 10.33)
	//} else if pp.diff.ARReal < 8.0 {
	//	approachRateFactor += 0.01 * (8.0 - pp.diff.ARReal)
	//}
	//
	//aimValue *= 1.0 + math.Min(approachRateFactor, approachRateFactor*(float64(pp.totalHits)/1000.0))

	// We want to give more reward for lower AR when it comes to aim and HD. This nerfs high AR and buffs lower AR.
	if pp.diff.Mods.Active(difficulty.Hidden) {
		aimValue *= 1.0 + 0.04*(12.0-pp.diff.ARReal)
	}

	if pp.diff.Mods.Active(difficulty.Flashlight) {
		flBonus := 1.0 + 0.35*math.Min(1.0, float64(pp.totalHits)/200.0)
		if pp.totalHits > 200 {
			flBonus += 0.3 * math.Min(1, (float64(pp.totalHits)-200.0)/300.0)
		}

		if pp.totalHits > 500 {
			flBonus += (float64(pp.totalHits) - 500.0) / 1200.0
		}

		aimValue *= flBonus
	}

	// Scale the aim value with accuracy _slightly_
	aimValue *= 0.5 + pp.accuracy/2.0
	// It is important to also consider accuracy difficulty when doing that
	aimValue *= 0.98 + math.Pow(pp.diff.ODReal, 2)/2500

	return aimValue
}

func (pp *PPv2) computeSpeedValue() float64 {
	speedValue := ppBase(pp.speedStrain)

	approachRateFactor := 0.0
	if pp.diff.ARReal > 10.33 {
		approachRateFactor += 0.225 * (pp.diff.ARReal - 10.33)
	} else if pp.diff.ARReal < 8.0 {
		approachRateFactor += 0.01 * (8.0 - pp.diff.ARReal)
	}

	speedValue *= 1.0 + approachRateFactor

	// Combo scaling
	if pp.maxCombo > 0 {
		speedValue *= math.Pow((math.Tan(math.Pi / 4 * (2 * (float64(pp.scoreMaxCombo) / float64(pp.maxCombo)) - 1)) + 1) / 2, 0.8)
	}

	// Penalize misses by assessing # of misses relative to the total # of objects. Default a 3% reduction for any # of misses.
	if pp.countMiss > 0 {
		speedValue *= 0.97 * math.Pow(1-math.Pow(float64(pp.countMiss)/float64(pp.totalHits), 0.775), float64(pp.countMiss))
	}

	// Scale the speed value with accuracy and OD
	speedValue *= (0.575 + math.Pow(pp.diff.ODReal, 2)/250) * math.Pow(pp.accuracy, (14.5-math.Max(pp.diff.ODReal, 8))/2)
	// Scale the speed value with # of 50s to punish doubletapping.

	mehMult := 0.0
	if float64(pp.countMeh) >= float64(pp.totalHits)/500 {
		mehMult = float64(pp.countMeh) - float64(pp.totalHits)/500.0
	}

	speedValue *= math.Pow(0.98, mehMult)

	return speedValue
}

func (pp *PPv2) computeAccuracyValue() float64 {
	sigmaCircle := 0.0
	sigmaSlider := 0.0
	sigmaTotal := 0.0

	zScore := 2.58
	sqrt2 := math.Sqrt(2.0)
	accMultiplier := 1200.0
	accScale := 1.3
	countHitCircles := pp.ncircles
	countSliders := pp.nsliders

	if countSliders > 0 {
		sliderConst := math.Sqrt(2.0 / float64(countSliders)) * zScore
		sliderProbability := (2.0 * pp.accuracy + math.Pow(sliderConst, 2.0) - sliderConst * math.Sqrt(4.0 * pp.accuracy + math.Pow(sliderConst, 2.0) - 4.0 * math.Pow(pp.accuracy, 2.0))) / (2.0 + 2.0 * math.Pow(sliderConst, 2.0))
		sigmaSlider = (199.5 - 10.0 * pp.diff.ODReal) / (sqrt2 * math.Erfinv(sliderProbability))
	}

	if countHitCircles > 0 {
		circleConst := math.Sqrt(2.0 / float64(countHitCircles)) * zScore
		circleProbability := (2.0 * pp.accuracy + math.Pow(circleConst, 2.0) - circleConst * math.Sqrt(4.0 * pp.accuracy + math.Pow(circleConst, 2.0) - 4.0 * math.Pow(pp.accuracy, 2.0))) / (2.0 + 2.0 * math.Pow(circleConst, 2.0))
		sigmaCircle = (79.5 - 6.0 * pp.diff.ODReal) / (sqrt2 * math.Erfinv(circleProbability))
	}

	if sigmaSlider == 0 {
		return accMultiplier * math.Pow(accScale, -sigmaCircle)
	}

	if sigmaCircle == 0 {
		return accMultiplier * math.Pow(accScale, -sigmaSlider)
	}

	sigmaTotal = 1.0 / (1.0 / sigmaCircle + 1.0 / sigmaSlider)

	accuracyValue := accMultiplier * math.Pow(accScale, -sigmaTotal)

	acc2ndScale := 0.5 + math.Pow(math.Sin(math.Max(0.0, math.Pi * (pp.accuracy - 0.75))), 2)
	accuracyValue *= acc2ndScale

	accuracyValue *= 0.95

	if pp.diff.Mods.Active(difficulty.Hidden) {
		accuracyValue *= 1.08
	}

	if pp.diff.Mods.Active(difficulty.Flashlight) {
		accuracyValue *= 1.02
	}

	return accuracyValue
}
