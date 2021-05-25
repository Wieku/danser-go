package xexxar

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/xexxar/preprocessing"
	"github.com/wieku/danser-go/app/xexxar/skills"
	"log"
	"math"
)

const (
	// Global stars multiplier
	StarScalingFactor float64 = 0.18

	// 50% of the difference between aim and speed is added to total
	// star rating to compensate for aim/speed only maps
	ExtremeScalingFactor float64 = 0.5

	DifficultyPower float64 = 0.75
)

type Stars struct {
	// Star rating, visible on osu!'s beatmap page
	Total float64

	// Aim stars, needed for Performance Points (aka PP) calculations
	Aim float64

	// Speed stars, needed for Performance Points (aka PP) calculations
	Speed float64
}

// Retrieves skills values and converts to Stars
func getStars(aim, speed *skills.Skill) Stars {
	aimVal := math.Pow(aim.DifficultyValue(), DifficultyPower) * StarScalingFactor
	speedVal := math.Pow(speed.DifficultyValue(), DifficultyPower) * StarScalingFactor

	// total stars
	total := aimVal + speedVal + math.Abs(speedVal-aimVal)*ExtremeScalingFactor

	return Stars{
		Total: total,
		Aim:   aimVal,
		Speed: speedVal,
	}
}

// Calculate final star rating of a map
func CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty) Stars {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff)
	speedSkill := skills.NewSpeedSkill(diff)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
	}

	return getStars(aimSkill, speedSkill)
}

// Calculate successive star ratings for every part of a beatmap
func CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty) []Stars {
	modString := (diff.Mods & difficulty.DifficultyAdjustMask).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff)
	speedSkill := skills.NewSpeedSkill(diff)

	stars := make([]Stars, 1, len(objects))

	lastProgress := -1

	for i, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)

		stars = append(stars, getStars(aimSkill, speedSkill))

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
