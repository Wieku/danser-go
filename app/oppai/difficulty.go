package oppai

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
)

/* ------------------------------------------------------------- */
/* difficulty calculator                                         */

// constants for difficulty calculator
const (
	/** almost the normalized circle diameter. */
	AlmostDiameter float64 = 90.0

	/**
	 * arbitrary thresholds to determine when a stream is spaced
	 * enough that it becomes hard to alternate.
	 */
	SingleSpacing float64 = 125.0

	/**
	 * max strains are weighted from highest to lowest, this is how
	 * much the weight decays.
	 */
	DecayWeight float64 = 0.9

	/**
	 * strains are calculated by analyzing the map in chunks and taking
	 * the peak strains in each chunk. this is the length of a strain
	 * interval in milliseconds
	 */
	StrainStep float64 = 400.0

	/** non-normalized diameter where the small circle buff starts. */
	CirclesizeBuffThreshold float64 = 30.0

	/** global stars multiplier. */
	StarScalingFactor float64 = 0.0675

	/** in osu! pixels */
	PlayfieldWidth  float64 = 512.0
	PlayfieldHeight float64 = 384.0

	/**
	 * 50% of the difference between aim and speed is added to total
	 * star rating to compensate for aim/speed only maps
	 */
	ExtremeScalingFactor float64 = 0.5

	MinSpeedBonus        float64 = 75.0
	MaxSpeedBonus        float64 = 45.0
	AngleBonusScale      float64 = 90.0
	AimTimingThreshold   float64 = 107.0
	SpeedAngleBonusBegin float64 = 5 * math.Pi / 6
	AimAngleBonusBegin   float64 = math.Pi / 3

	// DiffSpeed : strain index for speed
	DiffSpeed = 0

	// DiffAim : strain index for aim
	DiffAim = 1
)

// DecayBase : strain decay per interval.
var DecayBase = []float64{0.3, 0.15}

// WeightScaling : balances speed and aim.
var WeightScaling = []float64{1400.0, 26.25}

// PlayfieldCenter ...
var PlayfieldCenter = vector.NewVec2d(PlayfieldWidth/2.0, PlayfieldHeight/2.0)

func dSpacingWeight(Type int, distance float64, delta_time float64,
	prev_distance float64, prev_delta_time float64, angle float64) float64 {
	strain_time := math.Max(delta_time, 50.0)
	prev_strain_time := math.Max(prev_delta_time, 50.0)
	var angle_bonus float64
	switch Type {
	case DiffAim:
		result := 0.0
		if !math.IsNaN(angle) && angle > AimAngleBonusBegin {
			angle_bonus = math.Sqrt(
				math.Max(prev_distance-AngleBonusScale, 0.0) *
					math.Pow(math.Sin(angle-AimAngleBonusBegin), 2.0) *
					math.Max(distance-AngleBonusScale, 0.0))
			result = 1.5 * math.Pow(math.Max(0.0, angle_bonus), 0.99) /
				math.Max(AimTimingThreshold, prev_strain_time)
		}
		weighted_distance := math.Pow(distance, 0.99)
		return math.Max(result+
			weighted_distance/
				math.Max(AimTimingThreshold, strain_time),
			weighted_distance/strain_time)
	case DiffSpeed:
		distance := math.Min(distance, SingleSpacing)
		delta_time := math.Max(delta_time, MaxSpeedBonus)
		speed_bonus := 1.0
		if delta_time < MinSpeedBonus {
			speed_bonus += math.Pow((MinSpeedBonus-delta_time)/40.0, 2.0)
		}
		angle_bonus := 1.0
		if !math.IsNaN(angle) && angle < SpeedAngleBonusBegin {
			s := math.Sin(1.5 * (SpeedAngleBonusBegin - angle))
			angle_bonus += math.Pow(s, 2) / 3.57
			if angle < math.Pi/2.0 {
				angle_bonus = 1.28
				if distance < AngleBonusScale && angle < math.Pi/4.0 {
					angle_bonus += (1.0 - angle_bonus) *
						math.Min((AngleBonusScale-distance)/10.0, 1.0)
				} else if distance < AngleBonusScale {
					angle_bonus += (1.0 - angle_bonus) *
						math.Min((AngleBonusScale-distance)/10.0, 1.0) *
						math.Sin((math.Pi/2.0-angle)*4.0/math.Pi)
				}
			}
		}
		return ((1.0 + (speed_bonus-1.0)*0.75) * angle_bonus *
			(0.95 + speed_bonus*math.Pow(distance/SingleSpacing, 3.5))) /
			strain_time
	}
	panic("this diff type does not exist")
}

