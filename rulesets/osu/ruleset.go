package osu

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/render/batches"
	"math"
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/bmath/difficulty"
	"sort"
	"strconv"
	"fmt"
	"log"
	"github.com/wieku/danser/bmath"
)

type Grade int64

const (
	FadeIn   = 120
	FadeOut  = 600
	PostEmpt = 500
)

const (
	D = iota
	C
	B
	A
	S
	SH
	SS
	SSH
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
	Slider30,
	SliderMiss,
	SpinnerBonus,
	SpinnerScore HitResult
}{-1, 0, 50, 100, 300, 10, 30, -2, 1100, -3}

func (result HitResult) String() string {
	switch result {
	case -2:
		return "0"
	case -3:
		return "100"
	default:
		return strconv.Itoa(int(result))
	}
}

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
	Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch)
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
	maxCombo      int64
	modMultiplier float64
	numObjects    int64
	grade         Grade
	hits          map[HitResult]int64
}

type OsuRuleSet struct {
	beatMap         *beatmap.BeatMap
	cursors         map[*render.Cursor]*subSet
	scoreMultiplier float64

	queue     []hitobject
	processed []hitobject
	listener  func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result HitResult, comboResult ComboResult)
}

func NewOsuRuleset(beatMap *beatmap.BeatMap, cursors []*render.Cursor, mods []difficulty.Modifier) *OsuRuleSet {
	ruleset := new(OsuRuleSet)
	ruleset.beatMap = beatMap

	pauses := int64(0)
	for _, p := range beatMap.Pauses {
		pauses += p.GetBasicData().EndTime - p.GetBasicData().StartTime
	}

	drainTime := float64(beatMap.HitObjects[len(beatMap.HitObjects)-1].GetBasicData().EndTime-beatMap.HitObjects[0].GetBasicData().StartTime-pauses) / 1000
	ruleset.scoreMultiplier = math.Round((beatMap.HPDrainRate + beatMap.OverallDifficulty + beatMap.CircleSize + math.Max(math.Min(float64(len(beatMap.HitObjects))/float64(drainTime)*8, 16), 0)) / 38 * 5)

	ruleset.cursors = make(map[*render.Cursor]*subSet)

	var diffPlayers []*difficultyPlayer

	for i, cursor := range cursors {
		diff := difficulty.NewDifficulty(beatMap.HPDrainRate, beatMap.CircleSize, beatMap.OverallDifficulty, beatMap.AR)
		diff.SetMods(mods[i])

		player := &difficultyPlayer{cursor, diff, -1}
		diffPlayers = append(diffPlayers, player)

		grade := Grade(SS)

		if mods[i]&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			grade = SSH
		}

		ruleset.cursors[cursor] = &subSet{player, 0, 100, 0, 0, 0, mods[i].GetScoreMultiplier(), 0, grade, make(map[HitResult]int64)}
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
		if spinner, ok := obj.(*objects.Spinner); ok {
			rSpinner := new(Spinner)
			rSpinner.Init(ruleset, spinner, diffPlayers)
			ruleset.queue = append(ruleset.queue, rSpinner)
		}
	}

	sort.Slice(ruleset.queue, func(i, j int) bool {
		return ruleset.queue[i].GetFadeTime() < ruleset.queue[j].GetFadeTime()
	})

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

func (set *OsuRuleSet) Draw(time int64, batch *batches.SpriteBatch, color mgl32.Vec4) {
	if len(set.processed) > 0 {
		for i := len(set.processed)-1; i > 0; i-- {
			g := set.processed[i]
			g.Draw(time, color, batch)
		}
	}
}

func (set *OsuRuleSet) SendResult(time int64, cursor *render.Cursor, number int64, x, y float64, result HitResult, raw bool, comboResult ComboResult) {
	if result == HitResults.Ignore {
		return
	}

	subSet := set.cursors[cursor]

	combo := math.Max(float64(subSet.combo-1), 0.0)

	if result != HitResults.SliderMiss {

		increase := int64(result)

		if result == HitResults.SpinnerScore {
			increase = 100
		}

		if raw {
			subSet.score += increase
		} else {
			subSet.score += increase + int64(float64(increase)*combo*set.scoreMultiplier*subSet.modMultiplier/25.0)
		}
	}

	if result == HitResults.Hit50 || result == HitResults.Hit100 || result == HitResults.Hit300 || result == HitResults.Miss {
		subSet.rawScore += int64(result)
		subSet.hits[result]++
		subSet.numObjects++
	}

	if comboResult == ComboResults.Reset || result == HitResults.Miss {
		subSet.combo = 0
	} else if comboResult == ComboResults.Increase {
		subSet.combo++
	}

	if subSet.combo > subSet.maxCombo {
		subSet.maxCombo = subSet.combo
	}

	if subSet.numObjects == 0 {
		subSet.accuracy = 100
	} else {
		subSet.accuracy = 100 * float64(subSet.rawScore) / float64(subSet.numObjects*300)
	}

	ratio := float64(subSet.hits[HitResults.Hit300]) / float64(subSet.numObjects)

	if subSet.hits[HitResults.Hit300] == subSet.numObjects {
		if subSet.player.diff.Mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			subSet.grade = SSH
		} else {
			subSet.grade = SS
		}
	} else if ratio > 0.9 && float64(subSet.hits[HitResults.Hit50])/float64(subSet.numObjects) < 0.01 && subSet.hits[HitResults.Miss] == 0 {
		if subSet.player.diff.Mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			subSet.grade = SH
		} else {
			subSet.grade = S
		}
	} else if ratio > 0.8 && subSet.hits[HitResults.Miss] == 0 || ratio > 0.9 {
		subSet.grade = A
	} else if ratio > 0.7 && subSet.hits[HitResults.Miss] == 0 || ratio > 0.8 {
		subSet.grade = B
	} else if ratio > 0.6 {
		subSet.grade = C
	} else {
		subSet.grade = D
	}

	if set.listener != nil {
		set.listener(cursor, time, number, bmath.NewVec2d(x, y), result, comboResult)
	}

	if len(set.cursors) == 1 {
		log.Println("Got:", fmt.Sprintf("%3s", result), "Combo:", fmt.Sprintf("%4d", subSet.combo), "Max Combo:", fmt.Sprintf("%4d", subSet.maxCombo), "Score:", fmt.Sprintf("%9d", subSet.score), "Acc:", fmt.Sprintf("%3.2f%%", 100*float64(subSet.rawScore)/float64(subSet.numObjects*300)), subSet.hits)
	}
}

func (set *OsuRuleSet) SetListener(listener func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result HitResult, comboResult ComboResult)) {
	set.listener = listener
}

func (set *OsuRuleSet) GetResults(cursor *render.Cursor) (float64, int64, int64, Grade) {
	subSet := set.cursors[cursor]
	return subSet.accuracy, subSet.maxCombo, subSet.score, subSet.grade
}
