package difficulty

import (
	"fmt"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/rplpa"
	"math"
	"reflect"
	"slices"
	"strings"
)

const (
	HitFadeIn      = 400.0
	HitFadeOut     = 240.0
	HittableRange  = 400.0
	ResultFadeIn   = 120.0
	ResultFadeOut  = 600.0
	PostEmpt       = 500.0
	LzSpinBonusGap = 2
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

	HPMod           float64
	SpinnerRatio    float64
	LzSpinnerMinRPS float64
	LzSpinnerMaxRPS float64
	Speed           float64

	ARReal float64
	ODReal float64

	BaseModSpeed float64

	modSettings map[reflect.Type]any
	adjustPitch bool
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

	diff.Speed = 1

	diff.modSettings = make(map[reflect.Type]any)

	diff.calculate()

	return diff
}

func (diff *Difficulty) calculate() {
	diff.hp, diff.cs, diff.od, diff.ar = diff.baseHP, diff.baseCS, diff.baseOD, diff.baseAR
	if s, ok := diff.modSettings[rfType[DiffAdjustSettings]()].(DiffAdjustSettings); ok {
		diff.ar = s.ApproachRate
		diff.od = s.OverallDifficulty
		diff.hp = s.DrainRate
		diff.cs = s.CircleSize
	}

	hpDrain, cs, od, ar := diff.hp, diff.cs, diff.od, diff.ar

	if diff.Mods&HardRock > 0 {
		ar = min(ar*1.4, 10)
		cs = min(cs*1.3, 10)
		od = min(od*1.4, 10)
		hpDrain = min(hpDrain*1.4, 10)
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

	diff.TimeFadeIn = HitFadeIn * min(1, diff.PreemptU/450)

	diff.Hit50U = DifficultyRate(od, 200, 150, 100)
	diff.Hit100U = DifficultyRate(od, 140, 100, 60)
	diff.Hit300U = DifficultyRate(od, 80, 50, 20)

	diff.Hit50 = int64(diff.Hit50U)
	diff.Hit100 = int64(diff.Hit100U)
	diff.Hit300 = int64(diff.Hit300U)

	diff.SpinnerRatio = DifficultyRate(od, 3, 5, 7.5)
	diff.LzSpinnerMinRPS = DifficultyRate(od, 90, 150, 225) / 60
	diff.LzSpinnerMaxRPS = DifficultyRate(od, 250, 380, 430) / 60

	if diff.Mods&DoubleTime > 0 {
		diff.BaseModSpeed = 1.5
	} else if diff.Mods&HalfTime > 0 {
		diff.BaseModSpeed = 0.75
	} else {
		diff.BaseModSpeed = 1
	}

	diff.Speed = diff.BaseModSpeed

	if s, ok := diff.modSettings[rfType[SpeedSettings]()].(SpeedSettings); ok {
		if diff.BaseModSpeed != s.SpeedChange {
			diff.Speed = s.SpeedChange
		}

		diff.adjustPitch = s.AdjustPitch
	}

	diff.ARReal = DiffFromRate(diff.GetModifiedTime(diff.PreemptU), 1800, 1200, 450)
	diff.ODReal = DiffFromRate(diff.GetModifiedTime(diff.Hit300U), 80, 50, 20)
}

func (diff *Difficulty) SetMods(mods Modifier) {
	clear(diff.modSettings)

	diff.Mods = 0

	diff.AddMod(mods)
}

func (diff *Difficulty) AddMod(mods Modifier) {
	diff.Mods |= mods

	if mods.Active(HalfTime | Daycore) {
		diff.modSettings[rfType[SpeedSettings]()] = NewSpeedSettings(0.75, mods.Active(Daycore))
	} else if mods.Active(DoubleTime | Nightcore) {
		diff.modSettings[rfType[SpeedSettings]()] = NewSpeedSettings(1.5, mods.Active(Nightcore))
	}

	if mods.Active(Easy) {
		diff.modSettings[rfType[EasySettings]()] = NewEasySettings()
	}

	if mods.Active(Flashlight) {
		diff.modSettings[rfType[FlashlightSettings]()] = NewFlashlightSettings()
	}

	if mods.Active(DifficultyAdjust) {
		diff.modSettings[rfType[DiffAdjustSettings]()] = NewDiffAdjustSettings(diff.baseAR, diff.baseCS, diff.baseHP, diff.baseOD)
	}

	if mods.Active(Classic) {
		diff.modSettings[rfType[ClassicSettings]()] = NewClassicSettings()
	}

	diff.calculate()
}

func (diff *Difficulty) RemoveMod(mods Modifier) {
	if mods.Active(Nightcore) {
		mods |= DoubleTime
	}

	if mods.Active(Daycore) {
		mods |= HalfTime
	}

	if mods.Active(Perfect) {
		mods |= SuddenDeath
	}

	diff.Mods &= ^mods

	if mods.Active(HalfTime | Daycore | DoubleTime | Nightcore) {
		delete(diff.modSettings, rfType[SpeedSettings]())
	}

	if mods.Active(Easy) {
		delete(diff.modSettings, rfType[EasySettings]())
	}

	if mods.Active(Flashlight) {
		delete(diff.modSettings, rfType[FlashlightSettings]())
	}

	if mods.Active(DifficultyAdjust) {
		delete(diff.modSettings, rfType[DiffAdjustSettings]())
	}

	if mods.Active(Classic) {
		delete(diff.modSettings, rfType[ClassicSettings]())
	}

	diff.calculate()
}

func (diff *Difficulty) SetMods2(mods []rplpa.ModInfo) {
	clear(diff.modSettings)

	mComp := None

	for _, mInfo := range mods {
		mod := ParseFromAcronym(mInfo.Acronym)

		if mod != None {
			mComp |= mod

			if mod.Active(HalfTime | Daycore) {
				diff.modSettings[rfType[SpeedSettings]()] = parseConfig(NewSpeedSettings(0.75, mod.Active(Daycore)), mInfo.Settings)
			} else if mod.Active(DoubleTime | Nightcore) {
				diff.modSettings[rfType[SpeedSettings]()] = parseConfig(NewSpeedSettings(1.5, mod.Active(Nightcore)), mInfo.Settings)
			}

			if mod.Active(Easy) {
				diff.modSettings[rfType[EasySettings]()] = parseConfig(NewEasySettings(), mInfo.Settings)
			}

			if mod.Active(Flashlight) {
				diff.modSettings[rfType[FlashlightSettings]()] = parseConfig(NewFlashlightSettings(), mInfo.Settings)
			}

			if mod.Active(DifficultyAdjust) {
				diff.modSettings[rfType[DiffAdjustSettings]()] = parseConfig(NewDiffAdjustSettings(diff.baseAR, diff.baseCS, diff.baseHP, diff.baseOD), mInfo.Settings)
			}

			if mod.Active(Classic) {
				diff.modSettings[rfType[ClassicSettings]()] = parseConfig(NewClassicSettings(), mInfo.Settings)
			}
		}
	}

	if mComp.Active(Nightcore) {
		mComp |= DoubleTime
	}

	if mComp.Active(Perfect) {
		mComp |= SuddenDeath
	}

	if mComp.Active(Daycore) {
		mComp |= HalfTime
	}

	diff.Mods = mComp

	diff.calculate()
}

func (diff *Difficulty) ExportMods2() (mods []rplpa.ModInfo) {
	mComp := diff.Mods

	if mComp.Active(Nightcore) {
		mComp &= ^DoubleTime
	}

	if mComp.Active(Daycore) {
		mComp &= ^HalfTime
	}

	if mComp.Active(Perfect) {
		mComp &= ^SuddenDeath
	}

	for i := 0; i <= 62; i++ {
		mTest := Modifier(1 << i)

		if mComp.Active(mTest) {
			var modSettings map[string]any

			if cType, ok := modConfigs[mTest]; ok {
				if cConf, exists := diff.modSettings[cType]; exists {
					modSettings = exportConfig(cConf)
				}
			}

			mods = append(mods, rplpa.ModInfo{
				Acronym:  modsString[i],
				Settings: modSettings,
			})
		}
	}

	return
}

func (diff *Difficulty) CheckModActive(mods Modifier) bool {
	return diff.Mods&mods > 0
}

func (diff *Difficulty) GetModifiedTime(time float64) float64 {
	return time / diff.Speed
}

func (diff *Difficulty) GetSpeed() float64 {
	return diff.Speed
}

func (diff *Difficulty) AdjustsPitch() bool {
	return diff.adjustPitch
}

func (diff *Difficulty) GetPitch() float64 {
	if diff.adjustPitch && diff.Speed != 1 {
		return diff.Speed
	}

	return 1
}

func (diff *Difficulty) GetScoreMultiplier() float64 {
	baseMultiplier := (diff.Mods & (^(HalfTime | Daycore | DoubleTime | Nightcore | Flashlight))).GetScoreMultiplier()

	if diff.Mods.Active(Lazer) {
		value := math.Floor(diff.Speed*10)/10 - 1

		if diff.Speed >= 1 {
			baseMultiplier *= 1 + value/5
		} else {
			baseMultiplier *= 0.6 + value
		}
	} else {
		if diff.Speed > 1 {
			if diff.Mods.Active(ScoreV2) {
				baseMultiplier *= 1 + (0.40 * (diff.Speed - 1))
			} else {
				baseMultiplier *= 1 + (0.24 * (diff.Speed - 1))
			}
		} else if diff.Speed < 1 {
			if diff.Speed >= 0.75 {
				baseMultiplier *= 0.3 + 0.7*(1-(1-diff.Speed)/0.25)
			} else {
				baseMultiplier *= max(0, 0.3*(1-(0.75-diff.Speed)/0.75))
			}
		}
	}

	if diff.CheckModActive(Flashlight) {
		mult := 1.12

		if fl, ok := GetModConfig[FlashlightSettings](diff); ok {
			if fl != NewFlashlightSettings() {
				mult = 1
			}
		}

		baseMultiplier *= mult
	}

	return baseMultiplier
}

func (diff *Difficulty) GetModStringFull() []string {
	mods := (diff.Mods & (^DifficultyAdjust)).StringFull()

	var daBase []string

	if ar := diff.GetAR(); math.Abs(ar-diff.GetBaseAR()) > 0.001 {
		daBase = append(daBase, fmt.Sprintf("AR %s", mutils.FormatWOZeros(ar, 2)))
	}

	if od := diff.GetOD(); math.Abs(od-diff.GetBaseOD()) > 0.001 {
		daBase = append(daBase, fmt.Sprintf("OD %s", mutils.FormatWOZeros(od, 2)))
	}

	if cs := diff.GetCS(); math.Abs(cs-diff.GetBaseCS()) > 0.001 {
		daBase = append(daBase, fmt.Sprintf("CS %s", mutils.FormatWOZeros(cs, 2)))
	}

	if hp := diff.GetHP(); math.Abs(hp-diff.GetBaseHP()) > 0.001 {
		daBase = append(daBase, fmt.Sprintf("HP %s", mutils.FormatWOZeros(hp, 2)))
	}

	if len(daBase) > 0 {
		mods = slices.Insert(mods, 0, "DifficultyAdjust:"+strings.Join(daBase, ":"))
	}

	if cSpeed := diff.Speed; math.Abs(cSpeed-diff.BaseModSpeed) > 0.001 {
		toCheck := []Modifier{DoubleTime, Nightcore, HalfTime, Daycore}
		anyFound := false

		for _, ms := range toCheck {
			if i := slices.Index(mods, ms.StringFull()[0]); i != -1 {
				anyFound = true
				mods[i] += ":" + mutils.FormatWOZeros(cSpeed, 2) + "x"
			}
		}

		if !anyFound {
			mods = append(mods, fmt.Sprintf("DifficultyAdjust:%sx", mutils.FormatWOZeros(cSpeed, 2)))
		}
	}

	return mods
}

func (diff *Difficulty) GetModString() string {
	return diff.getModStringBase(diff.Mods)
}

func (diff *Difficulty) GetModStringMasked() string {
	return diff.getModStringBase(GetDiffMaskedMods(diff.Mods))
}

func (diff *Difficulty) getModStringBase(mod Modifier) string {
	mods := mod.String()

	if ar := diff.GetAR(); math.Abs(ar-diff.GetBaseAR()) > 0.001 {
		mods += fmt.Sprintf("AR%s", mutils.FormatWOZeros(ar, 2))
	}

	if od := diff.GetOD(); math.Abs(od-diff.GetBaseOD()) > 0.001 {
		mods += fmt.Sprintf("OD%s", mutils.FormatWOZeros(od, 2))
	}

	if cs := diff.GetCS(); math.Abs(cs-diff.GetBaseCS()) > 0.001 {
		mods += fmt.Sprintf("CS%s", mutils.FormatWOZeros(cs, 2))
	}

	if hp := diff.GetHP(); math.Abs(hp-diff.GetBaseHP()) > 0.001 {
		mods += fmt.Sprintf("HP%s", mutils.FormatWOZeros(hp, 2))
	}

	if cSpeed := diff.Speed; math.Abs(cSpeed-diff.BaseModSpeed) > 0.001 {
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

func (diff *Difficulty) Clone() *Difficulty {
	diff2 := *diff
	diff2.modSettings = make(map[reflect.Type]any)
	for k, v := range diff.modSettings {
		diff2.modSettings[k] = v
	}

	return &diff2
}

func (diff *Difficulty) Equals(diff2 *Difficulty) bool {
	if diff.Mods != diff2.Mods {
		return false
	}

	for k, v := range diff.modSettings {
		v2, ok := diff2.modSettings[k]

		if !ok {
			return false
		}

		rVal1 := reflect.ValueOf(&v).Elem().Elem()
		rVal2 := reflect.ValueOf(&v2).Elem().Elem()
		rType := reflect.TypeOf(v)

		for i := range rType.NumField() {
			sField := rType.Field(i)

			check := true

			switch sField.Type.Kind() {
			case reflect.Float64:
				check = math.Abs(rVal1.Field(i).Float()-rVal2.Field(i).Float()) < 0.00001
			case reflect.Int:
				check = rVal1.Field(i).Int() == rVal2.Field(i).Int()
			case reflect.Bool:
				check = rVal1.Field(i).Bool() == rVal2.Field(i).Bool()
			}

			if !check {
				return false
			}
		}
	}

	return true
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
