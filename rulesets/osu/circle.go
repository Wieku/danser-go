package osu

import (
	"github.com/wieku/danser/beatmap/objects"
	"math"
	"github.com/wieku/danser/bmath/difficulty"
)

type objstate struct {
	buttons  buttonState
	finished bool
}

type Circle struct {
	ruleSet           *OsuRuleSet
	hitCircle         *objects.Circle
	players           []*difficultyPlayer
	state             []objstate
	fadeStartRelative float64
}

func (circle *Circle) Init(ruleSet *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer) {
	circle.ruleSet = ruleSet
	circle.hitCircle = object.(*objects.Circle)
	circle.players = players
	circle.state = make([]objstate, len(players))

	circle.fadeStartRelative = 1000000
	for _, player := range circle.players {
		circle.fadeStartRelative = math.Min(circle.fadeStartRelative, player.diff.Preempt)
	}
}

func (circle *Circle) Update(time int64) bool {
	unfinished := 0
	for i, player := range circle.players {
		state := &circle.state[i]

		if !state.finished {
			unfinished++

			if player.cursorLock == -1 {
				state.buttons.Left = player.cursor.LeftButton
				state.buttons.Right = player.cursor.RightButton
			}

			xOffset := 0.0
			yOffset := 0.0
			if player.diff.Mods&difficulty.HardRock > 0 {
				data := circle.hitCircle.GetBasicData()
				xOffset = data.StackOffset.X + float64(data.StackIndex) * player.diff.CircleRadius/(10)
				yOffset = data.StackOffset.Y - float64(data.StackIndex) * player.diff.CircleRadius/(10)
			}

			if player.cursorLock == -1 || player.cursorLock == circle.hitCircle.GetBasicData().Number {
				clicked := (!state.buttons.Left && player.cursor.LeftButton) || (!state.buttons.Right && player.cursor.RightButton)

				if clicked && player.cursor.Position.Dst(circle.hitCircle.GetPosition().SubS(xOffset, yOffset)) <= player.diff.CircleRadius {
					hit := HitResults.Miss

					relative := int64(math.Abs(float64(time - circle.hitCircle.GetBasicData().EndTime)))
					if relative < player.diff.Hit300 {
						hit = HitResults.Hit300
					} else if relative < player.diff.Hit100 {
						hit = HitResults.Hit100
					} else if relative < player.diff.Hit50 {
						hit = HitResults.Hit50
					} else if relative > int64(player.diff.Preempt-player.diff.FadeIn) {
						hit = HitResults.Ignore
					}

					if hit != HitResults.Ignore {
						combo := ComboResults.Increase
						if hit == HitResults.Miss {
							combo = ComboResults.Reset
						} else {
							if len(circle.players) == 1 {
								circle.hitCircle.PlaySound()
							}
						}

						circle.ruleSet.SendResult(time, player.cursor, circle.hitCircle.GetBasicData().Number, circle.hitCircle.GetPosition().X, circle.hitCircle.GetPosition().Y, hit, false, combo)

						player.cursorLock = -1
						state.finished = true
						continue
					}
				}

				player.cursorLock = circle.hitCircle.GetBasicData().Number
			}

			if time > circle.hitCircle.GetBasicData().EndTime+player.diff.Hit50 {
				circle.ruleSet.SendResult(time, player.cursor, circle.hitCircle.GetBasicData().Number, circle.hitCircle.GetPosition().X, circle.hitCircle.GetPosition().Y, HitResults.Miss, false, ComboResults.Reset)
				player.cursorLock = -1
				state.finished = true
				continue
			}

			if player.cursorLock == circle.hitCircle.GetBasicData().Number {
				state.buttons.Left = player.cursor.LeftButton
				state.buttons.Right = player.cursor.RightButton
			}
		}

	}

	if len(circle.players) > 1 && time == circle.hitCircle.GetBasicData().StartTime {
		circle.hitCircle.PlaySound()
	}

	return unfinished == 0
}

func (circle *Circle) GetFadeTime() int64 {
	return circle.hitCircle.GetBasicData().StartTime - int64(circle.fadeStartRelative)
}
