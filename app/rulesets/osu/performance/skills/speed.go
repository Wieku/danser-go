package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	SingleSpacingThreshold float64 = 125.0
	MinSpeedBonus          float64 = 75.0 // ~200BPM
	MaxSpeedBonus          float64 = 45.0
	SpeedBalancingFactor   float64 = 40
	SpeedAngleBonusBegin   float64 = 5 * math.Pi / 6

	PiOver2 float64 = math.Pi / 2
	PiOver4 float64 = math.Pi / 4
)

func NewSpeedSkill(d *difficulty.Difficulty, experimental bool) *Skill {
	skill := NewSkill(d, experimental)
	skill.SkillMultiplier = 1400
	skill.StrainDecayBase = 0.3
	skill.ReducedSectionCount = 5
	skill.DifficultyMultiplier = 1.04
	skill.StrainValueOf = speedStrainValue

	return skill
}

func speedStrainValue(s *Skill, current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	distance := math.Min(SingleSpacingThreshold, current.TravelDistance+current.JumpDistance)
	deltaTime := math.Max(MaxSpeedBonus, current.DeltaTime)
	strainTime := current.StrainTime

	if s.Experimental {
		previous := s.GetPrevious(0)
		greatWindowFull := float64(s.diff.Hit300) / s.diff.Speed * 2
		speedWindowRatio := strainTime / greatWindowFull

		// Aim to nerf cheesy rhythms (Very fast consecutive doubles with large deltatimes between)
		if previous != nil && strainTime < greatWindowFull && previous.StrainTime > strainTime {
			strainTime = mutils.LerpF64(previous.StrainTime, strainTime, speedWindowRatio)
		}

		// Cap deltatime to the OD 300 hitwindow.
		// 0.93 is derived from making sure 260bpm OD8 streams aren't nerfed harshly, whilst 0.92 limits the effect of the cap.
		strainTime /= mutils.ClampF64((strainTime/greatWindowFull)/0.93, 0.92, 1)

		deltaTime = strainTime
	}

	speedBonus := 1.0

	if deltaTime < MinSpeedBonus {
		speedBonus = 1 + math.Pow((MinSpeedBonus-deltaTime)/SpeedBalancingFactor, 2.0)
	}

	angleBonus := 1.0
	if !math.IsNaN(current.Angle) && current.Angle < SpeedAngleBonusBegin {
		angleBonus = 1 + math.Pow(math.Sin(1.5*(SpeedAngleBonusBegin-current.Angle)), 2)/3.57

		if current.Angle < PiOver2 {
			angleBonus = 1.28
			if distance < AngleBonusScale && current.Angle < PiOver4 {
				angleBonus += (1.0 - angleBonus) *
					math.Min((AngleBonusScale-distance)/10.0, 1.0)
			} else if distance < AngleBonusScale {
				angleBonus += (1.0 - angleBonus) *
					math.Min((AngleBonusScale-distance)/10.0, 1.0) *
					math.Sin((PiOver2-current.Angle)/PiOver4)
			}
		}
	}

	return ((1.0 + (speedBonus-1.0)*0.75) * angleBonus *
		(0.95 + speedBonus*math.Pow(distance/SingleSpacingThreshold, 3.5))) /
		strainTime
}
