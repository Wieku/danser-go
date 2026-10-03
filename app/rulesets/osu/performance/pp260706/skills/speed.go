package skills

import (
	"math"
	"slices"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
)

const (
	speedSkillMultiplier float64 = 1.16
	speedStrainDecayBase float64 = 0.3
)

type SpeedSkill struct {
	*Harmonic

	currentDiff float64

	relevantCount *putils.LogisticSum
	topSliders    *putils.LogisticSum
}

func NewSpeedSkill(d *difficulty.Difficulty, stepCalc bool) *SpeedSkill {
	skill := &SpeedSkill{
		Harmonic: NewHarmonic(d, stepCalc),
	}

	skill.HarmonicScale = 20
	skill.DecayExponent = 0.9
	skill.DifficultyOf = skill.speedDifficulty
	skill.StrainDecay = skill.strainDecay
	skill.PostProcess = skill.postProcess

	skill.relevantCount = putils.NewLogisticSum(stepCalc, 6, 1, 1, func(previous, current float64) bool {
		return current > previous
	}, func(strains []float64) float64 {
		return slices.Max(strains) / 12
	})

	skill.topSliders = putils.NewLogisticSum(stepCalc, 0.88, 10, 1.1, func(previous, current float64) bool {
		return current != previous
	}, func(strains []float64) float64 {
		if skill.NoteWeightSum == 0 {
			return 0.0
		}

		return skill.DifficultyValue() / skill.NoteWeightSum
	})

	return skill
}

func (s *SpeedSkill) strainDecay(ms float64) float64 {
	return math.Pow(speedStrainDecayBase, ms/1000)
}

func (s *SpeedSkill) speedDifficulty(current *preprocessing.DifficultyObject) float64 {
	if s.diff.CheckModActive(difficulty.Relax) {
		return 0
	}

	decay := s.strainDecay(current.AdjustedDeltaTime)

	s.currentDiff *= decay
	s.currentDiff += s.calculateAdjustedDifficulty(current) * (1 - decay) * speedSkillMultiplier

	currentRhythm := evaluators.EvaluateRhythm(current)

	totalDiff := s.currentDiff * currentRhythm

	s.relevantCount.AddStrain(totalDiff)

	if current.IsSlider {
		s.topSliders.AddStrain(totalDiff)
	}

	return totalDiff
}

func (s *SpeedSkill) postProcess(current *preprocessing.DifficultyObject, strain float64, diffValue float64) {
	s.relevantCount.ProcessLastStrain(strain / 12)

	if s.NoteWeightSum == 0 {
		s.topSliders.ProcessLastStrain(0)
	} else if current.IsSlider {
		s.topSliders.ProcessLastStrain(diffValue / s.NoteWeightSum)
	} else {
		s.topSliders.UpdateDivider(diffValue / s.NoteWeightSum)
	}
}

func (s *SpeedSkill) RelevantNoteCount() float64 {
	return s.relevantCount.GetValue()
}

func (s *SpeedSkill) CountTopWeightedSliders() float64 {
	return s.topSliders.GetValue()
}

func (s *SpeedSkill) calculateAdjustedDifficulty(current *preprocessing.DifficultyObject) float64 {
	diffc := evaluators.EvaluateSpeed(current)

	if s.diff.CheckModActive(difficulty.Relax2) {
		diffc *= 0.5
	}

	return diffc
}