/**
* calculates the strain for one difficulty type and stores it in
* obj. this assumes that normpos is already computed.
* this also sets is_single if type is DIFF_SPEED
 */
func dStrain(Type int, obj *DiffObject, prev DiffObject, speedMul float64) {
	var value float64
	timeElapsed := float64(obj.Data.GetStartTime()-prev.Data.GetStartTime()) / speedMul
	var decay = math.Pow(DecayBase[Type], timeElapsed/1000.0)

	obj.DeltaTime = timeElapsed

	if (obj.Data.GetType() & (objects.SLIDER | objects.CIRCLE)) != 0 {
		var distance = obj.Normpos.Sub(prev.Normpos).Len()
		obj.DDistance = distance

		//if Type == DiffSpeed {
		//	obj.IsSingle = distance > SingleSpacing
		//}

		value = dSpacingWeight(Type, obj.DDistance, timeElapsed,
			prev.DDistance, prev.DeltaTime, obj.Angle)
		value *= WeightScaling[Type]
	}
	obj.Strains[Type] = prev.Strains[Type]*decay + value
}

// DiffCalc difficulty calculator
type DiffCalc struct {
	Total float64 // star rating
	Aim   float64 // aim stars

	Speed float64 // speed stars

	speedMul   float64
	Objects    []*DiffObject
	NumObjects int
	strains    []float64

	diff *difficulty.Difficulty
	//mapStats          *MapStats
}

type Stars struct {
	Total float64 // star rating
	Aim   float64 // aim stars
	Speed float64 // speed stars
}

func lengthBonus(stars float64, difficulty float64) float64 {
	return 0.32 + 0.5*
		(math.Log10(difficulty+stars)-math.Log10(stars))
}

type diffValues struct {
	Difficulty float64
	Total      float64
}

func (d *DiffCalc) calcIndividual(Type int) diffValues {
	d.strains = make([]float64, 0, 256)

	var strainStep = StrainStep * d.speedMul
	var intervalEnd = math.Ceil(float64(d.Objects[0].Data.GetStartTime())/strainStep) * strainStep
	var maxStrain float64

	// calculate all strains
	for i := 0; i < d.NumObjects; i++ {
		var obj = d.Objects[i]

		var prev *DiffObject
		if i > 0 {
			prev = d.Objects[i-1]
		}

		if prev != nil {
			dStrain(Type, obj, *prev, d.speedMul)
		}

		for float64(obj.Data.GetStartTime()) > intervalEnd {
			/* add max strain for this interval */
			d.strains = append(d.strains, maxStrain)

			if prev != nil {
				/* decay last object's strains until the next
				   interval and use that as the initial max
				   strain */

				var decay = math.Pow(DecayBase[Type],
					(intervalEnd-float64(prev.Data.GetStartTime()))/1000.0)
				maxStrain = prev.Strains[Type] * decay
			} else {
				maxStrain = 0.0
			}

			intervalEnd += strainStep
		}

		maxStrain = math.Max(maxStrain, obj.Strains[Type])
	}

	d.strains = append(d.strains, maxStrain)

	/* weight the top strains sorted from highest to lowest */
	weight := 1.0
	var total float64
	var difficulty float64

	reverseSortFloat64s(d.strains)

	for _, strain := range d.strains {
		total += math.Pow(strain, 1.2)
		difficulty += strain * weight
		weight *= DecayWeight
	}

	return diffValues{difficulty, total}
}

// DefaultSingletapThreshold default value for singletap_threshold.
const DefaultSingletapThreshold float64 = 125.0

