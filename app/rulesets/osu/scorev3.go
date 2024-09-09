package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
)

type scoreV3Processor struct {
	score         int64
	combo         int64
	modMultiplier float64

	comboPart    float64
	comboPartMax float64

	accPart    int64
	accPartMax int64

	hits    int64
	maxHits int64

	player *difficultyPlayer
	bonus  float64
}

func newScoreV3Processor() *scoreV3Processor {
	return &scoreV3Processor{}
}

func (s *scoreV3Processor) Init(beatMap *beatmap.BeatMap, player *difficultyPlayer) {
	s.player = player
	s.modMultiplier = player.diff.GetScoreMultiplier()

	s.comboPartMax = 0
	s.maxHits = 0

	for _, o := range beatMap.HitObjects {
		if o.GetType() == objects.SPINNER {
			s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetEndTime()), o.GetStartPosition(), nil))
		} else if slider, ok := o.(*objects.Slider); ok {
			s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetEndTime()), o.GetStartPosition(), nil))

			for i, p := range slider.ScorePointsLazer {
				if i == len(slider.ScorePoints)-1 {
					s.AddResult(createJudgementResult(SliderEnd, SliderEnd, Increase, int64(p.Time), p.Pos, nil))
				} else if p.IsReverse {
					s.AddResult(createJudgementResult(SliderRepeat, SliderRepeat, Increase, int64(p.Time), p.Pos, nil))
				} else {
					s.AddResult(createJudgementResult(SliderPoint, SliderPoint, Increase, int64(p.Time), p.Pos, nil))
				}
			}
		} else {
			s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetStartTime()), o.GetStartPosition(), nil))
		}
	}

	s.comboPartMax = s.comboPart
	s.maxHits = s.hits

	s.combo = 0
	s.hits = 0
	s.comboPart = 0
	s.bonus = 0
	s.accPart = 0
	s.accPartMax = 0
}

func (s *scoreV3Processor) AddResult(result JudgementResult) {
	if result.ComboResult == Reset || result.HitResult == Miss {
		s.combo = 0
	} else if result.ComboResult == Increase {
		s.combo++
	}

	if result.HitResult.IsBonus() {
		s.bonus += float64(result.HitResult.ScoreValueLazer())
	} else if result.HitResult.AffectsAccLZ() {
		s.comboPart += float64(result.MaxResult.ScoreValueLazer()) * math.Pow(float64(s.combo), 0.5)

		s.accPart += result.HitResult.ScoreValueLazer()
		s.accPartMax += result.MaxResult.ScoreValueLazer()

		s.hits++
	}

	if s.maxHits > 0 {
		acc := 1.0
		if s.hits > 0 {
			acc = float64(s.accPart) / float64(s.accPartMax)
		}

		cPart := s.comboPart / s.comboPartMax
		aPart := float64(s.hits) / float64(s.maxHits)

		s.score = int64(math.Round(math.Round(500000*acc*cPart+500000*math.Pow(acc, 5)*aPart+s.bonus) * s.modMultiplier))
	}
}

func (s *scoreV3Processor) ModifyResult(result HitResult, src HitObject) HitResult {
	return result
}

func (s *scoreV3Processor) GetScore() int64 {
	return s.score
}

func (s *scoreV3Processor) GetCombo() int64 {
	return s.combo
}
