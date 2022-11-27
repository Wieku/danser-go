package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/evaluators"
	preprocessing2 "github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/preprocessing"
	"math"
)

const (
	speedSkillMultiplier float64 = 1375
	speedStrainDecayBase float64 = 0.3
)

type SpeedSkill struct {
	*Skill

	currentStrain float64
	currentRhythm float64

	objectStrains []float64
}

func NewSpeedSkill(d *difficulty.Difficulty) *SpeedSkill {
	skill := &SpeedSkill{
		Skill:         NewSkill(d),
		objectStrains: make([]float64, 0),
	}

	skill.ReducedSectionCount = 5
	skill.DifficultyMultiplier = 1.04
	skill.StrainValueOf = skill.speedStrainValue
	skill.CalculateInitialStrain = skill.speedInitialStrain

	return skill
}

func (s *SpeedSkill) strainDecay(ms float64) float64 {
	return math.Pow(speedStrainDecayBase, ms/1000)
}

func (s *SpeedSkill) speedInitialStrain(time float64, current *preprocessing2.DifficultyObject) float64 {
	return (s.currentStrain * s.currentRhythm) * s.strainDecay(time-current.Previous(0).StartTime)
}

func (s *SpeedSkill) speedStrainValue(current *preprocessing2.DifficultyObject) float64 {
	s.currentStrain *= s.strainDecay(current.StrainTime)
	s.currentStrain += evaluators.EvaluateSpeed(current) * speedSkillMultiplier

	s.currentRhythm = evaluators.EvaluateRhythm(current)

	totalStrain := s.currentStrain * s.currentRhythm

	s.objectStrains = append(s.objectStrains, totalStrain)

	return totalStrain
}

func (s *SpeedSkill) RelevantNoteCount() (sum float64) {
	if len(s.objectStrains) == 0 {
		return
	}

	maxStrain := s.objectStrains[0]

	for _, strain := range s.objectStrains {
		if strain > maxStrain {
			maxStrain = strain
		}
	}

	if maxStrain == 0 {
		return
	}

	for _, strain := range s.objectStrains {
		sum += 1.0 / (1.0 + math.Exp(-(strain/maxStrain*12.0 - 6.0)))
	}

	return
}