// Calc calculates beatmap difficulty and stores it in total, aim,
/* speed, nsingles, nsingles_speed fields.
 * @param singletap_threshold the smallest milliseconds interval
 *        that will be considered singletappable. for example,
 *        125ms is 240 1/2 singletaps ((60000 / 240) / 2)
 * @return self
 */
func (d *DiffCalc) Calc() Stars {
	radius := d.diff.CircleRadius

	/* positions are normalized on circle radius so that we can
	calc as if everything was the same circlesize */
	scalingFactor := 52.0 / radius

	if radius < CirclesizeBuffThreshold {
		scalingFactor *= 1.0 +
			math.Min(CirclesizeBuffThreshold-radius, 5.0)/50.0
	}

	normalizedCenter := PlayfieldCenter.Scl(scalingFactor)

	var prev1 *DiffObject
	var prev2 *DiffObject

	/* calculate normalized positions */
	for i := 0; i < d.NumObjects; i++ {
		obj := d.Objects[i]
		if (obj.Data.GetType() & objects.SPINNER) != 0 {
			obj.Normpos = normalizedCenter

		} else {
			pos := obj.Data.GetStackedStartPositionMod(d.diff.Mods).Copy64()

			obj.Normpos = pos.Scl(scalingFactor)

			if i >= 2 && prev2 != nil && prev1 != nil {
				v1 := prev2.Normpos.Sub(prev1.Normpos)
				v2 := obj.Normpos.Sub(prev1.Normpos)
				dot := v1.Dot(v2)
				det := v1.X*v2.Y - v1.Y*v2.X
				obj.Angle = math.Abs(math.Atan2(float64(det), float64(dot)))
			} else {
				obj.Angle = math.NaN()
			}

			prev2 = prev1
			prev1 = obj
		}
	}

	/* speed and aim stars */
	aimvals := d.calcIndividual(DiffAim)
	d.Aim = aimvals.Difficulty

	speedvals := d.calcIndividual(DiffSpeed)
	d.Speed = speedvals.Difficulty

	d.Aim = math.Sqrt(d.Aim) * StarScalingFactor
	d.Speed = math.Sqrt(d.Speed) * StarScalingFactor

	if d.diff.Mods.Active(difficulty.TouchDevice) {
		d.Aim = math.Pow(d.Aim, 0.8)
	}

	/* total stars */
	d.Total = d.Aim + d.Speed +
		math.Abs(d.Speed-d.Aim)*ExtremeScalingFactor

	return Stars{
		Total: d.Total,
		Aim:   d.Aim,
		Speed: d.Speed,
	}
}

func CalcSingle(objects []objects.IHitObject, difficulty *difficulty.Difficulty) Stars {
	return newDiffCalc(objects, difficulty).Calc()
}

func CalcStep(objects []objects.IHitObject, diff *difficulty.Difficulty) []Stars {
	modString := (diff.Mods&difficulty.DifficultyAdjustMask).String()
	if modString == "" {
		modString = "NM"
	}

	log.Println("Calculating step SR for mods:", modString)

	d := newDiffCalc(objects, diff)
	stars := make([]Stars, 0, len(objects))
	sum := len(objects) * (len(objects) + 1) / 2
	lastProgress := -1

	for i := 1; i <= len(objects); i++ {
		d.NumObjects = i
		stars = append(stars, d.Calc())

		if len(objects) > 2500 {
			progress := (50 * i * (i + 1)) / sum

			if progress != lastProgress && progress%5 == 0 {
				log.Println(fmt.Sprintf("Progress: %d%%", progress))
			}

			lastProgress = progress
		}
	}

	log.Println("Calculations finished!")

	return stars
}

func newDiffCalc(objects []objects.IHitObject, d *difficulty.Difficulty) *DiffCalc {
	diffObjects := make([]*DiffObject, 0, len(objects))

	for _, o := range objects {
		diffObjects = append(diffObjects, &DiffObject{
			Data:    o,
			Strains: make([]float64, 2),
		})
	}

	return &DiffCalc{
		speedMul:   d.Speed,
		Objects:    diffObjects,
		NumObjects: len(diffObjects),
		diff:       d,
	}
}
