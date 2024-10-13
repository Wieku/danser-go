package pp211112

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp211112/preprocessing"
	skills2 "github.com/wieku/danser-go/app/rulesets/osu/performance/pp211112/skills"
	"log"
	"math"
)

const (
	// StarScalingFactor is a global stars multiplier
	StarScalingFactor float64 = 0.0675
	CurrentVersion    int     = 20211112
)

type DifficultyCalculator struct{}

func NewDifficultyCalculator() api.IDifficultyCalculator {
	return &DifficultyCalculator{}
}

// getStarsFromRawValues converts raw skill values to Attributes
func (diffCalc *DifficultyCalculator) getStarsFromRawValues(rawAim, rawAimNoSliders, rawSpeed, rawFlashlight float64, diff *difficulty.Difficulty, attr api.Attributes) api.Attributes {
	aimRating := math.Sqrt(rawAim) * StarScalingFactor
	aimRatingNoSliders := math.Sqrt(rawAimNoSliders) * StarScalingFactor
	speedRating := math.Sqrt(rawSpeed) * StarScalingFactor
	flashlightVal := math.Sqrt(rawFlashlight) * StarScalingFactor

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

	attr.Total = total
	attr.Aim = aimRating
	attr.SliderFactor = sliderFactor
	attr.Speed = speedRating
	attr.Flashlight = flashlightVal

	return attr
}

// Retrieves skill values and converts to Attributes
func (diffCalc *DifficultyCalculator) getStars(aim *skills2.AimSkill, aimNoSliders *skills2.AimSkill, speed *skills2.SpeedSkill, flashlight *skills2.Flashlight, diff *difficulty.Difficulty, attr api.Attributes) api.Attributes {
	attr = diffCalc.getStarsFromRawValues(
		aim.DifficultyValue(),
		aimNoSliders.DifficultyValue(),
		speed.DifficultyValue(),
		flashlight.DifficultyValue(),
		diff,
		attr,
	)

	attr.AimDifficultStrainCount = aim.CountDifficultStrains(diff.Speed)
	attr.SpeedDifficultStrainCount = speed.CountDifficultStrains(diff.Speed)

	return attr
}

func (diffCalc *DifficultyCalculator) addObjectToAttribs(o objects.IHitObject, attr *api.Attributes) {
	if s, ok := o.(*objects.Slider); ok {
		attr.Sliders++
		attr.MaxCombo += len(s.ScorePoints)
	} else if _, ok := o.(*objects.Circle); ok {
		attr.Circles++
	} else if _, ok := o.(*objects.Spinner); ok {
		attr.Spinners++
	}

	attr.MaxCombo++
	attr.ObjectCount++
}

// CalculateSingle calculates the final difficulty attributes of a map
func (diffCalc *DifficultyCalculator) CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty) api.Attributes {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, false)

	aimSkill := skills2.NewAimSkill(diff, true, false)
	aimNoSlidersSkill := skills2.NewAimSkill(diff, false, false)
	speedSkill := skills2.NewSpeedSkill(diff, false)
	flashlightSkill := skills2.NewFlashlightSkill(diff, false)

	attr := api.Attributes{}

	diffCalc.addObjectToAttribs(objects[0], &attr)

	for i, o := range diffObjects {
		diffCalc.addObjectToAttribs(objects[i+1], &attr)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	return diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, attr)
}

// CalculateStep calculates successive star ratings for every part of a beatmap
func (diffCalc *DifficultyCalculator) CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty) []api.Attributes {
	modString := (diff.Mods & difficulty.DifficultyAdjustMask).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, false)

	aimSkill := skills2.NewAimSkill(diff, true, false)
	aimNoSlidersSkill := skills2.NewAimSkill(diff, false, false)
	speedSkill := skills2.NewSpeedSkill(diff, false)
	flashlightSkill := skills2.NewFlashlightSkill(diff, false)

	stars := make([]api.Attributes, 1, len(objects))

	diffCalc.addObjectToAttribs(objects[0], &stars[0])

	lastProgress := -1

	for i, o := range diffObjects {
		attr := stars[i]
		diffCalc.addObjectToAttribs(objects[i+1], &attr)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)

		stars = append(stars, diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, attr))

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

func (diffCalc *DifficultyCalculator) CalculateStrainPeaks(objects []objects.IHitObject, diff *difficulty.Difficulty) api.StrainPeaks {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff, false)

	aimSkill := skills2.NewAimSkill(diff, true, false)
	speedSkill := skills2.NewSpeedSkill(diff, false)
	flashlightSkill := skills2.NewFlashlightSkill(diff, false)

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

	for i := 0; i < len(peaks.Aim); i++ {
		stars := diffCalc.getStarsFromRawValues(peaks.Aim[i], peaks.Aim[i], peaks.Speed[i], peaks.Flashlight[i], diff, api.Attributes{})
		peaks.Total[i] = stars.Total
	}

	return peaks
}

func (diffCalc *DifficultyCalculator) GetVersion() int {
	return CurrentVersion
}

func (diffCalc *DifficultyCalculator) GetVersionMessage() string {
	return "2021-11-12: https://osu.ppy.sh/home/news/2021-11-09-performance-points-star-rating-updates"
}
