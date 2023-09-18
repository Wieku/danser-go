package osu

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"math"
)

const FrameTime = 1000.0 / 60

type spinnerstate struct {
	lastAngle            float64
	requirement          int64
	rotationCount        int64
	lastRotationCount    int64
	scoringRotationCount int64
	rotationCountF       float32
	rotationCountFD      float64
	frameVariance        float64
	theoreticalVelocity  float64
	currentVelocity      float64
	finished             bool
	zeroCount            int64
	rpm                  float64
	updatedBefore        bool
}

type Spinner struct {
	ruleSet           *OsuRuleSet
	hitSpinner        *objects.Spinner
	players           []*difficultyPlayer
	state             map[*difficultyPlayer]*spinnerstate
	fadeStartRelative float64
	maxAcceleration   float64
}

func (spinner *Spinner) GetNumber() int64 {
	return spinner.hitSpinner.GetID()
}

func (spinner *Spinner) Init(ruleSet *OsuRuleSet, object objects.IHitObject, players []*difficultyPlayer) {
	spinner.ruleSet = ruleSet
	spinner.hitSpinner = object.(*objects.Spinner)
	spinner.players = players
	spinner.state = make(map[*difficultyPlayer]*spinnerstate)

	rSpinner := object.(*objects.Spinner)

	spinnerTime := int64(rSpinner.GetEndTime()) - int64(rSpinner.GetStartTime())

	spinner.fadeStartRelative = 100000

	for _, player := range spinner.players {
		spinner.state[player] = new(spinnerstate)
		spinner.fadeStartRelative = min(spinner.fadeStartRelative, player.diff.Preempt)
		spinner.state[player].requirement = int64(float64(spinnerTime) / 1000 * player.diff.SpinnerRatio)
		spinner.state[player].frameVariance = FrameTime
	}

	spinner.maxAcceleration = 0.00008 + max(0, (5000-float64(spinnerTime))/1000/2000)
}

func (spinner *Spinner) UpdateClickFor(*difficultyPlayer, int64) bool {
	return true
}

func (spinner *Spinner) UpdateFor(player *difficultyPlayer, time int64, _ bool) bool {
	numFinishedTotal := 0

	spinnerPosition := spinner.hitSpinner.GetStackedStartPosition()

	state := spinner.state[player]

	timeDiff := float64(time - player.cursor.LastFrameTime)
	if player.cursor.LastFrameTime == 0 {
		timeDiff = FrameTime
	}

	if !state.finished {
		numFinishedTotal++

		if player.cursor.IsReplayFrame && time > int64(spinner.hitSpinner.GetStartTime()) && time < int64(spinner.hitSpinner.GetEndTime()) {
			maxAccelThisFrame := player.diff.GetModifiedTime(spinner.maxAcceleration * timeDiff)

			if player.diff.CheckModActive(difficulty.SpunOut) || player.diff.CheckModActive(difficulty.Relax2) {
				state.currentVelocity = 0.03
			} else if state.theoreticalVelocity > state.currentVelocity {
				accel := maxAccelThisFrame
				if state.currentVelocity < 0 && player.diff.CheckModActive(difficulty.Relax) {
					accel /= 4
				}

				state.currentVelocity += min(state.theoreticalVelocity-state.currentVelocity, accel)
			} else {
				accel := -maxAccelThisFrame
				if state.currentVelocity > 0 && player.diff.CheckModActive(difficulty.Relax) {
					accel /= 4
				}

				state.currentVelocity += max(state.theoreticalVelocity-state.currentVelocity, accel)
			}

			state.currentVelocity = max(-0.05, min(state.currentVelocity, 0.05))

			if len(spinner.players) == 1 {
				if state.currentVelocity == 0 {
					spinner.hitSpinner.PauseSpinSample()
				} else {
					spinner.hitSpinner.StartSpinSample()
				}
			}

			decay1 := math.Pow(0.9, timeDiff/FrameTime)
			state.rpm = state.rpm*decay1 + (1.0-decay1)*(math.Abs(state.currentVelocity)*1000)/(math.Pi*2)*60

			mouseAngle := float64(player.cursor.RawPosition.Sub(spinnerPosition).AngleR())

			if !player.cursor.OldSpinnerScoring && !state.updatedBefore {
				state.lastAngle = mouseAngle
				state.updatedBefore = true
			}

			angleDiff := mouseAngle - state.lastAngle

			if mouseAngle-state.lastAngle < -math.Pi {
				angleDiff = (2 * math.Pi) + mouseAngle - state.lastAngle
			} else if state.lastAngle-mouseAngle < -math.Pi {
				angleDiff = (-2 * math.Pi) - state.lastAngle + mouseAngle
			}

			decay := math.Pow(0.999, timeDiff)
			state.frameVariance = decay*state.frameVariance + (1-decay)*timeDiff

			if angleDiff == 0 {
				state.zeroCount += 1

				if state.zeroCount < 2 {
					state.theoreticalVelocity /= 3
				} else {
					state.theoreticalVelocity = 0
				}
			} else {
				state.zeroCount = 0

				if (!player.gameDownState && !player.diff.CheckModActive(difficulty.Relax)) || time < int64(spinner.hitSpinner.GetStartTime()) || time > int64(spinner.hitSpinner.GetEndTime()) {
					angleDiff = 0
				}

				if math.Abs(angleDiff) < math.Pi {
					if player.diff.GetModifiedTime(state.frameVariance) > FrameTime*1.04 {
						if timeDiff > 0 {
							state.theoreticalVelocity = angleDiff / player.diff.GetModifiedTime(timeDiff)
						} else {
							state.theoreticalVelocity = 0
						}
					} else {
						state.theoreticalVelocity = angleDiff / FrameTime
					}
				} else {
					state.theoreticalVelocity = 0
				}
			}

			state.lastAngle = mouseAngle

			rotationAddition := state.currentVelocity * timeDiff

			state.rotationCountFD += rotationAddition
			state.rotationCountF += float32(math.Abs(float64(float32(rotationAddition)) / math.Pi))

			if len(spinner.players) == 1 {
				spinner.hitSpinner.SetRotation(player.diff.GetModifiedTime(state.rotationCountFD))
				spinner.hitSpinner.SetRPM(state.rpm)
				spinner.hitSpinner.UpdateCompletion(float64(state.rotationCountF) / float64(state.requirement))
			}

			state.rotationCount = int64(state.rotationCountF)

			if state.rotationCount != state.lastRotationCount {
				state.scoringRotationCount++

				if state.scoringRotationCount == spinner.getRequirementClear(player) && len(spinner.players) == 1 {
					spinner.hitSpinner.Clear()
				}

				if state.scoringRotationCount > state.requirement+3 && (state.scoringRotationCount-(state.requirement+3))%2 == 0 {
					if len(spinner.players) == 1 {
						spinner.hitSpinner.Bonus()
					}

					spinner.ruleSet.SendResult(time, player.cursor, spinner, spinnerPosition.X, spinnerPosition.Y, SpinnerBonus, Hold)
				} else if state.scoringRotationCount > 1 && state.scoringRotationCount%2 == 0 {
					spinner.ruleSet.SendResult(time, player.cursor, spinner, spinnerPosition.X, spinnerPosition.Y, SpinnerPoints, Hold)
				} else if state.scoringRotationCount > 1 {
					spinner.ruleSet.SendResult(time, player.cursor, spinner, spinnerPosition.X, spinnerPosition.Y, SpinnerSpin, Hold)
				}

				state.lastRotationCount = state.rotationCount
			}
		}
	}

	return numFinishedTotal == 0
}

