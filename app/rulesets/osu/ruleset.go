package osu

import (
	"fmt"
	"github.com/flesnuk/oppai5"
	"github.com/olekukonko/tablewriter"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const Tolerance2B = 3

type Grade int64

const (
	D = Grade(iota)
	C
	B
	A
	S
	SH
	SS
	SSH
	NONE
)

var GradesText = []string{"D", "C", "B", "A", "S", "SH", "SS", "SSH", "None"}

type ClickAction int64

const (
	Ignored = ClickAction(iota)
	Shake
	Click
)

type ComboResult int64

var ComboResults = struct {
	Reset,
	Hold,
	Increase ComboResult
}{0, 1, 2}

type buttonState struct {
	Left, Right bool
}

func (state buttonState) BothReleased() bool {
	return !(state.Left || state.Right)
}

type HitObject interface {
	Init(ruleset *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer)
	UpdateFor(player *difficultyPlayer, time int64) bool
	UpdateClickFor(player *difficultyPlayer, time int64) bool
	UpdatePostFor(player *difficultyPlayer, time int64) bool
	UpdatePost(time int64) bool
	IsHit(player *difficultyPlayer) bool
	GetFadeTime() int64
	GetNumber() int64
}

type difficultyPlayer struct {
	cursor          *graphics.Cursor
	diff            *difficulty.Difficulty
	DoubleClick     bool
	alreadyStolen   bool
	buttons         buttonState
	gameDownState   bool
	mouseDownButton Buttons
	lastButton      Buttons
	lastButton2     Buttons
	leftCond        bool
	leftCondE       bool
	rightCond       bool
	rightCondE      bool
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
	diff          []oppai.DiffCalc
	ppv2          *oppai.PPv2
	hits          map[HitResult]int64
	currentKatu   int
	currentBad    int
	hp            *HealthProcessor
}

type OsuRuleSet struct {
	beatMap         *beatmap.BeatMap
	cursors         map[*graphics.Cursor]*subSet
	scoreMultiplier float64

	ended bool

	oppaiMaps []*oppai.Map
	oppDiffs  map[difficulty.Modifier][]oppai.DiffCalc

	queue       []HitObject
	processed   []HitObject
	listener    func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result HitResult, comboResult ComboResult, pp float64, score int64)
	endlistener func(time int64, number int64)
}

