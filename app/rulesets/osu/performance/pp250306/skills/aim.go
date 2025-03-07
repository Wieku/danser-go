package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp250306/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp250306/preprocessing"
	"math"
)

const (
	aimSkillMultiplier float64 = 25.6
	aimStrainDecayBase float64 = 0.15
)

type AimSkill struct {
	*Skill
	withSliders   bool
	currentStrain float64

	maxStrain     float64
	sliderStrains []float64

	difficultSlidersV float64
}

func NewAimSkill(d *difficulty.Difficulty, withSliders, stepCalc bool) *AimSkill {
	skill := &AimSkill{Skill: NewSkill(d, stepCalc), withSliders: withSliders, sliderStrains: make([]float64, 0)}

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

	if current.IsSlider {
		skill.sliderStrains = append(skill.sliderStrains, skill.currentStrain)

		if !skill.stepCalc { // Don't need to precalculate difficult sliders for normal strain calc
			return skill.currentStrain
		}

		if skill.currentStrain > skill.maxStrain {
			skill.maxStrain = skill.currentStrain
			skill.difficultSlidersV = skill.getDifficultSliders()
		} else if skill.maxStrain != 0 {
			skill.difficultSlidersV += 1.0 / (1.0 + math.Exp(-(skill.currentStrain/skill.maxStrain*12.0 - 6.0)))
		}
	}

	return skill.currentStrain
}

func (skill *AimSkill) getDifficultSliders() (sum float64) {
	if len(skill.sliderStrains) == 0 {
		return
	}

	if skill.maxStrain == 0 {
		return
	}

	for _, strain := range skill.sliderStrains {
		sum += 1.0 / (1.0 + math.Exp(-(strain/skill.maxStrain*12.0 - 6.0)))
	}

	return
}

func (skill *AimSkill) GetDifficultSliders() float64 {
	if skill.stepCalc {
		return skill.difficultSlidersV
	}

	return skill.getDifficultSliders()
}
