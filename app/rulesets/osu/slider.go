package osu

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type Buttons int64

const Left = Buttons(1)
const Right = Buttons(2)

type sliderstate struct {
	downButton  Buttons
	isStartHit  bool
	isHit       bool
	points      []tickpoint
	scored      int
	missed      int
	slideStart  int64
	sliding     bool
	startResult HitResult
	endScored   bool
}

type tickpoint struct {
	time       int64
	scoreGiven HitResult
	edgeNum    int

	judged bool
}

type Slider struct {
	ruleSet           *OsuRuleSet
	hitSlider         *objects.Slider
	players           []*difficultyPlayer
	state             map[*difficultyPlayer]*sliderstate
	fadeStartRelative float64

	lastSliderTime int64
	sliderPosition vector.Vector2f
}

func (slider *Slider) GetNumber() int64 {
	return slider.hitSlider.GetID()
}

func (slider *Slider) IsSliding(player *difficultyPlayer) bool {
	return slider.state[player].sliding
}

func (slider *Slider) Init(ruleSet *OsuRuleSet, object objects.IHitObject, players []*difficultyPlayer) {
	slider.ruleSet = ruleSet
	slider.hitSlider = object.(*objects.Slider)
	slider.players = players
	slider.state = make(map[*difficultyPlayer]*sliderstate)

	rSlider := object.(*objects.Slider)

	slider.lastSliderTime = math.MinInt64
	slider.fadeStartRelative = 100000

	for _, player := range slider.players {
		slider.fadeStartRelative = min(slider.fadeStartRelative, player.diff.Preempt)
		slider.state[player] = new(sliderstate)
		slider.state[player].startResult = Miss

		edgeNumber := 1

		if player.diff.CheckModActive(difficulty.Lazer) {
			for _, point := range rSlider.ScorePointsLazer {
				if point.IsReverse || point.LastPoint {
					scoreGiven := SliderRepeat
					time := point.Time

					if point.LastPoint {
						scoreGiven = SliderEnd

						if player.lzNoSliderAcc {
							scoreGiven = LegacySliderEnd
						}

						time = rSlider.GetEndTime()
					}

					slider.state[player].points = append(slider.state[player].points, tickpoint{int64(time), scoreGiven, edgeNumber, false})
					edgeNumber++
				} else {
					slider.state[player].points = append(slider.state[player].points, tickpoint{int64(min(point.Time, rSlider.GetEndTime())), SliderPoint, -1, false})
				}
			}
			// TODO: min(point.Time, rSlider.GetEndTime()) is a hack, should be revisited later
		} else {
			for _, point := range rSlider.ScorePoints {
				if point.IsReverse {
					slider.state[player].points = append(slider.state[player].points, tickpoint{int64(point.Time), SliderRepeat, edgeNumber, false})
					edgeNumber++
				} else {
					slider.state[player].points = append(slider.state[player].points, tickpoint{int64(point.Time), SliderPoint, -1, false})
				}
			}

			if len(slider.state[player].points) > 0 {
				slider.state[player].points[len(slider.state[player].points)-1].time = max(int64(slider.hitSlider.GetStartTime())+int64(slider.hitSlider.GetEndTime()-slider.hitSlider.GetStartTime())/2, int64(slider.hitSlider.GetEndTime())-36) //slider ends 36ms before the real end for scoring
				slider.state[player].points[len(slider.state[player].points)-1].scoreGiven = SliderEnd
			}
		}

	}
}

