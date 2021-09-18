package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
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

	// How many DifficultyObjects should be kept.
	HistoryLength int

	// Number of sections which strain value will be reduced.
	ReducedSectionCount int

	// Multiplier applied to the section with the biggest strain.
	ReducedStrainBaseline float64

	// Final multiplier after calculations.
	DifficultyMultiplier float64

	// Keeps track of previous DifficultyObjects for strain section calculations
	Previous []*preprocessing.DifficultyObject

	// The current strain level
	CurrentStrain float64

	// Delegate to calculate strain value of skill
	StrainValueOf func(skill *Skill, obj *preprocessing.DifficultyObject) float64

	currentSectionPeak float64
	currentSectionEnd  float64

	strainPeaks []float64

	diff *difficulty.Difficulty
}

func NewSkill(d *difficulty.Difficulty) *Skill {
	return &Skill{
		DecayWeight:           0.9,
		SectionLength:         400,
		HistoryLength:         1,
		ReducedSectionCount:   10,
		ReducedStrainBaseline: 0.75,
		DifficultyMultiplier:  1.06,
		diff:                  d,
	}
}

func (skill *Skill) processInternal(current *preprocessing.DifficultyObject) {
	startTime := current.StartTime
	sectionLength := skill.SectionLength

	if len(skill.Previous) == 0 {
		skill.currentSectionEnd = math.Ceil(startTime/sectionLength) * sectionLength
	}

	for startTime > skill.currentSectionEnd {
		skill.saveCurrentPeak()
		skill.startNewSectionFrom(skill.currentSectionEnd)

		skill.currentSectionEnd += sectionLength
	}

	skill.CurrentStrain *= skill.strainDecay(current.DeltaTime)
	skill.CurrentStrain += skill.StrainValueOf(skill, current) * skill.SkillMultiplier

	skill.currentSectionPeak = math.Max(skill.CurrentStrain, skill.currentSectionPeak)
}

// Processes given DifficultyObject
func (skill *Skill) Process(current *preprocessing.DifficultyObject) {
	if len(skill.Previous) > skill.HistoryLength {
		skill.Previous = skill.Previous[len(skill.Previous)-skill.HistoryLength:]
	}

	skill.processInternal(current)

	skill.Previous = append(skill.Previous, current)
}

func (skill *Skill) GetPrevious() *preprocessing.DifficultyObject {
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

	strains := skill.GetCurrentStrainPeaks()
	reverseSortFloat64s(strains)

	numReduced := mutils.MinI(len(strains), skill.ReducedSectionCount)

	for i := 0; i < numReduced; i++ {
		scale := math.Log10(mutils.LerpF64(1, 10, mutils.ClampF64(float64(i) / float64(skill.ReducedSectionCount), 0, 1)))
		strains[i] *= mutils.LerpF64(skill.ReducedStrainBaseline, 1.0, scale)
	}

	reverseSortFloat64s(strains)

	for _, strain := range strains {
		diff += strain * weight
		weight *= skill.DecayWeight
	}

	return diff * skill.DifficultyMultiplier
}

func (skill *Skill) strainDecay(ms float64) float64 {
	return math.Pow(skill.StrainDecayBase, ms/1000)
}

func (skill *Skill) saveCurrentPeak() {
	skill.strainPeaks = append(skill.strainPeaks, skill.currentSectionPeak)
}

func (skill *Skill) startNewSectionFrom(end float64) {
	skill.currentSectionPeak = skill.CurrentStrain * skill.strainDecay(end-skill.GetPrevious().StartTime)
}

func reverseSortFloat64s(arr []float64) {
	sort.Float64s(arr)

	n := len(arr)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
}
