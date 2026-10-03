package skills

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
)

const (
	flSkillMultiplier float64 = 0.058
	flStrainDecayBase float64 = 0.15
)

type Flashlight struct {
	*StrainSkill
	currentStrain float64
	objectCount   func() int
}

func NewFlashlightSkill(d *difficulty.Difficulty, objectCount func() int) *Flashlight {
	skill := &Flashlight{StrainSkill: NewSkill(d, false)}

	skill.StrainValueOf = skill.flashlightStrainValue
	skill.CalculateInitialStrain = skill.flInitialStrain
	skill.CalculateDifficulty = skill.flDifficulty
	skill.objectCount = objectCount

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
	s.currentStrain += s.calculateAdjustedDifficulty(current) * flSkillMultiplier

	return s.currentStrain
}

func (s *Flashlight) calculateAdjustedDifficulty(current *preprocessing.DifficultyObject) float64 {
	diffc := evaluators.EvaluateFlashlight(current)

	if s.diff.CheckModActive(difficulty.TouchDevice) {
		diffc = math.Pow(diffc, 0.9)
	}

	if s.diff.CheckModActive(difficulty.Relax) {
		diffc *= 0.7
	}

	if s.diff.CheckModActive(difficulty.Relax2) {
		diffc *= 0.4
	}

	diffc *= 0.985 + putils.Powi(max(0, s.diff.ODReal), 2)/4000

	return diffc
}

func (s *Flashlight) flDifficulty() float64 {
	diff := 0.0

	for _, strain := range s.GetCurrentStrainPeaks() {
		diff += strain
	}

	totalObjects := s.objectCount()
	dMult := 0.7 + 0.1*min(1.0, float64(totalObjects)/200)
	if totalObjects > 200 {
		dMult += 0.2 * min(1, (float64(totalObjects)-200)/200)
	}

	return diff * dMult
}

func FlashlightDifficultyToPerformance(difficulty float64) float64 {
	return math.Pow(difficulty, 2.0) * 25.0
}
