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

	rhythmMultiplier float64 = 0.75
	historyTimeMax   float64 = 5000
)

type SpeedSkill struct {
	*Skill

	CurrentRhythm float64
}

func NewSpeedSkill(d *difficulty.Difficulty, experimental bool) *SpeedSkill {
	skill := &SpeedSkill{
		Skill: NewSkill(d, experimental),
	}

	skill.SkillMultiplier = 1400
	skill.StrainDecayBase = 0.3
	skill.ReducedSectionCount = 5
	skill.DifficultyMultiplier = 1.04
	skill.StrainValueOf = skill.speedStrainValue

	if experimental {
		skill.SkillMultiplier = 1375
		skill.HistoryLength = 32
		skill.StrainBonusOf = skill.speedStrainBonus
		skill.CalculateInitialStrain = func(time float64) float64 {
			return (skill.CurrentStrain * skill.CurrentRhythm) * skill.strainDecay(time-skill.GetPrevious(0).StartTime)
		}
	}

	return skill
}

func (s *SpeedSkill) speedStrainValue(current *preprocessing.DifficultyObject) float64 {
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

		//deltaTime = strainTime

		speedBonus := 1.0

		if strainTime < MinSpeedBonus {
			speedBonus = 1 + 0.75*math.Pow((MinSpeedBonus-strainTime)/SpeedBalancingFactor, 2.0)
		}

		return (speedBonus + speedBonus*math.Pow(distance/SingleSpacingThreshold, 3.5)) / strainTime
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

func (s *SpeedSkill) speedStrainBonus(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		s.CurrentRhythm = 0
		return 0
	}

	greatWindow := float64(s.diff.Hit300) / s.diff.Speed

	previousIslandSize := 0
	rhythmComplexitySum := 0.0
	islandSize := 1
	startRatio := 0.0 // store the ratio of the current start of an island to buff for tighter rhythms

	firstDeltaSwitch := false

	for i := len(s.Previous) - 2; i > 0; i-- {
		currObj := s.GetPrevious(i - 1)
		prevObj := s.GetPrevious(i)
		lastObj := s.GetPrevious(i + 1)

		currHistoricalDecay := math.Max(0, historyTimeMax-(current.StartTime-currObj.StartTime)) / historyTimeMax // scales note 0 to 1 from history to now

		if currHistoricalDecay != 0 {
			currHistoricalDecay = math.Min(float64(len(s.Previous)-i)/float64(len(s.Previous)), currHistoricalDecay) // either we're limited by time or limited by object count.

			currDelta := currObj.StrainTime
			prevDelta := prevObj.StrainTime
			lastDelta := lastObj.StrainTime
			currRatio := 1.0 + 6.0*math.Min(0.5, math.Pow(math.Sin(math.Pi/(math.Min(prevDelta, currDelta)/math.Max(prevDelta, currDelta))), 2)) // fancy function to calculate rhythmbonuses.

			windowPenalty := math.Min(1, math.Max(0, math.Abs(prevDelta-currDelta)-greatWindow*0.6)/(greatWindow*0.6))

			windowPenalty = math.Min(1, windowPenalty)

			effectiveRatio := windowPenalty * currRatio

			if firstDeltaSwitch {
				if !(prevDelta > 1.25*currDelta || prevDelta*1.25 < currDelta) {
					if islandSize < 7 {
						islandSize++ // island is still progressing, count size.
					}
				} else {
					if _, ok := currObj.BaseObject.(*preprocessing.LazySlider); ok { // bpm change is into slider, this is easy acc window
						effectiveRatio *= 0.125
					}

					if _, ok := prevObj.BaseObject.(*preprocessing.LazySlider); ok { // bpm change was from a slider, this is easier typically than circle -> circle
						effectiveRatio *= 0.25
					}

					if previousIslandSize == islandSize { // repeated island size (ex: triplet -> triplet)
						effectiveRatio *= 0.25
					}

					if previousIslandSize%2 == islandSize%2 { // repeated island polartiy (2 -> 4, 3 -> 5)
						effectiveRatio *= 0.50
					}

					if lastDelta > prevDelta+10 && prevDelta > currDelta+10 { // previous increase happened a note ago, 1/1->1/2-1/4, dont want to buff this.
						effectiveRatio *= 0.125
					}

					rhythmComplexitySum += math.Sqrt(effectiveRatio*startRatio) * currHistoricalDecay * math.Sqrt(4+float64(islandSize)) / 2 * math.Sqrt(4+float64(previousIslandSize)) / 2

					startRatio = effectiveRatio

					previousIslandSize = islandSize // log the last island size.

					if prevDelta*1.25 < currDelta { // we're slowing down, stop counting
						firstDeltaSwitch = false // if we're speeding up, this stays true and we keep counting island size.
					}

					islandSize = 1
				}
			} else if prevDelta > 1.25*currDelta { // we want to be speeding up.
				// Begin counting island until we change speed again.
				firstDeltaSwitch = true
				startRatio = effectiveRatio
				islandSize = 1
			}
		}
	}

	s.CurrentRhythm = math.Sqrt(4+rhythmComplexitySum*rhythmMultiplier) / 2 //produces multiplier that can be applied to strain. range [1, infinity) (not really though)

	return s.CurrentRhythm
}
