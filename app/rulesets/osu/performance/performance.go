package performance

import (
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930"
)

var diffCalc api.IDifficultyCalculator

func GetDifficultyCalculator() api.IDifficultyCalculator {
	if diffCalc == nil {
		diffCalc = pp220930.NewDifficultyCalculator()
	}

	return diffCalc
}

func CreatePPCalculator() api.IPerformanceCalculator {
	return pp220930.NewPPCalculator()
}
