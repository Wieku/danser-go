package performance

import (
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp211112"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007"
	"github.com/wieku/danser-go/app/settings"
)

var diffCalcInit func() api.IDifficultyCalculator
var ppCalcInit func() api.IPerformanceCalculator

func initConstructors() {
	if diffCalcInit != nil {
		return
	}

	switch settings.Gameplay.PPVersion {
	case "211112":
		diffCalcInit = pp211112.NewDifficultyCalculator
		ppCalcInit = pp211112.NewPPCalculator
	case "220930":
		diffCalcInit = pp220930.NewDifficultyCalculator
		ppCalcInit = pp220930.NewPPCalculator
	default:
		diffCalcInit = pp241007.NewDifficultyCalculator
		ppCalcInit = pp241007.NewPPCalculator
	}
}

var diffCalc api.IDifficultyCalculator

func GetDifficultyCalculator() api.IDifficultyCalculator {
	initConstructors()

	if diffCalc == nil {
		diffCalc = diffCalcInit()
	}

	return diffCalc
}

func CreatePPCalculator() api.IPerformanceCalculator {
	initConstructors()

	return ppCalcInit()
}
