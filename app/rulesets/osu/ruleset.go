package osu

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu/performance"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"log"
	"math"
	"sort"
	"strings"
)

const Tolerance2B = 3

type ClickAction uint8

const (
	Ignored = ClickAction(iota)
	Shake
	Click
)

type buttonState struct {
	Left, Right bool
}

func (state buttonState) BothReleased() bool {
	return !(state.Left || state.Right)
}

type HitObject interface {
	Init(ruleset *OsuRuleSet, object objects.IHitObject, players []*difficultyPlayer)
	UpdateFor(player *difficultyPlayer, time int64, processSliderEndsAhead bool) bool
	UpdateClickFor(player *difficultyPlayer, time int64) bool
	UpdatePostFor(player *difficultyPlayer, time int64, processSliderEndsAhead bool) bool
	UpdatePost(time int64) bool
	MissForcefully(player *difficultyPlayer, time int64)
	IsHit(player *difficultyPlayer) bool
	GetFadeTime() int64
	GetNumber() int64
	GetObject() objects.IHitObject
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

	maskedModString string

	lzLegacyNotelock bool
	lzNoSliderAcc    bool
	lzLegacySound    bool
}

type subSet struct {
	player *difficultyPlayer

	score          *Score
	hp             IHealthProcessor
	scoreProcessor scoreProcessor

	currentKatu int
	currentBad  int

	numObjects uint

	ppv2 api.IPerformanceCalculator

	recoveries int
	failed     bool
	sdpfFail   bool
	forceFail  bool
}

type hitListener func(cursor *graphics.Cursor, judgementResult JudgementResult, score Score)

type endListener func(time int64, number int64)

type failListener func(cursor *graphics.Cursor)

type OsuRuleSet struct {
	beatMap *beatmap.BeatMap
	cursors map[*graphics.Cursor]*subSet

	ended bool

	oppDiffs map[string][]api.Attributes

	queue        []HitObject
	processed    []HitObject
	hitListener  hitListener
	endListener  endListener
	failListener failListener
}

