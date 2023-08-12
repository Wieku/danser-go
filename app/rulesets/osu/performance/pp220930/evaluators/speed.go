package evaluators

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	speedSingleSpacingThreshold float64 = 125.0
	speedMinSpeedBonus          float64 = 75.0 // ~200BPM
	speedBalancingFactor        float64 = 40
)

func EvaluateSpeed(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	osuCurrObj := current
	osuPrevObj := current.Previous(0)
	osuNextObj := current.Next(0)

	strainTime := osuCurrObj.StrainTime
	doubletapness := 1.0

	if osuNextObj != nil {
		currDeltaTime := max(1, osuCurrObj.DeltaTime)
		nextDeltaTime := max(1, osuNextObj.DeltaTime)
		deltaDifference := math.Abs(nextDeltaTime - currDeltaTime)
		speedRatio := currDeltaTime / max(currDeltaTime, deltaDifference)
		windowRatio := math.Pow(min(1, currDeltaTime/osuCurrObj.GreatWindow), 2)
		doubletapness = math.Pow(speedRatio, 1-windowRatio)
	}

	// Cap deltatime to the OD 300 hitwindow.
	// 0.93 is derived from making sure 260bpm OD8 streams aren't nerfed harshly, whilst 0.92 limits the effect of the cap.
	strainTime /= mutils.Clamp((strainTime/osuCurrObj.GreatWindow)/0.93, 0.92, 1)

	// derive speedBonus for calculation
	speedBonus := 1.0

	if strainTime < speedMinSpeedBonus {
		speedBonus = 1 + 0.75*math.Pow((speedMinSpeedBonus-strainTime)/speedBalancingFactor, 2.0)
	}

	var travelDistance float64
	if osuPrevObj != nil {
		travelDistance = osuPrevObj.TravelDistance
	}

	distance := min(speedSingleSpacingThreshold, travelDistance+osuCurrObj.MinimumJumpDistance)

	return (speedBonus + speedBonus*math.Pow(distance/speedSingleSpacingThreshold, 3.5)) * doubletapness / strainTime
}
