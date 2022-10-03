package pp220930

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930/skills"
	"log"
	"math"
)

const (
	// StarScalingFactor is a global stars multiplier
	StarScalingFactor float64 = 0.0675
	CurrentVersion    int     = 20220930
)

type Attributes struct {
	// Total Star rating, visible on osu!'s beatmap page
	Total float64

	// Aim stars, needed for Performance Points (aka PP) calculations
	Aim float64

	// Speed stars, needed for Performance Points (aka PP) calculations
	Speed float64

	SpeedNoteCount float64

	// Flashlight stars, needed for Performance Points (aka PP) calculations
	Flashlight float64

	// SliderFactor is a ratio of Aim calculated without sliders to Aim with them
	SliderFactor float64

	ObjectCount int
	Circles     int
	Sliders     int
	Spinners    int
	MaxCombo    int
}

// StrainPeaks contains peaks of Aim, Speed and Flashlight skills, as well as peaks passed through star rating formula
type StrainPeaks struct {
	// Aim peaks
	Aim []float64

	// Speed peaks
	Speed []float64

	// Flashlight peaks
	Flashlight []float64

	// Total contains aim, speed and flashlight peaks passed through star rating formula
	Total []float64
}

// getStarsFromRawValues converts raw skill values to Attributes
func getStarsFromRawValues(rawAim, rawAimNoSliders, rawSpeed, rawFlashlight float64, diff *difficulty.Difficulty, attr Attributes) Attributes {
	aimRating := math.Sqrt(rawAim) * StarScalingFactor
	aimRatingNoSliders := math.Sqrt(rawAimNoSliders) * StarScalingFactor
	speedRating := math.Sqrt(rawSpeed) * StarScalingFactor
	flashlightRating := math.Sqrt(rawFlashlight) * StarScalingFactor

	sliderFactor := 1.0
	if aimRating > 0.00001 {
		sliderFactor = aimRatingNoSliders / aimRating
	}

	if diff.CheckModActive(difficulty.TouchDevice) {
		aimRating = math.Pow(aimRating, 0.8)
		flashlightRating = math.Pow(flashlightRating, 0.8)
	}

	if diff.CheckModActive(difficulty.Relax) {
		aimRating *= 0.9
		speedRating = 0
		flashlightRating *= 0.7
	}

	var total float64

	baseAimPerformance := ppBase(aimRating)
	baseSpeedPerformance := ppBase(speedRating)
	baseFlashlightPerformance := 0.0

	if diff.CheckModActive(difficulty.Flashlight) {
		baseFlashlightPerformance = math.Pow(flashlightRating, 2.0) * 25.0
	}

	basePerformance := math.Pow(
		math.Pow(baseAimPerformance, 1.1)+
			math.Pow(baseSpeedPerformance, 1.1)+
			math.Pow(baseFlashlightPerformance, 1.1),
		1.0/1.1,
	)

	if basePerformance > 0.00001 {
		total = math.Cbrt(PerformanceBaseMultiplier) * 0.027 * (math.Cbrt(100000/math.Pow(2, 1/1.1)*basePerformance) + 4)
	}

	attr.Total = total
	attr.Aim = aimRating
	attr.SliderFactor = sliderFactor
	attr.Speed = speedRating
	attr.Flashlight = flashlightRating

	return attr
}

// Retrieves skill values and converts to Attributes
func getStars(aim *skills.AimSkill, aimNoSliders *skills.AimSkill, speed *skills.SpeedSkill, flashlight *skills.Flashlight, diff *difficulty.Difficulty, attr Attributes) Attributes {
	attr = getStarsFromRawValues(
		aim.DifficultyValue(),
		aimNoSliders.DifficultyValue(),
		speed.DifficultyValue(),
		flashlight.DifficultyValue(),
		diff,
		attr,
	)

	attr.SpeedNoteCount = speed.RelevantNoteCount()

	return attr
}

func addObjectToAttribs(o objects.IHitObject, attr *Attributes) {
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
func CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty) Attributes {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff, true)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false)
	speedSkill := skills.NewSpeedSkill(diff)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	attr := Attributes{}

	addObjectToAttribs(objects[0], &attr)

	for i, o := range diffObjects {
		addObjectToAttribs(objects[i+1], &attr)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	return getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, attr)
}

// CalculateStep calculates successive star ratings for every part of a beatmap
func CalculateStep(objects []objects.IHitObject, diff *difficulty.Difficulty) []Attributes {
	modString := difficulty.GetDiffMaskedMods(diff.Mods).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff, true)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false)
	speedSkill := skills.NewSpeedSkill(diff)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	stars := make([]Attributes, 1, len(objects))

	addObjectToAttribs(objects[0], &stars[0])

	lastProgress := -1

	for i, o := range diffObjects {
		attr := stars[i]
		addObjectToAttribs(objects[i+1], &attr)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)

		stars = append(stars, getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, attr))

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

func CalculateStrainPeaks(objects []objects.IHitObject, diff *difficulty.Difficulty) StrainPeaks {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff, true)
	speedSkill := skills.NewSpeedSkill(diff)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	for _, o := range diffObjects {
		aimSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)
	}

	peaks := StrainPeaks{
		Aim:        aimSkill.GetCurrentStrainPeaks(),
		Speed:      speedSkill.GetCurrentStrainPeaks(),
		Flashlight: flashlightSkill.GetCurrentStrainPeaks(),
	}

	peaks.Total = make([]float64, len(peaks.Aim))

	for i := 0; i < len(peaks.Aim); i++ {
		stars := getStarsFromRawValues(peaks.Aim[i], peaks.Aim[i], peaks.Speed[i], peaks.Flashlight[i], diff, Attributes{})
		peaks.Total[i] = stars.Total
	}

	return peaks
}
