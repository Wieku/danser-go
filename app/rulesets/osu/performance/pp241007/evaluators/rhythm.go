package evaluators

import (
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
	"slices"
)

const (
	rhythmHistoryTimeMax    = 5000.0
	rhythmHistoryObjectsMax = 32
	rhythmMultiplier        = 0.95
	rhythmRatioMultiplier   = 12.0
	rhythmMinDeltaTime      = 25
)

func EvaluateRhythm(current *preprocessing.DifficultyObject) float64 {
	if current.IsSpinner {
		return 0
	}

	rhythmComplexitySum := 0.0

	deltaDifferenceEpsilon := current.GreatWindow * 0.3

	island := newIsland(deltaDifferenceEpsilon)
	previousIsland := newIsland(deltaDifferenceEpsilon)

	// we can't use dictionary here because we need to compare island with a tolerance
	// which is impossible to pass into the hash comparer
	islandCounts := make([]*pair, 0)

	startRatio := 0.0 // store the ratio of the current start of an island to buff for tighter rhythms

	firstDeltaSwitch := false

	historicalNoteCount := min(current.Index, rhythmHistoryObjectsMax)

	rhythmStart := 0

	for rhythmStart < historicalNoteCount-2 && current.StartTime-current.Previous(rhythmStart).StartTime < rhythmHistoryTimeMax {
		rhythmStart++
	}

	prevObj := current.Previous(rhythmStart)
	lastObj := current.Previous(rhythmStart + 1)

	for i := rhythmStart; i > 0; i-- {
		currObj := current.Previous(i - 1)

		// scales note 0 to 1 from history to now
		timeDecay := (rhythmHistoryTimeMax - (current.StartTime - currObj.StartTime)) / rhythmHistoryTimeMax
		noteDecay := float64(historicalNoteCount-i) / float64(historicalNoteCount)

		currHistoricalDecay := min(noteDecay, timeDecay) // either we're limited by time or limited by object count.

		currDelta := currObj.StrainTime
		prevDelta := prevObj.StrainTime
		lastDelta := lastObj.StrainTime

		// calculate how much current delta difference deserves a rhythm bonus
		// this function is meant to reduce rhythm bonus for deltas that are multiples of each other (i.e 100 and 200)
		deltaDifferenceRatio := min(prevDelta, currDelta) / max(prevDelta, currDelta)
		currRatio := 1.0 + rhythmRatioMultiplier*min(0.5, math.Pow(math.Sin(math.Pi/deltaDifferenceRatio), 2))

		// reduce ratio bonus if delta difference is too big
		fraction := max(prevDelta/currDelta, currDelta/prevDelta)
		fractionMultiplier := mutils.Clamp(2.0-fraction/8.0, 0.0, 1.0)

		windowPenalty := min(1, max(0, math.Abs(prevDelta-currDelta)-deltaDifferenceEpsilon)/deltaDifferenceEpsilon)

		effectiveRatio := windowPenalty * currRatio * fractionMultiplier

		if firstDeltaSwitch {
			if math.Abs(prevDelta-currDelta) < deltaDifferenceEpsilon {
				island.addDelta(int(currDelta))
			} else {
				// bpm change is into slider, this is easy acc window
				if currObj.IsSlider {
					effectiveRatio *= 0.125
				}

				// bpm change was from a slider, this is easier typically than circle -> circle
				// unintentional side effect is that bursts with kicksliders at the ends might have lower difficulty than bursts without sliders
				if prevObj.IsSlider {
					effectiveRatio *= 0.3
				}

				// repeated island polarity (2 -> 4, 3 -> 5)
				if island.isSimilarPolarity(previousIsland) {
					effectiveRatio *= 0.5
				}

				// previous increase happened a note ago, 1/1->1/2-1/4, dont want to buff this.
				if lastDelta > prevDelta+deltaDifferenceEpsilon && prevDelta > currDelta+deltaDifferenceEpsilon {
					effectiveRatio *= 0.125
				}

				// repeated island size (ex: triplet -> triplet)
				// TODO: remove this nerf since its staying here only for balancing purposes because of the flawed ratio calculation
				if previousIsland.deltaCount == island.deltaCount {
					effectiveRatio *= 0.5
				}

				if countIndex := slices.IndexFunc(islandCounts, func(p *pair) bool { return p.island.equals(island) }); countIndex != -1 {
					islandCount := islandCounts[countIndex]

					// only add island to island counts if they're going one after another
					if previousIsland.equals(island) {
						islandCount.count++
					}

					// repeated island (ex: triplet -> triplet)
					power := logistic(float64(island.delta), 2.75, 0.24, 14)
					effectiveRatio *= min(3.0/float64(islandCount.count), math.Pow(1.0/float64(islandCount.count), power))

					//islandCounts[countIndex] = (islandCount.Island, islandCount.Count);
				} else {
					islandCounts = append(islandCounts, newPair(island, 1))
				}

				// scale down the difficulty if the object is doubletappable
				doubletapness := prevObj.GetDoubletapness(currObj)
				effectiveRatio *= 1 - doubletapness*0.75

				rhythmComplexitySum += math.Sqrt(effectiveRatio*startRatio) * currHistoricalDecay

				startRatio = effectiveRatio

				previousIsland = island

				if prevDelta+deltaDifferenceEpsilon < currDelta { // we're slowing down, stop counting
					firstDeltaSwitch = false // if we're speeding up, this stays true and we keep counting island size.
				}

				island = newIslandD(int(currDelta), deltaDifferenceEpsilon)
			}
		} else if prevDelta > currDelta+deltaDifferenceEpsilon {
			// Begin counting island until we change speed again.
			firstDeltaSwitch = true

			// bpm change is into slider, this is easy acc window
			if currObj.IsSlider {
				effectiveRatio *= 0.6
			}

			// bpm change was from a slider, this is easier typically than circle -> circle
			// unintentional side effect is that bursts with kicksliders at the ends might have lower difficulty than bursts without sliders
			if prevObj.IsSlider {
				effectiveRatio *= 0.6
			}

			startRatio = effectiveRatio

			island = newIslandD(int(currDelta), deltaDifferenceEpsilon)
		}

		lastObj = prevObj
		prevObj = currObj
	}

	return math.Sqrt(4+rhythmComplexitySum*rhythmMultiplier) / 2 //produces multiplier that can be applied to strain. range [1, infinity) (not really though)
}

