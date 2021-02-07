package objects

import (
	"github.com/wieku/danser-go/app/settings"
	"strconv"
)

const FadeIn = 400.0
const FadeOut = 240.0

func GetObject(data []string) IHitObject {
	objTypeI, _ := strconv.Atoi(data[3])
	objType := Type(objTypeI)
	if (objType & CIRCLE) > 0 {
		return NewCircle(data)
	} else if (objType & SPINNER) > 0 {
		if settings.Objects.LoadSpinners || settings.KNOCKOUT || settings.PLAY {
			return NewSpinner(data)
		}
	} else if (objType & SLIDER) > 0 {
		sl := NewSlider(data)
		if sl == nil {
			return nil
		} else {
			return sl
		}
	}
	return nil
}

type Type int

const (
	CIRCLE = Type(1 << iota)
	SLIDER
	NEWCOMBO
	SPINNER
	LONGNOTE = Type(128) //only for mania, used to have correct number of sliders in database just in case
)
