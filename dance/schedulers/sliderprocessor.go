package schedulers

import "github.com/wieku/danser/beatmap/objects"

func objectPreProcess(hitobject objects.BaseObject, sliderDance bool) ([]objects.BaseObject, bool) {
	if s1, ok1 := hitobject.(*objects.Slider); ok1 && sliderDance {
		return s1.GetAsDummyCircles(), true
	}
	return nil, false
}

func preprocessQueue(index int, queue []objects.BaseObject, sliderDance bool) []objects.BaseObject {
	if arr, ok := objectPreProcess(queue[index], sliderDance); ok {
		if index < len(queue) -1 {
			return append(queue[:index], append(arr, queue[index+1:]...)...)
		} else {
			return append(queue[:index], arr...)
		}
	}
	return queue
}