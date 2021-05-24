package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/xexxar/preprocessing"
	"math"
)

const (
	SingleSpacingThreshold float64 = 125.0
	MinSpeedBonus          float64 = 75.0
	MaxSpeedBonus          float64 = 45.0
	SpeedBalancingFactor   float64 = 40
	SpeedAngleBonusBegin   float64 = 5 * math.Pi / 6

	PiOver2 float64 = math.Pi / 2
	PiOver4 float64 = math.Pi / 4

	averageLength int = 2
	tapStrainMultiplier float64 = 2.65
)

func NewSpeedSkill(d *difficulty.Difficulty) *Skill {
	skill := NewSkill(d)
	skill.StarsPerDouble = 1.075
	skill.HistoryLength = 16
	skill.StrainValueOf = speedStrainValue

	return skill
}

func isRatioEqual(ratio, a, b float64) bool {
	return a + 15 > ratio * b && a - 15 < ratio * b
}

func speedStrainValue(skill *Skill, current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok || len(skill.Previous) == 0 {
		return 0
	}

	strainValue := 0.25

	sumDeltaTime := 0.0

	if len(skill.Previous) < 8 {
		return 0
	}

	for i := 0; i < len(skill.Previous); i++ {
		if i < averageLength {
			sumDeltaTime += skill.GetPrevious(i).StrainTime
		}
	}

	avgDeltaTime := sumDeltaTime / math.Min(float64(len(skill.Previous)), float64(averageLength))

	// {doubles, triplets, quads, quints, 6-tuplets, 7 Tuplets, greater}
	islandSizes := []float64{0, 0, 0, 0, 0, 0, 0}
	islandTimes := []float64{0, 0, 0, 0, 0, 0, 0}

	islandSize := 0

	specialTransitionCount := 0.0

	firstDeltaSwitch := false

	for i := 1; i < len(skill.Previous); i++ {
		prevDelta := skill.GetPrevious(i - 1).StrainTime
		currDelta := skill.GetPrevious(i).StrainTime

		if isRatioEqual(1.5, prevDelta, currDelta) || isRatioEqual(1.5, currDelta, prevDelta) {
			_, ok := skill.GetPrevious(i-1).BaseObject.(*preprocessing.LazySlider)
			_, ok1 := skill.GetPrevious(i).BaseObject.(*preprocessing.LazySlider)

			if ok || ok1 {
				specialTransitionCount += 50.0 / math.Sqrt(prevDelta*currDelta) * (float64(i) / float64(skill.HistoryLength))
			} else {
				specialTransitionCount += 250.0 / math.Sqrt(prevDelta*currDelta) * (float64(i) / float64(skill.HistoryLength))
			}
		}

		if firstDeltaSwitch {
			if isRatioEqual(1.0, prevDelta, currDelta) {
				islandSize++ // island is still progressing, count size.
			} else if prevDelta > currDelta*1.25 { // we're speeding up
				if islandSize > 6 {
					islandTimes[6] = islandTimes[6] + 100.0/math.Sqrt(prevDelta*currDelta)*(float64(i) / float64(skill.HistoryLength))
					islandSizes[6] = islandSizes[6] + 1
				} else {
					islandTimes[islandSize] = islandTimes[islandSize] + 100.0/math.Sqrt(prevDelta*currDelta)*(float64(i) / float64(skill.HistoryLength))
					islandSizes[islandSize] = islandSizes[islandSize] + 1
				}

				islandSize = 0 // reset and count again, we sped up (usually this could only be if we did a 1/2 -> 1/3 -> 1/4) (or 1/1 -> 1/2 -> 1/4)
			} else { // we're not the same or speeding up, must be slowing down.
				if islandSize > 6 {
					islandTimes[6] = islandTimes[6] + 100.0/math.Sqrt(prevDelta*currDelta)*(float64(i) / float64(skill.HistoryLength))
					islandSizes[6] = islandSizes[6] + 1
				} else {
					islandTimes[islandSize] = islandTimes[islandSize] + 100.0/math.Sqrt(prevDelta*currDelta)*(float64(i) / float64(skill.HistoryLength))
					islandSizes[islandSize] = islandSizes[islandSize] + 1
				}

				firstDeltaSwitch = false // stop counting island until next speed up.
			}
		} else if prevDelta > 1.25*currDelta { // we want to be speeding up.
			// Begin counting island until we slow again.
			firstDeltaSwitch = true
			islandSize = 0
		}
	}

	rhythmComplexitySum := 0.0

	for i := 0; i < len(islandSizes); i++{
		if islandSizes[i] != 0 {
			rhythmComplexitySum += islandTimes[i] / math.Pow(islandSizes[i], .5)
		} // sum the total amount of rhythm variance, penalizing for repeated island sizes.
	}

	sliderCount := 1

	for i := 0; i < len(skill.Previous); i++ {
		_, ok := skill.GetPrevious(i).BaseObject.(*preprocessing.LazySlider)

		if ok {
			sliderCount++
		}
	}

	rhythmComplexitySum += specialTransitionCount // add in our 1.5 * transitions
	rhythmComplexitySum *= .75

	if 75/avgDeltaTime > 1 { // scale tap value for high BPM.
		strainValue += math.Pow(75/avgDeltaTime, 2)
	} else {
		strainValue += math.Pow(75/avgDeltaTime, 1)
	}

	skill.currentStrain *= computeDecay(.9, current.StrainTime)
	skill.currentStrain += strainValue * tapStrainMultiplier

	return skill.currentStrain * (float64(len(skill.Previous)) / float64(skill.HistoryLength)) * (math.Sqrt(4+rhythmComplexitySum) / 2)
}