func (slider *Slider) UpdateClickFor(player *difficultyPlayer, time int64) bool {
	state := slider.state[player]

	position := slider.hitSlider.GetStackedStartPositionMod(player.diff)

	clicked := player.leftCondE || player.rightCondE

	radius := float32(player.diff.CircleRadius)
	if player.diff.CheckModActive(difficulty.Relax2) {
		radius = 100
	}

	inRadius := player.cursor.RawPosition.Dst(position) <= radius

	if clicked && !state.isStartHit && (!state.isHit || player.diff.CheckModActive(difficulty.Lazer)) {
		action := slider.ruleSet.CanBeHit(time, slider, player)

		if inRadius {
			if action == Click {
				if player.leftCondE {
					player.leftCondE = false
				} else if player.rightCondE {
					player.rightCondE = false
				}

				if player.leftCond {
					state.downButton = Left
				} else if player.rightCond {
					state.downButton = Right
				} else {
					state.downButton = player.mouseDownButton
				}

				hit := SliderMiss
				combo := Reset

				state.startResult = slider.ruleSet.GetResultForDelta(player, math.Abs(float64(time)-slider.hitSlider.GetStartTime()))

				if state.startResult != Miss {
					hit = SliderStart
					combo = Increase
				}

				if hit != Ignore {
					if len(slider.players) == 1 {
						slider.hitSlider.HitEdge(0, float64(time), hit != SliderMiss)
					}

					state.isStartHit = true

					slider.ruleSet.PostHit(time, slider, player)

					if player.diff.CheckModActive(difficulty.Lazer) && !player.lzNoSliderAcc {
						slider.ruleSet.SendResult(player.cursor, createJudgementResult(state.startResult, Hit300, combo, time, position, slider))
					} else {
						slider.ruleSet.SendResult(player.cursor, createJudgementResult(hit, SliderStart, combo, time, position, slider))
					}

					if state.startResult != Miss && player.diff.CheckModActive(difficulty.Lazer) {
						slider.lazerPostHeadProcess(player, state, time)
					}
				}
			} else {
				player.leftCondE = false
				player.rightCondE = false
			}
		} else if action == Click {
			slider.ruleSet.SendResult(player.cursor, createJudgementResult(PositionalMiss, SliderStart, Hold, time, position, slider))
		}
	}

	return state.isStartHit
}

func (slider *Slider) lazerPostHeadProcess(player *difficultyPlayer, state *sliderstate, time int64) {
	sliderPosition := slider.hitSlider.GetStackedPositionAtMod(float64(time), player.diff)

	followRadiusFull := player.diff.CircleRadius * 2.4

	if player.cursor.RawPosition.Dst(sliderPosition) > float32(followRadiusFull) {
		return
	}

	allTicksInRange := true

	for _, point := range state.points {
		if point.time > time {
			break
		}

		currPos := slider.hitSlider.GetStackedPositionAtMod(float64(point.time), player.diff)

		if player.cursor.RawPosition.Dst(currPos) > float32(followRadiusFull) {
			allTicksInRange = false
			break
		}
	}

	slider.processTicksLazer(player, state, time, allTicksInRange, sliderPosition)

	if allTicksInRange || player.cursor.RawPosition.Dst(sliderPosition) <= float32(player.diff.CircleRadius) {
		state.sliding = true
		state.slideStart = time

		if len(slider.players) == 1 {
			slider.hitSlider.InitSlide(float64(time))
		}
	}
}

func (slider *Slider) UpdateFor(player *difficultyPlayer, time int64, processSliderEndsAhead bool) bool {
	lzMod := player.diff.CheckModActive(difficulty.Lazer)

	if lzMod {
		slider.processHeadMiss(player, time)
	}

	state := slider.state[player]

	if time != slider.lastSliderTime {
		slider.sliderPosition = slider.hitSlider.GetPositionAt(float64(time))
		slider.lastSliderTime = time
	}

	sliderPosition := objects.ModifyPosition(slider.hitSlider.HitObject, slider.sliderPosition, player.diff) // Calculate stacked position

	if time >= int64(slider.hitSlider.GetStartTime()) && ((!state.isHit && !lzMod) || (lzMod && state.isStartHit)) {
		mouseDownAcceptable := false
		mouseDownAcceptableSwap := player.gameDownState &&
			!(player.lastButton == (Left|Right) &&
				player.lastButton2 == player.mouseDownButton)

		if player.gameDownState {
			if state.downButton == Buttons(0) || (player.mouseDownButton != (Left|Right) && mouseDownAcceptableSwap) {
				state.downButton = Buttons(0)
				if player.leftCond {
					state.downButton = Left
				} else if player.rightCond {
					state.downButton = Right
				} else {
					state.downButton = player.mouseDownButton
				}

				mouseDownAcceptable = true
			} else if (player.mouseDownButton & state.downButton) > 0 {
				mouseDownAcceptable = true
			}
		} else {
			state.downButton = Buttons(0)
		}

		mouseDownAcceptable = mouseDownAcceptable || mouseDownAcceptableSwap || player.diff.CheckModActive(difficulty.Relax)

		radiusNeeded := player.diff.CircleRadius
		if state.sliding {
			radiusNeeded *= 2.4
		}

		allowable := mouseDownAcceptable && player.cursor.RawPosition.Dst(sliderPosition) <= float32(radiusNeeded)

		if allowable && !state.sliding {
			state.sliding = true
			state.slideStart = time

			if len(slider.players) == 1 {
				slider.hitSlider.InitSlide(float64(time))
			}
		}

		if player.diff.CheckModActive(difficulty.Lazer) {
			slider.processTicksLazer(player, state, time, allowable, sliderPosition)
		} else {
			slider.processTicksStable(player, state, time, allowable, sliderPosition, processSliderEndsAhead)
		}

		if !allowable && state.sliding && state.scored+state.missed < len(state.points) {
			if len(slider.players) == 1 {
				slider.hitSlider.KillSlide(float64(time))
			}

			state.sliding = false
		}
	}

	return true
}

