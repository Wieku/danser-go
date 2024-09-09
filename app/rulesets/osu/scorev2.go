package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
)

type scoreV2Processor struct {
	score         int64
	combo         int64
	modMultiplier float64
	comboPartMax  float64
	comboPart     float64

	rawScore int64
	hits     int64
	maxHits  int64

	player *difficultyPlayer
	bonus  float64

	accuracy float64
}

func newScoreV2Processor() *scoreV2Processor {
	return &scoreV2Processor{}
}

func (s *scoreV2Processor) Init(beatMap *beatmap.BeatMap, player *difficultyPlayer) {
	s.player = player
	s.modMultiplier = player.diff.GetScoreMultiplier()

	s.comboPartMax = 0
	s.maxHits = 0

	for _, o := range beatMap.HitObjects {
		if o.GetType() == objects.SPINNER {
			s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetEndTime()), o.GetStartPosition(), nil))
		} else if slider, ok := o.(*objects.Slider); ok {
			for j := 0; j < len(slider.TickReverse)+1; j++ {
				s.AddResult(createJudgementResult(SliderRepeat, SliderRepeat, Increase, 0, o.GetStartPosition(), nil))
			}

			for j := 0; j < len(slider.TickPoints); j++ {
				s.AddResult(createJudgementResult(SliderPoint, SliderPoint, Increase, 0, o.GetStartPosition(), nil))
			}

			s.AddResult(createJudgementResult(Hit300, Hit300, Hold, int64(o.GetEndTime()), o.GetEndPosition(), nil))
		} else {
			s.AddResult(createJudgementResult(Hit300, Hit300, Increase, int64(o.GetStartTime()), o.GetStartPosition(), nil))
		}
	}

	s.comboPartMax = s.comboPart
	s.maxHits = s.hits

	s.combo = 0
	s.rawScore = 0
	s.hits = 0
	s.comboPart = 0
	s.bonus = 0
	s.accuracy = 1
}

func (s *scoreV2Processor) AddResult(result JudgementResult) {
	if result.ComboResult == Reset || result.HitResult == Miss {
		s.combo = 0
	} else if result.ComboResult == Increase {
		s.combo++
	}

	scoreValue := result.HitResult.ScoreValueV2()

	if result.HitResult.IsBonus() {
		s.bonus += float64(scoreValue)
	} else {
		s.comboPart += float64(scoreValue) * (1 + float64(s.combo)/10)
	}

	if r := result.HitResult & BaseHitsM; r > 0 {
		s.rawScore += r.ScoreValueV2()
		s.hits++
	}

	if s.maxHits == 0 {
		return
	}

	acc := float32(1.0)
	acc2 := 1.0

	if s.hits > 0 {
		acc = float32(s.rawScore) / float32(s.hits*300)
		acc2 = float64(s.rawScore) / float64(s.hits*300)
	}

	s.score = int64(math.Round((s.comboPart/s.comboPartMax*700000 + math.Pow(float64(acc), 10)*(float64(s.hits)/float64(s.maxHits))*300000 + s.bonus) * s.modMultiplier))

	s.accuracy = acc2
}

func (s *scoreV2Processor) ModifyResult(result HitResult, src HitObject) HitResult {
	if result&BaseHitsM > 0 {
		if slider, ok := src.(*Slider); ok {
			startResult := slider.GetStartResult(s.player)

			if result&Hit300 > 0 && startResult&Hit300 > 0 {
				return Hit300
			} else if result&(Hit300|Hit100) > 0 && startResult&(Hit300|Hit100) > 0 {
				return Hit100
			} else if result != Miss {
				return Hit50
			}
		}
	}

	return result
}

func (s *scoreV2Processor) GetScore() int64 {
	return s.score
}

func (s *scoreV2Processor) GetCombo() int64 {
	return s.combo
}

func (s *scoreV2Processor) GetAccuracy() float64 {
	return s.accuracy
}
