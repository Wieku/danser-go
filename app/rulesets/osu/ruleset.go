package osu

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu/performance"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
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

type ComboResult uint8

const (
	Reset = ComboResult(iota)
	Hold
	Increase
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

type scoreProcessor interface {
	Init(beatMap *beatmap.BeatMap, player *difficultyPlayer)
	AddResult(result HitResult, comboResult ComboResult)
	ModifyResult(result HitResult, src HitObject) HitResult
	GetScore() int64
	GetCombo() int64
}

type Score struct {
	Score        int64
	Accuracy     float64
	Grade        Grade
	Combo        uint
	PerfectCombo bool
	Count300     uint
	CountGeki    uint
	Count100     uint
	CountKatu    uint
	Count50      uint
	CountMiss    uint
	CountSB      uint
	PP           performance.PPv2Results
}

type subSet struct {
	player *difficultyPlayer

	score          *Score
	hp             *HealthProcessor
	scoreProcessor scoreProcessor
	rawScore       int64
	currentKatu    int
	currentBad     int

	numObjects uint

	ppv2 *performance.PPv2

	recoveries int
	failed     bool
	sdpfFail   bool
}

type hitListener func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result HitResult, comboResult ComboResult, ppResults performance.PPv2Results, score int64)

type endListener func(time int64, number int64)

type failListener func(cursor *graphics.Cursor)

type OsuRuleSet struct {
	beatMap *beatmap.BeatMap
	cursors map[*graphics.Cursor]*subSet

	ended bool

	oppDiffs map[difficulty.Modifier][]performance.Attributes

	queue        []HitObject
	processed    []HitObject
	hitListener  hitListener
	endListener  endListener
	failListener failListener

	experimentalPP bool
}