func NewOsuRuleset(beatMap *beatmap.BeatMap, cursors []*graphics.Cursor, diffs []*difficulty.Difficulty) *OsuRuleSet {
	log.Println("Creating osu! ruleset...")

	ruleset := new(OsuRuleSet)
	ruleset.beatMap = beatMap
	ruleset.oppDiffs = make(map[string][]api.Attributes)

	log.Println("Using pp calc version", performance.GetDifficultyCalculator().GetVersionMessage())

	ruleset.cursors = make(map[*graphics.Cursor]*subSet)

	diffPlayers := make([]*difficultyPlayer, 0, len(cursors))

	for i, cursor := range cursors {
		diff := diffs[i]

		beatMap.CalculateStackLeniency(diff) // Calculate additional stack indexes for DA/EZ/HR/whatever that changes Preempt

		diff.Mods = diff.Mods | (beatMap.Diff.Mods & difficulty.ScoreV2) // if beatmap has ScoreV2 mod, force it for all players
		diff.Mods = diff.Mods | (beatMap.Diff.Mods & difficulty.Lazer)   // same for Lazer

		player := &difficultyPlayer{cursor: cursor, diff: diff, maskedModString: diff.GetModStringMasked()}
		diffPlayers = append(diffPlayers, player)

		lzLegacyHP := false

		if diff.CheckModActive(difficulty.Classic) {
			if s, ok := difficulty.GetModConfig[difficulty.ClassicSettings](diff); ok {
				player.lzLegacyNotelock = s.ClassicNoteLock
				player.lzNoSliderAcc = s.NoSliderHeadAccuracy
				player.lzLegacySound = s.AlwaysPlayTailSample
				lzLegacyHP = s.ClassicHealth
			}
		}

		if ruleset.oppDiffs[player.maskedModString] == nil {
			ruleset.oppDiffs[player.maskedModString] = performance.GetDifficultyCalculator().CalculateStep(ruleset.beatMap.HitObjects, player.diff)

			star := ruleset.oppDiffs[player.maskedModString][len(ruleset.oppDiffs[player.maskedModString])-1]

			log.Println("Stars:")
			log.Println("\tAim:  ", star.Aim)
			log.Println("\tSpeed:", star.Speed)

			if diff.CheckModActive(difficulty.Flashlight) {
				log.Println("\tFlash:", star.Flashlight)
			}

			log.Println("\tTotal:", star.Total)

			pp := performance.CreatePPCalculator()
			ppResults := pp.Calculate(star, api.PerfScore{CountGreat: -1, MaxCombo: -1, Accuracy: 1}, diff)

			log.Println("SS PP:")
			log.Println("\tAim:  ", ppResults.Aim)
			log.Println("\tTap:  ", ppResults.Speed)

			if diff.CheckModActive(difficulty.Flashlight) {
				log.Println("\tFlash:", star.Flashlight)
			}

			log.Println("\tAcc:  ", ppResults.Acc)
			log.Println("\tTotal:", ppResults.Total)
		}

		log.Println(fmt.Sprintf("Calculating HP rates for \"%s\"...", cursor.Name))

		var hp IHealthProcessor

		if diff.CheckModActive(difficulty.Lazer) && !lzLegacyHP {
			hp = NewHealthProcessorV2(beatMap, player)
		} else {
			hp = NewHealthProcessor(beatMap, player, !cursor.OldSpinnerScoring)
		}

		hp.CalculateRate()
		hp.ResetHp()

		log.Println("\tPassive drain rate:", hp.GetDrainRate()*1000)

		if hp2, ok := hp.(*HealthProcessor); ok {
			log.Println("\tNormal multiplier:", hp2.HpMultiplierNormal)
			log.Println("\tCombo end multiplier:", hp2.HpMultiplierComboEnd)
		}

		recoveries := 0
		if diff.CheckModActive(difficulty.Easy) {
			if conf, ok := difficulty.GetModConfig[difficulty.EasySettings](diff); ok {
				recoveries = conf.Retries
			} else {
				recoveries = 2
			}
		}

		hp.AddFailListener(func() {
			ruleset.failInternal(player)
		})

		var sc scoreProcessor

		if diff.CheckModActive(difficulty.Lazer) {
			sc = newScoreV3Processor()
		} else if diff.CheckModActive(difficulty.ScoreV2) {
			sc = newScoreV2Processor()
		} else {
			sc = newScoreV1Processor()
		}

		sc.Init(beatMap, player)

		ruleset.cursors[cursor] = &subSet{
			player: player,
			score: &Score{
				Accuracy: 1,
			},
			ppv2:           performance.CreatePPCalculator(),
			hp:             hp,
			recoveries:     recoveries,
			scoreProcessor: sc,
		}
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

	return ruleset
}

func (set *OsuRuleSet) Update(time int64) {
	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			if isDone := g.UpdatePost(time); isDone {
				if set.endListener != nil {
					set.endListener(time, g.GetNumber())
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
		set.printEndTable()

		set.ended = true
	}
}

func (set *OsuRuleSet) printEndTable() {
	cs := make([]*graphics.Cursor, 0)
	for c := range set.cursors {
		cs = append(cs, c)
	}

	sort.Slice(cs, func(i, j int) bool {
		return set.cursors[cs[i]].scoreProcessor.GetScore() > set.cursors[cs[j]].scoreProcessor.GetScore()
	})

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"#", "Player", "Score", "Accuracy", "Grade", "300", "100", "50", "Miss", "Combo", "Max Combo", "Mods", "PP"})

	for i, c := range cs {
		var data []string
		data = append(data, fmt.Sprintf("%d", i+1))
		data = append(data, c.Name)
		data = append(data, utils.Humanize(set.cursors[c].scoreProcessor.GetScore()))
		data = append(data, fmt.Sprintf("%.2f", set.cursors[c].score.Accuracy*100))
		data = append(data, set.cursors[c].score.Grade.String())
		data = append(data, utils.Humanize(set.cursors[c].score.Count300))
		data = append(data, utils.Humanize(set.cursors[c].score.Count100))
		data = append(data, utils.Humanize(set.cursors[c].score.Count50))
		data = append(data, utils.Humanize(set.cursors[c].score.CountMiss))
		data = append(data, utils.Humanize(set.cursors[c].scoreProcessor.GetCombo()))
		data = append(data, utils.Humanize(set.cursors[c].score.Combo))
		data = append(data, set.cursors[c].player.diff.GetModString())
		data = append(data, fmt.Sprintf("%.2f", set.cursors[c].score.PP.Total))
		table.Append(data)
	}

	table.Render()

	for _, s := range strings.Split(tableString.String(), "\n") {
		log.Println(s)
	}
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

	if len(set.processed) > 0 && !set.cursors[cursor].failed {
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

func (set *OsuRuleSet) UpdateNormalFor(cursor *graphics.Cursor, time int64, processSliderEndsAhead bool) {
	player := set.cursors[cursor].player

	wasSliderAlready := false

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			if !cursor.IsAutoplay && !cursor.IsPlayer {
				// TODO: recreate stable's hitobject "unloading" for replays

				s, isSlider := g.(*Slider)

				if isSlider {
					if wasSliderAlready {
						continue
					}

					if !s.IsHit(player) {
						wasSliderAlready = true
					}
				}
			}

			g.UpdateFor(player, time, processSliderEndsAhead)
		}
	}
}

