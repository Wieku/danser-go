package input

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu"
)

const leniency = 12

type RelaxInputProcessor struct {
	cursor  *graphics.Cursor
	ruleset *osu.OsuRuleSet
	wasLeft bool
}

func NewRelaxInputProcessor(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor) *RelaxInputProcessor {
	processor := new(RelaxInputProcessor)
	processor.cursor = cursor
	processor.ruleset = ruleset

	return processor
}

func (processor *RelaxInputProcessor) Update(time float64) {
	processed := processor.ruleset.GetProcessed()
	player := processor.ruleset.GetPlayer(processor.cursor)

	click := false

	currDiff := processor.ruleset.GetPlayerDifficulty(processor.cursor)
	isLazer := currDiff.CheckModActive(difficulty.Lazer)

	for _, o := range processed {
		circle, c1 := o.(*osu.Circle)
		slider, c2 := o.(*osu.Slider)
		_, c3 := o.(*osu.Spinner)

		if c3 || (c1 && circle.IsHit(player)) || (c2 && slider.IsStartHit(player)) {
			continue
		}

		obj := o.GetObject()

		if isLazer {
			pos := obj.GetStackedStartPositionMod(currDiff)

			if (!c2 || time <= obj.GetEndTime()) &&
				time >= obj.GetStartTime()-leniency &&
				pos.Dst(processor.cursor.RawPosition) <= float32(currDiff.CircleRadiusL) &&
				time-obj.GetStartTime() <= currDiff.Hit50U {
				click = true
			}
		} else if time > obj.GetStartTime()-leniency {
			click = true
		}
	}

	processor.cursor.LeftButton = click && !processor.wasLeft
	processor.cursor.RightButton = click && processor.wasLeft

	if click {
		processor.wasLeft = !processor.wasLeft
	}
}
