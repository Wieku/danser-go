package newCalc

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

// utility struct that has LazyEndPosition and LazyTravelDistance for difficulty calculations
type LazySlider struct {
	*objects.Slider

	diff *difficulty.Difficulty

	LazyEndPosition vector.Vector2f
	LazyTravelDistance float32
}

func NewLazySlider(slider *objects.Slider, d *difficulty.Difficulty) *LazySlider {
	decorated := &LazySlider{
		Slider:             slider,
		diff: d,
	}

	decorated.calculateEndPosition()

	return decorated
}

func (s *LazySlider) calculateEndPosition() {
	s.LazyEndPosition = s.GetStackedStartPositionMod(s.diff.Mods)

	approxFollowCircleRadius := float32(s.diff.CircleRadius) * 3

	compute := func(time float64) {
		difference := s.GetStackedPositionAtMod(time, s.diff.Mods).Sub(s.LazyEndPosition)
		dist := difference.Len()

		if dist > approxFollowCircleRadius {
			difference = difference.Nor()
			dist -= approxFollowCircleRadius
			s.LazyEndPosition = s.LazyEndPosition.Add(difference.Scl(dist))
			s.LazyTravelDistance += dist
		}
	}

	for i, p := range s.ScorePoints {
		time := p.Time
		if i == len(s.ScorePoints) - 1 {
			time = math.Floor(math.Max(s.StartTime+(s.EndTime-s.StartTime)/2, s.EndTime-36))
		}

		compute(time)
	}
}