func (set *OsuRuleSet) UpdatePostFor(cursor *graphics.Cursor, time int64, processSliderEndsAhead bool) {
	player := set.cursors[cursor].player

	if len(set.processed) > 0 {
		for i := 0; i < len(set.processed); i++ {
			g := set.processed[i]

			g.UpdatePostFor(player, time, processSliderEndsAhead)
		}
	}
}

func (set *OsuRuleSet) SendResult(cursor *graphics.Cursor, judgementResult JudgementResult) {
	subSet := set.cursors[cursor]

	if judgementResult.HitResult == Ignore || judgementResult.HitResult == PositionalMiss {
		if judgementResult.HitResult == PositionalMiss && set.hitListener != nil && !subSet.player.diff.Mods.Active(difficulty.Relax) {
			set.hitListener(cursor, judgementResult, *subSet.score)
		}

		return
	}

	if (subSet.player.diff.Mods.Active(difficulty.SuddenDeath|difficulty.Perfect) && judgementResult.ComboResult == Reset) ||
		(subSet.player.diff.Mods.Active(difficulty.Perfect) && (judgementResult.HitResult&BaseHitsM > 0 && judgementResult.HitResult&BaseHitsM != Hit300)) {
		if judgementResult.HitResult&BaseHitsM > 0 {
			judgementResult.HitResult = Miss
		} else if judgementResult.HitResult&(SliderHits) > 0 {
			judgementResult.HitResult = SliderMiss
		}

		judgementResult.ComboResult = Reset
		subSet.sdpfFail = true
	}

	judgementResult.HitResult = subSet.scoreProcessor.ModifyResult(judgementResult.HitResult, judgementResult.object)

	if judgementResult.ComboResult == Reset && judgementResult.HitResult != Miss { // skips missed slider "ends" as they don't reset combo
		subSet.score.CountSB++
	}

	bResult := judgementResult.HitResult & BaseHitsM

	if bResult > 0 {
		subSet.numObjects++
	}

	subSet.scoreProcessor.AddResult(judgementResult)
	subSet.score.AddResult(judgementResult)

	subSet.score.Score = subSet.scoreProcessor.GetScore()
	subSet.score.Combo = max(uint(subSet.scoreProcessor.GetCombo()), subSet.score.Combo)
	subSet.score.Accuracy = subSet.scoreProcessor.GetAccuracy()

	subSet.score.CalculateGrade(subSet.player.diff.Mods)

	index := max(1, subSet.numObjects) - 1

	diff := set.oppDiffs[subSet.player.maskedModString][index]

	subSet.score.PerfectCombo = uint(diff.MaxCombo) == subSet.score.Combo

	subSet.score.PP = subSet.ppv2.Calculate(diff, subSet.score.ToPerfScore(), subSet.player.diff)

	switch judgementResult.HitResult {
	case Hit100:
		subSet.currentKatu++
	case Hit50, Miss:
		subSet.currentBad++
	}

	if judgementResult.HitResult&BaseHitsM > 0 && (int(judgementResult.Number) == len(set.beatMap.HitObjects)-1 || (int(judgementResult.Number) < len(set.beatMap.HitObjects)-1 && set.beatMap.HitObjects[judgementResult.Number+1].IsNewCombo())) {
		allClicked := true

		// We don't want to give geki/katu if all objects in current combo weren't clicked
		index := sort.Search(len(set.processed), func(i int) bool {
			return set.processed[i].GetNumber() >= judgementResult.Number
		})

		for i := index - 1; i >= 0; i-- {
			obj := set.processed[i]

			if !obj.IsHit(subSet.player) {
				allClicked = false
				break
			}

			if set.beatMap.HitObjects[obj.GetNumber()].IsNewCombo() {
				break
			}
		}

		if judgementResult.HitResult&BaseHits > 0 {
			if subSet.currentKatu == 0 && subSet.currentBad == 0 && allClicked {
				judgementResult.HitResult |= GekiAddition
				subSet.score.CountGeki++
			} else if subSet.currentBad == 0 && allClicked {
				judgementResult.HitResult |= KatuAddition
				subSet.score.CountKatu++
			} else {
				judgementResult.HitResult |= MuAddition
			}
		}

		subSet.currentBad = 0
		subSet.currentKatu = 0
	}

	if subSet.sdpfFail {
		subSet.hp.Increase(-100000, true)
	} else {
		subSet.hp.AddResult(judgementResult)
	}

	if set.hitListener != nil {
		set.hitListener(cursor, judgementResult, *subSet.score)
	}

	if len(set.cursors) == 1 && judgementResult.HitResult != SliderFinish && !settings.RECORD {
		log.Println(fmt.Sprintf(
			"Got: %3d, Combo: %4d, Max Combo: %4d, Score: %9d, Acc: %6.2f%%, 300: %4d, 100: %3d, 50: %2d, miss: %2d, from: %d, at: %d, pos: %.0fx%.0f, pp: %.2f",
			judgementResult.HitResult.ScoreValueMod(subSet.player.diff.Mods),
			subSet.scoreProcessor.GetCombo(),
			subSet.score.Combo,
			subSet.scoreProcessor.GetScore(),
			subSet.score.Accuracy*100,
			subSet.score.Count300,
			subSet.score.Count100,
			subSet.score.Count50,
			subSet.score.CountMiss,
			judgementResult.Number,
			judgementResult.Time,
			judgementResult.Position.X,
			judgementResult.Position.Y,
			subSet.score.PP.Total,
		))
	}
}

