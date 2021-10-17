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

	// 50% of the difference between aim and speed is added to total
	// star rating to compensate for aim/speed only maps
	ExtremeScalingFactor float64 = 0.5
)

type Stars struct {
	// Star rating, visible on osu!'s beatmap page
	Total float64

	// Aim stars, needed for Performance Points (aka PP) calculations
	Aim float64

	// Speed stars, needed for Performance Points (aka PP) calculations
	Speed float64

	// Flashlight stars, needed for Performance Points (aka PP) calculations
	Flashlight float64
}

// Retrieves skills values and converts to Stars
func getStars(aim *skills.AimSkill, speed *skills.SpeedSkill, flashlight *skills.Flashlight, diff *difficulty.Difficulty, experimental bool) Stars {
	aimRating := math.Sqrt(aim.DifficultyValue()) * StarScalingFactor
	speedRating := math.Sqrt(speed.DifficultyValue()) * StarScalingFactor
	flashlightVal := math.Sqrt(flashlight.DifficultyValue()) * StarScalingFactor

	var total float64

	if experimental { // https://github.com/ppy/osu/pull/13986
		if diff.CheckModActive(difficulty.Flashlight) {
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
	} else { // Live as of 2021-07-27
		total = aimRating + speedRating + math.Abs(speedRating-aimRating)*ExtremeScalingFactor
	}

	return Stars{
		Total:      total,
		Aim:        aimRating,
		Speed:      speedRating,
		Flashlight: flashlightVal,
	}
}

// Calculate final star rating of a map
func CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty, experimental bool) Stars {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, experimental)

	aimSkill := skills.NewAimSkill(diff)
	speedSkill := skills.NewSpeedSkill(diff, experimental)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	return getStars(aimSkill, speedSkill, flashlightSkill, diff, experimental)
}

// Calculate successive star ratings for every part of a beatmap
func CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty, experimental bool) []Stars {
	modString := (diff.Mods & difficulty.DifficultyAdjustMask).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, experimental)

	aimSkill := skills.NewAimSkill(diff)
	speedSkill := skills.NewSpeedSkill(diff, experimental)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	stars := make([]Stars, 1, len(objects))

	lastProgress := -1

	for i, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)

		stars = append(stars, getStars(aimSkill, speedSkill, flashlightSkill, diff, experimental))

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
