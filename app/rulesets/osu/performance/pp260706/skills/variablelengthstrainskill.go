package skills

import (
	"math"
	"sort"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
)

type StrainPeak struct {
	Value         float64
	SectionLength float64
}

func newStrainPeak(value, sectionLength float64) StrainPeak {
	return StrainPeak{
		Value:         value,
		SectionLength: math.RoundToEven(sectionLength),
	}
}

type queuedStrain struct {
	value     float64
	startTime float64
}

// VariableLengthStrainSkill is similar to StrainSkill, but instead of strains having a fixed length,
// strains can be any length. A new StrainPeak is created for each DifficultyObject.
type VariableLengthStrainSkill struct {
	// The weight by which each strain value decays.
	DecayWeight float64

	// The maximum length of each strain section.
	MaxSectionLength float64

	// Delegate to calculate strain value of skill
	StrainValueOf func(obj *preprocessing.DifficultyObject) float64
	PostProcess   func(obj *preprocessing.DifficultyObject, strain, diffValue float64)

	CalculateInitialStrain func(time float64, current *preprocessing.DifficultyObject) float64

	// Delegate to decay recorded strains when reconstructing graph peaks.
	StrainDecay func(ms float64) float64

	CalculateDifficulty func() float64

	currentSectionPeak  float64
	currentSectionBegin float64
	currentSectionEnd   float64

	// The number of MaxSectionLength sections calculated such that enough of the difficulty value is preserved.
	maxStoredLength float64

	strainPeaks []StrainPeak
	totalLength float64

	// Graph samples retain their chronology independently of sorted and pruned difficulty peaks.
	objectStrains []timedStrain

	// Stores previous strains so that a high difficulty object followed by a lower difficulty object
	// gets a full strain instead of being cut short.
	queuedStrains []queuedStrain

	finalPeak *StrainPeak

	diffStrains *putils.LogisticSum

	difficulty float64

	diff *difficulty.Difficulty

	stepCalc bool
}

func NewVariableLengthStrainSkill(d *difficulty.Difficulty, stepCalc bool) *VariableLengthStrainSkill {
	skill := &VariableLengthStrainSkill{
		DecayWeight:      0.9,
		MaxSectionLength: 400,
		diff:             d,
		stepCalc:         stepCalc,
	}

	skill.maxStoredLength = 11 / (1 - skill.DecayWeight)

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

	return skill
}

// Process processes a DifficultyObject and updates current strain values accordingly.
func (skill *VariableLengthStrainSkill) Process(current *preprocessing.DifficultyObject) {
	if current.Index == 0 {
		skill.currentSectionBegin = current.StartTime
		skill.currentSectionEnd = skill.currentSectionBegin + skill.MaxSectionLength

		skill.currentSectionPeak = skill.StrainValueOf(current)
		skill.processStrain(current, skill.currentSectionPeak)

		return
	}

	skill.backfillPeaks(current)

	currentStrain := skill.StrainValueOf(current)

	if currentStrain > skill.currentSectionPeak {
		skill.queuedStrains = skill.queuedStrains[:0]

		skill.saveCurrentPeak(current.StartTime - skill.currentSectionBegin)

		skill.currentSectionBegin = current.StartTime
		skill.currentSectionEnd = skill.currentSectionBegin + skill.MaxSectionLength
		skill.currentSectionPeak = currentStrain
	} else {
		for len(skill.queuedStrains) > 0 && skill.queuedStrains[len(skill.queuedStrains)-1].value < currentStrain {
			skill.queuedStrains = skill.queuedStrains[:len(skill.queuedStrains)-1]
		}

		skill.queuedStrains = append(skill.queuedStrains, queuedStrain{value: currentStrain, startTime: current.StartTime})
	}

	skill.processStrain(current, currentStrain)
}

func (skill *VariableLengthStrainSkill) processStrain(current *preprocessing.DifficultyObject, currentStrain float64) {
	skill.objectStrains = append(skill.objectStrains, timedStrain{value: currentStrain, startTime: current.StartTime})
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

func (skill *VariableLengthStrainSkill) backfillPeaks(current *preprocessing.DifficultyObject) {
	for current.StartTime > skill.currentSectionEnd {
		skill.saveCurrentPeak(skill.currentSectionEnd - skill.currentSectionBegin)
		skill.currentSectionBegin = skill.currentSectionEnd

		if len(skill.queuedStrains) > 0 {
			strain := skill.queuedStrains[0]
			skill.queuedStrains = skill.queuedStrains[1:]

			skill.currentSectionEnd = strain.startTime + skill.MaxSectionLength
			skill.startNewSectionFrom(skill.currentSectionBegin, current)

			skill.currentSectionPeak = max(skill.currentSectionPeak, strain.value)
		} else {
			skill.currentSectionEnd = skill.currentSectionBegin + skill.MaxSectionLength
			skill.startNewSectionFrom(skill.currentSectionBegin, current)
		}
	}
}

func (skill *VariableLengthStrainSkill) saveCurrentPeak(sectionLength float64) {
	skill.removeFinalPeak()

	skill.addPeak(newStrainPeak(skill.currentSectionPeak, sectionLength))
	skill.totalLength += sectionLength

	for skill.totalLength > skill.maxStoredLength*skill.MaxSectionLength {
		last := len(skill.strainPeaks) - 1
		skill.totalLength -= skill.strainPeaks[last].SectionLength
		skill.strainPeaks = skill.strainPeaks[:last]
	}
}

func (skill *VariableLengthStrainSkill) addPeak(peak StrainPeak) {
	index := sort.Search(len(skill.strainPeaks), func(i int) bool {
		return skill.strainPeaks[i].Value <= peak.Value
	})

	skill.strainPeaks = append(skill.strainPeaks, StrainPeak{})
	copy(skill.strainPeaks[index+1:], skill.strainPeaks[index:])
	skill.strainPeaks[index] = peak
}

func (skill *VariableLengthStrainSkill) removeFinalPeak() {
	if skill.finalPeak == nil {
		return
	}

	for i, peak := range skill.strainPeaks {
		if peak == *skill.finalPeak {
			copy(skill.strainPeaks[i:], skill.strainPeaks[i+1:])
			skill.strainPeaks = skill.strainPeaks[:len(skill.strainPeaks)-1]
			break
		}
	}

	skill.finalPeak = nil
}

func (skill *VariableLengthStrainSkill) startNewSectionFrom(time float64, current *preprocessing.DifficultyObject) {
	skill.currentSectionPeak = skill.CalculateInitialStrain(time, current)
}

func (skill *VariableLengthStrainSkill) getCurrentStrainPeaks() []StrainPeak {
	if skill.finalPeak == nil {
		peak := newStrainPeak(skill.currentSectionPeak, skill.currentSectionEnd-skill.currentSectionBegin)
		skill.finalPeak = &peak
		skill.addPeak(peak)
	}

	return skill.strainPeaks
}

func (skill *VariableLengthStrainSkill) GetCurrentStrainPeaks() []float64 {
	return getStrainPeaks(skill.objectStrains, skill.StrainDecay)
}

func (skill *VariableLengthStrainSkill) DifficultyValue() float64 {
	if skill.stepCalc {
		return skill.difficulty
	}

	return skill.CalculateDifficulty()
}

func (skill *VariableLengthStrainSkill) CountTopWeightedStrains() float64 {
	return skill.diffStrains.GetValue()
}
