package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

type scoreV1Processor struct {
	score           int64
	combo           int64
	modMultiplier   float64
	scoreMultiplier float64
}

func newScoreV1Processor() *scoreV1Processor {
	return &scoreV1Processor{}
}

func (s *scoreV1Processor) Init(beatMap *beatmap.BeatMap, player *difficultyPlayer) {
	s.modMultiplier = player.diff.GetScoreMultiplier()

	pauses := int64(0)
	for _, p := range beatMap.Pauses {
		pauses += int64(p.GetEndTime() - p.GetStartTime())
	}

	drainTime := float32((int64(beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime()) - int64(beatMap.HitObjects[0].GetStartTime()) - pauses) / 1000)

	// HACK: we need to cast to float32 then to float64 to lose some precision but calculate them again as float64s to have matching results with osu!stable
	s.scoreMultiplier = math.RoundToEven((float64(float32(beatMap.Diff.GetHP())) + float64(float32(beatMap.Diff.GetOD())) + float64(float32(beatMap.Diff.GetCS())) + float64(mutils.Clamp(float32(len(beatMap.HitObjects))/drainTime*8, 0, 16))) / 38 * 5)
}

func (s *scoreV1Processor) AddResult(result HitResult, comboResult ComboResult) {
	combo := max(s.combo-1, 0)

	if result != SliderMiss && result != Miss {
		increase := result.ScoreValue()

		if result&RawHits > 0 {
			s.score += increase
		} else {
			s.score += increase + int64(float64(increase)*float64(combo)*s.scoreMultiplier*s.modMultiplier/25.0)
		}
	}

	if comboResult == Reset || result == Miss {
		s.combo = 0
	} else if comboResult == Increase {
		s.combo++
	}
}

func (s *scoreV1Processor) ModifyResult(result HitResult, _ HitObject) HitResult {
	return result
}

func (s *scoreV1Processor) GetScore() int64 {
	return s.score
}

func (s *scoreV1Processor) GetCombo() int64 {
	return s.combo
}
