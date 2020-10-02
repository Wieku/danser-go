package settings

var Skin = initSkin()

func initSkin() *skin {
	return &skin{
		CurrentSkin:       "default",
		UseColorsFromSkin: false,
	}
}

type skin struct {
	CurrentSkin       string
	UseColorsFromSkin bool
}
