package evaluators

import (
	"math"

	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260706/preprocessing"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/putils"
	"github.com/wieku/danser-go/framework/math/mutils"
)

const (
	readingWindowSize                 = 3000                                   // 3 seconds
	readingDistanceInfluenceThreshold = preprocessing.NormalizedDiameter * 1.5 // 1.5 circles distance between centers
)

func EvaluateReading(current *preprocessing.DifficultyObject, hidden bool) float64 {
	if current.IsSpinner || current.Index == 0 {
		return 0
	}

	currObj := current
	nextObj := current.Next(0)

	velocity := max(1, currObj.LazyJumpDistance/currObj.AdjustedDeltaTime) // Only allow velocity to buff

	currentVisibleObjectDensity := retrieveCurrentVisibleObjectDensity(currObj)
	pastObjectDifficultyInfluence := getPastObjectDifficultyInfluence(currObj)

	constantAngleNerfFactor := getConstantAngleNerfFactor(currObj)

	noteDensityDifficulty := calculateDensityDifficulty(nextObj, velocity, constantAngleNerfFactor, pastObjectDifficultyInfluence, currentVisibleObjectDensity)

	hiddenDifficulty := 0.0
	if hidden {
		hiddenDifficulty = calculateHiddenDifficulty(currObj, pastObjectDifficultyInfluence, currentVisibleObjectDensity, velocity, constantAngleNerfFactor)
	}

	preemptDifficulty := calculatePreemptDifficulty(velocity, constantAngleNerfFactor, currObj.Preempt)

	readingDifficulty := putils.Norm(1.5, preemptDifficulty, hiddenDifficulty, noteDensityDifficulty)

	// Having less time to process information is harder
	readingDifficulty *= readingHighBpmBonus(currObj.AdjustedDeltaTime)

	return readingDifficulty
}

// / <summary>
// / Calculates the density difficulty of the current object and how hard it is to aim it because of it based on:
// / <list type="bullet">
// / <item><description>cursor velocity to the current object,</description></item>
// / <item><description>how many times the current object's angle was repeated,</description></item>
// / <item><description>density of objects visible when the current object appears,</description></item>
// / <item><description>density of objects visible when the current object needs to be clicked,</description></item>
// / /// </list>
// / </summary>
func calculateDensityDifficulty(nextObj *preprocessing.DifficultyObject, velocity, constantAngleNerfFactor, pastObjectDifficultyInfluence, currentVisibleObjectDensity float64) float64 {
	const readingDensityMultiplier = 2.4
	const readingDensityDifficultyBase = 2.5

	// Consider future densities too because it can make the path the cursor takes less clear
	futureObjectDifficultyInfluence := math.Sqrt(currentVisibleObjectDensity)

	if nextObj != nil {
		// Reduce difficulty if movement to next object is small
		futureObjectDifficultyInfluence *= putils.Smootherstep(nextObj.LazyJumpDistance, 15, readingDistanceInfluenceThreshold)
	}

	// Value higher note densities exponentially
	noteDensityDifficulty := math.Pow(pastObjectDifficultyInfluence+futureObjectDifficultyInfluence, 1.7) * 0.4 * constantAngleNerfFactor * velocity

	// Award only denser than average maps.
	noteDensityDifficulty = max(0, noteDensityDifficulty-readingDensityDifficultyBase)

	// Apply a soft cap to general density reading to account for partial memorization
	noteDensityDifficulty = math.Pow(noteDensityDifficulty, 0.45) * readingDensityMultiplier

	return noteDensityDifficulty
}

// / <summary>
// / Calculates the difficulty of aiming the current object when the approach rate is very high based on:
// / <list type="bullet">
// / <item><description>cursor velocity to the current object,</description></item>
// / <item><description>how many times the current object's angle was repeated,</description></item>
// / <item><description>how many milliseconds elapse between the approach circle appearing and touching the inner circle</description></item>
// / </list>
// / </summary>
func calculatePreemptDifficulty(velocity, constantAngleNerfFactor, preempt float64) float64 {
	const readingPreemptBalancingFactor = 140000
	const readingPreemptStartingPoint = 500 // AR 9.66 in milliseconds

	// Arbitrary curve for the base value preempt difficulty should have as approach rate increases.
	// https://www.desmos.com/calculator/c175335a71
	preemptDifficulty := math.Pow((readingPreemptStartingPoint-preempt+math.Abs(preempt-readingPreemptStartingPoint))/2, 2.5) / readingPreemptBalancingFactor

	preemptDifficulty *= constantAngleNerfFactor * velocity

	return preemptDifficulty
}

