package skills

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/pp/xexxar/preprocessing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	snapStrainMultiplier   float64 = 24.5
	flowStrainMultiplier   float64 = 23.75
	sliderStrainMultiplier float64 = 75
	totalStrainMultiplier  float64 = 0.1675
	distanceConstant       float64 = 2.5
	baseDecayAim           float64 = 0.75
)

func NewAimSkill(d *difficulty.Difficulty) *Skill {
	skill := NewSkill(d)
	skill.StarsPerDouble = 1.1
	skill.HistoryLength = 2
	skill.currentStrain = 1
	skill.StrainValueOf = aimStrainValue

	return skill
}

func snapScaling(distance float64) float64 {
	if distance <= distanceConstant {
		return 1
	}

	return (distanceConstant + (distance-distanceConstant)*(math.Log(1+(distance-distanceConstant)/math.Sqrt(2))/math.Log(2))/(distance-distanceConstant)) / distance
}

func flowStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject, prevVector, currVector, nextVector vector.Vector2f) float64 {
	observedDistance := currVector.Sub(prevVector.Scl(0.1))

	prevAngularMomentumChange := math.Abs(osuCurrObj.Angle*currVector.Len64() - osuPrevObj.Angle*prevVector.Len64())
	nextAngularMomentumChange := math.Abs(osuCurrObj.Angle*currVector.Len64() - osuNextObj.Angle*nextVector.Len64())

	angularMomentumChange := math.Sqrt(math.Min(currVector.Len64(), prevVector.Len64()) * math.Abs(nextAngularMomentumChange-prevAngularMomentumChange) / (2 * math.Pi))

	momentumChange := math.Sqrt(math.Abs(currVector.Len64()-prevVector.Len64()) * math.Min(currVector.Len64(), prevVector.Len64()))

	strain := osuCurrObj.FlowProbability * (observedDistance.Len64() +
		momentumChange +
		angularMomentumChange*osuPrevObj.FlowProbability)

	strain *= math.Min(osuCurrObj.StrainTime/(osuCurrObj.StrainTime-20), osuPrevObj.StrainTime/(osuPrevObj.StrainTime-20))
	// buff high BPM slightly.

	return strain
}

func snapStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject, prevVector, currVector, nextVector vector.Vector2f) float64 {
	observedDistance := currVector.Add(prevVector.Scl(0.35))

	strain := (observedDistance.Len64() * snapScaling((observedDistance.Len64() * osuCurrObj.StrainTime) / 100)) * osuCurrObj.SnapProbability

	strain *= math.Min(osuCurrObj.StrainTime/(osuCurrObj.StrainTime-20), osuPrevObj.StrainTime/(osuPrevObj.StrainTime-20))

	return strain
}

func sliderStrainAt(osuPrevObj, osuCurrObj, osuNextObj *preprocessing.DifficultyObject) float64 {
	return osuPrevObj.TravelDistance / osuPrevObj.StrainTime
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

		nextVector := osuNextObj.DistanceVector.Scl(1 / float32(osuNextObj.StrainTime))
		currVector := osuCurrObj.DistanceVector.Scl(1 / float32(osuCurrObj.StrainTime))
		prevVector := osuPrevObj.DistanceVector.Scl(1 / float32(osuPrevObj.StrainTime))

		snapStrain := snapStrainAt(osuPrevObj, osuCurrObj, osuNextObj, prevVector, currVector, nextVector)
		flowStrain := flowStrainAt(osuPrevObj, osuCurrObj, osuNextObj, prevVector, currVector, nextVector)
		sliderStrain := sliderStrainAt(osuPrevObj, osuCurrObj, osuNextObj)

		skill.currentStrain *= computeDecay(baseDecayAim, osuCurrObj.StrainTime)
		skill.currentStrain += snapStrain * snapStrainMultiplier
		skill.currentStrain += flowStrain * flowStrainMultiplier
		skill.currentStrain += sliderStrain * sliderStrainMultiplier

		result = totalStrainMultiplier * skill.currentStrain
	}

	return result
}
