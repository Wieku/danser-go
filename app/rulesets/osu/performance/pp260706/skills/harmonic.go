package skills

import (
	"math"
	"slices"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
)

type Harmonic struct {
	NoteWeightSum float64

	HarmonicScale float64

	DecayExponent float64

	// Delegate to calculate strain value of skill
	DifficultyOf func(obj *preprocessing.DifficultyObject) float64
	PostProcess  func(obj *preprocessing.DifficultyObject, strain, diffValue float64)

	// Delegate to decay recorded strains when reconstructing graph peaks.
	StrainDecay func(ms float64) float64

	ApplyDifficultyTransformation func(difficulties []float64)

	difficulties        []timedStrain
	difficultiesNonZero []float64

	diffStrains *putils.LogisticSum

	difficulty float64

	diff *difficulty.Difficulty

	stepCalc bool
}

func NewHarmonic(d *difficulty.Difficulty, stepCalc bool) *Harmonic {
	skill := &Harmonic{
		HarmonicScale: 1.0,
		DecayExponent: 0.9,

		difficulties:        make([]timedStrain, 0),
		difficultiesNonZero: make([]float64, 0),
		diff:                d,
		stepCalc:            stepCalc,
	}

	skill.diffStrains = putils.NewLogisticSum(stepCalc,
		0.88,
		10,
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

// Processes given DifficultyObject
func (skill *Harmonic) Process(current *preprocessing.DifficultyObject) {
	currentDiff := skill.DifficultyOf(current)

	skill.difficulties = append(skill.difficulties, timedStrain{value: currentDiff, startTime: current.StartTime})
	if currentDiff > 0 {
		skill.difficultiesNonZero = append(skill.difficultiesNonZero, currentDiff)
	}

	skill.diffStrains.AddStrain(currentDiff)

	if !skill.stepCalc {
		return
	}

	skill.difficulty = skill.defaultDifficulty()

	if skill.NoteWeightSum == 0 {
		skill.diffStrains.ProcessLastStrain(0)
	} else {
		skill.diffStrains.ProcessLastStrain(skill.difficulty / skill.NoteWeightSum)
	}

	if skill.PostProcess != nil {
		skill.PostProcess(current, currentDiff, skill.difficulty)
	}
}

func (skill *Harmonic) defaultDifficulty() float64 {
	if len(skill.difficultiesNonZero) == 0 {
		return 0
	}

	diffValue := 0.0
	skill.NoteWeightSum = 0.0

	difficulties := slices.Clone(skill.difficultiesNonZero)

	if skill.ApplyDifficultyTransformation != nil {
		skill.ApplyDifficultyTransformation(difficulties)
	}

	slices.Sort(difficulties)

	for i := range len(difficulties) {
		note := difficulties[len(difficulties)-1-i]

		if note == 0 {
			break
		}

		// Use a harmonic sum that considers each note of the map according to a predefined weight.
		weight := (1 + (skill.HarmonicScale / (1 + float64(i)))) / (math.Pow(float64(i), skill.DecayExponent) + 1 + (skill.HarmonicScale / (1 + float64(i))))

		skill.NoteWeightSum += weight

		diffValue += note * weight
	}

	return diffValue
}

func (skill *Harmonic) GetCurrentStrainPeaks() []float64 {
	return getStrainPeaks(skill.difficulties, skill.StrainDecay)
}

func (skill *Harmonic) DifficultyValue() float64 {
	if skill.stepCalc {
		return skill.difficulty
	}

	return skill.defaultDifficulty()
}

func (skill *Harmonic) CountTopWeightedStrains() float64 {
	return skill.diffStrains.GetValue()
}

func HarmonicDifficultyToPerformance(difficulty float64) float64 {
	return 4.0 * math.Pow(difficulty, 3.0)
}
