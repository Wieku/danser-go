package difficulty

import (
	"fmt"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const (
	HitFadeIn     = 400.0
	HitFadeOut    = 240.0
	HittableRange = 400.0
	ResultFadeIn  = 120.0
	ResultFadeOut = 600.0
	PostEmpt      = 500.0
)

type Difficulty struct {
	ar float64
	od float64
	cs float64
	hp float64

	baseAR float64
	baseOD float64
	baseCS float64
	baseHP float64

	PreemptU      float64
	Preempt       float64
	TimeFadeIn    float64
	CircleRadiusU float64
	CircleRadius  float64
	Mods          Modifier

	Hit50U  float64
	Hit100U float64
	Hit300U float64

	Hit50  int64
	Hit100 int64
	Hit300 int64

	HPMod        float64
	SpinnerRatio float64
	Speed        float64

	ARReal      float64
	ODReal      float64
	CustomSpeed float64
}

func NewDifficulty(hp, cs, od, ar float64) *Difficulty {
	diff := new(Difficulty)

	diff.baseHP = hp
	diff.hp = hp

	diff.baseCS = cs
	diff.cs = cs

	diff.baseOD = od
	diff.od = od

	diff.baseAR = ar
	diff.ar = ar

	diff.CustomSpeed = 1

	diff.calculate()

	return diff
}

func (diff *Difficulty) calculate() {
	hpDrain, cs, od, ar := diff.hp, diff.cs, diff.od, diff.ar

	if diff.Mods&HardRock > 0 {
		ar = math.Min(ar*1.4, 10)
		cs = math.Min(cs*1.3, 10)
		od = math.Min(od*1.4, 10)
		hpDrain = math.Min(hpDrain*1.4, 10)
	}

	if diff.Mods&Easy > 0 {
		ar /= 2
		cs /= 2
		od /= 2
		hpDrain /= 2
	}

	diff.HPMod = hpDrain

	diff.CircleRadiusU = DifficultyRate(cs, 54.4, 32, 9.6)
	diff.CircleRadius = diff.CircleRadiusU * 1.00041 //some weird allowance osu has

	diff.PreemptU = DifficultyRate(ar, 1800, 1200, 450)
	diff.Preempt = math.Floor(diff.PreemptU)

	diff.TimeFadeIn = HitFadeIn * math.Min(1, diff.PreemptU/450)

	diff.Hit50U = DifficultyRate(od, 200, 150, 100)
	diff.Hit100U = DifficultyRate(od, 140, 100, 60)
	diff.Hit300U = DifficultyRate(od, 80, 50, 20)

	diff.Hit50 = int64(diff.Hit50U)
	diff.Hit100 = int64(diff.Hit100U)
	diff.Hit300 = int64(diff.Hit300U)

	diff.SpinnerRatio = DifficultyRate(od, 3, 5, 7.5)
	diff.Speed = 1.0 / diff.GetModifiedTime(1)

	diff.ARReal = DiffFromRate(diff.GetModifiedTime(diff.PreemptU), 1800, 1200, 450)
	diff.ODReal = DiffFromRate(diff.GetModifiedTime(diff.Hit300U), 80, 50, 20)
}

func (diff *Difficulty) SetMods(mods Modifier) {
	diff.Mods = mods
	diff.calculate()
}

func (diff *Difficulty) CheckModActive(mods Modifier) bool {
	return diff.Mods&mods > 0
}

func (diff *Difficulty) GetModifiedTime(time float64) float64 {
	if diff.Mods&DoubleTime > 0 {
		return time / (1.5 * diff.CustomSpeed)
	} else if diff.Mods&HalfTime > 0 {
		return time / (0.75 * diff.CustomSpeed)
	} else {
		return time / diff.CustomSpeed
	}
}

func (diff *Difficulty) GetScoreMultiplier() float64 {
	baseMultiplier := (diff.Mods & (^(HalfTime | Daycore | DoubleTime | Nightcore))).GetScoreMultiplier()

	if diff.Speed > 1 {
		baseMultiplier *= 1 + (0.24 * (diff.Speed - 1))
	} else if diff.Speed < 1 {
		if diff.Speed >= 0.75 {
			baseMultiplier *= 0.3 + 0.7*(1-(1-diff.Speed)/0.25)
		} else {
			baseMultiplier *= math.Max(0, 0.3*(1-(0.75-diff.Speed)/0.75))
		}
	}

	return baseMultiplier
}

func (diff *Difficulty) GetModStringFull() []string {
	mods := diff.Mods.StringFull()

	if ar := diff.GetAR(); ar != diff.GetBaseAR() {
		mods = append(mods, fmt.Sprintf("DA:AR%s", mutils.FormatWOZeros(ar, 2)))
	}

	if od := diff.GetOD(); od != diff.GetBaseOD() {
		mods = append(mods, fmt.Sprintf("DA:OD%s", mutils.FormatWOZeros(od, 2)))
	}

	if cs := diff.GetCS(); cs != diff.GetBaseCS() {
		mods = append(mods, fmt.Sprintf("DA:CS%s", mutils.FormatWOZeros(cs, 2)))
	}

	if hp := diff.GetHP(); hp != diff.GetBaseHP() {
		mods = append(mods, fmt.Sprintf("DA:HP%s", mutils.FormatWOZeros(hp, 2)))
	}

	if cSpeed := diff.CustomSpeed; cSpeed != 1 {
		mods = append(mods, fmt.Sprintf("DA:%sx", mutils.FormatWOZeros(cSpeed, 2)))
	}

	return mods
}

func (diff *Difficulty) GetModString() string {
	mods := diff.Mods.String()

	if ar := diff.GetAR(); ar != diff.GetBaseAR() {
		mods += fmt.Sprintf("AR%s", mutils.FormatWOZeros(ar, 2))
	}

	if od := diff.GetOD(); od != diff.GetBaseOD() {
		mods += fmt.Sprintf("OD%s", mutils.FormatWOZeros(od, 2))
	}

	if cs := diff.GetCS(); cs != diff.GetBaseCS() {
		mods += fmt.Sprintf("CS%s", mutils.FormatWOZeros(cs, 2))
	}

	if hp := diff.GetHP(); hp != diff.GetBaseHP() {
		mods += fmt.Sprintf("HP%s", mutils.FormatWOZeros(hp, 2))
	}

	if cSpeed := diff.CustomSpeed; cSpeed != 1 {
		mods += fmt.Sprintf("S%sx", mutils.FormatWOZeros(cSpeed, 2))
	}

	return mods
}

func (diff *Difficulty) GetBaseHP() float64 {
	return diff.baseHP
}

func (diff *Difficulty) GetHP() float64 {
	return diff.hp
}

func (diff *Difficulty) SetHP(hp float64) {
	diff.baseHP = hp
	diff.hp = hp
	diff.calculate()
}

func (diff *Difficulty) SetHPCustom(hp float64) {
	diff.hp = hp
	diff.calculate()
}

func (diff *Difficulty) GetBaseCS() float64 {
	return diff.baseCS
}

func (diff *Difficulty) GetCS() float64 {
	return diff.cs
}

func (diff *Difficulty) SetCS(cs float64) {
	diff.baseCS = cs
	diff.cs = cs
	diff.calculate()
}

func (diff *Difficulty) SetCSCustom(cs float64) {
	diff.cs = cs
	diff.calculate()
}

func (diff *Difficulty) GetBaseOD() float64 {
	return diff.baseOD
}

func (diff *Difficulty) GetOD() float64 {
	return diff.od
}

func (diff *Difficulty) SetOD(od float64) {
	diff.baseOD = od
	diff.od = od
	diff.calculate()
}

func (diff *Difficulty) SetODCustom(od float64) {
	diff.od = od
	diff.calculate()
}

func (diff *Difficulty) GetBaseAR() float64 {
	return diff.baseAR
}

func (diff *Difficulty) GetAR() float64 {
	return diff.ar
}

func (diff *Difficulty) SetAR(ar float64) {
	diff.baseAR = ar
	diff.ar = ar
	diff.calculate()
}

func (diff *Difficulty) SetARCustom(ar float64) {
	diff.ar = ar
	diff.calculate()
}

func (diff *Difficulty) SetCustomSpeed(speed float64) {
	diff.CustomSpeed = speed
	diff.calculate()
}

func DifficultyRate(diff, min, mid, max float64) float64 {
	diff = float64(float32(diff))

	if diff > 5 {
		return mid + (max-mid)*(diff-5)/5
	}

	if diff < 5 {
		return mid - (mid-min)*(5-diff)/5
	}

	return mid
}

func DiffFromRate(rate, min, mid, max float64) float64 {
	rate = float64(float32(rate))

	minStep := (min - mid) / 5
	maxStep := (mid - max) / 5

	if rate > mid {
		return -(rate - min) / minStep
	}

	return 5.0 - (rate-mid)/maxStep
}
