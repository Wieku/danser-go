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
	comboPartMax  float64
	comboPart     float64

	hitMap map[HitResult]int64

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
	s.hitMap = make(map[HitResult]int64)

	for _, o := range beatMap.HitObjects {
		if o.GetType() == objects.CIRCLE || o.GetType() == objects.SPINNER {
			s.AddResult(Hit300, Increase)
		} else if slider, ok := o.(*objects.Slider); ok {
			s.AddResult(Hit300, Increase)

			for i := range slider.ScorePoints {
				if i == len(slider.ScorePoints)-1 {
					s.AddResult(SliderEnd, Increase)
				} else {
					s.AddResult(SliderRepeat, Increase)
				}
			}
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

func (s *scoreV3Processor) AddResult(result HitResult, comboResult ComboResult) {
	if comboResult == Reset || result == Miss {
		s.combo = 0
	} else if comboResult == Increase {
		s.combo++
	}

	scoreValue := result.ScoreValueLazer()

	if result&(SpinnerPoints|SpinnerBonus) > 0 {
		s.bonus += float64(scoreValue)
	} else {
		s.comboPart += float64(scoreValue) * (1 + float64(s.combo)/10)
	}

	if result&BaseHitsM > 0 {
		s.hitMap[result&BaseHitsM]++
		s.hits++
	}

	if s.maxHits > 0 {
		acc := 1.0
		if s.hits > 0 {
			acc = float64(s.hitMap[Hit50]*50+s.hitMap[Hit100]*100+s.hitMap[Hit300]*300) / float64(s.hits*300)
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
