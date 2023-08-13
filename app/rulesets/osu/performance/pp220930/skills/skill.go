package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
	"sort"
)

type Skill struct {
	// The weight by which each strain value decays.
	DecayWeight float64

	// The length of each strain section.
	SectionLength float64

	// Number of sections which strain value will be reduced.
	ReducedSectionCount int

	// Multiplier applied to the section with the biggest strain.
	ReducedStrainBaseline float64

	// Final multiplier after calculations.
	DifficultyMultiplier float64

	// Delegate to calculate strain value of skill
	StrainValueOf func(obj *preprocessing.DifficultyObject) float64

	CalculateInitialStrain func(time float64, current *preprocessing.DifficultyObject) float64

	currentSectionPeak float64
	currentSectionEnd  float64

	strainPeaks []float64

	diff *difficulty.Difficulty
}

func NewSkill(d *difficulty.Difficulty) *Skill {
	skill := &Skill{
		DecayWeight:           0.9,
		SectionLength:         400,
		ReducedSectionCount:   10,
		ReducedStrainBaseline: 0.75,
		DifficultyMultiplier:  1.06,
		diff:                  d,
	}

	return skill
}

// Processes given DifficultyObject
func (skill *Skill) Process(current *preprocessing.DifficultyObject) {
	if current.Index == 0 {
		skill.currentSectionEnd = math.Ceil(current.StartTime/skill.SectionLength) * skill.SectionLength
	}

	for current.StartTime > skill.currentSectionEnd {
		skill.saveCurrentPeak()
		skill.startNewSectionFrom(skill.currentSectionEnd, current)
		skill.currentSectionEnd += skill.SectionLength
	}

	skill.currentSectionPeak = max(skill.StrainValueOf(current), skill.currentSectionPeak)
}

func (skill *Skill) GetCurrentStrainPeaks() []float64 {
	peaks := make([]float64, len(skill.strainPeaks)+1)
	copy(peaks, skill.strainPeaks)
	peaks[len(peaks)-1] = skill.currentSectionPeak

	return peaks
}

func (skill *Skill) DifficultyValue() float64 {
	diff := 0.0
	weight := 1.0

	strains := skill.GetCurrentStrainPeaks()
	reverseSortFloat64s(strains)

	numReduced := min(len(strains), skill.ReducedSectionCount)

	for i := 0; i < numReduced; i++ {
		scale := math.Log10(mutils.Lerp(1.0, 10.0, mutils.Clamp(float64(i)/float64(skill.ReducedSectionCount), 0, 1)))
		strains[i] *= mutils.Lerp(skill.ReducedStrainBaseline, 1.0, scale)
	}

	reverseSortFloat64s(strains)

	for _, strain := range strains {
		diff += strain * weight
		weight *= skill.DecayWeight
	}

	return diff * skill.DifficultyMultiplier
}

func (skill *Skill) saveCurrentPeak() {
	skill.strainPeaks = append(skill.strainPeaks, skill.currentSectionPeak)
}

func (skill *Skill) startNewSectionFrom(end float64, current *preprocessing.DifficultyObject) {
	skill.currentSectionPeak = skill.CalculateInitialStrain(end, current)
}

func reverseSortFloat64s(arr []float64) {
	sort.Float64s(arr)

	n := len(arr)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
}
