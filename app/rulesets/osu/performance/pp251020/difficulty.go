package pp251020

import (
	"log"
	"math"
	"time"

	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020/skills"
)

const (
	// StarScalingFactor is a global stars multiplier
	StarScalingFactor float64 = 0.0265
	CurrentVersion    int     = 20251020
)

type DifficultyCalculator struct{}

func NewDifficultyCalculator() api.IDifficultyCalculator {
	return &DifficultyCalculator{}
}

// getStarsFromRawValues converts raw skill values to Attributes
func (diffCalc *DifficultyCalculator) getStarsFromRawValues(aimDifficultyValue, aimNoSlidersDifficultyValue, speedDifficultyValue, rawFlashlight float64, totalHits int, diff *difficulty.Difficulty) api.Attributes {
	mechanicalDifficultyRating := calculateMechanicalDifficultyRating(aimDifficultyValue, speedDifficultyValue)
	sliderFactor := ternary(aimDifficultyValue > 0, CalculateDifficultyRating(aimNoSlidersDifficultyValue)/CalculateDifficultyRating(aimDifficultyValue), 1)

	oRatingCalculator := newOsuRatingCalculator(diff, totalHits, diff.ARReal, diff.ODReal, mechanicalDifficultyRating, sliderFactor)

	aimRating := oRatingCalculator.ComputeAimRating(aimDifficultyValue)
	speedRating := oRatingCalculator.ComputeSpeedRating(speedDifficultyValue)

	flashlightRating := 0.0
	if diff.CheckModActive(difficulty.Flashlight) {
		flashlightRating = oRatingCalculator.ComputeFlashlightRating(rawFlashlight)
	}

	baseAimPerformance := skills.DefaultDifficultyToPerformance(aimRating)
	baseSpeedPerformance := skills.DefaultDifficultyToPerformance(speedRating)
	baseFlashlightPerformance := skills.FlashlightDifficultyToPerformance(flashlightRating)

	basePerformance :=
		math.Pow(
			math.Pow(baseAimPerformance, 1.1)+
				math.Pow(baseSpeedPerformance, 1.1)+
				math.Pow(baseFlashlightPerformance, 1.1),
			1.0/1.1,
		)

	starRating := calculateStarRating(basePerformance)

	return api.Attributes{
		Total:        starRating,
		Aim:          aimRating,
		Speed:        speedRating,
		Flashlight:   flashlightRating,
		SliderFactor: sliderFactor,
		ObjectCount:  totalHits,
	}
}

// Retrieves skill values and converts to Attributes
func (diffCalc *DifficultyCalculator) getStars(aim *skills.AimSkill, aimWithoutSliders *skills.AimSkill, speed *skills.SpeedSkill, flashlight *skills.Flashlight, sim *scoreSimulator, diff *difficulty.Difficulty) api.Attributes {
	speedNotes := speed.RelevantNoteCount()

	aimDifficultStrainCount := aim.CountTopWeightedStrains()
	speedDifficultStrainCount := speed.CountTopWeightedStrains()

	aimNoSlidersTopWeightedSliderCount := aimWithoutSliders.CountTopWeightedSliders()
	aimNoSlidersDifficultStrainCount := aimWithoutSliders.CountTopWeightedStrains()

	aimTopWeightedSliderFactor := aimNoSlidersTopWeightedSliderCount / max(1, aimNoSlidersDifficultStrainCount-aimNoSlidersTopWeightedSliderCount)

	speedTopWeightedSliderCount := speed.CountTopWeightedSliders()
	speedTopWeightedSliderFactor := speedTopWeightedSliderCount / max(1, speedDifficultStrainCount-speedTopWeightedSliderCount)

	difficultSliders := aim.GetDifficultSliders()

	sliderNestedScorePerObject := sim.NestedScorePerObject
	legacyScoreBaseMultiplier := sim.ScoreMultiplier

	attr := diffCalc.getStarsFromRawValues(aim.DifficultyValue(), aimWithoutSliders.DifficultyValue(), speed.DifficultyValue(), flashlight.DifficultyValue(), sim.hitObjects, diff)

	//Total:        starRating,
	//		Aim:          aimRating,
	//		Speed:        speedRating,
	//		Flashlight:   flashlightRating,
	//		SliderFactor: sliderFactor,

	//DrainRate = drainRate,

	attr.SpeedNoteCount = speedNotes
	attr.AimDifficultStrainCount = aimDifficultStrainCount
	attr.AimDifficultSliderCount = difficultSliders
	attr.SpeedDifficultStrainCount = speedDifficultStrainCount
	attr.Circles = sim.circles
	attr.Sliders = sim.sliders
	attr.Spinners = sim.spinners
	attr.MaxCombo = int(sim.combo)

	attr.AimTopWeightedSliderFactor = aimTopWeightedSliderFactor
	attr.SpeedTopWeightedSliderFactor = speedTopWeightedSliderFactor
	attr.NestedScorePerObject = sliderNestedScorePerObject
	attr.LegacyScoreBaseMultiplier = legacyScoreBaseMultiplier
	attr.MaximumLegacyComboScore = sim.ComboScore

	return attr
}

