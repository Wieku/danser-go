package performance

import (
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp211112"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp250306"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp251020"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp260727"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp26xxxx"
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
	case "241007":
		diffCalcInit = pp241007.NewDifficultyCalculator
		ppCalcInit = pp241007.NewPPCalculator
	case "250306":
		diffCalcInit = pp250306.NewDifficultyCalculator
		ppCalcInit = pp250306.NewPPCalculator
	case "251020":
		diffCalcInit = pp251020.NewDifficultyCalculator
		ppCalcInit = pp251020.NewPPCalculator
	case "26xxxx":
		diffCalcInit = pp26xxxx.NewDifficultyCalculator
		ppCalcInit = pp26xxxx.NewPPCalculator
	default:
		diffCalcInit = pp260727.NewDifficultyCalculator
		ppCalcInit = pp260727.NewPPCalculator
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
