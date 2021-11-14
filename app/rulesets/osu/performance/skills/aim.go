package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	wideAngleMultiplier      float64 = 1.5
	acuteAngleMultiplier     float64 = 2.0
	sliderMultiplier         float64 = 1.5
	velocityChangeMultiplier float64 = 0.75
)

type AimSkill struct {
	*Skill
	withSliders bool
}

func NewAimSkill(d *difficulty.Difficulty, withSliders, experimental bool) *AimSkill {
	skill := &AimSkill{Skill: NewSkill(d, experimental), withSliders: withSliders}

	skill.SkillMultiplier = 23.25
	skill.StrainDecayBase = 0.15
	skill.HistoryLength = 2
	skill.StrainValueOf = skill.aimStrainValue

	return skill
}

func (skill *AimSkill) aimStrainValue(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok || len(skill.Previous) <= 1 {
		return 0
	}
	if _, ok := skill.GetPrevious(0).BaseObject.(*objects.Spinner); ok {
		return 0
	}

	osuCurrObj := current
	osuLastObj := skill.GetPrevious(0)
	osuLastLastObj := skill.GetPrevious(1)

	// Calculate the velocity to the current hitobject, which starts with a base distance / time assuming the last object is a hitcircle.
	currVelocity := osuCurrObj.JumpDistance / osuCurrObj.StrainTime

	// But if the last object is a slider, then we extend the travel velocity through the slider into the current object.
	if _, ok := osuLastObj.BaseObject.(*preprocessing.LazySlider); ok && skill.withSliders {
		movementVelocity := osuCurrObj.MovementDistance / osuCurrObj.MovementTime // calculate the movement velocity from slider end to current object
		travelVelocity := osuCurrObj.TravelDistance / osuCurrObj.TravelTime       // calculate the slider velocity from slider head to slider end.

		currVelocity = math.Max(currVelocity, movementVelocity+travelVelocity) // take the larger total combined velocity.
	}

	// As above, do the same for the previous hitobject.
	prevVelocity := osuLastObj.JumpDistance / osuLastObj.StrainTime

	if _, ok := osuLastLastObj.BaseObject.(*preprocessing.LazySlider); ok && skill.withSliders {
		movementVelocity := osuLastObj.MovementDistance / osuLastObj.MovementTime
		travelVelocity := osuLastObj.TravelDistance / osuLastObj.TravelTime

		prevVelocity = math.Max(prevVelocity, movementVelocity+travelVelocity)
	}

	wideAngleBonus := 0.0
	acuteAngleBonus := 0.0
	sliderBonus := 0.0
	velocityChangeBonus := 0.0

	aimStrain := currVelocity // Start strain with regular velocity.

	if math.Max(osuCurrObj.StrainTime, osuLastObj.StrainTime) < 1.25*math.Min(osuCurrObj.StrainTime, osuLastObj.StrainTime) { // If rhythms are the same.

		if !math.IsNaN(osuCurrObj.Angle) && !math.IsNaN(osuLastObj.Angle) && !math.IsNaN(osuLastLastObj.Angle) {
			currAngle := osuCurrObj.Angle
			lastAngle := osuLastObj.Angle
			lastLastAngle := osuLastLastObj.Angle

			// Rewarding angles, take the smaller velocity as base.
			angleBonus := math.Min(currVelocity, prevVelocity)

			wideAngleBonus = calcWideAngleBonus(currAngle)
			acuteAngleBonus = calcAcuteAngleBonus(currAngle)

			if osuCurrObj.StrainTime > 100 { // Only buff deltaTime exceeding 300 bpm 1/2.
				acuteAngleBonus = 0
			} else {
				acuteAngleBonus *= calcAcuteAngleBonus(lastAngle) * // Multiply by previous angle, we don't want to buff unless this is a wiggle type pattern.
					math.Min(angleBonus, 125/osuCurrObj.StrainTime) * // The maximum velocity we buff is equal to 125 / strainTime
					math.Pow(math.Sin(math.Pi/2*math.Min(1, (100-osuCurrObj.StrainTime)/25)), 2) * // scale buff from 150 bpm 1/4 to 200 bpm 1/4
					math.Pow(math.Sin(math.Pi/2*(mutils.ClampF64(osuCurrObj.JumpDistance, 50, 100)-50)/50), 2) // Buff distance exceeding 50 (radius) up to 100 (diameter).
			}

			// Penalize wide angles if they're repeated, reducing the penalty as the lastAngle gets more acute.
			wideAngleBonus *= angleBonus * (1 - math.Min(wideAngleBonus, math.Pow(calcWideAngleBonus(lastAngle), 3)))
			// Penalize acute angles if they're repeated, reducing the penalty as the lastLastAngle gets more obtuse.
			acuteAngleBonus *= 0.5 + 0.5*(1-math.Min(acuteAngleBonus, math.Pow(calcAcuteAngleBonus(lastLastAngle), 3)))
		}
	}

	if math.Max(prevVelocity, currVelocity) != 0 {
		// We want to use the average velocity over the whole object when awarding differences, not the individual jump and slider path velocities.
		prevVelocity = (osuLastObj.JumpDistance + osuLastObj.TravelDistance) / osuLastObj.StrainTime
		currVelocity = (osuCurrObj.JumpDistance + osuCurrObj.TravelDistance) / osuCurrObj.StrainTime

		// Scale with ratio of difference compared to 0.5 * max dist.
		distRatio := math.Pow(math.Sin(math.Pi/2*math.Abs(prevVelocity-currVelocity)/math.Max(prevVelocity, currVelocity)), 2)

		// Reward for % distance up to 125 / strainTime for overlaps where velocity is still changing.
		overlapVelocityBuff := math.Min(125/math.Min(osuCurrObj.StrainTime, osuLastObj.StrainTime), math.Abs(prevVelocity-currVelocity))

		// Reward for % distance slowed down compared to previous, paying attention to not award overlap
		nonOverlapVelocityBuff := math.Abs(prevVelocity-currVelocity) *
			// do not award overlap
			math.Pow(math.Sin(math.Pi/2*math.Min(1, math.Min(osuCurrObj.JumpDistance, osuLastObj.JumpDistance)/100)), 2)

		// Choose the largest bonus, multiplied by ratio.
		velocityChangeBonus = math.Max(overlapVelocityBuff, nonOverlapVelocityBuff) * distRatio

		// Penalize for rhythm changes.
		velocityChangeBonus *= math.Pow(math.Min(osuCurrObj.StrainTime, osuLastObj.StrainTime)/math.Max(osuCurrObj.StrainTime, osuLastObj.StrainTime), 2)
	}

	if osuCurrObj.TravelTime != 0 {
		// Reward sliders based on velocity.
		sliderBonus = osuCurrObj.TravelDistance / osuCurrObj.TravelTime
	}

	// Add in acute angle bonus or wide angle bonus + velocity change bonus, whichever is larger.
	aimStrain += math.Max(acuteAngleBonus*acuteAngleMultiplier, wideAngleBonus*wideAngleMultiplier+velocityChangeBonus*velocityChangeMultiplier)

	if skill.withSliders {
		// Add in additional slider velocity bonus.
		aimStrain += sliderBonus * sliderMultiplier
	}

	return aimStrain
}

func calcWideAngleBonus(angle float64) float64 {
	return math.Pow(math.Sin(3.0/4*(math.Min(5.0/6*math.Pi, math.Max(math.Pi/6, angle))-math.Pi/6)), 2)
}

func calcAcuteAngleBonus(angle float64) float64 {
	return 1 - calcWideAngleBonus(angle)
}
