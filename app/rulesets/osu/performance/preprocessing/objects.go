package preprocessing

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
)

// CreateDifficultyObjects creates difficulty objects needed for star rating calculations
func CreateDifficultyObjects(objsB []objects.IHitObject, d *difficulty.Difficulty, experimental bool) []*DifficultyObject {
	objs := make([]objects.IHitObject, 0, len(objsB))

	for _, o := range objsB {
		if s, ok := o.(*objects.Slider); ok {
			o = NewLazySlider(s, d, experimental)
		}

		objs = append(objs, o)
	}

	diffObjects := make([]*DifficultyObject, 0, len(objsB))

	for i := 1; i < len(objs); i++ {
		var lastLast, last, current objects.IHitObject

		if i > 1 {
			lastLast = objs[i-2]
		}

		last = objs[i-1]
		current = objs[i]

		diffObjects = append(diffObjects, NewDifficultyObject(current, lastLast, last, d, experimental))
	}

	return diffObjects
}
