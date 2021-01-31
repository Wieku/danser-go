package difficulty

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
)

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
	"K9",
	"RN", // Random
	"LM", // LastMod. Cinema?
	"K9",
	"K0",
	"K1",
	"K3",
	"K2",
	"V2",
	"LM",
	"DC",
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
	"Relax2",
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
}

func (modifier Modifier) GetScoreMultiplier() float64 {
	multiplier := 1.0

	if modifier&NoFail > 0 {
		multiplier *= 0.5
	}

	if modifier&Easy > 0 {
		multiplier *= 0.5
	}

	if modifier&HalfTime > 0 {
		multiplier *= 0.3
	}

	if modifier&Hidden > 0 {
		multiplier *= 1.06
	}

	if modifier&HardRock > 0 {
		multiplier *= 1.06
	}

	if modifier&DoubleTime > 0 {
		multiplier *= 1.12
	}

	if modifier&Flashlight > 0 {
		multiplier *= 1.12
	}

	if (modifier&Relax | modifier&Relax2) > 0 {
		multiplier = 0
	}

	if modifier&SpunOut > 0 {
		multiplier *= 0.9
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

	for i := 0; i < len(modsString); i++ {
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

	for i := 0; i < len(modsString); i++ {
		activated := mods&1 == 1
		if activated {
			s = append(s, modsStringFull[i])
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
		for index, availableMod := range modsString {
			if availableMod == mod {
				m |= 1 << uint(index)
				break
			}
		}
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

func (mods Modifier) Compatible() bool {
	if mods == None {
		return true
	}

	if mods.Active(HardRock) && mods.Active(Easy) {
		return false
	}

	if (mods.Active(Nightcore) || mods.Active(DoubleTime)) && (mods.Active(HalfTime) || mods.Active(Daycore)) {
		return false
	}

	if (mods.Active(Perfect) || mods.Active(SuddenDeath)) && mods.Active(NoFail) {
		return false
	}

	if mods.Active(Relax) && mods.Active(Relax2) {
		return false
	}

	if (mods.Active(Relax) || mods.Active(Relax2)) && mods.Active(Autoplay) {
		return false
	}

	return true
}
