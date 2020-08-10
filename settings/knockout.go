package settings

var Knockout = initKnockout()

func initKnockout() *knockout {
	return &knockout{
		Username:           "",
		MD5Pass:            "",
		ApiKey:             "",
		Mode:               ComboBreak,
		LocalReplays:       false,
		OnlineReplays:      true,
		ExcludeMods:        "EZHT",
		MaxPlayers:         50,
		BubbleMinimumCombo: 200,
		AddDanser:          false,
		DanserName:         "danser",
	}
}

type knockout struct {
	Username string
	MD5Pass  string
	ApiKey   string

	// Knockout mode. More info below
	Mode KnockoutMode

	// Whether to load local replays. They have to be put in ./replays/beatmapMD5/
	LocalReplays bool

	// Whether to load online replays using osu!api
	OnlineReplays bool

	// Exclude plays which contain one of the mods set here
	ExcludeMods string

	// Max players shown (excluding danser) on a map. Caps at 50.
	MaxPlayers int

	// Minimum combo before combo break to show a bubble in XReplays mode
	BubbleMinimumCombo int

	// Self explanatory
	AddDanser  bool
	DanserName string
}

type KnockoutMode int

const (
	// Players get knocked out when they lose a combo to a miss or slider break
	ComboBreak = KnockoutMode(iota)

	// Players get knocked as in ComboBreak but only when they reached their max combo on the map
	MaxCombo

	// Players won't get knocked out
	XReplays

	// Players scores other than 300's will be shown on the map (NOTE: this overrides MaxPlayer value)
	OneVsOne
)
