package skills

import (
	"math"
	"slices"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/evaluators"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	aimStrainDecayBase       = 0.2
	aimReducedSectionTime    = 4000
	aimReducedStrainBaseline = 0.727
)

type AimSkill struct {
	*VariableLengthStrainSkill
	withSliders   bool
	currentStrain float64

	diffSliders *putils.LogisticSum
	topSliders  *putils.LogisticSum
}

func NewAimSkill(d *difficulty.Difficulty, withSliders, stepCalc bool) *AimSkill {
	skill := &AimSkill{
		VariableLengthStrainSkill: NewVariableLengthStrainSkill(d, stepCalc),
		withSliders:               withSliders,
	}

	skill.StrainValueOf = skill.aimStrainValue
	skill.PostProcess = skill.postProcess
	skill.CalculateInitialStrain = skill.aimInitialStrain
	skill.StrainDecay = skill.strainDecay
	skill.CalculateDifficulty = skill.aimDifficulty

	skill.diffSliders = putils.NewLogisticSum(stepCalc, 6, 1, 1, func(previous, current float64) bool {
		return current > previous
	}, func(strains []float64) float64 {
		return slices.Max(strains) / 12
	})

	skill.topSliders = putils.NewLogisticSum(stepCalc, 0.88, 10, 1.1, func(previous, current float64) bool {
		return current != previous
	}, func(strains []float64) float64 {
		return skill.DifficultyValue() * (1 - skill.DecayWeight)
	})

	return skill
}

func (skill *AimSkill) strainDecay(ms float64) float64 {
	return math.Pow(aimStrainDecayBase, ms/1000)
}

func (skill *AimSkill) aimInitialStrain(time float64, current *preprocessing.DifficultyObject) float64 {
	return skill.currentStrain * skill.strainDecay(time-current.Previous(0).StartTime)
}

func (skill *AimSkill) aimStrainValue(current *preprocessing.DifficultyObject) float64 {
	if current.Diff.CheckModActive(difficulty.Relax2) {
		return 0
	}

	decay := skill.strainDecay(current.AdjustedDeltaTime)

	skill.currentStrain *= decay
	skill.currentStrain += skill.calculateAdjustedDifficulty(current) * (1 - decay)

	if current.IsSlider {
		skill.diffSliders.AddStrain(skill.currentStrain)
		skill.topSliders.AddStrain(skill.currentStrain)
	}

	return skill.currentStrain
}

func (skill *AimSkill) calculateAdjustedDifficulty(current *preprocessing.DifficultyObject) float64 {
	const aimSkillMultiplierSnap = 70.9
	const aimSkillMultiplierAgility = 2.35
	const aimSkillMultiplierFlow = 242.0

	snapDifficulty := evaluators.EvaluateSnapAim(current, skill.withSliders) * aimSkillMultiplierSnap
	agilityDifficulty := evaluators.EvaluateAgility(current) * aimSkillMultiplierAgility
	flowDifficulty := evaluators.EvaluateFlowAim(current, skill.withSliders) * aimSkillMultiplierFlow

	totalDifficulty := skill.calculateAimTotalValue(snapDifficulty, agilityDifficulty, flowDifficulty)

	totalDifficulty *= 0.985 + putils.Powi(max(0, current.Diff.ODReal), 2)/4000

	return totalDifficulty
}

