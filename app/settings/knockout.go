package settings

var Knockout = initKnockout()

func initKnockout() *knockout {
	return &knockout{
		Mode:                ComboBreak,
		SmokeEnabled:        false,
		GraceEndTime:        -10,
		BubbleMinimumCombo:  200,
		ExcludeMods:         "",
		MaxPlayers:          50,
		MinPlayers:          1,
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
	Mode KnockoutMode `combo:"0|Combo Break,1|Max Combo,2|Replay Showcase,3|Vs Mode,4|SS or Quit" liveedit:"false"`

	// In Mode = ComboBreak it won't knock out the player if they break combo before GraceEndTime (in seconds)
	GraceEndTime float64 `string:"true" min:"-10" max:"1000000" showif:"Mode=0"`

	// In Mode = XReplays it will show combo break bubble if combo was bigger than BubbleMinimumCombo
	BubbleMinimumCombo int `label:"Minimum combo to show break bubble" string:"true" min:"1" max:"1000000" showif:"Mode=2"`

	// Exclude plays which contain one of the mods set here
	ExcludeMods string `skip:"true" label:"Excluded mods (legacy)" tooltip:"Applicable only to classic knockout" liveedit:"false"`

	// Hide specific mods from being displayed in overlay (like NF)
	HideMods string `liveedit:"false"`

	// Max players shown (excluding danser) on a map. Caps at 50.
	MaxPlayers int `skip:"true" label:"Max players loaded (legacy)" string:"true" min:"0" max:"100" tooltip:"Applicable only to classic knockout"`

	// Min players shown on a map.
	MinPlayers int `label:"Minimum alive players" string:"true" min:"0" max:"100" showif:"Mode=0,1,4"`

	// Whether knocked out players should appear on map end
	RevivePlayersAtEnd bool `showif:"Mode=0,1,4"`

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

	SmokeEnabled bool `label:"Show cursor smoke in knockout mode"`

	// Self explanatory
	AddDanser  bool   `liveedit:"false"`
	DanserName string `label:"Danser's name" tooltip:"It's also used in danser replay mode" liveedit:"false"`
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