// / <summary>
// / Calculates the difficulty of aiming the current object when the hidden mod is active based on:
// / <list type="bullet">
// / <item><description>cursor velocity to the current object,</description></item>
// / <item><description>time the current object spends invisible,</description></item>
// / <item><description>density of objects visible when the current object appears,</description></item>
// / <item><description>density of objects visible when the current object needs to be clicked,</description></item>
// / <item><description>how many times the current object's angle was repeated,</description></item>
// / <item><description>if the current object is perfectly stacked to the previous one</description></item>
// / </list>
// / </summary>
func calculateHiddenDifficulty(currObj *preprocessing.DifficultyObject, pastObjectDifficultyInfluence, currentVisibleObjectDensity, velocity, constantAngleNerfFactor float64) float64 {
	const readingHiddenMultiplier = 0.28

	// Higher preempt means that time spent invisible is higher too, we want to reward that
	preemptFactor := math.Pow(currObj.Preempt, 2.2) * 0.01

	// Account for both past and current densities
	densityFactor := math.Pow(currentVisibleObjectDensity+pastObjectDifficultyInfluence, 3.3) * 3

	hiddenDifficulty := (preemptFactor + densityFactor) * constantAngleNerfFactor * velocity * 0.01

	// Apply a soft cap to general HD reading to account for partial memorization
	hiddenDifficulty = math.Pow(hiddenDifficulty, 0.4) * readingHiddenMultiplier

	var previousObj = currObj.Previous(0)

	// Buff perfect stacks only if current note is completely invisible at the time you click the previous note.
	if currObj.LazyJumpDistance == 0 && currObj.OpacityAt(previousObj.BaseObject.GetStartTime(), true) == 0 && previousObj.StartTime > currObj.StartTime-currObj.Preempt {
		hiddenDifficulty += readingHiddenMultiplier * 2500 / math.Pow(currObj.AdjustedDeltaTime, 1.5) // Perfect stacks are harder the less time between notes
	}

	return hiddenDifficulty
}

func getPastObjectDifficultyInfluence(currObj *preprocessing.DifficultyObject) float64 {
	pastObjectDifficultyInfluence := 0.0

	for _, loopObj := range retrievePastVisibleObjects(currObj) {
		loopDifficulty := currObj.OpacityAt(loopObj.BaseObject.GetStartTime(), false)

		// When aiming an object small distances mean previous objects may be cheesed, so it doesn't matter whether they were arranged confusingly.
		loopDifficulty *= putils.Smootherstep(loopObj.LazyJumpDistance, 15, readingDistanceInfluenceThreshold)

		// Account less for objects close to the max reading window
		timeBetweenCurrAndLoopObj := currObj.StartTime - loopObj.StartTime
		timeNerfFactor := getTimeNerfFactor(timeBetweenCurrAndLoopObj)

		loopDifficulty *= timeNerfFactor
		pastObjectDifficultyInfluence += loopDifficulty
	}

	return pastObjectDifficultyInfluence
}

// Returns a list of objects that are visible on screen at the point in time the current object becomes visible.
func retrievePastVisibleObjects(current *preprocessing.DifficultyObject) (yield []*preprocessing.DifficultyObject) {
	for i := 0; i < current.Index; i++ {
		hitObject := current.Previous(i)

		if hitObject == nil ||
			current.StartTime-hitObject.StartTime > readingWindowSize ||
			hitObject.StartTime < current.StartTime-current.Preempt { // Current object not visible at the time object needs to be clicked
			break
		}

		yield = append(yield, hitObject)
	}

	return
}

