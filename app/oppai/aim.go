package oppai

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
)

const (
	AngleBonusScale    float64 = 90.0
	AimTimingThreshold float64 = 107.0
	AimAngleBonusBegin float64 = math.Pi / 3
)

func NewAimSkill(useFixedCalculations bool, d *difficulty.Difficulty) *Skill {
	skill := NewSkill(useFixedCalculations, d)
	skill.SkillMultiplier = 26.25
	skill.StrainDecayBase = 0.15
	skill.StrainValueOf = aimStrainValue

	return skill
}

func aimStrainValue(skill *Skill, current *DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	result := 0.0

	if len(skill.Previous) > 0 {
		previous := skill.GetPrevious()

		if !math.IsNaN(current.Angle) && current.Angle > AimAngleBonusBegin {
			angleBonus := math.Sqrt(
				math.Max(previous.JumpDistance-AngleBonusScale, 0.0) *
					math.Pow(math.Sin(current.Angle-AimAngleBonusBegin), 2.0) *
					math.Max(current.JumpDistance-AngleBonusScale, 0.0))

			result = 1.5 * applyDiminishingExp(math.Max(0, angleBonus)) / math.Max(AimTimingThreshold, previous.StrainTime)
		}
	}

	jumpDistanceExp := applyDiminishingExp(current.JumpDistance)
	travelDistanceExp := applyDiminishingExp(current.TravelDistance)

	return math.Max(
		result+(jumpDistanceExp+travelDistanceExp+math.Sqrt(travelDistanceExp*jumpDistanceExp))/math.Max(current.StrainTime, AimTimingThreshold),
		(math.Sqrt(travelDistanceExp*jumpDistanceExp)+jumpDistanceExp+travelDistanceExp)/current.StrainTime)
}

func applyDiminishingExp(val float64) float64 {
	return math.Pow(val, 0.99)
}
