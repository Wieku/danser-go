package api

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
)

type IDifficultyCalculator interface {
	CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty) Attributes

	// CalculateStep calculates successive star ratings for every part of a beatmap
	CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty) []Attributes

	CalculateStrainPeaks(objects []objects.IHitObject, diff *difficulty.Difficulty) StrainPeaks

	GetVersion() int
	GetVersionMessage() string
}

type IPerformanceCalculator interface {
	Calculate(attribs Attributes, score PerfScore, diff *difficulty.Difficulty) PPv2Results
}
