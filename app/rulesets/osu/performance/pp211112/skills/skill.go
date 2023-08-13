package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp211112/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
	"sort"
)

type Skill struct {
	// Whether new pp changes should be considered
	Experimental bool

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
	StrainValueOf func(obj *preprocessing.DifficultyObject) float64

	// Delegate to calculate strain bonus of skill
	StrainBonusOf func(obj *preprocessing.DifficultyObject) float64

	CalculateInitialStrain func(time float64) float64

	currentSectionPeak float64
	currentSectionEnd  float64

	strainPeaks []float64

	diff *difficulty.Difficulty
}

func NewSkill(d *difficulty.Difficulty, experimental bool) *Skill {
	skill := &Skill{
		Experimental:          experimental,
		DecayWeight:           0.9,
		SectionLength:         400,
		HistoryLength:         1,
		ReducedSectionCount:   10,
		ReducedStrainBaseline: 0.75,
		DifficultyMultiplier:  1.06,
		diff:                  d,
	}

	skill.CalculateInitialStrain = func(time float64) float64 {
		return skill.CurrentStrain * skill.strainDecay(time-skill.GetPrevious(0).StartTime)
	}

	return skill
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
	skill.CurrentStrain += skill.StrainValueOf(current) * skill.SkillMultiplier

	tempStrain := skill.CurrentStrain

	if skill.StrainBonusOf != nil {
		tempStrain *= skill.StrainBonusOf(current)
	}

	skill.currentSectionPeak = max(tempStrain, skill.currentSectionPeak)
}

// Processes given DifficultyObject
func (skill *Skill) Process(current *preprocessing.DifficultyObject) {
	if len(skill.Previous) > skill.HistoryLength {
		skill.Previous = skill.Previous[len(skill.Previous)-skill.HistoryLength:]
	}

	skill.processInternal(current)

	skill.Previous = append(skill.Previous, current)
}

func (skill *Skill) GetPrevious(i int) *preprocessing.DifficultyObject {
	if len(skill.Previous)-1-i < 0 {
		return nil
	}

	return skill.Previous[len(skill.Previous)-1-i]
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

func (skill *Skill) strainDecay(ms float64) float64 {
	return math.Pow(skill.StrainDecayBase, ms/1000)
}

func (skill *Skill) saveCurrentPeak() {
	skill.strainPeaks = append(skill.strainPeaks, skill.currentSectionPeak)
}

func (skill *Skill) startNewSectionFrom(end float64) {
	skill.currentSectionPeak = skill.CalculateInitialStrain(end)
}

func (skill *Skill) CountDifficultStrains(clockRate float64) float64 {
	peaks := skill.GetCurrentStrainPeaks()

	var topStrain, realtimeCount float64

	for _, v := range peaks {
		topStrain = max(topStrain, v)
	}

	for _, v := range peaks {
		realtimeCount += math.Pow(v/topStrain, 4)
	}

	return realtimeCount * clockRate
}

func reverseSortFloat64s(arr []float64) {
	sort.Float64s(arr)

	n := len(arr)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
}
