package beatmap

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
)

//Original code by: https://github.com/ppy/osu/blob/master/osu.Game.Rulesets.Osu/Beatmaps/OsuBeatmapProcessor.cs

func isSpinner(obj objects.IHitObject) bool {
	_, ok2 := obj.(*objects.Spinner)
	return ok2
}

func isSlider(obj objects.IHitObject) bool {
	_, ok1 := obj.(*objects.Slider)
	return ok1
}

func calculateStackLeniency(b *BeatMap) {
	if !settings.Objects.StackEnabled {
		return
	}

	diffNM := difficulty.NewDifficulty(b.Diff.GetHPDrain(), b.Diff.GetCS(), b.Diff.GetOD(), b.Diff.GetAR())
	diffEZ := difficulty.NewDifficulty(b.Diff.GetHPDrain(), b.Diff.GetCS(), b.Diff.GetOD(), b.Diff.GetAR())
	diffHR := difficulty.NewDifficulty(b.Diff.GetHPDrain(), b.Diff.GetCS(), b.Diff.GetOD(), b.Diff.GetAR())

	diffEZ.SetMods(difficulty.Easy)
	diffHR.SetMods(difficulty.HardRock)

	processStacking(b.HitObjects, diffNM, b.StackLeniency)
	processStacking(b.HitObjects, diffEZ, b.StackLeniency)
	processStacking(b.HitObjects, diffHR, b.StackLeniency)

	for _, v := range b.HitObjects {
		v.UpdateStacking()
	}
}

func processStacking(hitObjects []objects.IHitObject, diff *difficulty.Difficulty, stackLeniency float64) {
	stack_distance := float32(3.0)
	stackThreshold := diff.Preempt * stackLeniency
	modifiers := diff.Mods

	for _, v := range hitObjects {
		v.SetStackIndex(0, diff.Mods)
	}

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

			if objectN.GetStartTime()-stackIHitObject.GetEndTime() > int64(stackThreshold) {
				break
			}

			if stackIHitObject.GetStartPosition().Dst(objectN.GetStartPosition()) < stack_distance || isSlider(stackIHitObject) && stackIHitObject.GetEndPosition().Dst(objectN.GetStartPosition()) < stack_distance {
				stackBaseIndex = n
				objectN.SetStackIndex(0, modifiers)
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

		if objectI.GetStackIndex(modifiers) != 0 || isSpinner(objectI) {
			continue
		}

		if !isSlider(objectI) && !isSpinner(objectI) {
			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinner(objectN) {
					continue
				}

				if objectI.GetStartTime()-objectN.GetEndTime() > int64(stackThreshold) {
					break
				}

				if n < extendedStartIndex {
					objectN.SetStackIndex(0, modifiers)
					extendedStartIndex = n
				}

				if isSlider(objectN) && objectN.GetEndPosition().Dst(objectI.GetStartPosition()) < stack_distance {
					offset := objectI.GetStackIndex(modifiers) - objectN.GetStackIndex(modifiers) + 1
					for j := n + 1; j <= i; j++ {
						objectJ := hitObjects[j]
						if objectN.GetEndPosition().Dst(objectJ.GetStartPosition()) < stack_distance {
							objectJ.SetStackIndex(objectJ.GetStackIndex(modifiers) - offset, modifiers)
						}
					}

					break
				}

				if objectN.GetStartPosition().Dst(objectI.GetStartPosition()) < stack_distance {
					objectN.SetStackIndex(objectI.GetStackIndex(modifiers) + 1, modifiers)
					objectI = objectN
				}

			}
		} else if isSlider(objectI) {

			for n--; n >= 0; n-- {
				objectN := hitObjects[n]

				if isSpinner(objectN) {
					continue
				}

				if objectI.GetStartTime()-objectN.GetStartTime() > int64(stackThreshold) {
					break
				}

				if objectN.GetEndPosition().Dst(objectI.GetStartPosition()) < stack_distance {
					objectN.SetStackIndex(objectI.GetStackIndex(modifiers) + 1, modifiers)
					objectI = objectN
				}

			}
		}

		for _, v := range hitObjects {
			v.SetStackOffset(-float32(v.GetStackIndex(modifiers)) * float32(diff.CircleRadius) / 10, modifiers)
		}
	}
}
