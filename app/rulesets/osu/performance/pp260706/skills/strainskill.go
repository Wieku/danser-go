package skills

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/collections"
)

type StrainSkill struct {
	// The weight by which each strain value decays.
	DecayWeight float64

	// The length of each strain section.
	SectionLength float64

	// Delegate to calculate strain value of skill
	StrainValueOf func(obj *preprocessing.DifficultyObject) float64
	PostProcess   func(obj *preprocessing.DifficultyObject, strain, diffValue float64)

	CalculateInitialStrain func(time float64, current *preprocessing.DifficultyObject) float64

	CalculateDifficulty func() float64

	currentSectionPeak float64
	currentSectionEnd  float64

	strainPeaks       []float64
	strainPeaksSorted *collections.SortedList[float64]

	diffStrains *putils.LogisticSum

	difficulty float64

	diff *difficulty.Difficulty

	stepCalc bool
}

func NewSkill(d *difficulty.Difficulty, stepCalc bool) *StrainSkill {
	skill := &StrainSkill{
		DecayWeight:       0.9,
		SectionLength:     400,
		strainPeaksSorted: collections.NewSortedList[float64](),
		diff:              d,
		stepCalc:          stepCalc,
	}

	skill.diffStrains = putils.NewLogisticSum(stepCalc,
		0.88,
		10,
		1.1,
		func(previous, current float64) bool {
			return previous != current
		},
		func(strains []float64) float64 {
			return skill.DifficultyValue() * (1 - skill.DecayWeight)
		},
	)

	skill.CalculateDifficulty = skill.defaultDifficulty

	return skill
}

// Processes given DifficultyObject
func (skill *StrainSkill) Process(current *preprocessing.DifficultyObject) {
	if current.Index == 0 {
		skill.currentSectionEnd = math.Ceil(current.StartTime/skill.SectionLength) * skill.SectionLength
	}

	skill.processSectionEnd(current)

	currentStrain := skill.StrainValueOf(current)

	skill.currentSectionPeak = max(currentStrain, skill.currentSectionPeak)

	skill.diffStrains.AddStrain(currentStrain)

	if !skill.stepCalc {
		return
	}

	skill.difficulty = skill.CalculateDifficulty()

	skill.diffStrains.ProcessLastStrain(skill.difficulty * (1 - skill.DecayWeight))

	if skill.PostProcess != nil {
		skill.PostProcess(current, currentStrain, skill.difficulty)
	}
}

func (skill *StrainSkill) processSectionEnd(nextObj *preprocessing.DifficultyObject) {
	for nextObj.StartTime > skill.currentSectionEnd {
		sectionsLeft := math.Floor((nextObj.StartTime - skill.currentSectionEnd) / skill.SectionLength)

		if skill.currentSectionPeak == 0 && sectionsLeft > 10 { // skip for maps with huge distances between objects
			newPeaks := make([]float64, len(skill.strainPeaks)+int(sectionsLeft))
			copy(newPeaks, skill.strainPeaks)
			skill.strainPeaks = newPeaks // just add it to temporal db, we don't need to add

			skill.currentSectionEnd += skill.SectionLength * sectionsLeft

			continue
		}

		skill.saveCurrentPeak()
		skill.startNewSectionFrom(skill.currentSectionEnd, nextObj)
		skill.currentSectionEnd += skill.SectionLength
	}
}

func (skill *StrainSkill) GetCurrentStrainPeaks() []float64 {
	peaks := make([]float64, len(skill.strainPeaks)+1)
	copy(peaks, skill.strainPeaks)
	peaks[len(peaks)-1] = skill.currentSectionPeak

	return peaks
}

func (skill *StrainSkill) getCurrentStrainPeaksSorted() []float64 {
	peaks := skill.strainPeaksSorted.CloneWithAddCap(1)

	peaks.Add(skill.currentSectionPeak)

	return peaks.Slice
}

func (skill *StrainSkill) defaultDifficulty() float64 {
	diffValue := 0.0
	weight := 1.0

	strains := skill.getCurrentStrainPeaksSorted()

	lastDiff := -math.MaxFloat64

	for i := range len(strains) {
		diffValue += strains[len(strains)-1-i] * weight
		weight *= skill.DecayWeight

		if math.Abs(diffValue-lastDiff) < math.SmallestNonzeroFloat64 { // escape when strain * weight calculates to 0
			break
		}

		lastDiff = diffValue
	}

	return diffValue
}

func (skill *StrainSkill) DifficultyValue() float64 {
	if skill.stepCalc {
		return skill.difficulty
	}

	return skill.CalculateDifficulty()
}

func (skill *StrainSkill) CountTopWeightedStrains() float64 {
	return skill.diffStrains.GetValue()
}

func (skill *StrainSkill) saveCurrentPeak() {
	skill.strainPeaks = append(skill.strainPeaks, skill.currentSectionPeak)

	if skill.currentSectionPeak > 0 {
		skill.strainPeaksSorted.Add(skill.currentSectionPeak)
	}
}

func (skill *StrainSkill) startNewSectionFrom(end float64, current *preprocessing.DifficultyObject) {
	skill.currentSectionPeak = skill.CalculateInitialStrain(end, current)
}

func DefaultDifficultyToPerformance(difficulty float64) float64 {
	return 4.0 * math.Pow(difficulty, 3.0)
}
