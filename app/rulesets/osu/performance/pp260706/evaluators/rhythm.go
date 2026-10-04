package evaluators

import (
	"math"

	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	rhythmHistoryTimeMax    = 5000.0
	rhythmHistoryObjectsMax = 32
	rhythmOverallMultiplier = 0.95
)

func EvaluateRhythm(current *preprocessing.DifficultyObject) float64 {
	if current.IsSpinner {
		return 0
	}

	rhythmComplexitySum := 0.0

	deltaDifferenceEpsilon := current.GreatWindow * 0.3

	island := newIslandD(math.MaxInt)
	previousIsland := newIslandD(math.MaxInt)

	islands := make([]*Island, 0)

	startDifficulty := 0.0 // store the difficulty of the current start of an island to buff for tighter rhythms

	firstDeltaSwitch := false

	historicalNoteCount := min(current.Index, rhythmHistoryObjectsMax)

	rhythmStart := 0

	for rhythmStart < historicalNoteCount-2 && current.StartTime-current.Previous(rhythmStart).StartTime < rhythmHistoryTimeMax {
		rhythmStart++
	}

	prevObj := current.Previous(rhythmStart)
	prevPrevObj := current.Previous(rhythmStart + 1)

	// we go from the furthest object back to the current one
	for i := rhythmStart; i > 0; i-- {
		currObj := current.Previous(i - 1)
		if currObj.IsSpinner {
			continue
		}

		// scales note 0 to 1 from history to now
		timeDecay := (rhythmHistoryTimeMax - (current.StartTime - currObj.StartTime)) / rhythmHistoryTimeMax
		noteDecay := float64(historicalNoteCount-i) / float64(historicalNoteCount)

		currHistoricalDecay := min(noteDecay, timeDecay) // either we're limited by time or limited by object count.

		// Use custom cap value to ensure that at this point delta time is actually zero
		const deltaMinValue = 1e-7

		currDelta := max(currObj.DeltaTime, deltaMinValue)
		prevDelta := max(prevObj.DeltaTime, deltaMinValue)

		deltaDifference := math.Abs(prevDelta - currDelta)

		// Make sure to always have the current island initialised - if we don't do it here it will only initialise on the next rhythm change
		if island.delta == math.MaxInt {
			island = newIslandD(int(currDelta))
		}

		// calculate how much current delta difference deserves a rhythm bonus
		// this function is meant to reduce rhythm bonus for deltas that are multiples of each other (i.e 100 and 200)
		deltaDifferenceRatio := max(prevDelta, currDelta) / min(prevDelta, currDelta)

		// reduce ratio bonus if delta difference is too big
		differenceMultiplier := mutils.Clamp(2.0-deltaDifferenceRatio/8.0, 0.0, 1.0)

		windowPenalty := mutils.Clamp((deltaDifference-deltaDifferenceEpsilon)/deltaDifferenceEpsilon, 0, 1)

		effectiveDifficulty := getEffectiveDifficulty(deltaDifferenceRatio) * windowPenalty * differenceMultiplier

		// if previous object is a slider it might be easier to tap since you don't have to do a whole tapping motion
		// while a full deltatime might end up some weird ratio the "unpress->tap" motion might be simple
		// for example a slider-circle-circle pattern should be evaluated as a regular triple and not as a single->double
		if prevObj.IsSlider {
			sliderLazyEndDelta := currObj.MinimumJumpTime
			sliderLazyDeltaDifferenceRatio := max(sliderLazyEndDelta, currDelta) / min(sliderLazyEndDelta, currDelta)

			sliderRealEndDelta := currObj.LastObjectEndDeltaTime
			sliderRealDeltaDifferenceRatio := max(sliderRealEndDelta, currDelta) / min(sliderRealEndDelta, currDelta)

			sliderEffectiveDifficulty := min(getEffectiveDifficulty(sliderLazyDeltaDifferenceRatio), getEffectiveDifficulty(sliderRealDeltaDifferenceRatio))
			effectiveDifficulty = min(sliderEffectiveDifficulty, effectiveDifficulty)
		}

		if deltaDifference < deltaDifferenceEpsilon {
			// island is still progressing
			island.addDelta(int(currDelta))
		}

		if firstDeltaSwitch {
			if deltaDifference > deltaDifferenceEpsilon {
				// bpm change is into slider, this is easy acc window
				if currObj.IsSlider {
					effectiveDifficulty *= 0.5
				}

				// repeated island polarity (2 -> 4, 3 -> 5)
				if island.isSimilarPolarity(previousIsland, deltaDifferenceEpsilon) {
					effectiveDifficulty *= 0.5
				}

				// previous increase happened a note ago, 1/1->1/2-1/4, dont want to buff this.
				if max(prevPrevObj.DeltaTime, deltaMinValue) > prevDelta+deltaDifferenceEpsilon && prevDelta > currDelta+deltaDifferenceEpsilon {
					effectiveDifficulty *= 0.125
				}

				// repeated island size (ex: triplet -> triplet)
				// TODO: remove this nerf since its staying here only for balancing purposes because of the flawed ratio calculation
				if previousIsland.deltaCount == island.deltaCount {
					effectiveDifficulty *= 0.5
				}

				isSpeedingUp := prevDelta > currDelta+deltaDifferenceEpsilon

				if isSpeedingUp {
					effectiveDifficulty *= 0.65
				}

				found := false

				for _, existingIsland := range islands {
					if existingIsland.almostEquals(island, deltaDifferenceEpsilon) {
						// only increase island occurrences if they're going one after another
						if previousIsland.almostEquals(island, deltaDifferenceEpsilon) {
							existingIsland.occurrences++
						}

						// repeated island (ex: triplet -> triplet)
						power := putils.Logistic(float64(island.delta), 58.33, 0.24, 2.75)
						effectiveDifficulty *= min(3.0/float64(existingIsland.occurrences), math.Pow(1.0/float64(existingIsland.occurrences), power))

						found = true
						break
					}
				}

				if !found && island.deltaCount > 0 {
					islands = append(islands, island)
				}

				// scale down the difficulty if the object is double-tappable
				effectiveDifficulty *= 1 - prevObj.CalculateDoubleTapFeasibility(currObj)*0.75

				if island.deltaCount > 1 {
					rhythmComplexitySum += math.Sqrt(effectiveDifficulty*startDifficulty) * currHistoricalDecay
				} else {
					// constant difficulty for single-note islands
					rhythmComplexitySum += 0.7 * currHistoricalDecay
				}

				startDifficulty = effectiveDifficulty

				if prevDelta+deltaDifferenceEpsilon < currDelta { // we're slowing down, stop counting
					firstDeltaSwitch = false // if we're speeding up, this stays true and we keep counting island size.
				}

				previousIsland = island
				island = newIslandD(int(currDelta))
			}
		} else if prevDelta > currDelta+deltaDifferenceEpsilon { // we're speeding up
			// Begin counting island until we change speed again.
			firstDeltaSwitch = true

			// bpm change is into slider, this is easy acc window
			if currObj.IsSlider {
				effectiveDifficulty *= 0.6
			}

			// bpm change was from a slider, this is easier typically than circle -> circle
			// unintentional side effect is that bursts with kicksliders at the ends might have lower difficulty than bursts without sliders
			if prevObj.IsSlider {
				effectiveDifficulty *= 0.6
			}

			startDifficulty = effectiveDifficulty

			island = newIslandD(int(currDelta))
		}

		prevPrevObj = prevObj
		prevObj = currObj
	}

	// If the current island is long we don't want the sum to have as big of an effect
	rhythmComplexitySum *= putils.ReverseLerp(float64(island.deltaCount), 22, 3)

	return math.Sqrt(4+rhythmComplexitySum*rhythmOverallMultiplier) / 2.0 // produces multiplier that can be applied to strain. range [1, infinity) (not really though)
}

