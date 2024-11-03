package beatmap

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
)

//Original code by: https://github.com/ppy/osu/blob/master/osu.Game.Rulesets.Osu/Beatmaps/OsuBeatmapProcessor.cs

const stackDistance = 3.0

func isSpinner(obj objects.IHitObject) bool {
	_, ok2 := obj.(*objects.Spinner)
	return ok2
}

func isSlider(obj objects.IHitObject) bool {
	_, ok1 := obj.(*objects.Slider)
	return ok1
}

func processStacking(hitObjects []objects.IHitObject, version int, diff *difficulty.Difficulty, stackLeniency float64) {
	stackThreshold := int64(math.Floor(diff.Preempt * stackLeniency))

	for _, v := range hitObjects {
		v.SetStackIndex(stackThreshold, 0)
	}

	if version >= 6 {
		applyNewStacking(hitObjects, stackThreshold)
	} else {
		applyOldStacking(hitObjects, stackThreshold)
	}

	for _, v := range hitObjects {
		if isSpinner(v) {
			v.SetStackIndex(stackThreshold, 0)
		}
	}
}

func applyNewStacking(hitObjects []objects.IHitObject, stackThreshold int64) {
	stackThresholdF := float64(stackThreshold)

	extendedEndIndex := len(hitObjects) - 1
	for i := len(hitObjects) - 1; i >= 0; i-- {
		stackBaseIndex := i

		for n := stackBaseIndex + 1; n < len(hitObjects); n++ {

			stackIHitObject := hitObjects[stackBaseIndex]
			if isSpinner(stackIHitObject) {
				break
			}

			objectN := hitObjects[n]
			if isSpinner(objectN) {
				continue
			}

			if objectN.GetStartTime()-stackIHitObject.GetEndTime() > stackThresholdF {
				break
			}

			if stackIHitObject.GetStartPosition().Dst(objectN.GetStartPosition()) < stackDistance || isSlider(stackIHitObject) && stackIHitObject.GetEndPosition().Dst(objectN.GetStartPosition()) < stackDistance {
				stackBaseIndex = n
				objectN.SetStackIndex(stackThreshold, 0)
			}
		}

		if stackBaseIndex > extendedEndIndex {
			extendedEndIndex = stackBaseIndex
			if extendedEndIndex == len(hitObjects)-1 {
				break
			}
		}

	}

	extendedStartIndex := 0
	for i := extendedEndIndex; i > 0; i-- {
		n := i

		objectI := hitObjects[i]

		if objectI.GetStackIndex(stackThreshold) != 0 || isSpinner(objectI) {
			continue
		}

		if !isSlider(objectI) && !isSpinner(objectI) {
			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinner(objectN) {
					continue
				}

				if objectI.GetStartTime()-objectN.GetEndTime() > stackThresholdF {
					break
				}

				if n < extendedStartIndex {
					objectN.SetStackIndex(stackThreshold, 0)
					extendedStartIndex = n
				}

				if isSlider(objectN) && objectN.GetEndPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					offset := objectI.GetStackIndex(stackThreshold) - objectN.GetStackIndex(stackThreshold) + 1
					for j := n + 1; j <= i; j++ {
						objectJ := hitObjects[j]
						if objectN.GetEndPosition().Dst(objectJ.GetStartPosition()) < stackDistance {
							objectJ.SetStackIndex(stackThreshold, objectJ.GetStackIndex(stackThreshold)-offset)
						}
					}

					break
				}

				if objectN.GetStartPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					objectN.SetStackIndex(stackThreshold, objectI.GetStackIndex(stackThreshold)+1)
					objectI = objectN
				}
			}
		} else if isSlider(objectI) {

			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinner(objectN) {
					continue
				}

				if objectI.GetStartTime()-objectN.GetStartTime() > stackThresholdF {
					break
				}

				if objectN.GetEndPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					objectN.SetStackIndex(stackThreshold, objectI.GetStackIndex(stackThreshold)+1)
					objectI = objectN
				}

			}
		}
	}
}

func applyOldStacking(hitObjects []objects.IHitObject, stackThreshold int64) {
	stackThresholdF := float64(stackThreshold)

	for i := 0; i < len(hitObjects); i++ {
		objectI := hitObjects[i]

		startTime := objectI.GetEndTime()

		if objectI.GetStackIndex(stackThreshold) == 0 || isSlider(objectI) {
			sliderStack := int64(0)

			for n := i + 1; n < len(hitObjects); n++ {
				objectN := hitObjects[n]

				if objectN.GetStartTime()-stackThresholdF > startTime {
					break
				}

				if objectN.GetStartPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					objectI.SetStackIndex(stackThreshold, objectI.GetStackIndex(stackThreshold)+1)
					startTime = objectN.GetEndTime()
				} else if objectN.GetStartPosition().Dst(objectI.GetEndPosition()) < stackDistance {
					sliderStack++
					objectN.SetStackIndex(stackThreshold, objectN.GetStackIndex(stackThreshold)-sliderStack)
					startTime = objectN.GetEndTime()
				}
			}
		}
	}
}
