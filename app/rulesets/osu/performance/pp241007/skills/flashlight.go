package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/preprocessing"
	"math"
)

const (
	flSkillMultiplier float64 = 0.05512
	flStrainDecayBase float64 = 0.15
)

type Flashlight struct {
	*Skill
	currentStrain float64
}

func NewFlashlightSkill(d *difficulty.Difficulty) *Flashlight {
	skill := &Flashlight{Skill: NewSkill(d, false)}

	skill.StrainValueOf = skill.flashlightStrainValue
	skill.CalculateInitialStrain = skill.flInitialStrain

	return skill
}

func (s *Flashlight) strainDecay(ms float64) float64 {
	return math.Pow(flStrainDecayBase, ms/1000)
}

func (s *Flashlight) flInitialStrain(time float64, current *preprocessing.DifficultyObject) float64 {
	return s.currentStrain * s.strainDecay(time-current.Previous(0).StartTime)
}

func (s *Flashlight) flashlightStrainValue(current *preprocessing.DifficultyObject) float64 {
	s.currentStrain *= s.strainDecay(current.DeltaTime)
	s.currentStrain += evaluators.EvaluateFlashlight(current) * flSkillMultiplier

	return s.currentStrain
}

func (s *Flashlight) DifficultyValue() float64 {
	diff := 0.0

	for _, strain := range s.GetCurrentStrainPeaks() {
		diff += strain
	}

	return diff
}

func FlashlightDifficultyToPerformance(difficulty float64) float64 {
	return math.Pow(difficulty, 2.0) * 25.0
}
