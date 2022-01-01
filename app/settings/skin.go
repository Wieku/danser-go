package settings

var Skin = initSkin()

func initSkin() *skin {
	return &skin{
		CurrentSkin:       "default",
		FallbackSkin:      "default",
		UseColorsFromSkin: false,
		UseBeatmapColors:  false,
		Cursor: &skinCursor{
			UseSkinCursor:    false,
			Scale:            1.0,
			TrailScale:       1.0,
			ForceLongTrail:   false,
			LongTrailLength:  2048,
			LongTrailDensity: 1.0,
		},
	}
}

type skin struct {
	CurrentSkin       string
	FallbackSkin      string
	UseColorsFromSkin bool
	UseBeatmapColors  bool

	Cursor *skinCursor
}

type skinCursor struct {
	UseSkinCursor    bool
	Scale            float64 `max:"2"`
	TrailScale       float64 `max:"2"`
	ForceLongTrail   bool
	LongTrailLength  int64
	LongTrailDensity float64
}
