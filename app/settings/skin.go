package settings

var Skin = initSkin()

func initSkin() *skin {
	return &skin{
		CurrentSkin:       "default",
		UseColorsFromSkin: false,
		Cursor: &skinCursor{
			UseSkinCursor:   false,
			Scale:           1.0,
			ForceLongTrail:  false,
			LongTrailLength: 2048,
		},
	}
}

type skin struct {
	CurrentSkin       string
	UseColorsFromSkin bool

	Cursor *skinCursor
}

type skinCursor struct {
	UseSkinCursor   bool
	Scale           float64
	ForceLongTrail  bool
	LongTrailLength int64
}
