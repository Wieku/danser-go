package osu

import (
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/beatmap"
	"math"
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/bmath/difficulty"
)

type HitResult int64
type ComboResult int64

var HitResults = struct {
	Ignore,
	Miss,
	Hit50,
	Hit100,
	Hit300,
	Slider10,
	Slider30 HitResult
}{-1, 0, 50, 100, 300, 10, 30}

var ComboResults = struct {
	Reset,
	Hold,
	Increase ComboResult
}{0, 1, 2}

type buttonState struct {
	Left, Right bool
}

type hitobject interface {
	Init(ruleset *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer)
	Update(time int64) bool
	GetFadeTime() int64
}

type difficultyPlayer struct {
	cursor     *render.Cursor
	diff       *difficulty.Difficulty
	cursorLock int64
}

type subSet struct {
	player        *difficultyPlayer
	rawScore      int64
	accuracy      float64
	score         int64
	combo         int64
	modMultiplier float64
	numObjects      int64
}

type OsuRuleSet struct {
	beatMap         *beatmap.BeatMap
	cursors         map[*render.Cursor]*subSet
	scoreMultiplier float64

	queue     []hitobject
	processed []hitobject
}

func NewOsuRuleset(beatMap *beatmap.BeatMap, cursors []*render.Cursor, mods []difficulty.Modifier) *OsuRuleSet {
	ruleset := new(OsuRuleSet)
	ruleset.beatMap = beatMap

	drainTime := beatMap.HitObjects[len(beatMap.HitObjects)-1].GetBasicData().EndTime - beatMap.HitObjects[0].GetBasicData().StartTime
	ruleset.scoreMultiplier = math.Round((beatMap.HPDrainRate + beatMap.OverallDifficulty + beatMap.CircleSize + math.Max(math.Min(float64(len(beatMap.HitObjects))/float64(drainTime)*8, 16), 0)) / 38 * 5)
	//diffPoints := int64(beatMap.HPDrainRate+beatMap.OverallDifficulty+beatMap.CircleSize)

	/*if diffPoints < 6 {
		ruleset.scoreMultiplier = 2
	} else if diffPoints < 13 {
		ruleset.scoreMultiplier = 3
	} else if diffPoints < 18 {
		ruleset.scoreMultiplier = 4
	} else if diffPoints < 25 {
		ruleset.scoreMultiplier = 5
	} else {
		ruleset.scoreMultiplier = 6
	}*/

	ruleset.cursors = make(map[*render.Cursor]*subSet)

	var diffPlayers []*difficultyPlayer

	for i, cursor := range cursors {
		diff := difficulty.NewDifficulty(beatMap.HPDrainRate, beatMap.CircleSize, beatMap.OverallDifficulty, beatMap.AR)
		diff.SetMods(mods[i])

		player := &difficultyPlayer{cursor, diff, -1}
		diffPlayers = append(diffPlayers, player)
		ruleset.cursors[cursor] = &subSet{player, 0, 100, 0, 0, mods[i].GetScoreMultiplier(), 0}
	}

	for _, obj := range beatMap.HitObjects {
		if circle, ok := obj.(*objects.Circle); ok {
			rCircle := new(Circle)
			rCircle.Init(ruleset, circle, diffPlayers)
			ruleset.queue = append(ruleset.queue, rCircle)
		}
		if slider, ok := obj.(*objects.Slider); ok {
			rSlider := new(Slider)
			rSlider.Init(ruleset, slider, diffPlayers)
			ruleset.queue = append(ruleset.queue, rSlider)
		}
	}

	return ruleset
}

func (set *OsuRuleSet) Update(time int64) {
	if len(set.queue) > 0 {
		for i := 0; i < len(set.queue); i++ {
			g := set.queue[i]
			if g.GetFadeTime() > time {
				break
			}

			set.processed = append(set.processed, g)

			if i < len(set.queue)-1 {
				set.queue = append(set.queue[:i], set.queue[i+1:]...)
			} else if i < len(set.queue) {
				set.queue = set.queue[:i]
			}

			i--
		}
	}

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			if isDone := g.Update(time); isDone {
				if i < len(set.processed)-1 {
					set.processed = append(set.processed[:i], set.processed[i+1:]...)
				} else if i < len(set.processed) {
					set.processed = set.processed[:i]
				}
				i--
			}
		}
	}

}

func (set *OsuRuleSet) SendResult(time int64, cursor *render.Cursor, x, y float64, result HitResult, raw bool, comboResult ComboResult) {
	if result == HitResults.Ignore {
		if comboResult == ComboResults.Reset {
			set.cursors[cursor].combo = 0
		}
		return
	}

	subSet := set.cursors[cursor]

	combo := math.Max(float64(subSet.combo-1), 0.0)

	if raw {
		subSet.score += int64(result)
	} else {
		subSet.score += int64(result) + int64(float64(result)*combo*set.scoreMultiplier*subSet.modMultiplier/25.0)
	}

	if result == HitResults.Hit50 || result == HitResults.Hit100 || result == HitResults.Hit300 {
		subSet.rawScore += int64(result)
		subSet.numObjects++
	}

	if comboResult == ComboResults.Reset || result == HitResults.Miss {
		subSet.combo = 0
	} else if comboResult == ComboResults.Increase {
		subSet.combo++
	}

	subSet.accuracy = 100*float64(subSet.rawScore)/float64(subSet.numObjects*300)
	//log.Println("Got:", result, "Combo:", subSet.combo, "Score:", subSet.score, "Acc:", fmt.Sprintf("%0.2f", 100*float64(subSet.rawScore)/float64(set.numObjects*300)))
}

func (set *OsuRuleSet) GetResults(cursor *render.Cursor) (float64, int64) {
	subSet := set.cursors[cursor]
	return subSet.accuracy, subSet.combo
}
