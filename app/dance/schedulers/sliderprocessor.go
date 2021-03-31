package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"sort"
)

func objectPreProcess(hitobject objects.IHitObject, sliderDance bool) ([]objects.IHitObject, bool) {
	if s1, ok1 := hitobject.(*objects.Slider); ok1 && sliderDance {
		return s1.GetAsDummyCircles(), true
	}

	return nil, false
}

func PreprocessQueue(index int, queue []objects.IHitObject, sliderDance bool) []objects.IHitObject {
	if arr, ok := objectPreProcess(queue[index], sliderDance); ok {
		if index < len(queue)-1 {
			queue1 := append(queue[:index], append(arr, queue[index+1:]...)...)

			sort.SliceStable(queue1, func(i, j int) bool { return queue1[i].GetStartTime() < queue1[j].GetStartTime() })

			return queue1
		} else {
			return append(queue[:index], arr...)
		}
	}

	return queue
}
