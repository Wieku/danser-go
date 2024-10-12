package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/preprocessing"
	"math"
)

const (
	aimSkillMultiplier float64 = 25.18
	aimStrainDecayBase float64 = 0.15
)

type AimSkill struct {
	*Skill
	withSliders   bool
	currentStrain float64
}

func NewAimSkill(d *difficulty.Difficulty, withSliders, stepCalc bool) *AimSkill {
	skill := &AimSkill{Skill: NewSkill(d, stepCalc), withSliders: withSliders}

	skill.StrainValueOf = skill.aimStrainValue
	skill.CalculateInitialStrain = skill.aimInitialStrain

	return skill
}

func (skill *AimSkill) strainDecay(ms float64) float64 {
	return math.Pow(aimStrainDecayBase, ms/1000)
}

func (skill *AimSkill) aimInitialStrain(time float64, current *preprocessing.DifficultyObject) float64 {
	return skill.currentStrain * skill.strainDecay(time-current.Previous(0).StartTime)
}

func (skill *AimSkill) aimStrainValue(current *preprocessing.DifficultyObject) float64 {
	skill.currentStrain *= skill.strainDecay(current.DeltaTime)
	skill.currentStrain += evaluators.EvaluateAim(current, skill.withSliders) * aimSkillMultiplier

	skill.objectStrains = append(skill.objectStrains, skill.currentStrain)

	return skill.currentStrain
}
