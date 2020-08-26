package schedulers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"sort"
)

func objectPreProcess(hitobject objects.BaseObject, sliderDance bool) ([]objects.BaseObject, bool) {
	if s1, ok1 := hitobject.(*objects.Slider); ok1 && sliderDance {
		return s1.GetAsDummyCircles(), true
	}
	return nil, false
}

func PreprocessQueue(index int, queue []objects.BaseObject, sliderDance bool) []objects.BaseObject {
	if arr, ok := objectPreProcess(queue[index], sliderDance); ok {
		if index < len(queue)-1 {
			queue1 := append(queue[:index], append(arr, queue[index+1:]...)...)
			sort.Slice(queue1, func(i, j int) bool { return queue1[i].GetBasicData().StartTime < queue1[j].GetBasicData().StartTime })
			return queue1
		} else {
			return append(queue[:index], arr...)
		}
	}
	return queue
}
