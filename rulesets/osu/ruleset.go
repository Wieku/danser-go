package osu

import (
	"fmt"
	"github.com/flesnuk/oppai5"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/olekukonko/tablewriter"
	"github.com/wieku/danser-go/beatmap"
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/difficulty"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/settings"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Grade int64

const (
	FadeIn      = 120
	FadeOut     = 600
	PostEmpt    = 500
	Tolerance2B = 3
)

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

type ClickAction int64

const (
	Ignored = ClickAction(iota)
	Shake
	Click
)

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

type hitobject interface {
	Init(ruleset *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer)
	UpdateFor(player *difficultyPlayer, time int64) bool
	UpdateClickFor(player *difficultyPlayer, time int64) bool
	UpdatePost(time int64) bool
	IsHit(player *difficultyPlayer) bool
	Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch)
	GetFadeTime() int64
	GetNumber() int64
}

type difficultyPlayer struct {
	cursor *render.Cursor
	diff   *difficulty.Difficulty
	//cursorLock    int64
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
}

type OsuRuleSet struct {
	beatMap         *beatmap.BeatMap
	cursors         map[*render.Cursor]*subSet
	scoreMultiplier float64

	ended bool

	oppaiMaps []*oppai.Map
	oppDiffs  map[int][]oppai.DiffCalc
	params    *oppai.Parameters

	queue       []hitobject
	processed   []hitobject
	listener    func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result HitResult, comboResult ComboResult, pp float64, score int64)
	endlistener func(time int64, number int64)
}

func NewOsuRuleset(beatMap *beatmap.BeatMap, cursors []*render.Cursor, mods []difficulty.Modifier) *OsuRuleSet {
	ruleset := new(OsuRuleSet)
	ruleset.beatMap = beatMap
	ruleset.oppDiffs = make(map[int][]oppai.DiffCalc)

	file, err := os.Open(settings.General.OsuSongsDir + string(os.PathSeparator) + beatMap.Dir + string(os.PathSeparator) + beatMap.File)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	for j := range beatMap.HitObjects {
		ruleset.oppaiMaps = append(ruleset.oppaiMaps, oppai.ParsebyNum(file, j+1))
		file.Seek(0, 0)
	}

	ruleset.params = &oppai.Parameters{}

	pauses := int64(0)
	for _, p := range beatMap.Pauses {
		pauses += p.GetBasicData().EndTime - p.GetBasicData().StartTime
	}

	drainTime := float64(beatMap.HitObjects[len(beatMap.HitObjects)-1].GetBasicData().EndTime-beatMap.HitObjects[0].GetBasicData().StartTime-pauses) / 1000
	ruleset.scoreMultiplier = math.Round((beatMap.Diff.GetHPDrain() + beatMap.Diff.GetOD() + beatMap.Diff.GetCS() + math.Max(math.Min(float64(len(beatMap.HitObjects))/float64(drainTime)*8, 16), 0)) / 38 * 5)

	ruleset.cursors = make(map[*render.Cursor]*subSet)

	var diffPlayers []*difficultyPlayer

	for i, cursor := range cursors {
		diff := difficulty.NewDifficulty(beatMap.Diff.GetHPDrain(), beatMap.Diff.GetCS(), beatMap.Diff.GetOD(), beatMap.Diff.GetAR())
		diff.SetMods(mods[i])

		player := &difficultyPlayer{cursor: cursor, diff: diff}
		diffPlayers = append(diffPlayers, player)

		grade := Grade(NONE)

		/*if mods[i]&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			grade = SSH
		}*/

		if ruleset.oppDiffs[int(mods[i])] == nil {
			diffs := make([]oppai.DiffCalc, 0)

			for _, m := range ruleset.oppaiMaps {
				diffs = append(diffs, (&oppai.DiffCalc{Beatmap: *m}).Calc(int(mods[i]), oppai.DefaultSingletapThreshold))
			}

			ruleset.oppDiffs[int(mods[i])] = diffs
		}

		ruleset.cursors[cursor] = &subSet{player, 0, 100, 0, 0, 0, mods[i].GetScoreMultiplier(), 0, grade, nil, &oppai.PPv2{}, make(map[HitResult]int64)}
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

			if isDone := g.UpdatePost(time); isDone {
				if set.endlistener != nil {
					set.endlistener(time, g.GetNumber())
				}
				if i < len(set.processed)-1 {
					set.processed = append(set.processed[:i], set.processed[i+1:]...)
				} else if i < len(set.processed) {
					set.processed = set.processed[:i]
				}
				i--
			}
		}
	}

	if len(set.queue) == 0 && len(set.processed) == 0 && !set.ended {
		cs := make([]*render.Cursor, 0)
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
			data = append(data, humanize(set.cursors[c].hits[HitResults.Hit300]))
			data = append(data, humanize(set.cursors[c].hits[HitResults.Hit100]))
			data = append(data, humanize(set.cursors[c].hits[HitResults.Hit50]))
			data = append(data, humanize(set.cursors[c].hits[HitResults.Miss]))
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

func (set *OsuRuleSet) UpdateClickFor(cursor *render.Cursor, time int64) {
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

func (set *OsuRuleSet) UpdateNormalFor(cursor *render.Cursor, time int64) {
	player := set.cursors[cursor].player

	wasSliderAlready := false

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			s, isSlider := g.(*Slider)

			if isSlider {
				if wasSliderAlready {
					continue
				}

				if !s.IsHit(player) {
					wasSliderAlready = true
				}
			}

			g.UpdateFor(player, time)
		}
	}
}