func (slider *Slider) processTicksStable(player *difficultyPlayer, state *sliderstate, time int64, allowable bool, sliderPosition vector.Vector2f, processSliderEndsAhead bool) {
	pointsPassed := 0

	for i, point := range state.points {
		if point.time > time && !(i == len(state.points)-1 && processSliderEndsAhead && point.time-time == 1) {
			break
		}

		pointsPassed++
	}

	if state.scored+state.missed < pointsPassed {
		index := state.scored + state.missed
		point := state.points[index]

		maxScore := SliderPoint
		if pointsPassed == len(state.points) {
			maxScore = SliderEnd
		} else if pointsPassed%(len(state.points)/len(slider.hitSlider.TickReverse)) == 0 {
			maxScore = SliderRepeat
		}

		scoreGiven := SliderMiss
		combo := Reset

		if allowable && state.slideStart <= point.time {
			state.scored++

			scoreGiven = maxScore
			combo = Increase
		} else {
			state.missed++

			if state.scored+state.missed == len(state.points) {
				combo = Hold
			}
		}

		slider.ruleSet.SendResult(player.cursor, createJudgementResult(scoreGiven, maxScore, combo, time, sliderPosition, slider))
	}
}

func (slider *Slider) processTicksLazer(player *difficultyPlayer, state *sliderstate, time int64, allowable bool, sliderPosition vector.Vector2f) {
	if !state.isStartHit {
		return
	}

	pointsPassed := 0

	for _, point := range state.points {
		pTime := point.time
		if point.scoreGiven&(SliderEnd|LegacySliderEnd) > 0 {
			pTime -= 36
		}

		if pTime > time {
			break
		}

		pointsPassed++
	}

	for index := state.scored + state.missed; index < pointsPassed; index++ {
		point := state.points[index]

		scoreGiven := Ignore
		combo := Reset

		if allowable {
			state.scored++

			scoreGiven = point.scoreGiven
			combo = Increase

			if point.scoreGiven&(SliderEnd|LegacySliderEnd) > 0 {
				state.endScored = true

				if player.lzNoSliderAcc {
					combo = Hold
				}
			}
		} else if time >= point.time {
			state.missed++

			scoreGiven = SliderMiss

			if point.scoreGiven&(SliderEnd|LegacySliderEnd) > 0 {
				combo = Hold
			}
		}

		if scoreGiven != Ignore {
			slider.ruleSet.SendResult(player.cursor, createJudgementResult(scoreGiven, point.scoreGiven, combo, time, sliderPosition, slider))
		}
	}
}

