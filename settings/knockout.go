package settings

var Knockout = initKnockout()

func initKnockout() *knockout {
	return &knockout{
		Username:      "",
		MD5Pass:       "",
		ApiKey:        "",
		Mode:          ComboBreak,
		LocalReplays:  false,
		OnlineReplays: true,
		ExcludeMods:   "EZHT",
		MaxPlayers:    50,
		IncludeDanser: false,
		DanserName:    "danser",
	}
}

type knockout struct {
	Username      string
	MD5Pass       string
	ApiKey        string
	Mode          KnockoutMode
	LocalReplays  bool
	OnlineReplays bool
	ExcludeMods   string
	MaxPlayers    int
	IncludeDanser bool
	DanserName    string
}

type KnockoutMode int

const (
	// Players get knocked out when they lose a combo to a miss or slider break
	ComboBreak = KnockoutMode(iota)

	// Players get knocked as in ComboBreak but only when they reached thier max combo on the map
	MaxCombo

	// Players won't get knocked out
	XReplays

	// Players scores other than 300's will be shown on the map (NOTE: this overrides MaxPlayer value)
	OneVsOne
)
