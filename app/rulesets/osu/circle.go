package osu

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
)

type objstate struct {
	isHit bool
}

type Circle struct {
	ruleSet           *OsuRuleSet
	hitCircle         *objects.Circle
	players           []*difficultyPlayer
	state             map[*difficultyPlayer]*objstate
	fadeStartRelative float64
}

func (circle *Circle) GetNumber() int64 {
	return circle.hitCircle.GetBasicData().Number
}

func (circle *Circle) Init(ruleSet *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer) {
	circle.ruleSet = ruleSet
	circle.hitCircle = object.(*objects.Circle)
	circle.players = players

	circle.state = make(map[*difficultyPlayer]*objstate)

	circle.fadeStartRelative = 1000000
	for _, player := range circle.players {
		circle.state[player] = new(objstate)
		circle.fadeStartRelative = math.Min(circle.fadeStartRelative, player.diff.Preempt)
	}
}

func (circle *Circle) UpdateFor(player *difficultyPlayer, time int64) bool {
	return true
}

func (circle *Circle) UpdateClickFor(player *difficultyPlayer, time int64) bool {
	state := circle.state[player]

	if !state.isHit {
		xOffset := float32(0.0)
		yOffset := float32(0.0)

		if player.diff.Mods&difficulty.HardRock > 0 {
			data := circle.hitCircle.GetBasicData()
			xOffset = data.StackOffset.X + float32(data.StackIndex)*float32(player.diff.CircleRadius)/10
			yOffset = data.StackOffset.Y - float32(data.StackIndex)*float32(player.diff.CircleRadius)/10
		}

		clicked := player.leftCondE || player.rightCondE
		inRange := player.cursor.Position.Dst(circle.hitCircle.GetPosition().SubS(xOffset, yOffset)) <= float32(player.diff.CircleRadius)

		if clicked && inRange {
			if player.leftCondE {
				player.leftCondE = false
			} else if player.rightCondE {
				player.rightCondE = false
			}

			action := circle.ruleSet.CanBeHit(time, circle, player)

			if action == Click {
				hit := Miss

				relative := int64(math.Abs(float64(time - circle.hitCircle.GetBasicData().EndTime)))
				if relative < player.diff.Hit300 {
					hit = Hit300
				} else if relative < player.diff.Hit100 {
					hit = Hit100
				} else if relative < player.diff.Hit50 {
					hit = Hit50
				}

				if hit != Ignore {
					combo := ComboResults.Increase
					if hit == Miss {
						combo = ComboResults.Reset
					} else {
						if len(circle.players) == 1 {
							circle.hitCircle.PlaySound()
						}
					}

					if len(circle.players) == 1 {
						circle.hitCircle.Arm(hit != Miss, time)
					}

					circle.ruleSet.SendResult(time, player.cursor, circle.hitCircle.GetBasicData().Number, circle.hitCircle.GetPosition().X, circle.hitCircle.GetPosition().Y, hit, false, combo)

					state.isHit = true
				}
			} else if action == Shake && len(circle.players) == 1 {
				circle.hitCircle.Shake(time)
			}
		}
	}

	return !state.isHit
}

func (circle *Circle) UpdatePostFor(player *difficultyPlayer, time int64) bool {
	state := circle.state[player]

	if time > circle.hitCircle.GetBasicData().EndTime+player.diff.Hit50 && !state.isHit {
		circle.ruleSet.SendResult(time, player.cursor, circle.hitCircle.GetBasicData().Number, circle.hitCircle.GetPosition().X, circle.hitCircle.GetPosition().Y, Miss, false, ComboResults.Reset)

		if len(circle.players) == 1 {
			circle.hitCircle.Arm(false, time)
		}

		state.isHit = true
	}

	return state.isHit
}

func (circle *Circle) UpdatePost(_ int64) bool {
	unfinished := 0

	for _, player := range circle.players {
		state := circle.state[player]

		if !state.isHit {
			unfinished++
		}
	}

	return unfinished == 0
}

func (circle *Circle) IsHit(player *difficultyPlayer) bool {
	return circle.state[player].isHit
}

func (circle *Circle) GetFadeTime() int64 {
	return circle.hitCircle.GetBasicData().StartTime - int64(circle.fadeStartRelative)
}