func NewOsuRuleset(beatMap *beatmap.BeatMap, cursors []*graphics.Cursor, mods []difficulty.Modifier) *OsuRuleSet {
	ruleset := new(OsuRuleSet)
	ruleset.beatMap = beatMap
	ruleset.oppDiffs = make(map[difficulty.Modifier][]oppai.DiffCalc)

	file, err := os.Open(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.File))

	if err != nil {
		panic("Could not open beatmap file.")
	}

	defer file.Close()

	for j := range beatMap.HitObjects {
		ruleset.oppaiMaps = append(ruleset.oppaiMaps, oppai.ParsebyNum(file, j+1))
		file.Seek(0, 0)
	}

	pauses := int64(0)
	for _, p := range beatMap.Pauses {
		pauses += p.GetBasicData().EndTime - p.GetBasicData().StartTime
	}

	drainTime := float32((beatMap.HitObjects[len(beatMap.HitObjects)-1].GetBasicData().EndTime - beatMap.HitObjects[0].GetBasicData().StartTime - pauses) / 1000)

	// HACK HACK HACK:
	// apparently .NET Framework treats doubles differently than other runtimes
	// so we need to subtract a small amount from the value to have proper scoreMultiplier in edge cases (like 4.5 before rounding)
	ruleset.scoreMultiplier = math.Round(float64((float32(beatMap.Diff.GetHPDrain())+float32(beatMap.Diff.GetOD())+float32(beatMap.Diff.GetCS())+bmath.ClampF32(float32(len(beatMap.HitObjects))/drainTime*8, 0, 16))/38*5) - 0.0000001)

	ruleset.cursors = make(map[*graphics.Cursor]*subSet)

	var diffPlayers []*difficultyPlayer

	for i, cursor := range cursors {
		diff := difficulty.NewDifficulty(beatMap.Diff.GetHPDrain(), beatMap.Diff.GetCS(), beatMap.Diff.GetOD(), beatMap.Diff.GetAR())
		diff.SetMods(mods[i])

		player := &difficultyPlayer{cursor: cursor, diff: diff}
		diffPlayers = append(diffPlayers, player)

		if ruleset.oppDiffs[mods[i]] == nil {
			diffs := make([]oppai.DiffCalc, 0)

			for _, m := range ruleset.oppaiMaps {
				diffs = append(diffs, (&oppai.DiffCalc{Beatmap: *m}).Calc(int(mods[i]), oppai.DefaultSingletapThreshold))
			}

			ruleset.oppDiffs[mods[i]] = diffs
		}

		hp := NewHealthProcessor(beatMap, diff)
		hp.CalculateRate()
		hp.ResetHp()

		ruleset.cursors[cursor] = &subSet{player, 0, 100, 0, 0, 0, mods[i].GetScoreMultiplier(), 0, NONE, nil, &oppai.PPv2{}, make(map[HitResult]int64), 0, 0, hp}
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

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			if isDone := g.UpdatePost(time); isDone {
				if set.endlistener != nil {
					set.endlistener(time, g.GetNumber())
				}

				set.processed = append(set.processed[:i], set.processed[i+1:]...)

				i--
			}
		}
	}

	if len(set.queue) > 0 {
		for i := 0; i < len(set.queue); i++ {
			g := set.queue[i]
			if g.GetFadeTime() > time {
				break
			}

			set.processed = append(set.processed, g)

			set.queue = append(set.queue[:i], set.queue[i+1:]...)

			i--
		}
	}

	for _, subSet := range set.cursors {
		subSet.hp.Update(time)
	}

	if len(set.queue) == 0 && len(set.processed) == 0 && !set.ended {
		cs := make([]*graphics.Cursor, 0)
		for c := range set.cursors {
			cs = append(cs, c)
		}
		sort.Slice(cs, func(i, j int) bool {
			return set.cursors[cs[i]].score > set.cursors[cs[j]].score
		})

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"#", "Player", "Score", "Accuracy", "Grade", "300", "100", "50", "Miss", "Combo", "Max Combo", "Mods", "PP"})

		for i, c := range cs {
			var data []string
			data = append(data, fmt.Sprintf("%d", i+1))
			data = append(data, c.Name)
			data = append(data, humanize(set.cursors[c].score))
			data = append(data, fmt.Sprintf("%.2f", set.cursors[c].accuracy))
			data = append(data, GradesText[set.cursors[c].grade])
			data = append(data, humanize(set.cursors[c].hits[Hit300]))
			data = append(data, humanize(set.cursors[c].hits[Hit100]))
			data = append(data, humanize(set.cursors[c].hits[Hit50]))
			data = append(data, humanize(set.cursors[c].hits[Miss]))
			data = append(data, humanize(set.cursors[c].combo))
			data = append(data, humanize(set.cursors[c].maxCombo))
			data = append(data, set.cursors[c].player.diff.Mods.String())
			data = append(data, fmt.Sprintf("%.2f", set.cursors[c].ppv2.Total))
			table.Append(data)
		}

		table.Render()

		for _, s := range strings.Split(tableString.String(), "\n") {
			log.Println(s)
		}

		set.ended = true
	}
}

func humanize(number int64) string {
	stringified := strconv.FormatInt(number, 10)

	a := len(stringified) % 3
	if a == 0 {
		a = 3
	}

	humanized := stringified[0:a]

	for i := a; i < len(stringified); i += 3 {
		humanized += "," + stringified[i:i+3]
	}

	return humanized
}

