package difficulty

import "math"

type Difficulty struct {
	hpDrain, cs, od, ar float64
	preempt, fadeIn     float64
	circleRadius        float64
	Mods                Modifier
	Hit50               int64
	Hit100              int64
	Hit300              int64
}

func NewDifficulty(hpDrain, cs, od, ar float64) *Difficulty {
	diff := new(Difficulty)
	diff.hpDrain = hpDrain
	diff.cs = cs
	diff.od = od
	diff.ar = ar
	diff.calculate()
	return diff
}

func (diff *Difficulty) calculate() {
	hpDrain, cs, od, ar := diff.hpDrain, diff.cs, diff.od, diff.ar

	if diff.Mods&HardRock > 0 {
		ar = math.Min(ar*1.4, 10)
		cs *= 1.3
		od *= 1.4
		hpDrain *= 1.4
	}

	if diff.Mods&Easy > 0 {
		ar /= 2
		cs /= 2
		od /= 2
		hpDrain /= 2
	}

	diff.circleRadius = 32 * (1.0 - 0.7*(cs-5)/5)
	diff.preempt = difficultyRate(ar, 1800, 1200, 450)
	diff.fadeIn = difficultyRate(ar, 1200, 800, 300)
	diff.Hit50 = int64(150 + 50 * (5 - od) / 5)
	diff.Hit100	= int64(100 + 40 * (5 - od) / 5)
	diff.Hit300	 = int64(50 + 30 * (5 - od) / 5)
}

func (diff *Difficulty) SetMods(mods Modifier) {
	diff.Mods = mods
	diff.calculate()
}

func difficultyRate(diff, min, mid, max float64) float64 {
	if diff > 5 {
		return mid + (max-mid)*(diff-5)/5
	}
	if diff < 5 {
		return mid - (mid-min)*(5-diff)/5
	}
	return mid
}
