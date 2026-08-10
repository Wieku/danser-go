package pp260727

import (
	"log"
	"path/filepath"

	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp26xxxx"
	"github.com/wieku/danser-go/app/settings"
)

const CurrentVersion = 20260729

type DifficultyCalculator struct {
	fallback api.IDifficultyCalculator
}

func NewDifficultyCalculator() api.IDifficultyCalculator {
	return &DifficultyCalculator{fallback: pp26xxxx.NewDifficultyCalculator()}
}

func (calculator *DifficultyCalculator) CalculateSingle(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) api.Attributes {
	variant, err := currentEngine.ensureVariantForPath(calculator.beatmapPath(bMap), diff)
	if err != nil || len(variant.attributes) == 0 {
		if err != nil {
			log.Printf("Current difficulty calculator unavailable, falling back to the March 2026 preview: %v", err)
		}
		return calculator.fallback.CalculateSingle(bMap, diff)
	}

	return variant.attributes[len(variant.attributes)-1]
}

func (calculator *DifficultyCalculator) CalculateStep(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) []api.Attributes {
	variant, err := currentEngine.ensureVariantForPath(calculator.beatmapPath(bMap), diff)
	if err != nil || len(variant.attributes) != len(bMap.HitObjects) {
		if err != nil {
			log.Printf("Current difficulty calculator unavailable, falling back to the March 2026 preview: %v", err)
		} else {
			log.Printf("Current difficulty calculator returned %d objects, expected %d; falling back to the March 2026 preview", len(variant.attributes), len(bMap.HitObjects))
		}
		return calculator.fallback.CalculateStep(bMap, diff)
	}

	attributes := make([]api.Attributes, len(variant.attributes))
	copy(attributes, variant.attributes)
	return attributes
}

func (calculator *DifficultyCalculator) CalculateStrainPeaks(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) api.StrainPeaks {
	variant, err := currentEngine.ensureVariantForPath(calculator.beatmapPath(bMap), diff)
	if err != nil {
		log.Printf("Current strain calculator unavailable, falling back to the March 2026 preview: %v", err)
		return calculator.fallback.CalculateStrainPeaks(bMap, diff)
	}

	return variant.strains
}

func (calculator *DifficultyCalculator) beatmapPath(bMap *beatmap.BeatMap) string {
	return filepath.Join(settings.General.GetSongsDir(), bMap.Dir, bMap.File)
}

func (calculator *DifficultyCalculator) GetVersion() int {
	return CurrentVersion
}

func (calculator *DifficultyCalculator) GetVersionMessage() string {
	return "2026-07-29: osu! 2026 Q2 SR & PP (reading/harmonic speed)"
}