// CalculateSingle calculates the final difficultyapi.Attributes of a map
func (diffCalc *DifficultyCalculator) CalculateSingle(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) api.Attributes {
	diffObjects := preprocessing.CreateDifficultyObjects(bMap.HitObjects, diff)

	aimSkill := skills.NewAimSkill(diff, true, false)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, false)
	speedSkill := skills.NewSpeedSkill(diff, false)
	flashlightSkill := skills.NewFlashlightSkill(diff)
	sim := newScoreSim(bMap, diff)

	for i, o := range diffObjects {
		sim.Add(o, i == 0)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	return diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, sim, diff)
}

// CalculateStep calculates successive star ratings for every part of a beatmap
func (diffCalc *DifficultyCalculator) CalculateStep(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) []api.Attributes {
	modString := difficulty.GetDiffMaskedMods(diff.Mods).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	startTime := time.Now()

	diffObjects := preprocessing.CreateDifficultyObjects(bMap.HitObjects, diff)

	aimSkill := skills.NewAimSkill(diff, true, true)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, false)
	speedSkill := skills.NewSpeedSkill(diff, true)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	stars := make([]api.Attributes, 1, len(bMap.HitObjects))

	sim := newScoreSim(bMap, diff)

	sim.AddFirst(bMap.HitObjects[0])
	stars[0] = diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, sim, diff)

	for _, o := range diffObjects {

		sim.Add(o, false)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)

		stars = append(stars, diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, sim, diff))
	}

	endTime := time.Now()

	log.Println("Calculations finished! Took ", endTime.Sub(startTime).Truncate(time.Millisecond).String())

	return stars
}

func (diffCalc *DifficultyCalculator) CalculateStrainPeaks(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) api.StrainPeaks {
	diffObjects := preprocessing.CreateDifficultyObjects(bMap.HitObjects, diff)

	aimSkill := skills.NewAimSkill(diff, true, false)
	speedSkill := skills.NewSpeedSkill(diff, false)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	peaks := api.StrainPeaks{
		Aim:        aimSkill.GetCurrentStrainPeaks(),
		Speed:      speedSkill.GetCurrentStrainPeaks(),
		Flashlight: flashlightSkill.GetCurrentStrainPeaks(),
	}

	peaks.Total = make([]float64, len(peaks.Aim))

	for i := range peaks.Aim {
		stars := diffCalc.getStarsFromRawValues(peaks.Aim[i], peaks.Aim[i], peaks.Speed[i], peaks.Flashlight[i], 1000, diff)
		peaks.Total[i] = stars.Total
	}

	return peaks
}

func (diffCalc *DifficultyCalculator) GetVersion() int {
	return CurrentVersion
}

func (diffCalc *DifficultyCalculator) GetVersionMessage() string {
	return "2025-10-20: https://osu.ppy.sh/home/news/2025-10-29-performance-points-star-rating-updates"
}

func calculateMechanicalDifficultyRating(aimDifficultyValue, speedDifficultyValue float64) float64 {
	aimValue := skills.DefaultDifficultyToPerformance(CalculateDifficultyRating(aimDifficultyValue))
	speedValue := skills.DefaultDifficultyToPerformance(CalculateDifficultyRating(speedDifficultyValue))

	totalValue := math.Pow(math.Pow(aimValue, 1.1)+math.Pow(speedValue, 1.1), 1/1.1)

	return calculateStarRating(totalValue)
}

func calculateStarRating(basePerformance float64) float64 {
	if basePerformance <= 0.00001 {
		return 0
	}

	return math.Cbrt(PerformanceBaseMultiplier) * StarScalingFactor * (math.Cbrt(100000/math.Pow(2, 1/1.1)*basePerformance) + 4)
}