func (set *OsuRuleSet) CanBeHit(time int64, object HitObject, player *difficultyPlayer) ClickAction {
	var clickAction ClickAction

	if player.cursor.IsAutoplay || (player.diff.CheckModActive(difficulty.Lazer) && !player.lzLegacyNotelock) {
		clickAction = set.CanBeHitLazer(time, object, player)
	} else {
		clickAction = set.CanBeHitStable(time, object, player)
	}

	if clickAction != Click {
		return clickAction
	}

	hitRange := difficulty.HittableRange
	if player.diff.CheckModActive(difficulty.Relax2) {
		hitRange -= 200
	}

	if math.Abs(float64(time-int64(object.GetObject().GetStartTime()))) >= hitRange {
		return Shake
	}

	return Click
}

func (set *OsuRuleSet) CanBeHitStable(time int64, object HitObject, player *difficultyPlayer) ClickAction {
	if _, ok := object.(*Circle); ok {
		index := -1

		for i, g := range set.processed {
			if g == object {
				index = i
			}
		}

		if index > 0 && set.processed[index-1].GetObject().GetStackIndexMod(player.diff) > 0 && !set.processed[index-1].IsHit(player) {
			return Ignored //don't shake the stacks
		}
	}

	for _, g := range set.processed {
		if !g.IsHit(player) {
			if g.GetNumber() != object.GetNumber() {
				if g.GetObject().GetEndTime()+Tolerance2B < object.GetObject().GetStartTime() {
					return Shake
				}
			} else {
				break
			}
		}
	}

	return Click
}

func (set *OsuRuleSet) CanBeHitLazer(time int64, object HitObject, player *difficultyPlayer) ClickAction {
	var lastObj HitObject
	var hitCheck func(player *difficultyPlayer) bool

	for _, g := range set.processed {
		if g.GetObject().GetStartTime() >= object.GetObject().GetStartTime() {
			break
		}

		if (g.GetObject().GetType() & (objects.CIRCLE | objects.SLIDER)) > 0 {
			if c, ok1 := g.(*Circle); ok1 {
				hitCheck = c.IsHit
			} else if s, ok2 := g.(*Slider); ok2 {
				hitCheck = s.IsStartHit
			}

			lastObj = g
		}
	}

	if !(lastObj == nil || hitCheck(player) || float64(time) >= lastObj.GetObject().GetStartTime()) {
		return Shake
	}

	return Click
}