func NewOsuRuleset(beatMap *beatmap.BeatMap, cursors []*graphics.Cursor, mods []difficulty.Modifier) *OsuRuleSet {
	log.Println("Creating osu! ruleset...")

	ruleset := new(OsuRuleSet)
	ruleset.beatMap = beatMap
	ruleset.oppDiffs = make(map[difficulty.Modifier][]performance.Attributes)

	if settings.Gameplay.UseLazerPP {
		log.Println("Using pp calc version 2022-01-23:")
		log.Println("\tRemove decay factor in Flashlight skill: https://github.com/ppy/osu/pull/15728")
		log.Println("\tMake speed skill consider only the shortest movement distance: https://github.com/ppy/osu/pull/15758")
		log.Println("\tFix cumulative strain time calculation in Flashlight skill: https://github.com/ppy/osu/pull/15867")
		log.Println("\tRemove combo scaling from Aim and Speed from osu! performance calculation: https://github.com/ppy/osu/pull/16280")
		log.Println("\tDon't floor effectiveMissCount: https://github.com/ppy/osu/pull/16331")
		ruleset.experimentalPP = true
	} else {
		log.Println("Using pp calc version 2021-11-09 with hotfix: https://osu.ppy.sh/home/news/2021-11-09-performance-points-star-rating-updates")
	}

	ruleset.cursors = make(map[*graphics.Cursor]*subSet)

	diffPlayers := make([]*difficultyPlayer, 0, len(cursors))

	for i, cursor := range cursors {
		diff := difficulty.NewDifficulty(beatMap.Diff.GetBaseHP(), beatMap.Diff.GetBaseCS(), beatMap.Diff.GetBaseOD(), beatMap.Diff.GetBaseAR())

		diff.SetHPCustom(beatMap.Diff.GetHP())
		diff.SetCSCustom(beatMap.Diff.GetCS())
		diff.SetODCustom(beatMap.Diff.GetOD())
		diff.SetARCustom(beatMap.Diff.GetAR())

		diff.SetMods(mods[i] | (beatMap.Diff.Mods & difficulty.ScoreV2)) // if beatmap has ScoreV2 mod, force it for all players
		diff.SetCustomSpeed(beatMap.Diff.CustomSpeed)

		player := &difficultyPlayer{cursor: cursor, diff: diff}
		diffPlayers = append(diffPlayers, player)

		if ruleset.oppDiffs[mods[i]&difficulty.DifficultyAdjustMask] == nil {
			ruleset.oppDiffs[mods[i]&difficulty.DifficultyAdjustMask] = performance.CalculateStep(ruleset.beatMap.HitObjects, diff, ruleset.experimentalPP)

			star := ruleset.oppDiffs[mods[i]&difficulty.DifficultyAdjustMask][len(ruleset.oppDiffs[mods[i]&difficulty.DifficultyAdjustMask])-1]

			log.Println("Stars:")
			log.Println("\tAim:  ", star.Aim)
			log.Println("\tSpeed:", star.Speed)

			if ruleset.experimentalPP && mods[i].Active(difficulty.Flashlight) {
				log.Println("\tFlash:", star.Flashlight)
			}

			log.Println("\tTotal:", star.Total)

			pp := &performance.PPv2{}
			pp.PPv2x(star, -1, -1, 0, 0, 0, diff, false)

			log.Println("SS PP:")
			log.Println("\tAim:  ", pp.Results.Aim)
			log.Println("\tTap:  ", pp.Results.Speed)

			if ruleset.experimentalPP && mods[i].Active(difficulty.Flashlight) {
				log.Println("\tFlash:", star.Flashlight)
			}

			log.Println("\tAcc:  ", pp.Results.Acc)
			log.Println("\tTotal:", pp.Results.Total)
		}

		log.Println(fmt.Sprintf("Calculating HP rates for \"%s\"...", cursor.Name))

		hp := NewHealthProcessor(beatMap, diff, !cursor.OldSpinnerScoring)
		hp.CalculateRate()
		hp.ResetHp()

		log.Println("\tPassive drain rate:", hp.PassiveDrain/2*1000)
		log.Println("\tNormal multiplier:", hp.HpMultiplierNormal)
		log.Println("\tCombo end multiplier:", hp.HpMultiplierComboEnd)

		recoveries := 0
		if diff.CheckModActive(difficulty.Easy) {
			recoveries = 2
		}

		hp.AddFailListener(func() {
			ruleset.failInternal(player)
		})

		var sc scoreProcessor

		if diff.CheckModActive(difficulty.ScoreV2) {
			sc = newScoreV2Processor()
		} else {
			sc = newScoreV1Processor()
		}

		sc.Init(beatMap, player)

		ruleset.cursors[cursor] = &subSet{
			player: player,
			score: &Score{
				Accuracy: 100,
			},
			ppv2:           &performance.PPv2{},
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
			data = append(data, fmt.Sprintf("%.2f", set.cursors[c].score.Accuracy))
			data = append(data, set.cursors[c].score.Grade.String())
			data = append(data, utils.Humanize(set.cursors[c].score.Count300))
			data = append(data, utils.Humanize(set.cursors[c].score.Count100))
			data = append(data, utils.Humanize(set.cursors[c].score.Count50))
			data = append(data, utils.Humanize(set.cursors[c].score.CountMiss))
			data = append(data, utils.Humanize(set.cursors[c].scoreProcessor.GetCombo()))
			data = append(data, utils.Humanize(set.cursors[c].score.Combo))
			data = append(data, set.cursors[c].player.diff.GetModString())
			data = append(data, fmt.Sprintf("%.2f", set.cursors[c].ppv2.Results.Total))
			table.Append(data)
		}

		table.Render()

		for _, s := range strings.Split(tableString.String(), "\n") {
			log.Println(s)
		}

		set.ended = true
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

func (set *OsuRuleSet) UpdateNormalFor(cursor *graphics.Cursor, time int64, processSliderEndsAhead bool) {
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

func (set *OsuRuleSet) SendResult(time int64, cursor *graphics.Cursor, src HitObject, x, y float32, result HitResult, comboResult ComboResult) {
	number := src.GetNumber()
	subSet := set.cursors[cursor]

	if result == Ignore || result == PositionalMiss {
		if result == PositionalMiss && set.hitListener != nil && !subSet.player.diff.Mods.Active(difficulty.Relax) {
			set.hitListener(cursor, time, number, vector.NewVec2f(x, y).Copy64(), result, comboResult, subSet.ppv2.Results, subSet.scoreProcessor.GetScore())
		}

		return
	}

	if (subSet.player.diff.Mods.Active(difficulty.SuddenDeath|difficulty.Perfect) && comboResult == Reset) ||
		(subSet.player.diff.Mods.Active(difficulty.Perfect) && (result&BaseHitsM > 0 && result&BaseHitsM != Hit300)) {
		if result&BaseHitsM > 0 {
			result = Miss
		} else if result&(SliderHits) > 0 {
			result = SliderMiss
		}

		comboResult = Reset
		subSet.sdpfFail = true
	}

	result = subSet.scoreProcessor.ModifyResult(result, src)
	subSet.scoreProcessor.AddResult(result, comboResult)

	subSet.score.Score = subSet.scoreProcessor.GetScore()

	if comboResult == Reset && result != Miss {
		subSet.score.CountSB++
	}

	bResult := result & BaseHitsM

	if bResult > 0 {
		subSet.rawScore += result.ScoreValue()

		switch bResult {
		case Hit300:
			subSet.score.Count300++
		case Hit100:
			subSet.score.Count100++
		case Hit50:
			subSet.score.Count50++
		case Miss:
			subSet.score.CountMiss++
		}

		subSet.numObjects++
	}

	subSet.score.Combo = mutils.Max(uint(subSet.scoreProcessor.GetCombo()), subSet.score.Combo)

	if subSet.numObjects == 0 {
		subSet.score.Accuracy = 100
	} else {
		subSet.score.Accuracy = 100 * float64(subSet.rawScore) / float64(subSet.numObjects*300)
	}

	ratio := float64(subSet.score.Count300) / float64(subSet.numObjects)

	if subSet.score.Count300 == subSet.numObjects {
		if subSet.player.diff.Mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			subSet.score.Grade = SSH
		} else {
			subSet.score.Grade = SS
		}
	} else if ratio > 0.9 && float64(subSet.score.Count50)/float64(subSet.numObjects) < 0.01 && subSet.score.CountMiss == 0 {
		if subSet.player.diff.Mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
			subSet.score.Grade = SH
		} else {
			subSet.score.Grade = S
		}
	} else if ratio > 0.8 && subSet.score.CountMiss == 0 || ratio > 0.9 {
		subSet.score.Grade = A
	} else if ratio > 0.7 && subSet.score.CountMiss == 0 || ratio > 0.8 {
		subSet.score.Grade = B
	} else if ratio > 0.6 {
		subSet.score.Grade = C
	} else {
		subSet.score.Grade = D
	}

	index := mutils.Max(1, subSet.numObjects) - 1

	diff := set.oppDiffs[subSet.player.diff.Mods&difficulty.DifficultyAdjustMask][index]

	subSet.score.PerfectCombo = uint(diff.MaxCombo) == subSet.score.Combo

	subSet.ppv2.PPv2x(diff, int(subSet.score.Combo), int(subSet.score.Count300), int(subSet.score.Count100), int(subSet.score.Count50), int(subSet.score.CountMiss), subSet.player.diff, set.experimentalPP)

	subSet.score.PP = subSet.ppv2.Results

	switch result {
	case Hit100:
		subSet.currentKatu++
	case Hit50, Miss:
		subSet.currentBad++
	}

	if result&BaseHitsM > 0 && (int(number) == len(set.beatMap.HitObjects)-1 || (int(number) < len(set.beatMap.HitObjects)-1 && set.beatMap.HitObjects[number+1].IsNewCombo())) {
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

			if set.beatMap.HitObjects[obj.GetNumber()].IsNewCombo() {
				break
			}
		}

		if result&BaseHits > 0 {
			if subSet.currentKatu == 0 && subSet.currentBad == 0 && allClicked {
				result |= GekiAddition
				subSet.score.CountGeki++
			} else if subSet.currentBad == 0 && allClicked {
				result |= KatuAddition
				subSet.score.CountKatu++
			} else {
				result |= MuAddition
			}
		}

		subSet.currentBad = 0
		subSet.currentKatu = 0
	}

	if subSet.sdpfFail {
		subSet.hp.Increase(-100000, true)
	} else {
		subSet.hp.AddResult(result)
	}

	if set.hitListener != nil {
		set.hitListener(cursor, time, number, vector.NewVec2f(x, y).Copy64(), result, comboResult, subSet.ppv2.Results, subSet.scoreProcessor.GetScore())
	}

	if len(set.cursors) == 1 && !settings.RECORD {
		log.Println(fmt.Sprintf(
			"Got: %3d, Combo: %4d, Max Combo: %4d, Score: %9d, Acc: %6.2f%%, 300: %4d, 100: %3d, 50: %2d, miss: %2d, from: %d, at: %d, pos: %.0fx%.0f, pp: %.2f",
			result.ScoreValue(),
			subSet.scoreProcessor.GetCombo(),
			subSet.score.Combo,
			subSet.scoreProcessor.GetScore(),
			subSet.score.Accuracy,
			subSet.score.Count300,
			subSet.score.Count100,
			subSet.score.Count50,
			subSet.score.CountMiss,
			number,
			time,
			x,
			y,
			subSet.ppv2.Results.Total,
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

		if index > 0 && set.beatMap.HitObjects[set.processed[index-1].GetNumber()].GetStackIndex(player.diff.Mods) > 0 && !set.processed[index-1].IsHit(player) {
			return Ignored //don't shake the stacks
		}
	}

	for _, g := range set.processed {
		if !g.IsHit(player) {
			if g.GetNumber() != object.GetNumber() {
				if set.beatMap.HitObjects[g.GetNumber()].GetEndTime()+Tolerance2B < set.beatMap.HitObjects[object.GetNumber()].GetStartTime() {
					return Shake
				}
			} else {
				break
			}
		}
	}

	hitRange := difficulty.HittableRange
	if player.diff.CheckModActive(difficulty.Relax2) {
		hitRange -= 200
	}

	if math.Abs(float64(time-int64(set.beatMap.HitObjects[object.GetNumber()].GetStartTime()))) >= hitRange {
		return Shake
	}

	return Click
}

func (set *OsuRuleSet) failInternal(player *difficultyPlayer) {
	subSet := set.cursors[player.cursor]

	if player.diff.CheckModActive(difficulty.NoFail | difficulty.Relax | difficulty.Relax2) {
		return
	}

	// EZ mod gives 2 additional lives
	if subSet.recoveries > 0 && !subSet.sdpfFail {
		subSet.hp.Increase(160, false)
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

	if time < int64(set.beatMap.HitObjects[len(set.beatMap.HitObjects)-1].GetEndTime())+subSet.player.diff.Hit50+20 {
		subSet.sdpfFail = true
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
