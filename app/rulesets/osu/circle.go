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
	return circle.hitCircle.GetID()
}

func (circle *Circle) Init(ruleSet *OsuRuleSet, object objects.IHitObject, players []*difficultyPlayer) {
	circle.ruleSet = ruleSet
	circle.hitCircle = object.(*objects.Circle)
	circle.players = players

	circle.state = make(map[*difficultyPlayer]*objstate)

	circle.fadeStartRelative = 1000000
	for _, player := range circle.players {
		circle.state[player] = new(objstate)
		circle.fadeStartRelative = min(circle.fadeStartRelative, player.diff.Preempt)
	}
}

func (circle *Circle) UpdateFor(_ *difficultyPlayer, _ int64, _ bool) bool {
	return true
}

func (circle *Circle) UpdateClickFor(player *difficultyPlayer, time int64) bool {
	state := circle.state[player]

	if !state.isHit {
		position := circle.hitCircle.GetStackedPositionAtMod(float64(time), player.diff)

		clicked := player.leftCondE || player.rightCondE

		radius := float32(player.diff.CircleRadius)
		if player.diff.CheckModActive(difficulty.Relax2) {
			radius = 100
		}

		inRange := player.cursor.RawPosition.Dst(position) <= radius

		if clicked {
			action := circle.ruleSet.CanBeHit(time, circle, player)

			if inRange {
				if action == Click {
					if player.leftCondE {
						player.leftCondE = false
					} else if player.rightCondE {
						player.rightCondE = false
					}

					hit := circle.ruleSet.GetResultForDelta(player, math.Abs(float64(time)-circle.hitCircle.GetEndTime()))

					if hit != Ignore {
						combo := Increase
						if hit == Miss {
							combo = Reset
						} else {
							if len(circle.players) == 1 {
								circle.hitCircle.PlaySound()
							}
						}

						if len(circle.players) == 1 {
							circle.hitCircle.Arm(hit != Miss, float64(time))
						}

						circle.ruleSet.PostHit(time, circle, player)
						circle.ruleSet.SendResult(player.cursor, createJudgementResult(hit, Hit300, combo, time, position, circle))

						state.isHit = true
					}
				} else {
					player.leftCondE = false
					player.rightCondE = false

					if action == Shake && len(circle.players) == 1 {
						circle.hitCircle.Shake(float64(time))
					}
				}
			} else if action == Click {
				circle.ruleSet.SendResult(player.cursor, createJudgementResult(PositionalMiss, Hit300, Hold, time, position, circle))
			}
		}
	}

	return !state.isHit
}

func (circle *Circle) UpdatePostFor(player *difficultyPlayer, time int64, _ bool) bool {
	state := circle.state[player]

	if time > int64(circle.hitCircle.GetEndTime())+player.diff.Hit50 && !state.isHit {
		position := circle.hitCircle.GetStackedPositionAtMod(float64(time), player.diff)
		circle.ruleSet.SendResult(player.cursor, createJudgementResult(Miss, Hit300, Reset, time, position, circle))

		if len(circle.players) == 1 {
			circle.hitCircle.Arm(false, float64(time))
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

func (circle *Circle) MissForcefully(player *difficultyPlayer, time int64) {
	state := circle.state[player]

	if !state.isHit {
		position := circle.hitCircle.GetStackedPositionAtMod(float64(time), player.diff)

		if len(circle.players) == 1 {
			circle.hitCircle.Arm(false, float64(time))
		}

		circle.ruleSet.SendResult(player.cursor, createJudgementResult(Miss, Hit300, Reset, time, position, circle))

		state.isHit = true
	}
}

func (circle *Circle) IsHit(player *difficultyPlayer) bool {
	return circle.state[player].isHit
}

func (circle *Circle) GetFadeTime() int64 {
	return int64(circle.hitCircle.GetStartTime() - circle.fadeStartRelative)
}

func (circle *Circle) GetObject() objects.IHitObject {
	return circle.hitCircle
}
