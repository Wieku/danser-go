package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	lzMinimumHealthError = 0.01
	lzMinHealthTarget    = 0.99
	lzMidHealthTarget    = 0.9
	lzMaxHealthTarget    = 0.4
)

type HealthProcessorV2 struct {
	beatMap *beatmap.BeatMap
	player  *difficultyPlayer

	passiveDrain float64

	health float64

	drains   []drainPeriod
	lastTime int64

	failListeners []FailListener

	countGood int
	countBad  int
}

func NewHealthProcessorV2(beatMap *beatmap.BeatMap, player *difficultyPlayer) *HealthProcessorV2 {
	hp := &HealthProcessorV2{
		beatMap: beatMap,
		player:  player,
	}

	hp.calculateDrainPeriods()

	return hp
}

func (hp *HealthProcessorV2) calculateDrainPeriods() {
	breakCount := len(hp.beatMap.Pauses)

	breakNumber := 0
	lastDrainStart := int64(hp.beatMap.HitObjects[0].GetStartTime())
	lastDrainEnd := int64(hp.beatMap.HitObjects[0].GetStartTime())

	for _, o := range hp.beatMap.HitObjects {
		if breakCount > 0 && breakNumber < breakCount {
			pause := hp.beatMap.Pauses[breakNumber]
			if pause.GetStartTime() >= float64(lastDrainEnd) && pause.GetEndTime() <= o.GetStartTime() {
				breakNumber++

				hp.drains = append(hp.drains, drainPeriod{lastDrainStart, lastDrainEnd})

				lastDrainStart = int64(o.GetStartTime())
			}
		}

		lastDrainEnd = int64(o.GetEndTime())
	}

	hp.drains = append(hp.drains, drainPeriod{lastDrainStart, lastDrainEnd})
}

func (hp *HealthProcessorV2) CalculateRate() { //nolint:gocyclo
	type healthIncrease struct {
		time     float64
		increase float64
	}

	targetMinimumHealth := mutils.Clamp(difficulty.DifficultyRate(hp.player.diff.HPMod, lzMinHealthTarget, lzMidHealthTarget, lzMaxHealthTarget), 0, 1)

	breakCount := len(hp.beatMap.Pauses)

	healthIncreases := make([]healthIncrease, 0)

	for _, o := range hp.beatMap.HitObjects {
		time := o.GetStartTime()

		if _, ok := o.(*objects.Spinner); ok {
			time = o.GetEndTime()
		}

		s, ok := o.(*objects.Slider)

		hr := Hit300
		if ok && hp.player.lzNoSliderAcc {
			hr = SliderStart
		}

		healthIncreases = append(healthIncreases, healthIncrease{
			time:     time,
			increase: hp.getHPResult(hr),
		})

		if ok {
			for j := 0; j < len(s.ScorePointsLazer); j++ {
				sc := s.ScorePoints[j]

				result := SliderPoint

				if j == len(s.ScorePointsLazer)-1 {
					result = SliderEnd
				} else if sc.IsReverse {
					result = SliderRepeat
				}

				healthIncreases = append(healthIncreases, healthIncrease{
					time:     sc.Time,
					increase: hp.getHPResult(result),
				})
			}

			if hp.player.lzNoSliderAcc {
				healthIncreases = append(healthIncreases, healthIncrease{
					time:     s.GetEndTime(),
					increase: hp.getHPResult(Hit300),
				})
			}
		}
	}

	adjustment := 1
	hp.passiveDrain = 1.0

	// Although we expect the following loop to converge within 30 iterations (health within 1/2^31 accuracy of the target),
	// we'll still keep a safety measure to avoid infinite loops by detecting overflows.
	for {
		currentHealth := 1.0
		lowestHealth := 1.0
		currentBreak := 0

		for i := 0; i < len(healthIncreases); i++ {
			currentTime := healthIncreases[i].time

			lastTime := float64(hp.drains[0].start)
			if i > 0 {
				lastTime = healthIncreases[i-1].time
			}

			for currentBreak < breakCount && hp.beatMap.Pauses[currentBreak].GetEndTime() <= currentTime {
				// If two hitobjects are separated by a break period, there is no drain for the full duration between the hitobjects.
				// This differs from legacy (version < 8) beatmaps which continue draining until the break section is entered,
				// but this shouldn't have a noticeable impact in practice.
				lastTime = currentTime
				currentBreak++
			}

			// Apply health adjustments
			currentHealth -= (currentTime - lastTime) * hp.passiveDrain
			lowestHealth = min(lowestHealth, currentHealth)
			currentHealth = min(1, currentHealth+healthIncreases[i].increase)

			// Common scenario for when the drain rate is definitely too harsh
			if lowestHealth < 0 {
				break
			}
		}

		// Stop if the resulting health is within a reasonable offset from the target
		if mutils.Abs(lowestHealth-targetMinimumHealth) <= lzMinimumHealthError {
			break
		}

		// This effectively works like a binary search - each iteration the search space moves closer to the target, but may exceed it.
		adjustment *= 2
		hp.passiveDrain += 1.0 / float64(adjustment) * mutils.Signum(lowestHealth-targetMinimumHealth)
	}
}

