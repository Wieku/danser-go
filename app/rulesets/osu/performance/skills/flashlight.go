package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
	"math"
)

type Flashlight struct {
	*Skill
}

func NewFlashlightSkill(d *difficulty.Difficulty, experimental bool) *Flashlight {
	skill := &Flashlight{NewSkill(d, experimental)}
	skill.SkillMultiplier = 0.15
	skill.StrainDecayBase = 0.15
	skill.DecayWeight = 1
	skill.HistoryLength = 10
	skill.StrainValueOf = skill.flashlightStrainValue

	return skill
}

func (s *Flashlight) flashlightStrainValue(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	scalingFactor := 52.0 / s.diff.CircleRadiusU
	smallDistNerf := 1.0
	cumulativeStrainTime := 0.0

	result := 0.0

	for i := 0; i < len(s.Previous); i++ {
		previous := s.GetPrevious(i)

		if _, ok := previous.BaseObject.(*objects.Spinner); ok {
			continue
		}

		jumpDistance := float64(current.BaseObject.GetStackedStartPositionMod(s.diff.Mods).Dst(previous.BaseObject.GetStackedEndPositionMod(s.diff.Mods)))

		cumulativeStrainTime += previous.StrainTime

		// We want to nerf objects that can be easily seen within the Flashlight circle radius.
		if i == 0 {
			smallDistNerf = math.Min(1.0, jumpDistance/75.0)
		}

		// We also want to nerf stacks so that only the first object of the stack is accounted for.
		stackNerf := math.Min(1.0, (previous.JumpDistance/scalingFactor)/25.0)

		result += math.Pow(0.8, float64(i)) * stackNerf * scalingFactor * jumpDistance / cumulativeStrainTime
	}

	return math.Pow(smallDistNerf*result, 2.0)
}
