package difficulty

import "github.com/wieku/rplpa"

type Modifier int64

const (
	None   = Modifier(iota)
	NoFail = Modifier(1 << (iota - uint(1)))
	Easy
	TouchDevice
	Hidden
	HardRock
	SuddenDeath
	DoubleTime
	Relax
	HalfTime
	Nightcore // Only set along with DoubleTime. i.e: NC only gives 576
	Flashlight
	Autoplay
	SpunOut
	Relax2  // Autopilot
	Perfect // Only set along with SuddenDeath. i.e: PF only gives 16416
	Key4
	Key5
	Key6
	Key7
	Key8
	FadeIn
	Random
	Cinema
	Target
	Key9
	KeyCoop
	Key1
	Key3
	Key2
	ScoreV2
	LastMod
	Daycore
	Lazer
	Classic
	DifficultyAdjust
	Mirror
	Traceable

	// DifficultyAdjustMask is outdated, use GetDiffMaskedMods instead
	DifficultyAdjustMask    = HardRock | Easy | DoubleTime | Nightcore | HalfTime | Daycore | Flashlight | Relax
	difficultyAdjustMaskNew = HardRock | Easy | DoubleTime | HalfTime | Hidden | Flashlight | Relax | Relax2 | TouchDevice
)

// GetDiffMaskedMods should be used instead of DifficultyAdjustMask. Hidden affects reading difficulty since 260706.
func GetDiffMaskedMods(mods Modifier) Modifier {
	//Probably redundant
	if mods.Active(Nightcore) {
		mods = (mods & (^Nightcore)) | DoubleTime
	}

	if mods.Active(Daycore) {
		mods = (mods & (^Daycore)) | HalfTime
	}

	base := difficultyAdjustMaskNew & mods

	return base
}

var modsString = [...]string{
	"NF",
	"EZ",
	"TD",
	"HD",
	"HR",
	"SD",
	"DT",
	"RX",
	"HT",
	"NC",
	"FL",
	"AT", // Auto.
	"SO",
	"AP", // Autopilot.
	"PF",
	"K4",
	"K5",
	"K6",
	"K7",
	"K8",
	"FI",
	"RN", // Random
	"CN",
	"TG",
	"K9",
	"K0",
	"K1",
	"K3",
	"K2",
	"V2",
	"LM",
	"DC",
	"LZ",
	"CL",
	"DA",
	"MR",
	"TC",
}

var modsStringFull = [...]string{
	"NoFail",
	"Easy",
	"TouchDevice",
	"Hidden",
	"HardRock",
	"SuddenDeath",
	"DoubleTime",
	"Relax",
	"HalfTime",
	"Nightcore",
	"Flashlight",
	"Autoplay",
	"SpunOut",
	"AutoPilot",
	"Perfect",
	"Key4",
	"Key5",
	"Key6",
	"Key7",
	"Key8",
	"FadeIn",
	"Random",
	"Cinema",
	"Target",
	"Key9",
	"KeyCoop",
	"Key1",
	"Key3",
	"Key2",
	"ScoreV2",
	"LastMod",
	"Daycore",
	"Lazer",
	"Classic",
	"DifficultyAdjust",
	"Mirror",
	"Traceable",
}

func (mods Modifier) GetScoreMultiplier() float64 {
	multiplier := 1.0

	if mods&NoFail > 0 && mods&ScoreV2 == 0 {
		multiplier *= 0.5
	}

	if mods&Easy > 0 {
		multiplier *= 0.5
	}

	if mods&HalfTime > 0 {
		multiplier *= 0.3
	}

	if mods&Hidden > 0 {
		multiplier *= 1.06
	}

	if mods&HardRock > 0 {
		if mods&ScoreV2 > 0 {
			multiplier *= 1.10
		} else {
			multiplier *= 1.06
		}
	}

	if mods&DoubleTime > 0 {
		if mods&ScoreV2 > 0 {
			multiplier *= 1.20
		} else {
			multiplier *= 1.12
		}
	}

	if mods&Flashlight > 0 {
		multiplier *= 1.12
	}

	if (mods&Relax | mods&Relax2) > 0 {
		if mods&Lazer > 0 {
			multiplier *= 0.1
		} else {
			multiplier = 0
		}
	}

	if mods&SpunOut > 0 {
		multiplier *= 0.9
	}

	if mods&Classic > 0 {
		multiplier *= 0.96
	}

	if mods&DifficultyAdjust > 0 {
		multiplier *= 0.5
	}

	return multiplier
}

func (mods Modifier) GetScoreMultiplierV2() float64 {
	multiplier := 1.0

	if mods&NoFail > 0 {
		multiplier *= 0.5
	}

	if mods&HardRock > 0 {
		multiplier *= 1.09
	}

	if mods&Traceable > 0 {
		multiplier *= 1.02
	}

	if (mods&Relax | mods&Relax2) > 0 {
		multiplier *= 0.1
	}

	if mods&SpunOut > 0 {
		multiplier *= 0.95
	}

	return multiplier
}