func (set *OsuRuleSet) GetResultForDelta(player *difficultyPlayer, delta float64) HitResult {
	if player.diff.CheckModActive(difficulty.Lazer) {
		if delta <= player.diff.Hit300U {
			return Hit300
		} else if delta <= player.diff.Hit100U {
			return Hit100
		} else if delta <= player.diff.Hit50U {
			return Hit50
		}
	} else {
		if int64(delta) < player.diff.Hit300 {
			return Hit300
		} else if int64(delta) < player.diff.Hit100 {
			return Hit100
		} else if int64(delta) < player.diff.Hit50 {
			return Hit50
		}
	}

	return Miss
}

func (set *OsuRuleSet) PostHit(time int64, object HitObject, player *difficultyPlayer) {
	if (!player.cursor.IsAutoplay && !player.diff.CheckModActive(difficulty.Lazer)) || (object.GetObject().GetType()&(objects.CIRCLE|objects.SLIDER)) == 0 {
		return
	}

	for _, g := range set.processed {
		if g.GetObject().GetStartTime() >= object.GetObject().GetStartTime() {
			break
		}

		if (g.GetObject().GetType() & (objects.CIRCLE | objects.SLIDER)) > 0 {
			var hitCheck func(player *difficultyPlayer) bool

			if c, ok1 := g.(*Circle); ok1 {
				hitCheck = c.IsHit
			} else if s, ok2 := g.(*Slider); ok2 {
				hitCheck = s.IsStartHit
			}

			if !hitCheck(player) {
				g.MissForcefully(player, time)
			}
		}
	}
}

func (set *OsuRuleSet) failInternal(player *difficultyPlayer) {
	subSet := set.cursors[player.cursor]

	if player.cursor.IsReplay && settings.Gameplay.IgnoreFailsInReplays {
		return
	}

	if !subSet.forceFail && player.diff.CheckModActive(difficulty.NoFail|difficulty.Relax|difficulty.Relax2) {
		return
	}

	// EZ mod gives 2 additional lives
	if subSet.recoveries > 0 && !subSet.sdpfFail && !subSet.forceFail {
		subSet.hp.IncreaseRelative(0.8, false)
		subSet.recoveries--

		return
	}

	// actual fail
	if set.failListener != nil && !subSet.failed {
		set.failListener(player.cursor)
	}

	subSet.failed = true
}

func (set *OsuRuleSet) PlayerStopped(cursor *graphics.Cursor, time int64) {
	subSet := set.cursors[cursor]

	// Let's believe in hp system. 1ms just in case for slider calculation inconsistencies
	if time < int64(set.beatMap.HitObjects[len(set.beatMap.HitObjects)-1].GetEndTime())-1 /*+subSet.player.diff.Hit50+20*/ {
		subSet.forceFail = true
		subSet.hp.Increase(-10000, true)
	}
}

func (set *OsuRuleSet) SetListener(listener hitListener) {
	set.hitListener = listener
}

func (set *OsuRuleSet) SetEndListener(listener endListener) {
	set.endListener = listener
}

func (set *OsuRuleSet) SetFailListener(listener failListener) {
	set.failListener = listener
}

func (set *OsuRuleSet) GetScore(cursor *graphics.Cursor) Score {
	return *(set.cursors[cursor].score)
}

func (set *OsuRuleSet) GetHP(cursor *graphics.Cursor) float64 {
	subSet := set.cursors[cursor]
	return subSet.hp.GetHealth()
}

func (set *OsuRuleSet) GetPlayer(cursor *graphics.Cursor) *difficultyPlayer {
	subSet := set.cursors[cursor]
	return subSet.player
}

func (set *OsuRuleSet) GetPlayerDifficulty(cursor *graphics.Cursor) *difficulty.Difficulty {
	subSet := set.cursors[cursor]
	return subSet.player.diff
}

func (set *OsuRuleSet) GetProcessed() []HitObject {
	return set.processed
}

func (set *OsuRuleSet) GetBeatMap() *beatmap.BeatMap {
	return set.beatMap
}
