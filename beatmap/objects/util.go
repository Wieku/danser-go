package objects

import (
	"strconv"
)

func GetObject(data []string) BaseObject {
	objType, _ := strconv.ParseInt(data[3], 10, 64)
	if (objType & CIRCLE) > 0 {
		return NewCircle(data)
	} else if (objType & SLIDER) > 0 {
		sl := NewSlider(data)
		if sl == nil {
			return nil
		} else {
			return sl
		}
	} else if (objType & SPINNNER) > 0 {
		return NewSpinner(data)
	}
	return nil
}

const (
	CIRCLE int64 = 1
	SLIDER int64 = 2
	SPINNNER int64 = 8
)