func (mods Modifier) String() (s string) {
	if mods.Active(Nightcore) {
		mods &= ^DoubleTime
	}

	if mods.Active(Daycore) {
		mods &= ^HalfTime
	}

	if mods.Active(Perfect) {
		mods &= ^SuddenDeath
	}

	for i := range len(modsString) {
		activated := mods&1 == 1
		if activated {
			s += modsString[i]
		}

		mods >>= 1
	}

	return
}

func (mods Modifier) StringFull() (s []string) {
	if mods.Active(Nightcore) {
		mods &= ^DoubleTime
	}

	if mods.Active(Daycore) {
		mods &= ^HalfTime
	}

	if mods.Active(Perfect) {
		mods &= ^SuddenDeath
	}

	return mods.StringFull2()
}

func (mods Modifier) StringFull2() (s []string) {
	for i := range len(modsString) {
		activated := mods&1 == 1
		if activated {
			s = append(s, modsStringFull[i])
		}

		mods >>= 1
	}

	return
}

func ParseFromAcronym(mod string) (m Modifier) {
	for index, availableMod := range modsString {
		if availableMod == mod {
			m = 1 << uint(index)
			break
		}
	}

	return
}

func (mods Modifier) ConvertToModInfoList() (mi []rplpa.ModInfo) {
	if mods.Active(Nightcore) {
		mods &= ^DoubleTime
	}

	if mods.Active(Daycore) {
		mods &= ^HalfTime
	}

	if mods.Active(Perfect) {
		mods &= ^SuddenDeath
	}

	for i := range len(modsString) {
		if mods&1 == 1 {
			mi = append(mi, rplpa.ModInfo{
				Acronym:  modsString[i],
				Settings: make(map[string]any),
			})
		}

		mods >>= 1
	}

	return
}

func ParseMods(mods string) (m Modifier) {
	modsSl := make([]string, len(mods)/2)
	for n, modPart := range mods {
		modsSl[n/2] += string(modPart)
	}

	for _, mod := range modsSl {
		m |= ParseFromAcronym(mod)
	}

	if m.Active(Nightcore) {
		m |= DoubleTime
	}

	if m.Active(Perfect) {
		m |= SuddenDeath
	}

	if m.Active(Daycore) {
		m |= HalfTime
	}

	return
}

func (mods Modifier) Active(mod Modifier) bool {
	return mods&mod > 0
}

func (mod Modifier) GetStableIncompatibleMods() Modifier {
	incompat := mod.GetCommonIncompatibleMods()

	switch mod {
	case NoFail, SuddenDeath, Perfect:
		return incompat | Relax | Relax2
	case Relax, Relax2:
		return incompat | NoFail | SuddenDeath | Perfect
	}

	return incompat
}

func (mod Modifier) GetLazerIncompatibleMods() Modifier {
	return mod.GetCommonIncompatibleMods()
}

func (mod Modifier) GetCommonIncompatibleMods() Modifier {
	switch mod {
	case Easy:
		return HardRock
	case HardRock:
		return Easy | Mirror
	case Mirror:
		return HardRock
	case NoFail:
		return SuddenDeath | Perfect
	case SuddenDeath:
		return NoFail | Perfect
	case Perfect:
		return NoFail | SuddenDeath
	case DoubleTime:
		return Nightcore | HalfTime | Daycore
	case Nightcore:
		return DoubleTime | HalfTime | Daycore
	case HalfTime:
		return Daycore | DoubleTime | Nightcore
	case Daycore:
		return HalfTime | DoubleTime | Nightcore
	case Hidden:
		return Traceable
	case Traceable:
		return Hidden
	case Relax:
		return Autoplay
	case Relax2:
		return SpunOut | Autoplay | TouchDevice
	case SpunOut:
		return Relax2 | Autoplay
	case Autoplay:
		return Relax | Relax2 | SpunOut | TouchDevice
	case TouchDevice:
		return Autoplay | Relax2
	case Lazer:
		return ScoreV2
	case ScoreV2:
		return Lazer | Classic
	case Classic:
		return ScoreV2
	}

	return None
}

func (mods Modifier) GetIncompatibleMods(lazer bool) Modifier {
	if lazer {
		return mods.GetLazerIncompatibleMods()
	}

	return mods.GetStableIncompatibleMods()
}

func (mods Modifier) GetIncompatibleCombo() Modifier {
	if mods == None {
		return None
	}

	if mods.Active(Target) {
		return Target
	}

	if mods.Active(Cinema) {
		return Cinema
	}

	// These flags also contain their stable counterparts when imported from a replay.
	if mods.Active(Nightcore) {
		mods &= ^DoubleTime
	}
	if mods.Active(Daycore) {
		mods &= ^HalfTime
	}
	if mods.Active(Perfect) {
		mods &= ^SuddenDeath
	}

	for i := range len(modsString) {
		mod := Modifier(1 << i)

		incompat := mods & mod.GetIncompatibleMods(mods.Active(Lazer))
		if mods.Active(mod) && incompat > 0 {
			return mod | incompat
		}
	}

	return None
}

func (mods Modifier) Compatible() bool {
	return mods.GetIncompatibleCombo() == None
}