func getEffectiveDifficulty(deltaDifferenceRatio float64) float64 {
	const rhythmRatioDifficultyMultiplier = 26.0

	// Take only the fractional part of the value since we're only interested in punishing multiples
	deltaDifferenceFraction := deltaDifferenceRatio - math.Trunc(deltaDifferenceRatio)

	return 1.0 + rhythmRatioDifficultyMultiplier*min(0.5, putils.SmoothstepBellCurvex(deltaDifferenceFraction))
}

type pair struct {
	island *Island
	count  int
}

func newPair(island *Island, count int) *pair {
	return &pair{island: island, count: count}
}

type Island struct {
	//deltaDifferenceEpsilon float64
	delta       int
	deltaCount  int
	occurrences int
}

//func newIsland(epsilon float64) *Island {
//	return &Island{
//		deltaDifferenceEpsilon: epsilon,
//		delta:                  maxInt,
//	}
//}

func newIslandD(delta int) *Island {
	return &Island{
		delta:       max(delta, preprocessing.MinDeltaTime),
		deltaCount:  1,
		occurrences: 1,
	}
}

func (island *Island) addDelta(delta int) {
	if island.delta == math.MaxInt {
		island.delta = max(delta, preprocessing.MinDeltaTime)
	}

	island.deltaCount++
}

func (island *Island) isSimilarPolarity(other *Island, epsilon float64) bool {
	if island.deltaCount <= 1 || other.deltaCount <= 1 {
		return false
	}

	return math.Abs(float64(island.delta-other.delta)) < epsilon &&
		island.deltaCount%2 == other.deltaCount%2
}

func (island *Island) almostEquals(other *Island, epsilon float64) bool {
	return math.Abs(float64(island.delta-other.delta)) < epsilon && island.deltaCount == other.deltaCount
}
