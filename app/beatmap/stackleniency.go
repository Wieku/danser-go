package beatmap

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
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

func calculateStackLeniency(b *BeatMap) {
	if !settings.Objects.StackEnabled {
		return
	}

	diffNM := difficulty.NewDifficulty(b.Diff.GetHPDrain(), b.Diff.GetCS(), b.Diff.GetOD(), b.Diff.GetAR())
	diffEZ := difficulty.NewDifficulty(b.Diff.GetHPDrain(), b.Diff.GetCS(), b.Diff.GetOD(), b.Diff.GetAR())
	diffHR := difficulty.NewDifficulty(b.Diff.GetHPDrain(), b.Diff.GetCS(), b.Diff.GetOD(), b.Diff.GetAR())

	diffEZ.SetMods(difficulty.Easy)
	diffHR.SetMods(difficulty.HardRock)

	processStacking(b.HitObjects, b.Version, diffNM, b.StackLeniency)
	processStacking(b.HitObjects, b.Version, diffEZ, b.StackLeniency)
	processStacking(b.HitObjects, b.Version, diffHR, b.StackLeniency)

	for _, v := range b.HitObjects {
		v.UpdateStacking()
	}
}

func processStacking(hitObjects []objects.IHitObject, version int, diff *difficulty.Difficulty, stackLeniency float64) {
	stackThreshold := math.Floor(diff.Preempt * stackLeniency)
	modifiers := diff.Mods

	for _, v := range hitObjects {
		v.SetStackIndex(0, diff.Mods)
	}

	if version >= 6 {
		applyNewStacking(hitObjects, modifiers, stackThreshold)
	} else {
		applyOldStacking(hitObjects, modifiers, stackThreshold)
	}

	for _, v := range hitObjects {
		if isSpinner(v) {
			v.SetStackIndex(0, difficulty.None)
			v.SetStackIndex(0, difficulty.Easy)
			v.SetStackIndex(0, difficulty.HardRock)
		} else {
			v.SetStackOffset(-float32(v.GetStackIndex(modifiers)) * float32(diff.CircleRadius) / 10, modifiers)
		}
	}
}

func applyNewStacking(hitObjects []objects.IHitObject, modifiers difficulty.Modifier, stackThreshold float64) {
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

			if objectN.GetStartTime()-stackIHitObject.GetEndTime() > stackThreshold {
				break
			}

			if stackIHitObject.GetStartPosition().Dst(objectN.GetStartPosition()) < stackDistance || isSlider(stackIHitObject) && stackIHitObject.GetEndPosition().Dst(objectN.GetStartPosition()) < stackDistance {
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

				if objectI.GetStartTime()-objectN.GetEndTime() > stackThreshold {
					break
				}

				if n < extendedStartIndex {
					objectN.SetStackIndex(0, modifiers)
					extendedStartIndex = n
				}

				if isSlider(objectN) && objectN.GetEndPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					offset := objectI.GetStackIndex(modifiers) - objectN.GetStackIndex(modifiers) + 1
					for j := n + 1; j <= i; j++ {
						objectJ := hitObjects[j]
						if objectN.GetEndPosition().Dst(objectJ.GetStartPosition()) < stackDistance {
							objectJ.SetStackIndex(objectJ.GetStackIndex(modifiers) - offset, modifiers)
						}
					}

					break
				}

				if objectN.GetStartPosition().Dst(objectI.GetStartPosition()) < stackDistance {
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

				if objectI.GetStartTime()-objectN.GetStartTime() > stackThreshold {
					break
				}

				if objectN.GetEndPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					objectN.SetStackIndex(objectI.GetStackIndex(modifiers) + 1, modifiers)
					objectI = objectN
				}

			}
		}
	}
}

func applyOldStacking(hitObjects []objects.IHitObject, modifiers difficulty.Modifier, stackThreshold float64) {
	for i := 0; i < len(hitObjects); i++ {
		objectI := hitObjects[i]

		startTime := objectI.GetEndTime()

		if objectI.GetStackIndex(modifiers) == 0 || isSlider(objectI) {
			sliderStack := int64(0)

			for n := i+1; n < len(hitObjects); n++ {
				objectN := hitObjects[n]

				if objectN.GetStartTime() - stackThreshold > startTime {
					break
				}

				if objectN.GetStartPosition().Dst(objectI.GetStartPosition()) < stackDistance {
					objectI.SetStackIndex(objectI.GetStackIndex(modifiers) + 1, modifiers)
					startTime = objectN.GetEndTime()
				} else if objectN.GetStartPosition().Dst(objectI.GetEndPosition()) < stackDistance {
					sliderStack++
					objectN.SetStackIndex(objectN.GetStackIndex(modifiers) - sliderStack, modifiers)
					startTime = objectN.GetEndTime()
				}
			}
		}
	}
}