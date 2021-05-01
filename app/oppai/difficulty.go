package oppai

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
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
}

// Retrieves skills values and converts to Stars
func getStars(aim, speed *Skill, diff *difficulty.Difficulty) Stars {
	aimVal := math.Sqrt(aim.DifficultyValue()) * StarScalingFactor
	speedVal := math.Sqrt(speed.DifficultyValue()) * StarScalingFactor

	if diff.Mods.Active(difficulty.TouchDevice) {
		aimVal = math.Pow(aimVal, 0.8)
	}

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
	diffObjects := createObjects(objects, diff)

	aimSkill := NewAimSkill(false, diff)
	speedSkill := NewSpeedSkill(false, diff)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
	}

	return getStars(aimSkill, speedSkill, diff)
}

// Calculate successive star ratings for every part of a beatmap
func CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty) []Stars {
	modString := (diff.Mods & difficulty.DifficultyAdjustMask).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	diffObjects := createObjects(objects, diff)

	aimSkill := NewAimSkill(false, diff)
	speedSkill := NewSpeedSkill(false, diff)

	stars := make([]Stars, 1, len(objects))

	lastProgress := -1

	for i, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)

		stars = append(stars, getStars(aimSkill, speedSkill, diff))

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

// Creates difficulty objects needed for star rating calculations
func createObjects(objsB []objects.IHitObject, d *difficulty.Difficulty) []*DifficultyObject {
	objs := make([]objects.IHitObject, 0, len(objsB))

	for _, o := range objsB {
		if s, ok := o.(*objects.Slider); ok {
			o = NewLazySlider(s, d)
		}

		objs = append(objs, o)
	}

	diffObjects := make([]*DifficultyObject, 0, len(objsB))

	for i := 1; i < len(objs); i++ {
		var lastLast, last, current objects.IHitObject

		if i > 1 {
			lastLast = objs[i-2]
		}

		last = objs[i-1]
		current = objs[i]

		diffObjects = append(diffObjects, NewDifficultyObject(current, lastLast, last, d))
	}

	return diffObjects
}
