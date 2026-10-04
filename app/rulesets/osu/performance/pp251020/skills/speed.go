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
	speedSkillMultiplier float64 = 1.47
	speedStrainDecayBase float64 = 0.3
)

type SpeedSkill struct {
	*Skill

	currentStrain float64
	currentRhythm float64
	maxStrain     float64

	relevantCount *putils.LogisticSum
	topSliders    *putils.LogisticSum
}

func NewSpeedSkill(d *difficulty.Difficulty, stepCalc bool) *SpeedSkill {
	skill := &SpeedSkill{
		Skill: NewSkill(d, stepCalc),
	}

	skill.ReducedSectionCount = 5
	skill.StrainValueOf = skill.speedStrainValue
	skill.CalculateInitialStrain = skill.speedInitialStrain
	skill.PostProcess = skill.postProcess

	skill.relevantCount = putils.NewLogisticSum(stepCalc, 6, 1, 1, func(previous, current float64) bool {
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

func (s *SpeedSkill) strainDecay(ms float64) float64 {
	return math.Pow(speedStrainDecayBase, ms/1000)
}

func (s *SpeedSkill) speedInitialStrain(time float64, current *preprocessing.DifficultyObject) float64 {
	return (s.currentStrain * s.currentRhythm) * s.strainDecay(time-current.Previous(0).StartTime)
}

func (s *SpeedSkill) speedStrainValue(current *preprocessing.DifficultyObject) float64 {
	s.currentStrain *= s.strainDecay(current.AdjustedDeltaTime)
	s.currentStrain += evaluators.EvaluateSpeed(current) * speedSkillMultiplier

	s.currentRhythm = evaluators.EvaluateRhythm(current)

	totalStrain := s.currentStrain * s.currentRhythm

	s.relevantCount.AddStrain(totalStrain)

	if current.IsSlider {
		s.topSliders.AddStrain(totalStrain)
	}

	return totalStrain
}

func (s *SpeedSkill) postProcess(current *preprocessing.DifficultyObject, strain float64, diffValue float64) {
	s.relevantCount.ProcessLastStrain(strain / 12)
	s.topSliders.ProcessLastStrain(diffValue / 10)
}

func (s *SpeedSkill) RelevantNoteCount() float64 {
	return s.relevantCount.GetValue()
}

func (s *SpeedSkill) CountTopWeightedSliders() float64 {
	return s.topSliders.GetValue()
}