func (skill *AimSkill) calculateAimTotalValue(snapDifficulty, agilityDifficulty, flowDifficulty float64) float64 {
	const aimSkillMultiplierTotal = 1.12
	const combinedSnapNormExponent = 1.2

	// We compare flow to combined snap and agility because snap by itself doesn't have enough difficulty to be above flow on streams
	// Agility on the other hand is supposed to measure the rate of cursor velocity changes while snapping
	// So snapping every circle on a stream requires an enormous amount of agility at which point it's easier to flow
	combinedSnapDifficulty := putils.Norm(combinedSnapNormExponent, snapDifficulty, agilityDifficulty)

	pSnap := calculateSnapFlowProbability(flowDifficulty / combinedSnapDifficulty)
	pFlow := 1 - pSnap

	if skill.diff.CheckModActive(difficulty.TouchDevice) {
		// we don't adjust agility here since agility represents TD difficulty in a decent enough way
		snapDifficulty = math.Pow(snapDifficulty, 0.89)
		combinedSnapDifficulty = putils.Norm(combinedSnapNormExponent, snapDifficulty, agilityDifficulty)
	}

	if skill.diff.CheckModActive(difficulty.Relax) {
		combinedSnapDifficulty *= 0.75
		flowDifficulty *= 0.6
	}

	totalDifficulty := combinedSnapDifficulty*pSnap + flowDifficulty*pFlow

	totalStrain := totalDifficulty * aimSkillMultiplierTotal

	return totalStrain
}

// A function that turns the ratio of snap : flow into the probability of snapping/flowing
// It has the constraints:
// P(snap) + P(flow) = 1 (the object is always either snapped or flowed)
// P(snap) = f(snap/flow), P(flow) = f(flow/snap) (ie snap and flow are symmetric and reversible)
// Therefore: f(x) + f(1/x) = 1
// 0 <= f(x) <= 1 (cannot have negative or greater than 100% probability of snapping or flowing)
// This logistic function is a solution, which fits nicely with the general idea of interpolation and provides a tuneable constant
func calculateSnapFlowProbability(ratio float64) float64 {
	const k = 7.27

	if ratio == 0 {
		return 0
	}

	if math.IsNaN(ratio) {
		return 1
	}

	return putils.LogisticE(-k*math.Log(ratio), 1)
}

func (skill *AimSkill) postProcess(current *preprocessing.DifficultyObject, strain float64, diffValue float64) {
	if current.IsSlider {
		skill.diffSliders.ProcessLastStrain(strain / 12)
	}

	skill.topSliders.ProcessLastStrain(diffValue * (1 - skill.DecayWeight))
}

func (skill *AimSkill) aimDifficulty() float64 {
	diffValue := 0.0
	time := 0.0

	for _, strain := range skill.getReducedStrainPeaks() {
		startTime := time
		endTime := time + strain.SectionLength/skill.MaxSectionLength

		weight := math.Pow(skill.DecayWeight, startTime) - math.Pow(skill.DecayWeight, endTime)

		diffValue += strain.Value * weight
		time = endTime
	}

	return diffValue / (1 - skill.DecayWeight)
}

func (skill *AimSkill) getReducedStrainPeaks() []StrainPeak {
	strains := slices.DeleteFunc(slices.Clone(skill.getCurrentStrainPeaks()), func(peak StrainPeak) bool {
		return peak.Value <= 0
	})

	const chunkSize = 20
	time := 0.0
	skipCount := 0

	for len(strains) > skipCount && time < aimReducedSectionTime {
		strain := strains[skipCount]

		for addedTime := 0.0; addedTime < strain.SectionLength; addedTime += chunkSize {
			scale := math.Log10(mutils.Lerp(1.0, 10.0, mutils.Clamp((time+addedTime)/aimReducedSectionTime, 0.0, 1.0)))

			strains = append(strains, newStrainPeak(
				strain.Value*mutils.Lerp(aimReducedStrainBaseline, 1.0, scale),
				min(chunkSize, strain.SectionLength-addedTime),
			))
		}

		time += strain.SectionLength
		skipCount++
	}

	strains = strains[skipCount:]
	slices.SortFunc(strains, func(a, b StrainPeak) int {
		if a.Value > b.Value {
			return -1
		}

		if a.Value < b.Value {
			return 1
		}

		return 0
	})

	return strains
}

func (skill *AimSkill) GetDifficultSliders() float64 {
	return skill.diffSliders.GetValue()
}

func (skill *AimSkill) CountTopWeightedSliders() float64 {
	return skill.topSliders.GetValue()
}
