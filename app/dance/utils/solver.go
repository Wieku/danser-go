package utils

import (
	"cmp"
	"math"
	"slices"

	"github.com/wieku/danser-go/app/beatmap/objects"
)

func Solve2B(queue []objects.IHitObject) []objects.IHitObject {
	// Resolving 2B conflicts
	for i := range queue {
		s, ok := queue[i].(*objects.Slider)
		if !ok {
			continue
		}

		found := false

		// We need to loop backwards to look for overlapping spinners (p) that are separated by circles:
		// --ppppppppppppppppp------
		// ----------c--c-----------
		// ---------------ssssssss--
		// Looking just by i-1 (like i+1 in forward detection) wouldn't detect that scenario because objects
		// are not sorted by end times
		for j := i - 1; j >= 0; j-- {
			if o := queue[j]; o.GetEndTime() >= s.GetStartTime() {
				queue = PreprocessQueue(i, queue, true)
				found = true
				break
			}
		}

		// If no conflict was detected in the past then look one object ahead, no looping is needed in this scenario
		if !found && i+1 < len(queue) {
			if o := queue[i+1]; o.GetStartTime() <= s.GetEndTime() {
				queue = PreprocessQueue(i, queue, true)
			}
		}
	}

	// Second 2B pass for spinners
	for i := 0; i < len(queue); i++ {
		spinner, ok := queue[i].(*objects.Spinner)
		if !ok {
			continue
		}

		var subSpinners []objects.IHitObject
		startTime := spinner.GetStartTime()

		// Adjust spinner's start time if it overlaps with previous circle/slider point
		if i-1 >= 0 && math.Abs(queue[i-1].GetEndTime()-startTime) < 1 {
			startTime += 30
		}

		for j := i + 1; j < len(queue); j++ {
			nextObj := queue[j]

			if nextObj.GetStartTime() > spinner.GetEndTime() {
				break
			}

			// Generate a spinner if gap is large enough
			if endTime := nextObj.GetStartTime() - 30; endTime > startTime {
				subSpinners = append(subSpinners, objects.NewDummySpinner(startTime, endTime))
			}

			startTime = nextObj.GetEndTime() + 30
		}

		if startTime == spinner.GetStartTime() {
			continue
		}

		// Generate a spinner if there's still time left
		if spinner.GetEndTime() > startTime {
			subSpinners = append(subSpinners, objects.NewDummySpinner(startTime, spinner.GetEndTime()))
		}

		queue = append(queue[:i], append(subSpinners, queue[i+1:]...)...)

		if len(subSpinners) == 0 {
			i--
			continue
		}

		slices.SortStableFunc(queue, func(a, b objects.IHitObject) int { return cmp.Compare(a.GetStartTime(), b.GetStartTime()) })
	}

	return queue
}
