package objects

import (
	"github.com/wieku/danser-go/app/settings"
	"strconv"
)

const FadeIn = 400.0
const FadeOut = 240.0

func GetObject(data []string) BaseObject {
	objType, _ := strconv.ParseInt(data[3], 10, 64)
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

const (
	CIRCLE   int64 = 1
	SLIDER   int64 = 2
	SPINNER  int64 = 8
	LONGNOTE int64 = 128 //only for mania, used to have correct number of sliders in database just in case
)