func (hp *HealthProcessorV2) ResetHp() {
	hp.health = 1
}

func (hp *HealthProcessorV2) AddResult(result JudgementResult) {
	normal := result.HitResult & (^Additions)

	hpRes := hp.getHPResult(normal)
	if normal == SliderMiss && result.ComboResult == Hold { //Missed slider ends don't decrease hp
		hpRes = 0
	}

	if (result.fromSliderFinish || result.object.GetObject().GetType() != objects.SLIDER) && result.object.GetObject().IsNewCombo() {
		hp.countGood = 0
		hp.countBad = 0
	}

	switch normal {
	case SliderMiss, Hit100:
		hp.countGood++
	case Hit50, Miss:
		hp.countBad++
	}

	if (result.HitResult&(BaseHits|SliderFinish)) > 0 && (result.fromSliderFinish || result.object.GetObject().GetType() != objects.SLIDER) && result.object.GetObject().IsLastCombo() {
		if hp.countGood == 0 && hp.countBad == 0 {
			hpRes += 0.07
		} else if hp.countBad == 0 {
			hpRes += 0.05
		} else {
			hpRes += 0.03
		}
	}

	hp.Increase(hpRes, true)
}

func (hp *HealthProcessorV2) getHPResult(result HitResult) float64 {
	normal := result & (^Additions)

	switch normal {
	case SliderMiss:
		return difficulty.DifficultyRate(hp.player.diff.HPMod, -0.02, -0.075, -0.14)
	case Miss:
		return difficulty.DifficultyRate(hp.player.diff.HPMod, -0.03, -0.125, -0.2)
	case Hit50:
		return 0.002
	case Hit100:
		return 0.011
	case Hit300:
		return 0.03
	case SliderPoint:
		return 0.015
	case SliderStart, SliderRepeat, LegacySliderEnd, SliderEnd:
		return 0.02
	case SpinnerSpin, SpinnerPoints:
		return 0.0085
	case SpinnerBonus:
		return 0.01
	}

	return 0
}

func (hp *HealthProcessorV2) Increase(amount float64, fromHitObject bool) {
	hp.health = mutils.Clamp(hp.health+amount, 0.0, 1)

	if hp.health <= 0 && fromHitObject {
		for _, f := range hp.failListeners {
			f()
		}
	}
}

func (hp *HealthProcessorV2) IncreaseRelative(amount float64, fromHitObject bool) {
	hp.Increase(amount, fromHitObject)
}

func (hp *HealthProcessorV2) Update(time int64) {
	drainTime := false

	for _, d := range hp.drains {
		if d.start <= time && d.end >= time {
			drainTime = true
			break
		}
	}

	if drainTime && time > hp.lastTime {
		hp.Increase(-hp.passiveDrain*float64(time-hp.lastTime), false)
	}

	hp.lastTime = time
}

func (hp *HealthProcessorV2) AddFailListener(listener FailListener) {
	hp.failListeners = append(hp.failListeners, listener)
}

func (hp *HealthProcessorV2) GetHealth() float64 {
	return hp.health
}

func (hp *HealthProcessorV2) GetDrainRate() float64 {
	return hp.passiveDrain
}
