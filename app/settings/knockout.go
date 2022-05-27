package settings

var Knockout = initKnockout()

func initKnockout() *knockout {
	return &knockout{
		Mode:                ComboBreak,
		GraceEndTime:        -10,
		BubbleMinimumCombo:  200,
		ExcludeMods:         "",
		MaxPlayers:          50,
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
	Mode KnockoutMode `combo:"0|Combo Break,1|Max Combo,2|Replay Showcase,3|Vs Mode,4|SS or Quit"`

	// In Mode = ComboBreak it won't knock out the player if they break combo before GraceEndTime (in seconds)
	GraceEndTime float64 `string:"true" min:"-10" max:"1000000"`

	// In Mode = XReplays it will show combo break bubble if combo was bigger than BubbleMinimumCombo
	BubbleMinimumCombo int `string:"true" min:"1" max:"1000000"`

	// Exclude plays which contain one of the mods set here
	ExcludeMods string

	// Hide specific mods from being displayed in overlay (like NF)
	HideMods string

	// Max players shown (excluding danser) on a map. Caps at 50.
	MaxPlayers int `min:"0" max:"100"`

	// Whether knocked out players should appear on map end
	RevivePlayersAtEnd bool

	// Whether scores should be sorted in real time
	LiveSort bool

	// Whether players should be sorted by Score or PP
	SortBy string `combo:"Score,PP,Accuracy"`

	// Whether knockout overlay (player list with stats) should be hidden in breaks
	HideOverlayOnBreaks bool

	//Minimum cursor size (when all players are alive)
	MinCursorSize float64 `min:"1" max:"20"`

	//Maximum cursor size (when there is only 1 player left)
	MaxCursorSize float64 `min:"1" max:"20"`

	// Self explanatory
	AddDanser  bool
	DanserName string
}

type KnockoutMode int

const (
	// Players get knocked out when they lose a combo to a miss or slider break
	ComboBreak = KnockoutMode(iota)

	// ComboBreak but only when they reached their max combo on the map
	MaxCombo

	// Players won't get knocked out
	XReplays

	// XReplays but Player scores other than 300's will be shown on the map
	OneVsOne

	// Forced Perfect mod
	SSOrQuit
)
