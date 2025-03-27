package evaluators

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp250306/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	speedSingleSpacingThreshold float64 = preprocessing.NormalizedRadius * 2 * 1.25
	speedMinSpeedBonus          float64 = 200 // 200 BPM 1/4th
	speedBalancingFactor        float64 = 40
	speedDistanceMultiplier     float64 = 0.9
)

func EvaluateSpeed(current *preprocessing.DifficultyObject) float64 {
	if current.IsSpinner {
		return 0
	}

	osuCurrObj := current
	osuPrevObj := current.Previous(0)

	strainTime := osuCurrObj.StrainTime
	doubletapness := 1.0 - osuCurrObj.GetDoubletapness(current.Next(0))

	// Cap deltatime to the OD 300 hitwindow.
	// 0.93 is derived from making sure 260bpm OD8 streams aren't nerfed harshly, whilst 0.92 limits the effect of the cap.
	strainTime /= mutils.Clamp((strainTime/osuCurrObj.GreatWindow)/0.93, 0.92, 1)

	// speedBonus will be 0.0 for BPM < 200
	speedBonus := 0.0

	// Add additional scaling bonus for streams/bursts higher than 200bpm
	if putils.MillisecondsToBPMD(strainTime) > speedMinSpeedBonus {
		speedBonus = 0.75 * math.Pow((putils.BPMToMillisecondsD(speedMinSpeedBonus)-strainTime)/speedBalancingFactor, 2.0)
	}

	var travelDistance float64
	if osuPrevObj != nil {
		travelDistance = osuPrevObj.TravelDistance
	}

	// Cap distance at single_spacing_threshold
	distance := min(speedSingleSpacingThreshold, travelDistance+osuCurrObj.MinimumJumpDistance)

	// Max distance bonus is 1 * `distance_multiplier` at single_spacing_threshold
	distanceBonus := math.Pow(distance/speedSingleSpacingThreshold, 3.95) * speedDistanceMultiplier

	if current.Diff.CheckModActive(difficulty.Relax2) {
		distanceBonus = 0
	}

	// Base difficulty with all bonuses
	difficulty := (1.0 + speedBonus + distanceBonus) * 1000 / strainTime

	// Apply penalty if there's doubletappable doubles
	return difficulty * doubletapness
}
