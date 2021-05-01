package newCalc

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
)

// utility struct that has LazyEndPosition and LazyTravelDistance for difficulty calculations
type LazySlider struct {
	*objects.Slider

	diff *difficulty.Difficulty

	LazyEndPosition vector.Vector2f
	LazyTravelDistance float64
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
	startPos := s.GetStackedStartPositionMod(s.diff.Mods)

	s.LazyEndPosition = startPos

	approxFollowCircleRadius := s.diff.CircleRadius * 3

	compute := func(time float64) {
		difference := startPos.Add(s.GetStackedPositionAtMod(time, s.diff.Mods)).Sub(s.LazyEndPosition)
		dist := float64(difference.Len())

		if dist > approxFollowCircleRadius {
			difference = difference.Nor()
			dist -= approxFollowCircleRadius
			s.LazyEndPosition = s.LazyEndPosition.Add(difference.Copy64().Scl(dist).Copy32())
			s.LazyTravelDistance += dist
		}
	}

	for _, p := range s.ScorePoints {
		compute(p.Time)
	}
}