func (set *OsuRuleSet) UpdateClickFor(cursor *graphics.Cursor, time int64) {
	player := set.cursors[cursor].player

	player.alreadyStolen = false

	if player.cursor.IsReplayFrame || player.cursor.IsPlayer {

		player.leftCond = !player.buttons.Left && player.cursor.LeftButton
		player.rightCond = !player.buttons.Right && player.cursor.RightButton

		player.leftCondE = player.leftCond
		player.rightCondE = player.rightCond

		if player.buttons.Left != player.cursor.LeftButton || player.buttons.Right != player.cursor.RightButton {
			player.gameDownState = player.cursor.LeftButton || player.cursor.RightButton

			player.lastButton2 = player.lastButton
			player.lastButton = player.mouseDownButton
			player.mouseDownButton = Buttons(0)
			if player.cursor.LeftButton {
				player.mouseDownButton |= Left
			}
			if player.cursor.RightButton {
				player.mouseDownButton |= Right
			}
		}
	}

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			g.UpdateClickFor(player, time)
		}
	}

	if player.cursor.IsReplayFrame || player.cursor.IsPlayer {
		player.buttons.Left = player.cursor.LeftButton
		player.buttons.Right = player.cursor.RightButton
	}
}

func (set *OsuRuleSet) UpdateNormalFor(cursor *graphics.Cursor, time int64) {
	player := set.cursors[cursor].player

	//wasSliderAlready := false

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			/*s, isSlider := g.(*Slider)

			if isSlider {
				if wasSliderAlready {
					continue
				}

				if !s.IsHit(player) {
					wasSliderAlready = true
				}
			}*/

			g.UpdateFor(player, time)
		}
	}
}

func (set *OsuRuleSet) UpdatePostFor(cursor *graphics.Cursor, time int64) {
	player := set.cursors[cursor].player

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			g.UpdatePostFor(player, time)
		}
	}
}

func (set *OsuRuleSet) SendResult(time int64, cursor *graphics.Cursor, number int64, x, y float32, result HitResult, raw bool, comboResult ComboResult) {
	if result == Ignore {
		return
	}

	subSet := set.cursors[cursor]

	combo := bmath.MaxI64(subSet.combo-1, 0)

	if result != SliderMiss {

		increase := result.ScoreValue()

		if raw {
			subSet.score += increase
		} else {
			subSet.score += increase + int64(float64(increase)*float64(combo)*set.scoreMultiplier*subSet.modMultiplier/25.0)
		}
	}

	if result&BaseHitsM > 0 {
		subSet.rawScore += result.ScoreValue()
		subSet.hits[result]++
		subSet.numObjects++
	}

	if comboResult == ComboResults.Reset || result == Miss {
		subSet.combo = 0
	} else if comboResult == ComboResults.Increase {
		subSet.combo++
	}

	subSet.maxCombo = bmath.MaxI64(subSet.combo, subSet.maxCombo)

	if subSet.numObjects == 0 {
		subSet.accuracy = 100
	} else {
		subSet.accuracy = 100 * float64(subSet.rawScore) / float64(subSet.numObjects*300)
	}

	ratio := float64(subSet.hits[Hit300]) / float64(subSet.numObjects)

	if subSet.hits[Hit300] == subSet.numObjects {
		if subSet.player.diff.Mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			subSet.grade = SSH
		} else {
			subSet.grade = SS
		}
	} else if ratio > 0.9 && float64(subSet.hits[Hit50])/float64(subSet.numObjects) < 0.01 && subSet.hits[Miss] == 0 {
		if subSet.player.diff.Mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			subSet.grade = SH
		} else {
			subSet.grade = S
		}
	} else if ratio > 0.8 && subSet.hits[Miss] == 0 || ratio > 0.9 {
		subSet.grade = A
	} else if ratio > 0.7 && subSet.hits[Miss] == 0 || ratio > 0.8 {
		subSet.grade = B
	} else if ratio > 0.6 {
		subSet.grade = C
	} else {
		subSet.grade = D
	}

	index := bmath.MaxI64(0, subSet.numObjects-1)

	diff := set.oppDiffs[subSet.player.diff.Mods][index]

	subSet.ppv2.PPv2WithMods(diff.Aim, diff.Speed, set.oppaiMaps[index], int(subSet.player.diff.Mods), int(subSet.hits[Hit300]), int(subSet.hits[Hit100]), int(subSet.hits[Hit50]), int(subSet.hits[Miss]), int(subSet.maxCombo))

	switch result {
	case Hit100:
		subSet.currentKatu++
	case Hit50, Miss:
		subSet.currentBad++
	}

	if result&BaseHitsM > 0 && (int(number) == len(set.beatMap.HitObjects)-1 || (int(number) < len(set.beatMap.HitObjects)-1 && set.beatMap.HitObjects[number+1].GetBasicData().NewCombo)) {
		allClicked := true

		// We don't want to give geki/katu if all objects in current combo weren't clicked
		index := sort.Search(len(set.processed), func(i int) bool {
			return set.processed[i].GetNumber() >= number
		})

		for i := index - 1; i >= 0; i-- {
			obj := set.processed[i]

			if !obj.IsHit(subSet.player) {
				allClicked = false
				break
			}

			if set.beatMap.HitObjects[obj.GetNumber()].GetBasicData().NewCombo {
				break
			}
		}

		if subSet.currentKatu == 0 && subSet.currentBad == 0 && allClicked {
			result |= GekiAddition
		} else if subSet.currentBad == 0 && allClicked {
			result |= KatuAddition
		} else {
			result |= MuAddition
		}

		subSet.currentBad = 0
		subSet.currentKatu = 0
	}

	subSet.hp.AddResult(result)

	if set.listener != nil {
		set.listener(cursor, time, number, vector.NewVec2f(x, y).Copy64(), result, comboResult, subSet.ppv2.Total, subSet.score)
	}

	if len(set.cursors) == 1 {
		log.Println(fmt.Sprintf(
			"Got: %3d, Combo: %4d, Max Combo: %4d, Score: %9d, Acc: %6.2f%%, 300: %4d, 100: %3d, 50: %2d, miss: %2d, from: %d, at: %d, pos: %.0fx%.0f, pp: %.2f",
			result.ScoreValue(),
			subSet.combo,
			subSet.maxCombo,
			subSet.score,
			subSet.accuracy,
			subSet.hits[Hit300],
			subSet.hits[Hit100],
			subSet.hits[Hit50],
			subSet.hits[Miss],
			number,
			time,
			x,
			y,
			subSet.ppv2.Total,
		))
	}
}

