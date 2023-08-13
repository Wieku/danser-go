package evaluators

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/preprocessing"
	"math"
)

const (
	rhythmMultiplier float64 = 0.75
	historyTimeMax   float64 = 5000
)

func EvaluateRhythm(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	previousIslandSize := 0
	rhythmComplexitySum := 0.0
	islandSize := 1
	startRatio := 0.0 // store the ratio of the current start of an island to buff for tighter rhythms

	firstDeltaSwitch := false

	historicalNoteCount := min(current.Index, 32)

	rhythmStart := 0

	for rhythmStart < historicalNoteCount-2 && current.StartTime-current.Previous(rhythmStart).StartTime < historyTimeMax {
		rhythmStart++
	}

	for i := rhythmStart; i > 0; i-- {
		currObj := current.Previous(i - 1)
		prevObj := current.Previous(i)
		lastObj := current.Previous(i + 1)

		currHistoricalDecay := (historyTimeMax - (current.StartTime - currObj.StartTime)) / historyTimeMax // scales note 0 to 1 from history to now

		if currHistoricalDecay != 0 {
			currHistoricalDecay = min(float64(historicalNoteCount-i)/float64(historicalNoteCount), currHistoricalDecay) // either we're limited by time or limited by object count.

			currDelta := currObj.StrainTime
			prevDelta := prevObj.StrainTime
			lastDelta := lastObj.StrainTime
			currRatio := 1.0 + 6.0*min(0.5, math.Pow(math.Sin(math.Pi/(min(prevDelta, currDelta)/max(prevDelta, currDelta))), 2)) // fancy function to calculate rhythmbonuses.

			windowPenalty := min(1, max(0, math.Abs(prevDelta-currDelta)-currObj.GreatWindow*0.3)/(currObj.GreatWindow*0.3))

			windowPenalty = min(1, windowPenalty)

			effectiveRatio := windowPenalty * currRatio

			if firstDeltaSwitch {
				if !(prevDelta > 1.25*currDelta || prevDelta*1.25 < currDelta) {
					if islandSize < 7 {
						islandSize++ // island is still progressing, count size.
					}
				} else {
					if _, ok := currObj.BaseObject.(*preprocessing.LazySlider); ok { // bpm change is into slider, this is easy acc window
						effectiveRatio *= 0.125
					}

					if _, ok := prevObj.BaseObject.(*preprocessing.LazySlider); ok { // bpm change was from a slider, this is easier typically than circle -> circle
						effectiveRatio *= 0.25
					}

					if previousIslandSize == islandSize { // repeated island size (ex: triplet -> triplet)
						effectiveRatio *= 0.25
					}

					if previousIslandSize%2 == islandSize%2 { // repeated island polarity (2 -> 4, 3 -> 5)
						effectiveRatio *= 0.50
					}

					if lastDelta > prevDelta+10 && prevDelta > currDelta+10 { // previous increase happened a note ago, 1/1->1/2-1/4, dont want to buff this.
						effectiveRatio *= 0.125
					}

					rhythmComplexitySum += math.Sqrt(effectiveRatio*startRatio) * currHistoricalDecay * math.Sqrt(4+float64(islandSize)) / 2 * math.Sqrt(4+float64(previousIslandSize)) / 2

					startRatio = effectiveRatio

					previousIslandSize = islandSize // log the last island size.

					if prevDelta*1.25 < currDelta { // we're slowing down, stop counting
						firstDeltaSwitch = false // if we're speeding up, this stays true and we keep counting island size.
					}

					islandSize = 1
				}
			} else if prevDelta > 1.25*currDelta { // we want to be speeding up.
				// Begin counting island until we change speed again.
				firstDeltaSwitch = true
				startRatio = effectiveRatio
				islandSize = 1
			}
		}
	}

	return math.Sqrt(4+rhythmComplexitySum*rhythmMultiplier) / 2 //produces multiplier that can be applied to strain. range [1, infinity) (not really though)
}
