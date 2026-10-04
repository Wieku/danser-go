package skills

import (
	"math"
	"slices"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
)

const (
	aimSkillMultiplier float64 = 26
	aimStrainDecayBase float64 = 0.15
)

type AimSkill struct {
	*Skill
	withSliders   bool
	currentStrain float64

	diffSliders *putils.LogisticSum
	topSliders  *putils.LogisticSum
}

func NewAimSkill(d *difficulty.Difficulty, withSliders, stepCalc bool) *AimSkill {
	skill := &AimSkill{
		Skill:       NewSkill(d, stepCalc),
		withSliders: withSliders,
	}

	skill.StrainValueOf = skill.aimStrainValue
	skill.PostProcess = skill.postProcess
	skill.CalculateInitialStrain = skill.aimInitialStrain

	skill.diffSliders = putils.NewLogisticSum(stepCalc, 6, 1, 1, func(previous, current float64) bool {
		return current > previous
	}, func(strains []float64) float64 {
		return slices.Max(strains) / 12
	})

	skill.topSliders = putils.NewLogisticSum(stepCalc, 0.88, 10, 1.1, func(previous, current float64) bool {
		return current != previous
	}, func(strains []float64) float64 {
		return skill.DifficultyValue() / 10
	})

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

	if current.IsSlider {
		skill.diffSliders.AddStrain(skill.currentStrain)
		skill.topSliders.AddStrain(skill.currentStrain)
	}

	return skill.currentStrain
}

func (skill *AimSkill) postProcess(current *preprocessing.DifficultyObject, strain float64, diffValue float64) {
	if current.IsSlider {
		skill.diffSliders.ProcessLastStrain(strain / 12)
	}

	skill.topSliders.ProcessLastStrain(diffValue / 10)
}

func (skill *AimSkill) GetDifficultSliders() float64 {
	return skill.diffSliders.GetValue()
}

func (skill *AimSkill) CountTopWeightedSliders() float64 {
	return skill.topSliders.GetValue()
}
