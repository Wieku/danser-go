package settings

var Knockout = initKnockout()

func initKnockout() *knockout {
	return &knockout{
		Mode:                ComboBreak,
		ExcludeMods:         "EZHT",
		MaxPlayers:          50,
		BubbleMinimumCombo:  200,
		RevivePlayersAtEnd:  false,
		LiveSort:            true,
		SortBy:              "Score",
		HideOverlayOnBreaks: false,
		MinCursorSize:       3.0,
		MaxCursorSize:       7.0,
		AddDanser:           false,
		DanserName:          "danser",
	}
}

type knockout struct {
	// Knockout mode. More info below
	Mode KnockoutMode

	// Exclude plays which contain one of the mods set here
	ExcludeMods string

	// Hide specific mods from being displayed in overlay (like NF)
	HideMods string

	// Max players shown (excluding danser) on a map. Caps at 50.
	MaxPlayers int

	// Minimum combo before combo break to show a bubble in XReplays mode
	BubbleMinimumCombo int

	// Whether knocked out players should appear on map end
	RevivePlayersAtEnd bool

	// Whether scores should be sorted in real time
	LiveSort bool

	// Whether players should be sorted by Score or PP
	SortBy string

	// Whether knockout overlay (player list with stats) should be hidden in breaks
	HideOverlayOnBreaks bool

	//Minimum cursor size (when all players are alive)
	MinCursorSize float64

	//Maximum cursor size (when there is only 1 player left)
	MaxCursorSize float64

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

	//Other result than 300 will knock the player out
	SSOrQuit

	// Players get knocked out if they fail
	Fail
)