func (spinner *Spinner) UpdatePostFor(player *difficultyPlayer, time int64, _ bool) bool {
	state := spinner.state[player]

	if time >= int64(spinner.hitSpinner.GetEndTime()) && !state.finished {
		hit := Miss
		combo := Reset

		if (!player.cursor.OldSpinnerScoring && spinner.state[player].requirement == 0) || state.scoringRotationCount >= spinner.getRequirementGreat(player) {
			hit = Hit300
		} else if state.scoringRotationCount >= spinner.getRequirementOk(player) {
			hit = Hit100
		} else if state.scoringRotationCount >= spinner.getRequirementMeh(player) {
			hit = Hit50
		}

		if hit != Miss {
			combo = Increase
		}

		if len(spinner.players) == 1 {
			spinner.hitSpinner.StopSpinSample()
			spinner.hitSpinner.Hit(float64(time), hit != Miss)
		}

		spinner.ruleSet.SendResult(time, player.cursor, spinner, spinner.hitSpinner.GetPosition().X, spinner.hitSpinner.GetPosition().Y, hit, combo)

		state.finished = true
	}

	return state.finished
}

func (spinner *Spinner) UpdatePost(_ int64) bool {
	numFinishedTotal := 0

	for _, player := range spinner.players {
		state := spinner.state[player]

		if !state.finished {
			numFinishedTotal++
		}
	}

	return numFinishedTotal == 0
}

func (spinner *Spinner) IsHit(pl *difficultyPlayer) bool {
	return spinner.state[pl].finished
}

func (spinner *Spinner) GetFadeTime() int64 {
	return int64(spinner.hitSpinner.GetStartTime() - spinner.fadeStartRelative)
}

// new vs old spinner handling helpers
func (spinner *Spinner) getRequirementMeh(player *difficultyPlayer) int64 {
	if player.cursor.OldSpinnerScoring {
		return spinner.state[player].requirement
	}

	return spinner.state[player].requirement / 4
}

func (spinner *Spinner) getRequirementOk(player *difficultyPlayer) int64 {
	if player.cursor.OldSpinnerScoring {
		return spinner.state[player].requirement + 1
	}

	return spinner.state[player].requirement - 1
}

func (spinner *Spinner) getRequirementGreat(player *difficultyPlayer) int64 {
	if player.cursor.OldSpinnerScoring {
		return spinner.state[player].requirement + 2
	}

	return spinner.state[player].requirement + 1
}

func (spinner *Spinner) getRequirementClear(player *difficultyPlayer) int64 {
	if player.cursor.OldSpinnerScoring {
		return spinner.state[player].requirement + 1
	}

	return spinner.state[player].requirement
}
