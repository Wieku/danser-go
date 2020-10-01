package settings

var Skin = initSkin()

func initSkin() *skin {
	return &skin{
		CurrentSkin: "default",
	}
}

type skin struct {
	CurrentSkin string
}
