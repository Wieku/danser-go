package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/mutils"
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

type FailListener func()

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

	drains            []drain
	lastTime          int64
	lowerSpinnerDrain bool

	spinners      []*objects.Spinner
	spinnerActive bool

	playing bool

	failListeners []FailListener
}

func NewHealthProcessor(beatMap *beatmap.BeatMap, diff *difficulty.Difficulty, lowerSpinnerDrain bool) *HealthProcessor {
	proc := &HealthProcessor{
		beatMap:           beatMap,
		diff:              diff,
		lowerSpinnerDrain: lowerSpinnerDrain,
	}

	for _, o := range beatMap.HitObjects {
		if s, ok := o.(*objects.Spinner); ok {
			proc.spinners = append(proc.spinners, s)
		}
	}

	return proc
}

func (hp *HealthProcessor) CalculateRate() { //nolint:gocyclo
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

		lastTime := int64(hp.beatMap.HitObjects[0].GetStartTime()) - int64(hp.diff.Preempt)
		fail = false

		breakNumber := 0

		comboTooLowCount := 0

		for i, o := range hp.beatMap.HitObjects {
			localLastTime := lastTime

			breakTime := int64(0)

			if breakCount > 0 && breakNumber < breakCount {
				pause := hp.beatMap.Pauses[breakNumber]
				if pause.GetStartTime() >= float64(localLastTime) && pause.GetEndTime() <= o.GetStartTime() {
					if hp.beatMap.Version < 8 {
						breakTime = int64(pause.Length())
					} else {
						breakTime = int64(pause.GetEndTime()) - localLastTime
					}
					breakNumber++
				}
			}

			hp.Increase(-hp.PassiveDrain*(o.GetStartTime()-float64(lastTime+breakTime)), false)

			lastTime = int64(o.GetEndTime())

			lowestHp = min(lowestHp, hp.Health)

			if hp.Health <= lowestHpEver {
				fail = true
				hp.PassiveDrain *= 0.96

				break
			}

			decr := hp.PassiveDrain * (o.GetEndTime() - o.GetStartTime())
			hpUnder := min(0, hp.Health-decr)

			hp.Increase(-decr, false)

			if s, ok := o.(*objects.Slider); ok {
				for j := 0; j < len(s.TickReverse)+1; j++ {
					hp.AddResult(SliderRepeat)
				}

				for j := 0; j < len(s.TickPoints); j++ {
					hp.AddResult(SliderPoint)
				}
			} else if s, ok := o.(*objects.Spinner); ok {
				requirement := int((s.GetEndTime() - s.GetStartTime()) / 1000 * hp.diff.SpinnerRatio)
				for j := 0; j < requirement; j++ {
					hp.AddResult(SpinnerSpin)
				}
			}

			//noinspection GoBoolExpressions - false positive
			if hpUnder < 0 && hp.Health+hpUnder <= lowestHpEver {
				fail = true
				hp.PassiveDrain *= 0.96

				break
			}

			if i == len(hp.beatMap.HitObjects)-1 || hp.beatMap.HitObjects[i+1].IsNewCombo() {
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

	breakNumber := 0
	lastDrainStart := int64(hp.beatMap.HitObjects[0].GetStartTime()) - int64(hp.diff.Preempt)
	lastDrainEnd := int64(hp.beatMap.HitObjects[0].GetStartTime()) - int64(hp.diff.Preempt)

	for _, o := range hp.beatMap.HitObjects {
		if breakCount > 0 && breakNumber < breakCount {
			pause := hp.beatMap.Pauses[breakNumber]
			if pause.GetStartTime() >= float64(lastDrainEnd) && pause.GetEndTime() <= o.GetStartTime() {
				breakNumber++

				if hp.beatMap.Version < 8 {
					lastDrainEnd = int64(pause.GetStartTime())
				}

				hp.drains = append(hp.drains, drain{lastDrainStart, lastDrainEnd})

				lastDrainStart = int64(o.GetStartTime())
			}
		}

		lastDrainEnd = int64(o.GetEndTime())
	}

	hp.drains = append(hp.drains, drain{lastDrainStart, lastDrainEnd})

	hp.ResetHp()
	hp.playing = true
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

	hp.Increase(hpAdd, true)
}

func (hp *HealthProcessor) Increase(amount float64, fromHitObject bool) {
	hp.HealthUncapped = max(0.0, hp.HealthUncapped+amount)
	hp.Health = mutils.Clamp(hp.Health+amount, 0.0, MaxHp)

	if hp.playing && hp.Health <= 0 && fromHitObject {
		for _, f := range hp.failListeners {
			f()
		}
	}
}

func (hp *HealthProcessor) ReducePassive(amount int64) {
	scale := 1.0
	if hp.spinnerActive && hp.lowerSpinnerDrain {
		scale = 0.25
	}

	if hp.diff.CheckModActive(difficulty.HalfTime) {
		scale *= 0.75
	}

	hp.Increase(-hp.PassiveDrain*float64(amount)*scale, false)
}

func (hp *HealthProcessor) Update(time int64) {
	drainTime := false

	for _, d := range hp.drains {
		if d.start <= time && d.end >= time {
			drainTime = true
			break
		}
	}

	hp.spinnerActive = false
	for _, d := range hp.spinners {
		if d.GetStartTime() > float64(time) {
			break
		}

		if d.GetStartTime() <= float64(time) && float64(time) <= d.GetEndTime() {
			hp.spinnerActive = true
			break
		}
	}

	if drainTime && time > hp.lastTime {
		hp.ReducePassive(time - hp.lastTime)
	}

	hp.lastTime = time
}

func (hp *HealthProcessor) AddFailListener(listener FailListener) {
	hp.failListeners = append(hp.failListeners, listener)
}
