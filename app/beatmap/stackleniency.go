package beatmap

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
)

//Original code by: https://github.com/ppy/osu/blob/master/osu.Game.Rulesets.Osu/Beatmaps/OsuBeatmapProcessor.cs

func isSpinnerBreak(obj objects.BaseObject) bool {
	_, ok2 := obj.(*objects.Pause)
	return ok2
}

func isSlider(obj objects.BaseObject) bool {
	_, ok1 := obj.(*objects.Slider)
	return ok1
}

func calculateStackLeniency(b *BeatMap) {
	stack_distance := float32(3.0)

	hitObjects := b.HitObjects

	if !settings.Objects.StackEnabled {
		return
	}

	for _, v := range hitObjects {
		v.GetBasicData().StackIndex = 0
	}

	extendedEndIndex := len(hitObjects) - 1
	for i := len(hitObjects) - 1; i >= 0; i-- {
		stackBaseIndex := i

		for n := stackBaseIndex + 1; n < len(hitObjects); n++ {

			stackBaseObject := hitObjects[stackBaseIndex]
			if isSpinnerBreak(stackBaseObject) {
				break
			}

			objectN := hitObjects[n]
			if isSpinnerBreak(objectN) {
				continue
			}

			stackThreshold := b.Diff.Preempt * b.StackLeniency

			if objectN.GetBasicData().StartTime-stackBaseObject.GetBasicData().EndTime > int64(stackThreshold) {
				break
			}

			if stackBaseObject.GetBasicData().StartPos.Dst(objectN.GetBasicData().StartPos) < stack_distance || isSlider(stackBaseObject) && stackBaseObject.GetBasicData().EndPos.Dst(objectN.GetBasicData().StartPos) < stack_distance {
				stackBaseIndex = n
				objectN.GetBasicData().StackIndex = 0
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

		if objectI.GetBasicData().StackIndex != 0 || isSpinnerBreak(objectI) {
			continue
		}

		stackThreshold := b.Diff.Preempt * b.StackLeniency

		if _, ok := objectI.(*objects.Circle); ok {
			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinnerBreak(objectN) {
					continue
				}

				if objectI.GetBasicData().StartTime-objectN.GetBasicData().EndTime > int64(stackThreshold) {
					break
				}

				if n < extendedStartIndex {
					objectN.GetBasicData().StackIndex = 0
					extendedStartIndex = n
				}

				if isSlider(objectN) && objectN.GetBasicData().EndPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					offset := objectI.GetBasicData().StackIndex - objectN.GetBasicData().StackIndex + 1
					for j := n + 1; j <= i; j++ {
						objectJ := hitObjects[j]
						if objectN.GetBasicData().EndPos.Dst(objectJ.GetBasicData().StartPos) < stack_distance {
							objectJ.GetBasicData().StackIndex -= offset
						}
					}

					break
				}

				if objectN.GetBasicData().StartPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					objectN.GetBasicData().StackIndex = objectI.GetBasicData().StackIndex + 1
					objectI = objectN
				}

			}
		} else if isSlider(objectI) {

			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinnerBreak(objectN) {
					continue
				}

				if objectI.GetBasicData().StartTime-objectN.GetBasicData().StartTime > int64(stackThreshold) {
					break
				}

				if objectN.GetBasicData().EndPos.Dst(objectI.GetBasicData().StartPos) < stack_distance {
					objectN.GetBasicData().StackIndex = objectI.GetBasicData().StackIndex + 1
					objectI = objectN
				}

			}
		}

	}

	for _, v := range hitObjects {
		if !isSpinnerBreak(v) {
			sc := -float32(v.GetBasicData().StackIndex) * float32(b.Diff.CircleRadius) / 10
			v.GetBasicData().StackOffset = vector.NewVec2f(sc, sc)
			v.GetBasicData().StartPos = v.GetBasicData().StartPos.Add(v.GetBasicData().StackOffset)
			v.GetBasicData().EndPos = v.GetBasicData().EndPos.Add(v.GetBasicData().StackOffset)
			v.UpdateStacking()
		}
	}
}
