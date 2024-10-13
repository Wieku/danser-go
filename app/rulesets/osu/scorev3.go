package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"math"
)

type scoreV3Processor struct {
	score    int64
	combo    int64
	accuracy float64

	modMultiplier float64

	bonus int64

	comboPart    float64
	comboPartMax float64

	accPart    int64
	accPartMax int64

	hits          int64
	maxHits       int64
	basicHitCount int64

	player *difficultyPlayer
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
			if player.lzNoSliderAcc {
				s.AddResult(createJudgementResult(SliderStart, SliderStart, Increase, int64(o.GetStartTime()), o.GetStartPosition(), nil))
			} else {
				s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetStartTime()), o.GetStartPosition(), nil))
			}

			for i, p := range slider.ScorePointsLazer {
				if i == len(slider.ScorePoints)-1 {
					if player.lzNoSliderAcc {
						s.AddResult(createJudgementResult(LegacySliderEnd, LegacySliderEnd, Hold, int64(p.Time), p.Pos, nil))
					} else {
						s.AddResult(createJudgementResult(SliderEnd, SliderEnd, Increase, int64(p.Time), p.Pos, nil))
					}
				} else if p.IsReverse {
					s.AddResult(createJudgementResult(SliderRepeat, SliderRepeat, Increase, int64(p.Time), p.Pos, nil))
				} else {
					s.AddResult(createJudgementResult(SliderPoint, SliderPoint, Increase, int64(p.Time), p.Pos, nil))
				}
			}

			if player.lzNoSliderAcc {
				s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetEndTime()), o.GetStartPosition(), nil))
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
	s.accuracy = 1
	s.basicHitCount = 0
}

func (s *scoreV3Processor) AddResult(result JudgementResult) {
	if result.ComboResult == Reset || result.HitResult == Miss {
		s.combo = 0
	} else if result.ComboResult == Increase {
		s.combo++
	}

	if result.HitResult.IsBonus() {
		s.bonus += result.HitResult.ScoreValueLazer()
	} else if result.HitResult.AffectsAccLZ() {
		s.accPart += result.HitResult.ScoreValueLazer()
		s.accPartMax += result.MaxResult.ScoreValueLazer()

		s.hits++
	}

	if result.HitResult&BaseHitsM > 0 {
		s.basicHitCount++
	}

	// slider end misses (not classic mod!) don't propagate combo score
	if result.HitResult.AffectsAccLZ() && !(result.HitResult == SliderMiss && result.MaxResult == SliderEnd) {
		s.comboPart += float64(result.MaxResult.ScoreValueLazer()) * math.Pow(float64(s.combo), 0.5)
	}

	if s.maxHits == 0 {
		return
	}

	acc := 1.0
	if s.hits > 0 && s.accPartMax > 0 {
		acc = float64(s.accPart) / float64(s.accPartMax)
	}

	s.accuracy = acc

	comboProgress := s.comboPart / s.comboPartMax
	accProgress := float64(s.hits) / float64(s.maxHits)

	s.score = int64(math.Round(math.Round(500000*acc*comboProgress+500000*math.Pow(acc, 5)*accProgress+float64(s.bonus)) * s.modMultiplier))
}

func (s *scoreV3Processor) ModifyResult(result HitResult, src HitObject) HitResult {
	return result
}

func (s *scoreV3Processor) GetScore() int64 {
	if settings.Gameplay.LazerClassicScore {
		return int64(math.Round((math.Pow(float64(s.basicHitCount), 2)*32.57 + 100000) * float64(s.score) / 1000000))
	}

	return s.score
}

func (s *scoreV3Processor) GetCombo() int64 {
	return s.combo
}

func (s *scoreV3Processor) GetAccuracy() float64 {
	return s.accuracy
}
