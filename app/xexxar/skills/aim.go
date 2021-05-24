package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/xexxar/preprocessing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	AngleBonusScale    float64 = 90.0
	AimTimingThreshold float64 = 107.0
	AimAngleBonusBegin float64 = math.Pi / 3

	decayExcessThreshold   float64 = 500
	snapStrainMultiplier   float64 = 10
	flowStrainMultiplier   float64 = 16.25
	hybridStrainMultiplier float64 = 8.25
	sliderStrainMultiplier float64 = 75
	totalStrainMultiplier  float64 = 0.1675
)

func NewAimSkill(d *difficulty.Difficulty) *Skill {
	skill := NewSkill(d)
	skill.StarsPerDouble = 1.1
	skill.HistoryLength = 2
	skill.currentStrain = 1
	skill.StrainValueOf = aimStrainValue

	return skill
}

func computeDecay(baseDecay, ms float64) float64 {
	decay := 0.0

	if ms < decayExcessThreshold {
		decay = baseDecay
	} else {
		decay = math.Pow(math.Pow(baseDecay, 1000.0 / math.Min(ms, decayExcessThreshold)), ms / 1000)
	}

	return decay
}

func snapScaling(distance float64) float64 {
	if distance == 0.0 {
		return 0
	}

	return (3.75 * (math.Log(distance / 3.75 + 1) / math.Log(2))) / distance
}

func flowStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject, prevVector, currVector, nextVector vector.Vector2f) float64 {
	//nextDiffVector := currVector.Sub(nextVector)
	prevDiffVector := prevVector.Sub(currVector)

	minDistance := 0.0

	angle := (osuCurrObj.Angle + osuNextObj.Angle) / 2

	minBoyDistance := math.Min(float64(currVector.Len()), float64(prevVector.Len()))

	if angle < math.Pi / 4 {
		minDistance = 0
	} else if angle > math.Pi / 2 {
		minDistance = math.Max(0, math.Min(100 / osuCurrObj.StrainTime, float64(prevDiffVector.Len()) - math.Max(float64(currVector.Len()), float64(prevVector.Len()))/2))
	} else {
		minDistance = math.Min(1, math.Max(0, osuCurrObj.JumpDistance - 75) / 50) *
			math.Pow(math.Sin(2 * (angle - math.Pi / 4)), 2) *
			math.Max(0, math.Min(math.Min(100 / osuCurrObj.StrainTime, minBoyDistance), float64(prevDiffVector.Len()) - math.Max(float64(currVector.Len()), float64(prevVector.Len()))/2))
	}

	strain := float64(prevVector.Len()) * osuPrevObj.FlowProbability +
		float64(currVector.Len()) * osuCurrObj.FlowProbability +
		math.Min(math.Min(float64(currVector.Len()), float64(prevVector.Len())), math.Abs(float64(currVector.Len()) - float64(prevVector.Len()))) * osuCurrObj.FlowProbability * osuPrevObj.FlowProbability +
		minDistance * osuCurrObj.FlowProbability * osuPrevObj.FlowProbability

	strain *= math.Min(osuCurrObj.StrainTime / (osuCurrObj.StrainTime - 10), osuPrevObj.StrainTime / (osuPrevObj.StrainTime - 10))

	return strain
}

func snapStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject, prevVector, currVector, nextVector vector.Vector2f) float64 {
	currVector = osuCurrObj.DistanceVector.Scl(float32(snapScaling(osuCurrObj.JumpDistance / 104))).Scl(1/float32(osuCurrObj.StrainTime))
	prevVector = osuPrevObj.DistanceVector.Scl(float32(snapScaling(osuPrevObj.JumpDistance / 104))).Scl(1/float32(osuPrevObj.StrainTime))

	//nextDiffVector := currVector.Add(nextVector)
	prevDiffVector := prevVector.Add(currVector)

	angleDistance := math.Max(0, float64(prevDiffVector.Len()) - math.Max(float64(currVector.Len()), float64(prevVector.Len())))

	angleAdjustment := 0.0

	currDistance := float64(currVector.Len()) * osuCurrObj.SnapProbability + float64(prevVector.Len()) * osuPrevObj.SnapProbability

	angle := math.Abs(osuCurrObj.Angle)

	if angle < math.Pi / 3 {
		angleAdjustment -= 0.2 * math.Abs(currDistance - angleDistance) * math.Pow(math.Sin(math.Pi / 2 - angle * 1.5), 2)
	} else {
		angleAdjustment += angleDistance * (1 + 0.5 * math.Pow(math.Sin(angle - math.Pi / 4), 2))
	}

	strain := currDistance + angleAdjustment * osuCurrObj.SnapProbability * osuPrevObj.SnapProbability

	strain *= math.Min(osuCurrObj.StrainTime / (osuCurrObj.StrainTime - 20), osuPrevObj.StrainTime / (osuPrevObj.StrainTime - 20))

	return strain
}

func hybridStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject, prevVector, currVector, nextVector vector.Vector2f) float64 {
	flowToSnapVector := prevVector.Sub(currVector)
	snapToFlowVector := currVector.Add(nextVector)

	flowToSnapStrain := float64(flowToSnapVector.Len()) * osuCurrObj.SnapProbability * osuPrevObj.FlowProbability
	snapToFlowStrain := float64(snapToFlowVector.Len()) * osuCurrObj.SnapProbability * osuNextObj.FlowProbability

	strain := math.Max(math.Sqrt(flowToSnapStrain * math.Sqrt(float64(currVector.Len()) * float64(prevVector.Len()))), math.Sqrt(snapToFlowStrain * math.Sqrt(float64(currVector.Len()) * float64(nextVector.Len()))))

	return strain
}

func sliderStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject) float64 {
	strain := osuPrevObj.TravelDistance / osuPrevObj.StrainTime

	return strain
}

func aimStrainValue(skill *Skill, current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok {
		return 0
	}

	result := 0.0

	if len(skill.Previous) > 1 {
		osuNextObj := current
		osuCurrObj := skill.GetPrevious(0)
		osuPrevObj := skill.GetPrevious(1)

		//nextVector := osuNextObj.DistanceVector.Scl(1 / float32(osuNextObj.StrainTime))
		currVector := osuCurrObj.DistanceVector.Scl(1 / float32(osuCurrObj.StrainTime))
		prevVector := osuPrevObj.DistanceVector.Scl(1 / float32(osuPrevObj.StrainTime))

		snapStrain := snapStrainAt(osuPrevObj, osuCurrObj, osuNextObj, prevVector, currVector, prevVector)
		flowStrain := flowStrainAt(osuPrevObj, osuCurrObj, osuNextObj, prevVector, currVector, prevVector)
		hybridStrain := hybridStrainAt(osuPrevObj, osuCurrObj, osuNextObj, prevVector, currVector, prevVector)
		sliderStrain := sliderStrainAt(osuPrevObj, osuCurrObj, osuNextObj)


		skill.currentStrain *= computeDecay(0.75, math.Max(50, osuCurrObj.StrainTime))
		skill.currentStrain += snapStrain * snapStrainMultiplier
		skill.currentStrain += flowStrain * flowStrainMultiplier
		skill.currentStrain += hybridStrain * hybridStrainMultiplier
		skill.currentStrain += sliderStrain * sliderStrainMultiplier

		result = totalStrainMultiplier * skill.currentStrain
	}

	return result
}
