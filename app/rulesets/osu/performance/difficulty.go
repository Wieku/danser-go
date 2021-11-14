package performance

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/skills"
	"log"
	"math"
)

const (
	// Global stars multiplier
	StarScalingFactor float64 = 0.0675
)

type Stars struct {
	// Star rating, visible on osu!'s beatmap page
	Total float64

	// Aim stars, needed for Performance Points (aka PP) calculations
	Aim float64

	// SliderFactor is the ratio of Aim calculated without sliders to Aim
	SliderFactor float64

	// Speed stars, needed for Performance Points (aka PP) calculations
	Speed float64

	// Flashlight stars, needed for Performance Points (aka PP) calculations
	Flashlight float64
}

type StrainPeaks struct {
	// Aim peaks
	Aim []float64

	// Speed peaks
	Speed []float64

	// Flashlight peaks
	Flashlight []float64
}

// Retrieves skills values and converts to Stars
func getStars(aim *skills.AimSkill, aimNoSliders *skills.AimSkill, speed *skills.SpeedSkill, flashlight *skills.Flashlight, diff *difficulty.Difficulty, experimental bool) Stars {
	aimRating := math.Sqrt(aim.DifficultyValue()) * StarScalingFactor
	aimRatingNoSliders := math.Sqrt(aimNoSliders.DifficultyValue()) * StarScalingFactor
	speedRating := math.Sqrt(speed.DifficultyValue()) * StarScalingFactor
	flashlightVal := math.Sqrt(flashlight.DifficultyValue()) * StarScalingFactor

	sliderFactor := 1.0
	if aimRating > 0.00001 {
		sliderFactor = aimRatingNoSliders / aimRating
	}

	var total float64

	if diff.CheckModActive(difficulty.Relax) {
		speedRating = 0.0
	}

	baseAimPerformance := ppBase(aimRating)
	baseSpeedPerformance := ppBase(speedRating)
	baseFlashlightPerformance := 0.0

	if diff.CheckModActive(difficulty.Flashlight) {
		baseFlashlightPerformance = math.Pow(flashlightVal, 2.0) * 25.0
	}

	basePerformance := math.Pow(
		math.Pow(baseAimPerformance, 1.1)+
			math.Pow(baseSpeedPerformance, 1.1)+
			math.Pow(baseFlashlightPerformance, 1.1),
		1.0/1.1,
	)

	if basePerformance > 0.00001 {
		total = math.Cbrt(1.12) * 0.027 * (math.Cbrt(100000/math.Pow(2, 1/1.1)*basePerformance) + 4)
	}

	return Stars{
		Total:        total,
		Aim:          aimRating,
		SliderFactor: sliderFactor,
		Speed:        speedRating,
		Flashlight:   flashlightVal,
	}
}

// CalculateSingle calculates the final star rating of a map
func CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty, experimental bool) Stars {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, experimental)

	aimSkill := skills.NewAimSkill(diff, true, experimental)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, experimental)
	speedSkill := skills.NewSpeedSkill(diff, experimental)
	flashlightSkill := skills.NewFlashlightSkill(diff, experimental)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	return getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, experimental)
}

func CalculateStrainPeaks(objects []objects.IHitObject, diff *difficulty.Difficulty, experimental bool) StrainPeaks {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, experimental)

	aimSkill := skills.NewAimSkill(diff, true, experimental)
	speedSkill := skills.NewSpeedSkill(diff, experimental)
	flashlightSkill := skills.NewFlashlightSkill(diff, experimental)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	return StrainPeaks{
		Aim:        aimSkill.GetCurrentStrainPeaks(),
		Speed:      speedSkill.GetCurrentStrainPeaks(),
		Flashlight: flashlightSkill.GetCurrentStrainPeaks(),
	}
}

// CalculateStep calculates successive star ratings for every part of a beatmap
func CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty, experimental bool) []Stars {
	modString := (diff.Mods & difficulty.DifficultyAdjustMask).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, experimental)

	aimSkill := skills.NewAimSkill(diff, true, experimental)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, experimental)
	speedSkill := skills.NewSpeedSkill(diff, experimental)
	flashlightSkill := skills.NewFlashlightSkill(diff, experimental)

	stars := make([]Stars, 1, len(objects))

	lastProgress := -1

	for i, o := range diffObjects {
		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)

		stars = append(stars, getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, experimental))

		if len(diffObjects) > 2500 {
			progress := (100 * i) / (len(diffObjects) - 1)

			if progress != lastProgress && progress%5 == 0 {
				log.Println(fmt.Sprintf("Progress: %d%%", progress))
			}

			lastProgress = progress
		}
	}

	log.Println("Calculations finished!")

	return stars
}
