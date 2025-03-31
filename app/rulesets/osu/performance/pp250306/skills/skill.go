package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp250306/preprocessing"
	"github.com/wieku/danser-go/framework/collections"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
	"slices"
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

	// Delegate to calculate strain value of skill
	StrainValueOf func(obj *preprocessing.DifficultyObject) float64

	CalculateInitialStrain func(time float64, current *preprocessing.DifficultyObject) float64

	currentSectionPeak float64
	currentSectionEnd  float64

	peakWeights []float64

	strainPeaks       []float64
	strainPeaksSorted *collections.SortedList[float64]

	objectStrains        []float64
	difficultStrainCount float64

	difficulty     float64
	lastDifficulty float64

	diff *difficulty.Difficulty

	stepCalc bool
}

func NewSkill(d *difficulty.Difficulty, stepCalc bool) *Skill {
	skill := &Skill{
		DecayWeight:           0.9,
		SectionLength:         400,
		ReducedSectionCount:   10,
		ReducedStrainBaseline: 0.75,
		objectStrains:         make([]float64, 0),
		strainPeaksSorted:     collections.NewSortedList[float64](),
		diff:                  d,
		stepCalc:              stepCalc,
		lastDifficulty:        -math.MaxFloat64,
	}

	return skill
}

// Processes given DifficultyObject
func (skill *Skill) Process(current *preprocessing.DifficultyObject) {
	if current.Index == 0 {
		skill.currentSectionEnd = math.Ceil(current.StartTime/skill.SectionLength) * skill.SectionLength
	}

	skill.processSectionEnd(current)

	currentStrain := skill.StrainValueOf(current)

	skill.currentSectionPeak = max(currentStrain, skill.currentSectionPeak)

	if !skill.stepCalc {
		return
	}

	skill.difficultyValue()

	if skill.lastDifficulty != skill.difficulty {
		skill.difficultStrainCount = skill.countDifficultStrains()
	} else if skill.difficulty != 0 {
		skill.difficultStrainCount += 1.1 / (1 + math.Exp(-10*(currentStrain/(skill.difficulty/10)-0.88)))
	}

	skill.lastDifficulty = skill.difficulty
}

func (skill *Skill) processSectionEnd(nextObj *preprocessing.DifficultyObject) {
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

func (skill *Skill) GetCurrentStrainPeaks() []float64 {
	peaks := make([]float64, len(skill.strainPeaks)+1)
	copy(peaks, skill.strainPeaks)
	peaks[len(peaks)-1] = skill.currentSectionPeak

	return peaks
}

func (skill *Skill) getCurrentStrainPeaksSorted() []float64 {
	peaks := skill.strainPeaksSorted.CloneWithAddCap(1)

	peaks.Add(skill.currentSectionPeak)

	return peaks.Slice
}

func (skill *Skill) difficultyValue() float64 {
	if skill.peakWeights == nil { //Precalculated peak weights
		skill.peakWeights = make([]float64, skill.ReducedSectionCount)
		for i := range skill.ReducedSectionCount {
			scale := math.Log10(mutils.Lerp(1.0, 10.0, mutils.Clamp(float64(i)/float64(skill.ReducedSectionCount), 0, 1)))
			skill.peakWeights[i] = mutils.Lerp(skill.ReducedStrainBaseline, 1.0, scale)
		}
	}

	skill.difficulty = 0.0
	weight := 1.0

	strains := skill.getCurrentStrainPeaksSorted()

	lowest := strains[len(strains)-1]

	sectionsReduced := min(len(strains), skill.ReducedSectionCount)

	for i := range sectionsReduced {
		strains[len(strains)-1-i] *= skill.peakWeights[i]
		lowest = min(lowest, strains[len(strains)-1-i])
	}

	// Search for lowest strain that's higher or equal than lowest reduced strain to avoid unnecessary sorting
	idx, _ := slices.BinarySearch(strains[:len(strains)-sectionsReduced], lowest)
	slices.Sort(strains[idx:])

	lastDiff := -math.MaxFloat64

	for i := range len(strains) {
		skill.difficulty += strains[len(strains)-1-i] * weight
		weight *= skill.DecayWeight

		if math.Abs(skill.difficulty-lastDiff) < math.SmallestNonzeroFloat64 { // escape when strain * weight calculates to 0
			break
		}

		lastDiff = skill.difficulty
	}

	return skill.difficulty
}

func (skill *Skill) DifficultyValue() float64 {
	if skill.stepCalc {
		return skill.difficulty
	}

	return skill.difficultyValue()
}

func (skill *Skill) countDifficultStrains() float64 {
	if skill.difficulty == 0 {
		return 0
	}

	consistentTopStrain := skill.difficulty / 10 // What would the top strain be if all strain values were identical
	// Use a weighted sum of all strains. Constants are arbitrary and give nice values

	sum := 0.0

	for _, s := range skill.objectStrains {
		sum += 1.1 / (1 + math.Exp(-10*(s/consistentTopStrain-0.88)))
	}

	return sum
}

func (skill *Skill) CountDifficultStrains() float64 {
	if skill.stepCalc {
		return skill.difficultStrainCount
	}

	return skill.countDifficultStrains()
}

func (skill *Skill) saveCurrentPeak() {
	skill.strainPeaks = append(skill.strainPeaks, skill.currentSectionPeak)

	if skill.currentSectionPeak > 0 {
		skill.strainPeaksSorted.Add(skill.currentSectionPeak)
	}
}

func (skill *Skill) startNewSectionFrom(end float64, current *preprocessing.DifficultyObject) {
	skill.currentSectionPeak = skill.CalculateInitialStrain(end, current)
}

func DefaultDifficultyToPerformance(difficulty float64) float64 {
	return math.Pow(5.0*max(1.0, difficulty/0.0675)-4.0, 3.0) / 100000.0
}