func (set *OsuRuleSet) Draw(time int64, batch *batches.SpriteBatch, color mgl32.Vec4) {
	if len(set.processed) > 0 {
		for i := len(set.processed) - 1; i > 0; i-- {
			g := set.processed[i]
			g.Draw(time, color, batch)
		}
	}
}

func (set *OsuRuleSet) SendResult(time int64, cursor *render.Cursor, number int64, x, y float32, result HitResult, raw bool, comboResult ComboResult) {
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

	set.params.N300 = uint16(subSet.hits[HitResults.Hit300])
	set.params.N100 = uint16(subSet.hits[HitResults.Hit100])
	set.params.N50 = uint16(subSet.hits[HitResults.Hit50])
	set.params.Misses = uint16(subSet.hits[HitResults.Miss])
	set.params.Mods = uint32(subSet.player.diff.Mods)
	set.params.Combo = uint16(subSet.maxCombo)

	index := subSet.numObjects - 1
	if index < 0 {
		index = 0
	}

	diff := set.oppDiffs[int(subSet.player.diff.Mods)][index]

	subSet.ppv2.PPv2WithMods(diff.Aim, diff.Speed, set.oppaiMaps[index], int(subSet.player.diff.Mods), int(subSet.hits[HitResults.Hit300]), int(subSet.hits[HitResults.Hit100]), int(subSet.hits[HitResults.Hit50]), int(subSet.hits[HitResults.Miss]), int(subSet.maxCombo)) //oppai.PPInfo(set.oppaiMap, set.params).PP.Total

	if set.listener != nil {
		set.listener(cursor, time, number, bmath.NewVec2f(x, y).Copy64(), result, comboResult, subSet.ppv2.Total*1.00013679674 /** 1.00050243137 */ /** 1.00018787323*/ /** 1.02046696933*/ /**1.02730112005*/, subSet.score)
	}

	if len(set.cursors) == 1 {
		log.Println("Got:", fmt.Sprintf("%3s", result), "Combo:", fmt.Sprintf("%4d", subSet.combo), "Max Combo:", fmt.Sprintf("%4d", subSet.maxCombo), "Score:", fmt.Sprintf("%9d", subSet.score), "Acc:", fmt.Sprintf("%3.2f%%", 100*float64(subSet.rawScore)/float64(subSet.numObjects*300)), subSet.hits, "from:", number, "at:", time, "pos:", x, y)
	}
}

func (set *OsuRuleSet) CanBeHit(time int64, object hitobject, player *difficultyPlayer) ClickAction {
	if _, ok := object.(*Circle); ok {
		index := -1

		for i, g := range set.processed {
			if g == object {
				index = i
			}
		}

		if index > 0 && set.beatMap.HitObjects[set.processed[index-1].GetNumber()].GetBasicData().StackIndex > 0 && !set.processed[index-1].IsHit(player) {
			return Ignored //TODO: this should not shake the object
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

func (set *OsuRuleSet) SetListener(listener func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result HitResult, comboResult ComboResult, pp float64, score int64)) {
	set.listener = listener
}

func (set *OsuRuleSet) SetEndListener(endlistener func(time int64, number int64)) {
	set.endlistener = endlistener
}

func (set *OsuRuleSet) GetResults(cursor *render.Cursor) (float64, int64, int64, Grade) {
	subSet := set.cursors[cursor]
	return subSet.accuracy, subSet.maxCombo, subSet.score, subSet.grade
}

func (set *OsuRuleSet) GetBeatMap() *beatmap.BeatMap {
	return set.beatMap
}
