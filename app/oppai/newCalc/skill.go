package newCalc

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"math"
	"sort"
)

type Skill struct {
	// Strain values are multiplied by this number for the given skill. Used to balance the value of different skills between each other.
	SkillMultiplier float64

	// Determines how quickly strain decays for the given skill.
	// For example a value of 0.15 indicates that strain decays to 15% of its original value in one second.
	StrainDecayBase float64

	// The weight by which each strain value decays.
	DecayWeight float64

	// The length of each strain section.
	SectionLength float64

	// How many DifficultyObjects should be kept
	HistoryLength int

	// Keeps track of previous DifficultyObjects for strain section calculations
	Previous []*DifficultyObject

	// The current strain level
	CurrentStrain float64

	// Delegate to calculate strain value of skill
	StrainValueOf func(skill *Skill, obj *DifficultyObject) float64

	currentSectionPeak float64
	currentSectionEnd  float64

	strainPeaks []float64

	// Should fixed clock rate calculations be used, set to false to use current osu!stable calculations (2021.01)
	fixedCalculations bool

	diff              *difficulty.Difficulty
}

func NewSkill(useFixedCalculations bool, d *difficulty.Difficulty) *Skill {
	return &Skill{
		DecayWeight:       0.9,
		SectionLength:     400,
		HistoryLength:     1,
		fixedCalculations: useFixedCalculations,
		diff:              d,
	}
}

func (skill *Skill) processInternal(current *DifficultyObject) {
	var startTime float64
	if skill.fixedCalculations {
		startTime = current.StartTime
	} else {
		startTime = current.BaseObject.GetStartTime()
	}

	if len(skill.Previous) == 0 {
		skill.currentSectionEnd = math.Ceil(startTime/skill.SectionLength) * skill.SectionLength
	}

	for startTime > skill.currentSectionEnd {
		skill.saveCurrentPeak()
		skill.startNewSectionFrom(skill.currentSectionEnd)

		if skill.fixedCalculations {
			skill.currentSectionEnd += skill.SectionLength
		} else {
			skill.currentSectionEnd += skill.SectionLength * skill.diff.Speed
		}
	}

	skill.CurrentStrain *= skill.strainDecay(current.DeltaTime)
	skill.CurrentStrain += skill.StrainValueOf(skill, current) * skill.SkillMultiplier

	skill.currentSectionPeak = math.Max(skill.CurrentStrain, skill.currentSectionPeak)
}

// Processes given DifficultyObject
func (skill *Skill) Process(current *DifficultyObject) {
	if len(skill.Previous) > skill.HistoryLength {
		skill.Previous = skill.Previous[len(skill.Previous)-skill.HistoryLength:]
	}

	skill.processInternal(current)

	skill.Previous = append(skill.Previous, current)
}

func (skill *Skill) GetPrevious() *DifficultyObject {
	if len(skill.Previous) == 0 {
		return nil
	}

	return skill.Previous[len(skill.Previous)-1]
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

	strains := reverseSortFloat64s(skill.GetCurrentStrainPeaks())

	for _, strain := range strains {
		diff += strain * weight
		weight *= skill.DecayWeight
	}

	return diff
}

func (skill *Skill) strainDecay(ms float64) float64 {
	return math.Pow(skill.StrainDecayBase, ms/1000)
}

func (skill *Skill) saveCurrentPeak() {
	skill.strainPeaks = append(skill.strainPeaks, skill.currentSectionPeak)
}

func (skill *Skill) startNewSectionFrom(end float64) {
	var startTime float64
	if skill.fixedCalculations {
		startTime = skill.GetPrevious().StartTime
	} else {
		startTime = skill.GetPrevious().BaseObject.GetStartTime()
	}

	skill.currentSectionPeak = skill.CurrentStrain * skill.strainDecay(end-startTime)
}

func reverseSortFloat64s(arr []float64) []float64 {
	x := make([]float64, len(arr))
	copy(x, arr)

	sort.Float64s(x)

	n := len(x)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		x[i], x[j] = x[j], x[i]
	}

	return x
}