func (set *OsuRuleSet) CanBeHit(time int64, object HitObject, player *difficultyPlayer) ClickAction {
	if _, ok := object.(*Circle); ok {
		index := -1

		for i, g := range set.processed {
			if g == object {
				index = i
			}
		}

		if index > 0 && set.beatMap.HitObjects[set.processed[index-1].GetNumber()].GetBasicData().StackIndex > 0 && !set.processed[index-1].IsHit(player) {
			return Ignored //don't shake the stacks
		}
	}

	for _, g := range set.processed {
		if !g.IsHit(player) {
			if g.GetNumber() != object.GetNumber() {
				if set.beatMap.HitObjects[g.GetNumber()].GetBasicData().EndTime+Tolerance2B < set.beatMap.HitObjects[object.GetNumber()].GetBasicData().StartTime {
					return Shake
				}
			} else {
				break
			}
		}
	}

	if math.Abs(float64(time-set.beatMap.HitObjects[object.GetNumber()].GetBasicData().StartTime)) >= difficulty.HittableRange {
		return Shake
	}
	return Click
}

func (set *OsuRuleSet) SetListener(listener func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result HitResult, comboResult ComboResult, pp float64, score int64)) {
	set.listener = listener
}

func (set *OsuRuleSet) SetEndListener(endlistener func(time int64, number int64)) {
	set.endlistener = endlistener
}

func (set *OsuRuleSet) GetResults(cursor *graphics.Cursor) (float64, int64, int64, Grade) {
	subSet := set.cursors[cursor]
	return subSet.accuracy, subSet.maxCombo, subSet.score, subSet.grade
}

func (set *OsuRuleSet) GetHP(cursor *graphics.Cursor) float64 {
	subSet := set.cursors[cursor]
	return subSet.hp.Health / MaxHp
}

func (set *OsuRuleSet) GetPlayer(cursor *graphics.Cursor) *difficultyPlayer {
	subSet := set.cursors[cursor]
	return subSet.player
}

func (set *OsuRuleSet) GetProcessed() []HitObject {
	return set.processed
}

func (set *OsuRuleSet) GetBeatMap() *beatmap.BeatMap {
	return set.beatMap
}
