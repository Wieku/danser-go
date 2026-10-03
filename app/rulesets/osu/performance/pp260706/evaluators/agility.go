package evaluators

import (
	"math"

	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
)

const (
	distanceCap float64 = preprocessing.NormalizedDiameter * 1.2
)

func EvaluateAgility(current *preprocessing.DifficultyObject) float64 {
	if current.IsSpinner {
		return 0
	}

	osuCurrObj := current
	osuPrevObj := current.Previous(0)

	travelDistance := 0.0
	if osuPrevObj != nil {
		travelDistance = osuPrevObj.LazyTravelDistance
	}

	distance := travelDistance + osuCurrObj.LazyJumpDistance

	distanceScaled := min(distance, distanceCap) / distanceCap

	agilityDifficulty := distanceScaled * 1000 / osuCurrObj.AdjustedDeltaTime

	agilityDifficulty *= math.Pow(osuCurrObj.SmallCircleBonus, 1.5)

	agilityDifficulty *= agilityHighBpmBonus(osuCurrObj.AdjustedDeltaTime)

	return agilityDifficulty
}

func agilityHighBpmBonus(ms float64) float64 {
	return 1 / (1 - math.Pow(0.2, ms/1000))
}