func (slider *Slider) UpdatePostFor(player *difficultyPlayer, time int64, processSliderEndsAhead bool) bool {
	state := slider.state[player]

	if !player.diff.CheckModActive(difficulty.Lazer) {
		slider.processHeadMiss(player, time)
	}

	if (time >= int64(slider.hitSlider.GetEndTime()) || (processSliderEndsAhead && int64(slider.hitSlider.GetEndTime())-time == 1)) && !state.isHit {
		if len(slider.players) == 1 && !state.isStartHit && !player.diff.CheckModActive(difficulty.Lazer) {
			slider.hitSlider.ArmStart(false, float64(time))
		}

		if state.startResult != Miss {
			state.scored++
		}

		rate := float64(state.scored) / float64(len(state.points)+1)

		if len(slider.players) == 1 {
			lzActive := player.diff.CheckModActive(difficulty.Lazer)

			if ((!lzActive || player.lzLegacySound) && rate > 0) || (lzActive && !player.lzLegacySound && state.endScored) {
				slider.hitSlider.HitEdge(len(slider.hitSlider.TickReverse), float64(time), true)
			}
		}

		position := slider.hitSlider.GetStackedEndPositionMod(player.diff)

		if player.diff.CheckModActive(difficulty.Lazer) && !player.lzNoSliderAcc {
			hit := Ignore
			if state.scored > 0 {
				hit = SliderFinish
			}

			slider.ruleSet.SendResult(player.cursor, createJudgementResultF(hit, SliderFinish, Hold, time, position, slider, true))
		} else {
			hit := Miss

			if rate == 1.0 {
				hit = Hit300
			} else if rate >= 0.5 {
				hit = Hit100
			} else if rate > 0 {
				hit = Hit50
			}

			combo := Reset
			if hit != Miss {
				combo = Hold

				if player.diff.CheckModActive(difficulty.Lazer) {
					combo = Increase
				}
			}

			slider.ruleSet.SendResult(player.cursor, createJudgementResultF(hit, Hit300, combo, time, position, slider, true))
		}

		state.isHit = true
	}

	return state.isHit
}

func (slider *Slider) processHeadMiss(player *difficultyPlayer, time int64) {
	state := slider.state[player]

	if time > int64(slider.hitSlider.GetStartTime())+player.diff.Hit50 && !state.isStartHit {
		if len(slider.players) == 1 && !state.isHit { //don't fade if slider already ended (and armed the start)
			slider.hitSlider.ArmStart(false, float64(time))
		}

		position := slider.hitSlider.GetStackedStartPositionMod(player.diff)

		if player.diff.CheckModActive(difficulty.Lazer) && !player.lzNoSliderAcc {
			slider.ruleSet.SendResult(player.cursor, createJudgementResult(Miss, Hit300, Reset, time, position, slider))
		} else {
			slider.ruleSet.SendResult(player.cursor, createJudgementResult(SliderMiss, SliderStart, Reset, time, position, slider))
		}

		if player.leftCond {
			state.downButton = Left
		} else if player.rightCond {
			state.downButton = Right
		} else {
			state.downButton = player.mouseDownButton
		}

		state.isStartHit = true
		state.startResult = Miss
	}
}

func (slider *Slider) UpdatePost(_ int64) bool {
	numFinishedTotal := 0

	for _, player := range slider.players {
		state := slider.state[player]

		if !state.isHit || !state.isStartHit {
			numFinishedTotal++
		}
	}

	return numFinishedTotal == 0
}

func (slider *Slider) MissForcefully(player *difficultyPlayer, time int64) {
	state := slider.state[player]

	if !state.isStartHit {
		position := slider.hitSlider.GetStackedStartPositionMod(player.diff)

		if len(slider.players) == 1 {
			slider.hitSlider.HitEdge(0, float64(time), false)
		}

		slider.ruleSet.SendResult(player.cursor, createJudgementResult(Miss, Hit300, Reset, time, position, slider))

		state.isStartHit = true
		state.startResult = Miss
	}
}

func (slider *Slider) IsHit(pl *difficultyPlayer) bool {
	return slider.state[pl].isHit
}

func (slider *Slider) IsStartHit(pl *difficultyPlayer) bool {
	return slider.state[pl].isStartHit
}

func (slider *Slider) GetStartResult(pl *difficultyPlayer) HitResult {
	return slider.state[pl].startResult
}

func (slider *Slider) GetFadeTime() int64 {
	return int64(slider.hitSlider.GetStartTime() - slider.fadeStartRelative)
}

func (slider *Slider) GetObject() objects.IHitObject {
	return slider.hitSlider
}