type pair struct {
	island *Island
	count  int
}

func newPair(island *Island, count int) *pair {
	return &pair{island: island, count: count}
}

type Island struct {
	deltaDifferenceEpsilon float64
	delta                  int
	deltaCount             int
}

func newIsland(epsilon float64) *Island {
	return &Island{
		deltaDifferenceEpsilon: epsilon,
		delta:                  math.MaxInt,
	}
}

func newIslandD(delta int, epsilon float64) *Island {
	return &Island{
		deltaDifferenceEpsilon: epsilon,
		delta:                  max(delta, rhythmMinDeltaTime),
		deltaCount:             1,
	}
}

func (island *Island) addDelta(delta int) {
	if island.delta == math.MaxInt {
		island.delta = max(delta, rhythmMinDeltaTime)
	}

	island.deltaCount++
}

func (island *Island) isSimilarPolarity(other *Island) bool {
	return island.deltaCount%2 == other.deltaCount%2
}

func (island *Island) equals(other *Island) bool {
	if other == nil {
		return false
	}

	return math.Abs(float64(island.delta-other.delta)) < island.deltaDifferenceEpsilon && island.deltaCount == other.deltaCount
}

func logistic(x, maxValue, multiplier, offset float64) float64 {
	return maxValue / (1 + math.Pow(math.E, offset-(multiplier*x)))
}
