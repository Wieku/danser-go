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

	hitMap map[HitResult]int64

	hits    int64
	maxHits int64

	player *difficultyPlayer
	bonus  float64
}

func newScoreV2Processor() *scoreV2Processor {
	return &scoreV2Processor{}
}

func (s *scoreV2Processor) Init(beatMap *beatmap.BeatMap, player *difficultyPlayer) {
	s.player = player
	s.modMultiplier = player.diff.GetScoreMultiplier()

	s.comboPartMax = 0
	s.maxHits = 0
	s.hitMap = make(map[HitResult]int64)

	for _, o := range beatMap.HitObjects {
		if o.GetType() == objects.CIRCLE || o.GetType() == objects.SPINNER {
			s.AddResult(Hit300, Increase)
		} else if slider, ok := o.(*objects.Slider); ok {
			for j := 0; j < len(slider.TickReverse)+1; j++ {
				s.AddResult(SliderRepeat, Increase)
			}

			for j := 0; j < len(slider.TickPoints); j++ {
				s.AddResult(SliderPoint, Increase)
			}

			s.AddResult(Hit300, Hold)
		}
	}

	s.comboPartMax = s.comboPart
	s.maxHits = s.hits

	s.combo = 0
	s.hits = 0
	s.comboPart = 0
	s.bonus = 0
	s.hitMap = make(map[HitResult]int64)
}

func (s *scoreV2Processor) AddResult(result HitResult, comboResult ComboResult) {
	if comboResult == Reset || result == Miss {
		s.combo = 0
	} else if comboResult == Increase {
		s.combo++
	}

	scoreValue := scoreValueV2(result)

	if result&(SpinnerPoints|SpinnerBonus) > 0 {
		s.bonus += float64(scoreValue)
	} else {
		s.comboPart += float64(scoreValue) * (1 + float64(s.combo)/10)
	}

	if result&BaseHitsM > 0 {
		s.hitMap[result]++
		s.hits++
	}

	if s.maxHits > 0 {
		acc := float32(1.0)
		if s.hits > 0 {
			acc = float32(s.hitMap[Hit50]*50+s.hitMap[Hit100]*100+s.hitMap[Hit300]*300) / float32(s.hits*300)
		}

		s.score = int64(math.Round((s.comboPart/s.comboPartMax*700000 + math.Pow(float64(acc), 10)*(float64(s.hits)/float64(s.maxHits))*300000 + s.bonus) * s.modMultiplier))
	}
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

func scoreValueV2(result HitResult) int64 {
	scoreVal := result.ScoreValue()
	if result&SpinnerBonus > 0 {
		scoreVal = 500
	}

	return scoreVal
}
