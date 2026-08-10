package pp260727

import (
	"log"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp26xxxx"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

type PPv2 struct {
	fallback api.IPerformanceCalculator
}

func NewPPCalculator() api.IPerformanceCalculator {
	return &PPv2{fallback: pp26xxxx.NewPPCalculator()}
}

func (calculator *PPv2) Calculate(attributes api.Attributes, score api.PerfScore, diff *difficulty.Difficulty) api.PPv2Results {
	attributes.MaxCombo = max(1, attributes.MaxCombo)

	if score.MaxCombo < 0 {
		score.MaxCombo = attributes.MaxCombo
	}
	if score.CountGreat < 0 {
		score.CountGreat = attributes.ObjectCount - score.CountOk - score.CountMeh - score.CountMiss
	}
	if score.SliderEnd < 0 {
		score.SliderEnd = attributes.Sliders
	}

	index := max(0, attributes.ObjectCount-1)
	largeTickMisses := max(0, score.SliderBreaks)
	largeTickHits := max(0, score.SliderTickTotal-largeTickMisses)

	result, err := currentEngine.calculate(diff, index, map[string]any{
		"totalScore":      score.Score,
		"isLegacyScore":   !diff.CheckModActive(difficulty.Lazer),
		"accuracy":        score.Accuracy,
		"maxCombo":        score.MaxCombo,
		"sliderEndHits":   score.SliderEnd,
		"comboBreaks":     score.SliderBreaks,
		"largeTickHits":   largeTickHits,
		"largeTickMisses": largeTickMisses,
		"greats":          score.CountGreat,
		"oks":             score.CountOk,
		"mehs":            score.CountMeh,
		"misses":          score.CountMiss,
	})
	if err != nil {
		log.Printf("Current performance calculator unavailable, falling back to the March 2026 preview: %v", err)
		return calculator.fallback.Calculate(attributes, score, diff)
	}

	return api.PPv2Results{
		Aim:        result.Performance.Aim,
		Speed:      result.Performance.Speed,
		Acc:        result.Performance.Accuracy,
		Flashlight: result.Performance.Flashlight,
		Cognition:  combineCognition(result.Performance.Reading, result.Performance.Flashlight),
		Total:      result.Performance.PP,
	}
}

func combineCognition(reading, flashlight float64) float64 {
	if reading <= 0 {
		return flashlight
	}
	if flashlight <= 0 {
		return reading
	}

	return putils.Norm(1.1, reading, flashlight*mutils.Clamp(flashlight/reading, 0.25, 1))
}
