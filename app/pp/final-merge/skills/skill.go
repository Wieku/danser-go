package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/pp/xexxar/preprocessing"
	"math"
)

const (
	decayExcessThreshold float64 = 500
)

type Skill struct {
	// How many DifficultyObjects should be kept
	HistoryLength int

	// Keeps track of previous DifficultyObjects for strain section calculations
	Previous []*preprocessing.DifficultyObject

	// Delegate to calculate strain value of skill
	StrainValueOf func(skill *Skill, obj *preprocessing.DifficultyObject) float64

	diff *difficulty.Difficulty

	currentStrain float64

	StarsPerDouble float64

	strains []float64
	times   []float64

	targetFcPrecision float64
	targetFcTime      float64
}

func NewSkill(d *difficulty.Difficulty) *Skill {
	return &Skill{
		HistoryLength:     1,
		diff:              d,
		targetFcPrecision: 0.01,
		targetFcTime:      30 * 60 * 1000,
	}
}

func (skill *Skill) difficultyExponent() float64 {
	return 1.0 / math.Log2(skill.StarsPerDouble)
}

func (skill *Skill) processInternal(current *preprocessing.DifficultyObject) {
	skill.strains = append(skill.strains, skill.StrainValueOf(skill, current))
	skill.times = append(skill.times, current.StartTime)
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
	if len(skill.Previous)-i <= 0 {
		return nil
	}

	return skill.Previous[len(skill.Previous)-1-i]
}

func (skill *Skill) calculateDifficultyValue() float64 {
	difficultyExponent := skill.difficultyExponent()
	SR := 0.0

	for i := 0; i < len(skill.strains); i++ {
		SR += math.Pow(skill.strains[i], difficultyExponent)
	}

	return math.Pow(SR, 1.0/difficultyExponent)
}

func (skill *Skill) DifficultyValue() float64 {
	return skill.fcTimeSkillLevel(skill.calculateDifficultyValue())
}

/// <summary>
/// The probability a player of the given skill full combos a map of the given difficulty.
/// </summary>
/// <param name="skill">The skill level of the player.</param>
/// <param name="difficulty">The difficulty of a range of notes.</param>
func (skill *Skill) fcProbability(skll, difficulty float64) float64 {
	return math.Exp(-math.Pow(difficulty/math.Max(1e-10, skll), skill.difficultyExponent()))
}

/// <summary>
/// Approximates the skill level of a player that can FC a map with the given <paramref name="difficulty"/>,
/// if their probability of success in doing so is equal to <paramref name="probability"/>.
/// </summary>
func (skill *Skill) skillLevel(probability, difficulty float64) float64 {
	return difficulty * math.Pow(-math.Log(probability), -1/skill.difficultyExponent())
}

// A implementation to not overbuff length with a long map. Longer maps = more retries.
func (skill *Skill) expectedTargetTime(totalDifficulty float64) float64 {
	targetTime := 0.0

	for i := 1; i < len(skill.strains); i++ {
		targetTime += math.Min(2000, skill.times[i]-skill.times[i-1]) * (skill.strains[i] / totalDifficulty)
	}

	return targetTime
}

func (skill *Skill) expectedFcTime(skll float64) float64 {
	lastTimestamp := skill.times[0] - 5.0 // time taken to retry map
	fcTime := 0.0

	for i := 0; i < len(skill.strains); i++ {
		dt := skill.times[i] - lastTimestamp
		lastTimestamp = skill.times[i]
		fcTime = (fcTime + dt) / skill.fcProbability(skll, skill.strains[i])
	}

	return fcTime - (skill.times[len(skill.times)-1] - skill.times[0])
}

func (skill *Skill) fcTimeSkillLevel(totalDifficulty float64) float64 {
	lengthEstimate := 0.4 * (skill.times[len(skill.times)-1] - skill.times[0])

	skill.targetFcTime += 45 * math.Max(0, skill.expectedTargetTime(totalDifficulty) - 60000) // for every 30 seconds past 3 mins, add 5 mins to estimated time to FC.

	fcProb := lengthEstimate / skill.targetFcTime

	skillLevel := skill.skillLevel(fcProb, totalDifficulty)

	for i := 0; i < 5; i++ {
		fcTime := skill.expectedFcTime(skillLevel)
		lengthEstimate = fcTime * fcProb
		fcProb = lengthEstimate / skill.targetFcTime
		skillLevel = skill.skillLevel(fcProb, totalDifficulty)

		if math.Abs(fcTime-skill.targetFcTime) < skill.targetFcPrecision*skill.targetFcTime {
			break //enough precision
		}
	}

	return skillLevel
}

func computeDecay(baseDecay, ms float64) float64 {
	decay := 0.0

	if ms < decayExcessThreshold {
		decay = baseDecay
	} else {
		decay = math.Pow(math.Pow(baseDecay, 1000.0/math.Min(ms, decayExcessThreshold)), ms/1000)
	}

	return decay
}
