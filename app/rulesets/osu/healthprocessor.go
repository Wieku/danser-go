package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"log"
	"math"
)

const (
	HpMu   = 6.0
	HpKatu = 10.0
	HpGeki = 14.0

	Hp50  = 0.4
	Hp100 = 2.2
	Hp300 = 6.0

	HpSliderTick   = 3.0
	HpSliderRepeat = 4.0

	HpSpinnerSpin  = 1.7
	HpSpinnerBonus = 2.0

	MaxHp = 200.0
)

type drain struct {
	start, end int64
}

type HealthProcessor struct {
	beatMap *beatmap.BeatMap
	diff    *difficulty.Difficulty

	PassiveDrain         float64
	HpMultiplierNormal   float64
	HpMultiplierComboEnd float64

	Health         float64
	HealthUncapped float64

	drains   []drain
	lastTime int64
}

func NewHealthProcessor(beatMap *beatmap.BeatMap, diff *difficulty.Difficulty) *HealthProcessor {
	return &HealthProcessor{
		beatMap: beatMap,
		diff:    diff,
	}
}

func (hp *HealthProcessor) CalculateRate() {
	lowestHpEver := difficulty.DifficultyRate(hp.diff.HPMod, 195, 160, 60)
	lowestHpComboEnd := difficulty.DifficultyRate(hp.diff.HPMod, 198, 170, 80)
	lowestHpEnd := difficulty.DifficultyRate(hp.diff.HPMod, 198, 180, 80)
	hpRecoveryAvailable := difficulty.DifficultyRate(hp.diff.HPMod, 8, 4, 0)

	hp.PassiveDrain = 0.05
	hp.HpMultiplierNormal = 1.0
	hp.HpMultiplierComboEnd = 1.0

	fail := true

	breakCount := len(hp.beatMap.Pauses)

	for fail {
		hp.ResetHp()

		lowestHp := hp.Health

		lastTime := hp.beatMap.HitObjects[0].GetBasicData().StartTime - int64(hp.diff.Preempt)
		fail = false

		breakNumber := 0

		comboTooLowCount := 0

		for i, o := range hp.beatMap.HitObjects {

			localLastTime := lastTime

			breakTime := int64(0)

			if breakCount > 0 && breakNumber < breakCount {
				pause := hp.beatMap.Pauses[breakNumber]
				if pause.GetBasicData().StartTime >= localLastTime && pause.GetBasicData().EndTime <= o.GetBasicData().StartTime {
					//TODO: calculations for beatmap version < 8
					breakTime = pause.GetBasicData().EndTime - localLastTime
					breakNumber++
				}
			}

			hp.Increase(-hp.PassiveDrain * float64(o.GetBasicData().StartTime-lastTime-breakTime))

			lastTime = o.GetBasicData().EndTime

			lowestHp = math.Min(lowestHp, hp.Health)

			if hp.Health <= lowestHpEver {
				fail = true
				hp.PassiveDrain *= 0.96
				break
			}

			hp.Increase(-hp.PassiveDrain * float64(o.GetBasicData().EndTime-o.GetBasicData().StartTime))

			if s, ok := o.(*objects.Slider); ok {
				for j := 0; j < len(s.TickReverse)+1; j++ {
					hp.AddResult(SliderRepeat)
				}

				for j := 0; j < len(s.TickPoints); j++ {
					hp.AddResult(SliderPoint)
				}
			} else if s, ok := o.(*objects.Spinner); ok {
				requirement := int(float64(s.GetBasicData().EndTime-s.GetBasicData().StartTime) / 1000 * hp.diff.SpinnerRatio)
				for j := 0; j < requirement; j++ {
					hp.AddResult(SpinnerSpin)
				}
			}

			if i == len(hp.beatMap.HitObjects)-1 || hp.beatMap.HitObjects[i+1].GetBasicData().NewCombo {
				hp.AddResult(Hit300g)

				if hp.Health < lowestHpComboEnd {
					comboTooLowCount++
					if comboTooLowCount > 2 {
						hp.HpMultiplierComboEnd *= 1.07
						hp.HpMultiplierNormal *= 1.03
						fail = true

						break
					}
				}
			} else {
				hp.AddResult(Hit300)
			}
		}

		if !fail && hp.Health < lowestHpEnd {
			fail = true
			hp.PassiveDrain *= 0.94
			hp.HpMultiplierComboEnd *= 1.01
			hp.HpMultiplierNormal *= 1.01
		}

		recovery := (hp.HealthUncapped - MaxHp) / float64(len(hp.beatMap.HitObjects))
		if !fail && recovery < hpRecoveryAvailable {
			fail = true
			hp.PassiveDrain *= 0.96
			hp.HpMultiplierComboEnd *= 1.02
			hp.HpMultiplierNormal *= 1.01
		}
	}

	log.Println("drainrate", hp.PassiveDrain/2*1000)
	log.Println("normalmult", hp.HpMultiplierNormal)
	log.Println("combomult", hp.HpMultiplierComboEnd)

	breakNumber := 0
	lastDrainStart := hp.beatMap.HitObjects[0].GetBasicData().StartTime - int64(hp.diff.Preempt)
	lastDrainEnd := hp.beatMap.HitObjects[0].GetBasicData().StartTime - int64(hp.diff.Preempt)

	for _, o := range hp.beatMap.HitObjects {

		if breakCount > 0 && breakNumber < breakCount {
			pause := hp.beatMap.Pauses[breakNumber]
			if pause.GetBasicData().StartTime >= lastDrainEnd && pause.GetBasicData().EndTime <= o.GetBasicData().StartTime {
				breakNumber++

				hp.drains = append(hp.drains, drain{lastDrainStart, lastDrainEnd})

				lastDrainStart = o.GetBasicData().StartTime
			}
		}

		lastDrainEnd = o.GetBasicData().EndTime
	}

	hp.drains = append(hp.drains, drain{lastDrainStart, lastDrainEnd})

	hp.ResetHp()
}

