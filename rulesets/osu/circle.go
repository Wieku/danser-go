package osu

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/beatmap/objects"
	"github.com/wieku/danser-go/bmath/difficulty"
	"github.com/wieku/danser-go/render/batches"
	"math"
)

type Renderable interface {
	Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch)
	DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch)
}

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
	//if len(players) > 1 {
	//	circle.renderable = NewHitCircleSprite(*difficulty.NewDifficulty(players[0].diff.GetHPDrain(), players[0].diff.GetCS(), players[0].diff.GetOD(), players[0].diff.GetAR()), object.GetBasicData().StartPos, object.GetBasicData().StartTime)
	//} else {
	//	circle.renderable = NewHitCircleSprite(*players[0].diff, object.GetBasicData().StartPos, object.GetBasicData().StartTime)
	//}

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

		//if circle.GetNumber() == 53 {
		//	log.Println("click", time, circle.hitCircle.GetBasicData().Number, circle.hitCircle.GetBasicData().StartTime, circle.hitCircle.GetBasicData().EndTime, circle.hitCircle.GetBasicData().EndPos, player.cursor.LeftButton, player.cursor.RightButton, circle.ruleSet.CanBeHit(time, circle, player), player.cursor.Position, circle.hitCircle.GetBasicData().StartPos.SubS(xOffset, yOffset), player.cursor.Position.Dst(circle.hitCircle.GetBasicData().StartPos.SubS(xOffset, yOffset)), player.diff.CircleRadius, player.cursor.Position.Dst(circle.hitCircle.GetBasicData().StartPos.SubS(xOffset, yOffset)) <= float32(player.diff.CircleRadius))
		//}

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

				hit := HitResults.Miss

				relative := int64(math.Abs(float64(time - circle.hitCircle.GetBasicData().EndTime)))
				if relative < player.diff.Hit300 {
					hit = HitResults.Hit300
				} else if relative < player.diff.Hit100 {
					hit = HitResults.Hit100
				} else if relative < player.diff.Hit50 {
					hit = HitResults.Hit50
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

					if len(circle.players) == 1 {
						circle.hitCircle.Arm(hit != HitResults.Miss, time)
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

func (circle *Circle) UpdatePost(time int64) bool {
	unfinished := 0
	for _, player := range circle.players {
		state := circle.state[player]

		if !state.isHit {
			unfinished++

			if time > circle.hitCircle.GetBasicData().EndTime+player.diff.Hit50 {
				circle.ruleSet.SendResult(time, player.cursor, circle.hitCircle.GetBasicData().Number, circle.hitCircle.GetPosition().X, circle.hitCircle.GetPosition().Y, HitResults.Miss, false, ComboResults.Reset)
				if len(circle.players) == 1 {
					circle.hitCircle.Arm(false, time)
				}

				state.isHit = true
			}
		}

	}

	if len(circle.players) > 1 && time == circle.hitCircle.GetBasicData().StartTime {
		//circle.hitCircle.PlaySound()
		//circle.renderable.Hit(time)
	}

	return unfinished == 0
}

func (circle *Circle) IsHit(player *difficultyPlayer) bool {
	return circle.state[player].isHit
}

func (circle *Circle) GetFadeTime() int64 {
	return circle.hitCircle.GetBasicData().StartTime - int64(circle.fadeStartRelative)
}

func (self *Circle) Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	//self.renderable.Draw(time, color, batch)
}

func (self *Circle) DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {
	//self.renderable.DrawApproach(time, color, batch)
}
