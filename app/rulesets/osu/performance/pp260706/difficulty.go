package pp260706

import (
	"log"
	"math"
	"time"

	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/skills"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	CurrentVersion int = 20260706
)

type DifficultyCalculator struct{}

func NewDifficultyCalculator() api.IDifficultyCalculator {
	return &DifficultyCalculator{}
}

// getStarsFromRawValues converts raw skill values to Attributes
func (diffCalc *DifficultyCalculator) getStarsFromRawValues(aimDifficultyValue, aimNoSlidersDifficultyValue, speedDifficultyValue, rawFlashlight, rawReading float64, totalHits int, diff *difficulty.Difficulty) api.Attributes {
	aimRating := CalculateAimDifficultyRating(aimDifficultyValue)
	aimNoSlidersRating := CalculateAimDifficultyRating(aimNoSlidersDifficultyValue)

	sliderFactor := ternary(aimDifficultyValue > 0, aimNoSlidersRating/aimRating, 1)

	speedRating := CalculateDifficultyRating(speedDifficultyValue)

	flashlightRating := 0.0
	if diff.CheckModActive(difficulty.Flashlight) {
		flashlightRating = CalculateDifficultyRating(rawFlashlight)
	}

	readingRating := CalculateDifficultyRating(rawReading)

	baseAimPerformance := skills.DefaultDifficultyToPerformance(aimRating)
	baseSpeedPerformance := skills.HarmonicDifficultyToPerformance(speedRating)
	baseReadingPerformance := skills.HarmonicDifficultyToPerformance(readingRating)
	baseFlashlightPerformance := skills.FlashlightDifficultyToPerformance(flashlightRating)
	baseCognitionPerformance := sumCognitionDifficulty(baseReadingPerformance, baseFlashlightPerformance)

	basePerformance := putils.Norm(1.1, baseAimPerformance, baseSpeedPerformance, baseCognitionPerformance)

	starRating := calculateStarRating(basePerformance)

	return api.Attributes{
		Total:        starRating,
		Aim:          aimRating,
		Speed:        speedRating,
		Flashlight:   flashlightRating,
		Reading:      readingRating,
		SliderFactor: sliderFactor,
		ObjectCount:  totalHits,
	}
}

func sumCognitionDifficulty(reading, flashlight float64) float64 {
	if reading <= 0 {
		return flashlight
	}

	if flashlight <= 0 {
		return reading
	}

	// Nerf flashlight value in cognition sum when reading is greater than flashlight
	return putils.Norm(1.1, reading, flashlight*mutils.Clamp(flashlight/reading, 0.25, 1.0))
}

// Retrieves skill values and converts to Attributes
func (diffCalc *DifficultyCalculator) getStars(aim *skills.AimSkill, aimWithoutSliders *skills.AimSkill, speed *skills.SpeedSkill, flashlight *skills.Flashlight, reading *skills.ReadingSkill, sim *scoreSimulator, diff *difficulty.Difficulty) api.Attributes {
	attr := diffCalc.getStarsFromRawValues(aim.DifficultyValue(), aimWithoutSliders.DifficultyValue(), speed.DifficultyValue(), flashlight.DifficultyValue(), reading.DifficultyValue(), sim.hitObjects, diff)

	aimDifficultStrainCount := aim.CountTopWeightedStrains()
	speedDifficultStrainCount := speed.CountTopWeightedStrains()
	readingDifficultNoteCount := reading.CountTopWeightedStrains()

	speedNotes := speed.RelevantNoteCount()

	aimNoSlidersTopWeightedSliderCount := aimWithoutSliders.CountTopWeightedSliders()
	aimNoSlidersDifficultStrainCount := aimWithoutSliders.CountTopWeightedStrains()

	aimTopWeightedSliderFactor := aimNoSlidersTopWeightedSliderCount / max(1, aimNoSlidersDifficultStrainCount-aimNoSlidersTopWeightedSliderCount)

	speedTopWeightedSliderCount := speed.CountTopWeightedSliders()
	speedTopWeightedSliderFactor := speedTopWeightedSliderCount / max(1, speedDifficultStrainCount-speedTopWeightedSliderCount)

	difficultSliders := aim.GetDifficultSliders()

	sliderNestedScorePerObject := sim.NestedScorePerObject
	legacyScoreBaseMultiplier := sim.ScoreMultiplier

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

	attr.ReadingDifficultNoteCount = readingDifficultNoteCount

	return attr
}

// CalculateSingle calculates the final difficultyapi.Attributes of a map
func (diffCalc *DifficultyCalculator) CalculateSingle(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) api.Attributes {
	diffObjects := preprocessing.CreateDifficultyObjects(bMap.HitObjects, diff)

	aimSkill := skills.NewAimSkill(diff, true, false)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, false)
	speedSkill := skills.NewSpeedSkill(diff, false)
	flashlightSkill := skills.NewFlashlightSkill(diff, func() int {
		return len(bMap.HitObjects)
	})
	readingSkill := skills.NewReadingSkill(diff, false)
	sim := newScoreSim(bMap, diff)

	for i, o := range diffObjects {
		sim.Add(o, i == 0)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
		readingSkill.Process(o)
	}

	return diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, readingSkill, sim, diff)
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
	flashlightSkill := skills.NewFlashlightSkill(diff, func() int {
		return len(bMap.HitObjects) // weird
	})
	readingSkill := skills.NewReadingSkill(diff, true)

	stars := make([]api.Attributes, 1, len(bMap.HitObjects))

	sim := newScoreSim(bMap, diff)

	sim.AddFirst(bMap.HitObjects[0])
	stars[0] = diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, readingSkill, sim, diff)

	for _, o := range diffObjects {
		sim.Add(o, false)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
		readingSkill.Process(o)

		stars = append(stars, diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, readingSkill, sim, diff))
	}

	log.Println("Calculations finished! Took ", time.Since(startTime).Truncate(time.Millisecond).String())

	return stars
}

func (diffCalc *DifficultyCalculator) CalculateStrainPeaks(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) api.StrainPeaks {
	diffObjects := preprocessing.CreateDifficultyObjects(bMap.HitObjects, diff)

	aimSkill := skills.NewAimSkill(diff, true, false)
	speedSkill := skills.NewSpeedSkill(diff, false)
	flashlightSkill := skills.NewFlashlightSkill(diff, func() int {
		return 0
	})
	readingSkill := skills.NewReadingSkill(diff, false)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
		readingSkill.Process(o)
	}

	peaks := api.StrainPeaks{
		Aim:        aimSkill.GetCurrentStrainPeaks(),
		Speed:      speedSkill.GetCurrentStrainPeaks(),
		Flashlight: flashlightSkill.GetCurrentStrainPeaks(),
	}
	readingPeaks := readingSkill.GetCurrentStrainPeaks()

	peaks.Total = make([]float64, len(peaks.Aim))

	for i := range peaks.Aim {
		stars := diffCalc.getStarsFromRawValues(peaks.Aim[i], peaks.Aim[i], peaks.Speed[i], peaks.Flashlight[i], readingPeaks[i], len(bMap.HitObjects), diff)
		peaks.Total[i] = stars.Total
	}

	return peaks
}

func (diffCalc *DifficultyCalculator) GetVersion() int {
	return CurrentVersion
}

func (diffCalc *DifficultyCalculator) GetVersionMessage() string {
	return "Not yet released 2026 changes"
}

func calculateStarRating(basePerformance float64) float64 {
	return math.Cbrt(basePerformance * PerformanceBaseMultiplier)
}
