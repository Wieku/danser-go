package osu

import (
	"github.com/wieku/danser/beatmap/objects"
	"math"
	"github.com/wieku/danser/bmath"
	"sort"
)

type objstateS struct {
	buttons   buttonState
	finished  bool
	lastPoint tickpoint
	scored    int64
}

type tickpoint struct {
	time       int64
	point      bmath.Vector2d
	scoreGiven HitResult
}

type Slider struct {
	ruleSet           *OsuRuleSet
	hitSlider         *objects.Slider
	players           []*difficultyPlayer
	state             []objstateS
	points            []tickpoint
	fadeStartRelative float64
	lastTime          int64
}

func (slider *Slider) Init(ruleSet *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer) {
	slider.ruleSet = ruleSet
	slider.hitSlider = object.(*objects.Slider)
	slider.players = players
	slider.state = make([]objstateS, len(players))

	for i, player := range slider.players {
		slider.fadeStartRelative = math.Min(slider.fadeStartRelative, player.diff.Preempt)
		time := int64(math.Max(float64(slider.hitSlider.GetBasicData().StartTime), float64(slider.hitSlider.GetBasicData().EndTime-36))) //slider ends 36ms before the real end for scoring
		slider.state[i].lastPoint = tickpoint{time, slider.hitSlider.GetPointAt(time), HitResults.Slider30}
	}

	rSlider := object.(*objects.Slider)

	for _, point := range rSlider.TickReverse {
		slider.points = append(slider.points, tickpoint{point.Time, point.Pos, HitResults.Slider30})
	}

	for _, point := range rSlider.TickPoints {
		slider.points = append(slider.points, tickpoint{point.Time, point.Pos, HitResults.Slider10})
	}

	sort.Slice(slider.points, func(i, j int) bool { return slider.points[i].time < slider.points[j].time })

}

func (slider *Slider) Update(time int64) bool {
	numFinished := 0

	for i, player := range slider.players {
		state := &slider.state[i]
		if !state.finished && (player.cursorLock == -1 || player.cursorLock == slider.hitSlider.GetBasicData().Number) {
			if ((!state.buttons.Left && player.cursor.LeftButton) || (!state.buttons.Right && player.cursor.RightButton)) && player.cursor.Position.Dst(slider.hitSlider.GetBasicData().StartPos) <= player.diff.CircleRadius {
				relative := int64(math.Abs(float64(time - slider.hitSlider.GetBasicData().StartTime)))

				if relative <= player.diff.Hit50 {
					slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().StartPos.X, slider.hitSlider.GetBasicData().StartPos.Y, HitResults.Slider30, true, ComboResults.Increase)
					slider.state[i].scored++
				} else if relative >= int64(player.diff.Preempt-player.diff.FadeIn) {
					slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetBasicData().StartPos.X, slider.hitSlider.GetBasicData().StartPos.Y, HitResults.Ignore, true, ComboResults.Reset)
				}

				player.cursorLock = -1
				state.finished = true
				continue
			}

			if time > slider.hitSlider.GetBasicData().StartTime+player.diff.Hit50 {
				slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, HitResults.Ignore, true, ComboResults.Reset)
				player.cursorLock = -1
				state.finished = true
				continue
			}

			player.cursorLock = slider.hitSlider.GetBasicData().Number
		}

		state.buttons.Left = player.cursor.LeftButton
		state.buttons.Right = player.cursor.RightButton
	}

	for j, point := range slider.points {
		if point.time <= slider.lastTime {
			continue
		}
		if numFinished > 0 {
			break
		}
		numFinished++
		for i, player := range slider.players {

			state := &slider.state[i]

			if (j > 0 && time >= point.time) || (j == len(slider.points)-1 && time >= state.lastPoint.time) {

				subPoint := point
				if slider.lastTime < state.lastPoint.time {
					if j == len(slider.points)-1 {
						subPoint = state.lastPoint
					}
					if (player.cursor.LeftButton || player.cursor.RightButton) && player.cursor.Position.Dst(subPoint.point) <= player.diff.CircleRadius*2.4 {
						slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, subPoint.scoreGiven, true, ComboResults.Increase)
						slider.state[i].scored++
					} else {
						combo := ComboResults.Reset
						if j == len(slider.points)-1 {
							combo = ComboResults.Hold
						}
						slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, HitResults.Ignore, true, combo)
					}
				}

				if j == len(slider.points)-1 && time >= point.time {
					rate := float64(slider.state[i].scored) / float64(len(slider.points))
					hit := HitResults.Miss

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
						slider.ruleSet.SendResult(time, player.cursor, slider.hitSlider.GetPosition().X, slider.hitSlider.GetPosition().Y, hit, false, combo)

						player.cursorLock = -1
						state.finished = true
					}
				}
			}

			state.buttons.Left = player.cursor.LeftButton
			state.buttons.Right = player.cursor.RightButton

		}

	}

	slider.lastTime = time

	return numFinished == 0
}

func (slider *Slider) GetFadeTime() int64 {
	return slider.hitSlider.GetBasicData().StartTime - int64(slider.fadeStartRelative)
}
