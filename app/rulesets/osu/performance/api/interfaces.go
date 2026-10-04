package api

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
)

type IDifficultyCalculator interface {
	CalculateSingle(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) Attributes

	// CalculateStep calculates successive star ratings for every part of a beatmap
	CalculateStep(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) []Attributes

	CalculateStrainPeaks(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) StrainPeaks

	GetVersion() int
	GetVersionMessage() string
}

type IPerformanceCalculator interface {
	Calculate(attribs Attributes, score PerfScore, diff *difficulty.Difficulty) PPv2Results
}
