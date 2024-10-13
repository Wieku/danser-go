package pp241007

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007/skills"
	"log"
	"math"
	"time"
)

const (
	// StarScalingFactor is a global stars multiplier
	StarScalingFactor float64 = 0.0675
	CurrentVersion    int     = 20241007
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

	baseAimPerformance := skills.DefaultDifficultyToPerformance(aimRating)
	baseSpeedPerformance := skills.DefaultDifficultyToPerformance(speedRating)
	baseFlashlightPerformance := 0.0

	if diff.CheckModActive(difficulty.Flashlight) {
		baseFlashlightPerformance = skills.FlashlightDifficultyToPerformance(flashlightRating)
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
func (diffCalc *DifficultyCalculator) getStars(aim *skills.AimSkill, aimNoSliders *skills.AimSkill, speed *skills.SpeedSkill, flashlight *skills.Flashlight, diff *difficulty.Difficulty, attr api.Attributes) api.Attributes {
	attr = diffCalc.getStarsFromRawValues(
		aim.DifficultyValue(),
		aimNoSliders.DifficultyValue(),
		speed.DifficultyValue(),
		flashlight.DifficultyValue(),
		diff,
		attr,
	)

	attr.SpeedNoteCount = speed.RelevantNoteCount()
	attr.AimDifficultStrainCount = aim.CountDifficultStrains()
	attr.SpeedDifficultStrainCount = speed.CountDifficultStrains()

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

// CalculateSingle calculates the final difficultyapi.Attributes of a map
func (diffCalc *DifficultyCalculator) CalculateSingle(objects []objects.IHitObject, diff *difficulty.Difficulty) api.Attributes {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff, true, false)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, false)
	speedSkill := skills.NewSpeedSkill(diff, false)
	flashlightSkill := skills.NewFlashlightSkill(diff)

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
	modString := difficulty.GetDiffMaskedMods(diff.Mods).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	startTime := time.Now()

	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

	aimSkill := skills.NewAimSkill(diff, true, true)
	aimNoSlidersSkill := skills.NewAimSkill(diff, false, false)
	speedSkill := skills.NewSpeedSkill(diff, true)
	flashlightSkill := skills.NewFlashlightSkill(diff)

	stars := make([]api.Attributes, 1, len(objects))

	diffCalc.addObjectToAttribs(objects[0], &stars[0])

	for i, o := range diffObjects {
		attr := stars[i]
		diffCalc.addObjectToAttribs(objects[i+1], &attr)

		aimSkill.Process(o)
		aimNoSlidersSkill.Process(o)
		speedSkill.Process(o)
		flashlightSkill.Process(o)

		stars = append(stars, diffCalc.getStars(aimSkill, aimNoSlidersSkill, speedSkill, flashlightSkill, diff, attr))
	}

	endTime := time.Now()

	log.Println("Calculations finished! Took ", endTime.Sub(startTime).Truncate(time.Millisecond).String())

	return stars
}

func (diffCalc *DifficultyCalculator) CalculateStrainPeaks(objects []objects.IHitObject, diff *difficulty.Difficulty) api.StrainPeaks {
	diffObjects := preprocessing.CreateDifficultyObjects(objects, diff)

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
	return "2024-10-07: no post yet"
}
