package objects

import (
	"github.com/wieku/danser-go/settings"
	"strconv"
)

func GetObject(data []string) BaseObject {
	objType, _ := strconv.ParseInt(data[3], 10, 64)
	if (objType & CIRCLE) > 0 {
		return NewCircle(data)
	} else if (objType & SPINNER) > 0 {
		if settings.Objects.LoadSpinners || settings.KNOCKOUT != "" {return NewSpinner(data)}
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
	CIRCLE int64 = 1
	SLIDER int64 = 2
	SPINNER int64 = 8
)
