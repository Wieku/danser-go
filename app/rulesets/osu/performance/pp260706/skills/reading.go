package skills

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	readingSkillMultiplier float64 = 2.5
	readingStrainDecayBase float64 = 0.8
)

type ReadingSkill struct {
	*Harmonic

	objects []*preprocessing.DifficultyObject

	currentDiff float64

	reducedDuration  *float64
	reducedNoteCount int
}

func NewReadingSkill(d *difficulty.Difficulty, stepCalc bool) *ReadingSkill {
	skill := &ReadingSkill{
		Harmonic: NewHarmonic(d, stepCalc),
	}

	skill.DifficultyOf = skill.readingDifficulty
	skill.StrainDecay = skill.strainDecay
	skill.ApplyDifficultyTransformation = skill.applyDifficultyTransformation

	skill.diffStrains = putils.NewLogisticSum(stepCalc,
		1.15,
		5,
		1.1,
		func(previous, current float64) bool {
			return previous != current
		},
		func(strains []float64) float64 {
			if skill.NoteWeightSum == 0 {
				return 0
			}

			return skill.DifficultyValue() / skill.NoteWeightSum
		},
	)

	return skill
}

func (s *ReadingSkill) applyDifficultyTransformation(difficulties []float64) {
	const reducedDifficultyBaseLine = 0.0 // Assume the first seconds are completely memorised

	for i := 0; i < min(len(difficulties), s.reducedNoteCount); i++ {
		scale := math.Log10(mutils.Lerp(1.0, 10.0, mutils.Clamp(float64(i)/float64(s.reducedNoteCount), 0.0, 1.0)))
		difficulties[i] *= mutils.Lerp(reducedDifficultyBaseLine, 1.0, scale)
	}
}

func (s *ReadingSkill) strainDecay(ms float64) float64 {
	return math.Pow(readingStrainDecayBase, ms/1000)
}

func (s *ReadingSkill) readingDifficulty(current *preprocessing.DifficultyObject) float64 {
	const reducedDifficultyDuration = 60 * 1000

	s.objects = append(s.objects, current)

	decay := s.strainDecay(current.DeltaTime)

	s.currentDiff *= decay
	s.currentDiff += s.calculateAdjustedDifficulty(current) * (1 - decay) * readingSkillMultiplier

	// This currently operates under the assumption that `ObjectDifficultyOf` is called once per object, and in order.
	// Under that assumption, we can trust that `current.StartTime` refers to the start time of the first object in the case that `reducedDuration` is yet to be set.
	if s.reducedDuration == nil {
		s.reducedDuration = new(current.StartTime + reducedDifficultyDuration)
	}

	// This relies on the same assumption, as calling in order means that we can safely increase the note count until we reach the first object after the reduced duration.
	if current.StartTime <= *s.reducedDuration {
		s.reducedNoteCount++
	}

	return s.currentDiff
}

func (s *ReadingSkill) calculateAdjustedDifficulty(current *preprocessing.DifficultyObject) float64 {
	diffc := evaluators.EvaluateReading(current, s.diff.HiddenFadesObjects())

	if s.diff.CheckModActive(difficulty.TouchDevice) {
		diffc = math.Pow(diffc, 0.89)
	}

	if s.diff.CheckModActive(difficulty.Relax) {
		diffc *= 0.4
	}

	if s.diff.CheckModActive(difficulty.Relax2) {
		diffc *= 0.1
	}

	diffc *= 0.825 + math.Pow(max(0, s.diff.ODReal), 2.2)/1125

	return diffc
}