func (hp *HealthProcessor) ResetHp() {
	hp.Health = MaxHp
	hp.HealthUncapped = MaxHp
}

func (hp *HealthProcessor) AddResult(result HitResult) {
	normal := result & (^Additions)
	addition := result & Additions

	hpAdd := 0.0

	switch normal {
	case SliderMiss:
		hpAdd += difficulty.DifficultyRate(hp.diff.HPMod, -4.0, -15.0, -28.0)
	case Miss:
		hpAdd += difficulty.DifficultyRate(hp.diff.HPMod, -6.0, -25.0, -40.0)
	case Hit50:
		hpAdd += hp.HpMultiplierNormal * difficulty.DifficultyRate(hp.diff.HPMod, 8*Hp50, Hp50, Hp50)
	case Hit100:
		hpAdd += hp.HpMultiplierNormal * difficulty.DifficultyRate(hp.diff.HPMod, 8*Hp100, Hp100, Hp100)
	case Hit300:
		hpAdd += hp.HpMultiplierNormal * Hp300
	case SliderPoint:
		hpAdd += hp.HpMultiplierNormal * HpSliderTick
	case SliderStart, SliderRepeat, SliderEnd:
		hpAdd += hp.HpMultiplierNormal * HpSliderRepeat
	case SpinnerSpin, SpinnerPoints:
		hpAdd += hp.HpMultiplierNormal * HpSpinnerSpin
	case SpinnerBonus:
		hpAdd += hp.HpMultiplierNormal * HpSpinnerBonus
	}

	switch addition {
	case MuAddition:
		hpAdd += hp.HpMultiplierComboEnd * HpMu
	case KatuAddition:
		hpAdd += hp.HpMultiplierComboEnd * HpKatu
	case GekiAddition:
		hpAdd += hp.HpMultiplierComboEnd * HpGeki
	}

	hp.Increase(hpAdd)
}

func (hp *HealthProcessor) Increase(amount float64) {
	hp.HealthUncapped = math.Max(0.0, hp.HealthUncapped+amount)
	hp.Health = bmath.ClampF64(hp.Health+amount, 0.0, MaxHp)
}

func (hp *HealthProcessor) ReducePassive(amount int64) {
	hp.Increase(-hp.PassiveDrain * float64(amount))
}

func (hp *HealthProcessor) Update(time int64) {
	drainTime := false

	for _, d := range hp.drains {
		if d.start <= time && d.end >= time {
			drainTime = true
			break
		}
	}

	if drainTime {
		hp.ReducePassive(time - hp.lastTime)
	}

	hp.lastTime = time
}
