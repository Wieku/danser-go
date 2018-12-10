package osu

import (
	"github.com/wieku/danser/beatmap/objects"
	"math"
	"github.com/wieku/danser/bmath"
	"sort"
	"github.com/wieku/danser/bmath/difficulty"
)

type objstateS struct {
	buttons    buttonState
	finished   bool
	points     []tickpoint
	scored     int64
	missed     int64
	slideStart int64
	sliding    bool
}

type tickpoint struct {
	time       int64
	point      bmath.Vector2d
	scoreGiven HitResult
	edgeNum 	int
}

type Slider struct {
	ruleSet           *OsuRuleSet
	hitSlider         *objects.Slider
	players           []*difficultyPlayer
	state             []*objstateS
	fadeStartRelative float64
	lastTime          int64
}

func (slider *Slider) Init(ruleSet *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer) {
	slider.ruleSet = ruleSet
	slider.hitSlider = object.(*objects.Slider)
	slider.players = players
	slider.state = make([]*objstateS, len(players))

	rSlider := object.(*objects.Slider)

	slider.fadeStartRelative = 100000

	for i, player := range slider.players {
		slider.fadeStartRelative = math.Min(slider.fadeStartRelative, player.diff.Preempt)
		time := int64(math.Max(float64(slider.hitSlider.GetBasicData().StartTime)+float64((slider.hitSlider.GetBasicData().EndTime-slider.hitSlider.GetBasicData().StartTime)/2), float64(slider.hitSlider.GetBasicData().EndTime-36))) //slider ends 36ms before the real end for scoring
		slider.state[i] = new(objstateS)

		edgeNumber := 0
		for g, point := range rSlider.TickReverse {
			if g < len(rSlider.TickReverse)-1 {
				slider.state[i].points = append(slider.state[i].points, tickpoint{point.Time, point.Pos, HitResults.Slider30, edgeNumber})
				edgeNumber++
			}
		}

		slider.state[i].points = append(slider.state[i].points, tickpoint{time, slider.hitSlider.GetPointAt(time), HitResults.Slider30, edgeNumber})

		for _, point := range rSlider.TickPoints {
			slider.state[i].points = append(slider.state[i].points, tickpoint{point.Time, point.Pos, HitResults.Slider10, -1})
		}

		sort.Slice(slider.state[i].points, func(g, h int) bool { return slider.state[i].points[g].time < slider.state[i].points[h].time })
	}

}

func (slider *Slider) Update(time int64) bool {
	numFinishedTotal := 0

	sliderPosition := slider.hitSlider.GetPointAt(time)

	for i, player := range slider.players {
		state := slider.state[i]

		xOffset := 0.0
		yOffset := 0.0
		if player.diff.Mods&difficulty.HardRock > 0 {
			data := slider.hitSlider.GetBasicData()
			xOffset = data.StackOffset.X + float64(data.StackIndex)*player.diff.CircleRadius/(10)
			yOffset = data.StackOffset.Y - float64(data.StackIndex)*player.diff.CircleRadius/(10)
		}

		/*if !player.cursor.IsReplayFrame {
			continue
		}*/

		if !state.finished {
			numFinishedTotal++

			if player.cursorLock == -1 {
				state.buttons.Left = player.cursor.LeftButton
				state.buttons.Right = player.cursor.RightButton
			}

			if player.cursorLock == -1 || player.cursorLock == slider.hitSlider.GetBasicData().Number {
				clicked := (!state.buttons.Left && player.cursor.LeftButton) || (!state.buttons.Right && player.cursor.RightButton)

				if clicked && player.cursor.Position.Dst(slider.hitSlider.GetBasicData().StartPos.SubS(xOffset, yOffset)) <= player.diff.CircleRadius {
					hit := HitResults.SliderMiss
					combo := ComboResults.Reset

					relative := int64(math.Abs(float64(time - slider.hitSlider.GetBasicData().StartTime)))

					if relative < player.diff.Hit50 {
						hit = HitResults.Slider30
						state.scored++
						combo = ComboResults.Increase
					} else if relative > int64(player.diff.Preempt-player.diff.FadeIn) {
						hit = HitResults.Ignore
						combo = ComboResults.Hold
					} else {
						state.missed++
					}

					if hit != HitResults.Ignore {
						if hit != HitResults.SliderMiss && len(slider.players) == 1 {
							slider.hitSlider.PlayEdgeSample(0)
						}
						slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().Number, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, hit, true, combo)

						player.cursorLock = -1
						state.finished = true
						continue
					}
				}

				player.cursorLock = slider.hitSlider.GetBasicData().Number
			}

			if time > slider.hitSlider.GetBasicData().StartTime+player.diff.Hit50 {
				slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().Number, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, HitResults.SliderMiss, true, ComboResults.Reset)
				player.cursorLock = -1
				state.missed++
				state.finished = true
				continue
			}

			if player.cursorLock == slider.hitSlider.GetBasicData().Number {
				state.buttons.Left = player.cursor.LeftButton
				state.buttons.Right = player.cursor.RightButton
			}
		}

		if !state.finished {
			continue
		}

		radiusNeeded := player.diff.CircleRadius

		if state.sliding {
			radiusNeeded *= 2.4
		}

		allowable := (player.cursor.LeftButton || player.cursor.RightButton) && player.cursor.Position.Dst(sliderPosition.SubS(xOffset, yOffset)) <= radiusNeeded

		if allowable && !state.sliding {
			state.sliding = true
			state.slideStart = time
		}

		for j, point := range state.points {
			//We want to catch up with ticks overlapped by slider start hit window
			if int64(j) < state.scored+state.missed {
				continue
			}

			if point.time > time {
				break
			}

			if j > 0 && time >= point.time {
				if allowable && state.slideStart <= point.time {
					if len(slider.players) == 1 && j < len(state.points)-1 {
						if point.edgeNum == -1 {
							slider.hitSlider.PlayTick()
						} else {
							slider.hitSlider.PlayEdgeSample(point.edgeNum)
						}
					}
					slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().Number, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, point.scoreGiven, true, ComboResults.Increase)
					state.scored++
				} else {
					combo := ComboResults.Reset
					if j == len(state.points)-1 {
						combo = ComboResults.Hold
					}
					state.missed++
					slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().Number, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, HitResults.SliderMiss, true, combo)
				}

			}

		}

		if time >= slider.hitSlider.GetBasicData().EndTime {
			rate := float64(state.scored) / float64(len(state.points))
			hit := HitResults.Miss

			if rate > 0 {
				slider.hitSlider.PlayEdgeSample(len(slider.hitSlider.TickReverse)-1)
			}

			if rate == 1.0 {
				hit = HitResults.Hit300
			} else if rate >= 0.5 {
				hit = HitResults.Hit100
			} else if rate > 0 {
				hit = HitResults.Hit50
			}

			if hit != HitResults.Ignore {
				combo := ComboResults.Hold
				if hit == HitResults.Miss {
					combo = ComboResults.Reset
				}
				slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().Number, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, hit, false, combo)

				state.finished = true
			}
		} else {
			numFinishedTotal++
		}

		if !allowable && state.sliding && state.scored+state.missed < int64(len(state.points)) {
			state.sliding = false
		}

		state.buttons.Left = player.cursor.LeftButton
		state.buttons.Right = player.cursor.RightButton
	}

	slider.lastTime = time

	return numFinishedTotal == 0
}

func (slider *Slider) GetFadeTime() int64 {
	return slider.hitSlider.GetBasicData().StartTime - int64(slider.fadeStartRelative)
}
