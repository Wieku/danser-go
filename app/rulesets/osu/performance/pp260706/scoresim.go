package pp260706

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/framework/math/mutils"
)

type scoreSimulator struct {
	breaks []*beatmap.Pause
	diff   *difficulty.Difficulty

	ScoreMultiplier float64

	combo      int64
	ComboScore int64

	nestedScore          int64
	NestedScorePerObject float64

	hitObjects int
	circles    int
	sliders    int
	spinners   int
	oStartTime float64
}

func newScoreSim(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) *scoreSimulator {
	sim := &scoreSimulator{
		breaks: bMap.Pauses,
		diff:   diff,
	}

	pauses := int64(0)
	for _, p := range bMap.Pauses {
		pauses += int64(p.GetEndTime() - p.GetStartTime())
	}

	drainTime := float32((int64(bMap.HitObjects[len(bMap.HitObjects)-1].GetStartTime()) - int64(bMap.HitObjects[0].GetStartTime()) - pauses) / 1000)

	// HACK: we need to cast to float32 then to float64 to lose some precision but calculate them again as float64s to have matching results with osu!stable
	sim.ScoreMultiplier = math.RoundToEven((float64(float32(diff.GetHP())) + float64(float32(diff.GetOD())) + float64(float32(diff.GetCS())) + float64(mutils.Clamp(float32(len(bMap.HitObjects))/drainTime*8, 0, 16))) / 38 * 5)

	return sim
}

func (s *scoreSimulator) AddFirst(obj objects.IHitObject) {
	s.oStartTime = obj.GetStartTime()
	s.add(obj)
}

func (s *scoreSimulator) Add(obj *preprocessing.DifficultyObject, first bool) {
	if first {
		s.AddFirst(obj.LastObject)
	}

	s.add(obj.BaseObject)
}

func (s *scoreSimulator) add(obj objects.IHitObject) {
	s.hitObjects++

	if slider, ok := obj.(*preprocessing.LazySlider); ok {
		obj = slider.Slider
	}

	if slider, ok := obj.(*objects.Slider); ok {
		s.sliders++
		for _, p := range slider.ScorePointsLazer {
			if !p.IsReverse && !p.LastPoint {
				s.nestedScore += 10
			}
		}

		s.nestedScore += (int64(slider.RepeatCount) + 1) * 30

		s.combo += int64(len(slider.ScorePointsLazer) + 1)
	} else if obj.GetType() == objects.SPINNER {
		s.spinners++
		s.nestedScore += calculateSpinnerScore(obj)
	} else {
		s.circles++
	}

	s.ComboScore += int64(float64(max(0, s.combo-1)) * (300.0 / 25 * s.ScoreMultiplier))

	s.NestedScorePerObject = float64(s.nestedScore) / float64(s.hitObjects)

	if obj.GetType() != objects.SLIDER {
		s.combo++
	}
}

func calculateSpinnerScore(spinner objects.IHitObject) int64 {
	const spinScore = 100
	const bonusSpinScore = 1000

	// The spinner object applies a lenience because gameplay mechanics differ from osu-stable.
	// We'll redo the calculations to match osu-stable here...
	const maximumRotationsPerSecond = 477.0 / 60

	// Normally, this value depends on the final overall difficulty. For simplicity, we'll only consider the worst case that maximises bonus score.
	// As we're primarily concerned with computing the maximum theoretical final score,
	// this will have the final effect of slightly underestimating bonus score achieved on stable when converting from score V1.
	const minimumRotationsPerSecond = 3

	secondsDuration := spinner.GetDuration() / 1000

	// The total amount of half spins possible for the entire spinner.
	totalHalfSpinsPossible := int(secondsDuration * maximumRotationsPerSecond * 2)
	// The amount of half spins that are required to successfully complete the spinner (i.e. get a 300).
	halfSpinsRequiredForCompletion := int(secondsDuration * minimumRotationsPerSecond)
	// To be able to receive bonus points, the spinner must be rotated another 1.5 times.
	halfSpinsRequiredBeforeBonus := halfSpinsRequiredForCompletion + 3

	var score int64

	fullSpins := totalHalfSpinsPossible / 2

	// Normal spin score
	score += int64(spinScore * fullSpins)

	bonusSpins := (totalHalfSpinsPossible - halfSpinsRequiredBeforeBonus) / 2

	// Reduce amount of bonus spins because we want to represent the more average case, rather than the best one.
	bonusSpins = max(0, bonusSpins-fullSpins/2)

	score += int64(bonusSpinScore * bonusSpins)

	return score
}