// Returns the density of objects visible at the point in time the current object needs to be clicked capped by the reading window.
func retrieveCurrentVisibleObjectDensity(current *preprocessing.DifficultyObject) float64 {
	visibleObjectCount := 0.0

	hitObject := current.Next(0)

	for hitObject != nil {
		if hitObject.StartTime-current.StartTime > readingWindowSize ||
			current.StartTime < hitObject.StartTime-hitObject.Preempt { // Object not visible at the time current object needs to be clicked.
			break
		}

		timeBetweenCurrAndLoopObj := hitObject.StartTime - current.StartTime
		timeNerfFactor := getTimeNerfFactor(timeBetweenCurrAndLoopObj)

		visibleObjectCount += hitObject.OpacityAt(current.BaseObject.GetStartTime(), false) * timeNerfFactor

		hitObject = hitObject.Next(0)
	}

	return visibleObjectCount
}

// Returns a factor of how often the current object's angle has been repeated in a certain time frame.
// It does this by checking the difference in angle between current and past objects and sums them based on a range of similarity.
// https://www.desmos.com/calculator/eb057a4822
func getConstantAngleNerfFactor(current *preprocessing.DifficultyObject) float64 {
	const readingMinimumAngleRelevancyTime = 2000 // 2 seconds
	const readingMaximumAngleRelevancyTime = 200

	constantAngleCount := 0.0
	index := 0
	currentTimeGap := 0.0

	loopObjPrev0 := current
	var loopObjPrev1, loopObjPrev2 *preprocessing.DifficultyObject

	for currentTimeGap < readingMinimumAngleRelevancyTime {
		var loopObj = current.Previous(index)

		if loopObj == nil {
			break
		}

		// Account less for objects that are close to the time limit.
		longIntervalFactor := 1 - putils.ReverseLerp(loopObj.AdjustedDeltaTime, readingMaximumAngleRelevancyTime, readingMinimumAngleRelevancyTime)

		if !math.IsNaN(loopObj.Angle) && !math.IsNaN(current.Angle) {
			angleDifference := math.Abs(current.Angle - loopObj.Angle)
			angleDifferenceAlternating := math.Pi

			if !math.IsNaN(loopObjPrev0.Angle) && loopObjPrev1 != nil && !math.IsNaN(loopObjPrev1.Angle) && loopObjPrev2 != nil && !math.IsNaN(loopObjPrev2.Angle) {
				angleDifferenceAlternating = math.Abs(loopObjPrev1.Angle - loopObj.Angle)
				angleDifferenceAlternating += math.Abs(loopObjPrev2.Angle - loopObjPrev0.Angle)

				weight := 1.0

				// Be sure that one of the angles is very sharp, when other is wide
				weight *= putils.ReverseLerp(min(loopObj.Angle, loopObjPrev0.Angle)*180/math.Pi, 20, 5)
				weight *= putils.ReverseLerp(max(loopObj.Angle, loopObjPrev0.Angle)*180/math.Pi, 60, 120)

				// Lerp between max angle difference and rescaled alternating difference, with more harsh scaling compared to normal difference
				angleDifferenceAlternating = mutils.Lerp(math.Pi, 0.1*angleDifferenceAlternating, weight)
			}

			stackFactor := putils.Smootherstep(loopObj.LazyJumpDistance, 0, preprocessing.NormalizedRadius)

			constantAngleCount += math.Cos(3*min(putils.DegreesToRadians(30), min(angleDifference, angleDifferenceAlternating)*stackFactor)) * longIntervalFactor
		}

		currentTimeGap = current.StartTime - loopObj.StartTime
		index++

		loopObjPrev2 = loopObjPrev1
		loopObjPrev1 = loopObjPrev0
		loopObjPrev0 = loopObj
	}

	return mutils.Clamp(2/constantAngleCount, 0.2, 1)
}

// Returns a nerfing factor for when objects are very distant in time, affecting reading less.
func getTimeNerfFactor(deltaTime float64) float64 {
	return mutils.Clamp(2-deltaTime/(readingWindowSize/2), 0, 1)
}

func readingHighBpmBonus(ms float64) float64 {
	return 1 / (1 - math.Pow(0.8, ms/1000))
}
