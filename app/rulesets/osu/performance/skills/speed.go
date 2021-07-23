package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
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
)

func NewSpeedSkill(useFixedCalculations bool, d *difficulty.Difficulty) *Skill {
	skill := NewSkill(useFixedCalculations, d)
	skill.SkillMultiplier = 1400
	skill.StrainDecayBase = 0.3
	skill.StrainValueOf = speedStrainValue

	return skill
}

func speedStrainValue(_ *Skill, current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	distance := math.Min(SingleSpacingThreshold, current.TravelDistance+current.JumpDistance)
	deltaTime := math.Max(MaxSpeedBonus, current.DeltaTime)

	speedBonus := 1.0
	if deltaTime < MinSpeedBonus {
		speedBonus = 1 + math.Pow((MinSpeedBonus-deltaTime)/SpeedBalancingFactor, 2.0)
	}

	angleBonus := 1.0
	if !math.IsNaN(current.Angle) && current.Angle < SpeedAngleBonusBegin {
		angleBonus = 1 + math.Pow(math.Sin(1.5*(SpeedAngleBonusBegin-current.Angle)), 2)/3.57

		if current.Angle < PiOver2 {
			angleBonus = 1.28
			if distance < AngleBonusScale && current.Angle < PiOver4 {
				angleBonus += (1.0 - angleBonus) *
					math.Min((AngleBonusScale-distance)/10.0, 1.0)
			} else if distance < AngleBonusScale {
				angleBonus += (1.0 - angleBonus) *
					math.Min((AngleBonusScale-distance)/10.0, 1.0) *
					math.Sin((PiOver2-current.Angle)/PiOver4)
			}
		}
	}

	return ((1.0 + (speedBonus-1.0)*0.75) * angleBonus *
		(0.95 + speedBonus*math.Pow(distance/SingleSpacingThreshold, 3.5))) /
		current.StrainTime
}
