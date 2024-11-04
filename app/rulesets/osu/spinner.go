package osu

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

const FrameTime = 1000.0 / 60
const countDuration = 595

type spinnerstate struct {
	lastAngle            float64
	lastAngle32          float32
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

	rotationCountFPrev                       float32
	totalAccumulatedRotation                 float32
	currentSpinMaxRotation                   float32
	totalAccumulatedRotationAtLastCompletion float32
	maximumBonusSpins                        int64
	lastTime                                 int64
}

func (s *spinnerstate) currentSpinRotation() float32 {
	return s.totalAccumulatedRotation - s.totalAccumulatedRotationAtLastCompletion
}

func (s *spinnerstate) totalRotation() float32 {
	return 360*float32(s.rotationCount) + s.currentSpinMaxRotation
}

func (s *spinnerstate) getCompletion() float32 {
	return s.totalRotation() / 360 / float32(s.requirement)
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
		spinner.state[player].frameVariance = FrameTime

		if player.diff.CheckModActive(difficulty.Lazer) {
			spinner.state[player].requirement = int64(player.diff.LzSpinnerMinRPS*float64(spinnerTime)/1000 + 0.0001)
			spinner.state[player].maximumBonusSpins = max(0, int64(player.diff.LzSpinnerMaxRPS*float64(spinnerTime)/1000+0.0001)-spinner.state[player].requirement-difficulty.LzSpinBonusGap)
		} else {
			spinner.state[player].requirement = int64(float64(spinnerTime) / 1000 * player.diff.SpinnerRatio)
		}
	}

	spinner.maxAcceleration = 0.00008 + max(0, (5000-float64(spinnerTime))/1000/2000)
}

func (spinner *Spinner) UpdateClickFor(*difficultyPlayer, int64) bool {
	return true
}

func (spinner *Spinner) UpdateFor(player *difficultyPlayer, time int64, _ bool) bool {
	state := spinner.state[player]

	if !state.finished {
		if player.diff.CheckModActive(difficulty.Lazer) {
			spinner.processLazer(player, time)
		} else {
			spinner.processStable(player, time)
		}
	}

	return state.finished
}

func (spinner *Spinner) processStable(player *difficultyPlayer, time int64) {
	spinnerPosition := spinner.hitSpinner.GetStartPosition()

	state := spinner.state[player]

	timeDiff := float64(time - player.cursor.LastFrameTime)
	if player.cursor.LastFrameTime == 0 {
		timeDiff = FrameTime
	}

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
					spinner.hitSpinner.Bonus(1000)
				}

				spinner.ruleSet.SendResult(player.cursor, createJudgementResult(SpinnerBonus, SpinnerBonus, Hold, time, spinnerPosition, spinner))
			} else if state.scoringRotationCount > 1 && state.scoringRotationCount%2 == 0 {
				spinner.ruleSet.SendResult(player.cursor, createJudgementResult(SpinnerPoints, SpinnerPoints, Hold, time, spinnerPosition, spinner))
			} else if state.scoringRotationCount > 1 {
				spinner.ruleSet.SendResult(player.cursor, createJudgementResult(SpinnerSpin, SpinnerSpin, Hold, time, spinnerPosition, spinner))
			}

			state.lastRotationCount = state.rotationCount
		}
	}
}

func (spinner *Spinner) processLazer(player *difficultyPlayer, time int64) {
	spinnerPosition := spinner.hitSpinner.GetStartPosition()

	state := spinner.state[player]

	if time >= int64(spinner.hitSpinner.GetStartTime()) && time <= int64(spinner.hitSpinner.GetEndTime()) {
		timeDiff := float64(time - state.lastTime)
		state.lastTime = time

		thisAngle := player.cursor.RawPosition.Sub(spinnerPosition).Angle()

		var delta float32 = 0.0

		if state.updatedBefore {
			delta = thisAngle - state.lastAngle32
		}

		state.updatedBefore = true

		if delta > 180 {
			delta -= 360
		}

		if delta < -180 {
			delta += 360
		}

		var deltaRPM float32 = 0

		if player.gameDownState || player.diff.CheckModActive(difficulty.Relax) {
			delta *= float32(player.diff.GetSpeed())

			if delta != 0 {
				state.totalAccumulatedRotation += delta

				state.currentSpinMaxRotation = max(state.currentSpinMaxRotation, mutils.Abs(state.currentSpinRotation()))

				// Handle the case where the user has completed another spin.
				// Note that this does could be an `if` rather than `while` if the above assertion held true.
				// It is a `while` loop to handle tests which throw larger values at this method.
				for state.currentSpinMaxRotation >= 360 {
					direction := mutils.Signum(state.currentSpinRotation())

					state.rotationCount++

					// Incrementing the last completion point will cause `currentSpinRotation` to
					// hold the remaining spin that needs to be considered.
					state.totalAccumulatedRotationAtLastCompletion += float32(direction) * 360

					// Reset the current max as we are entering a new spin.
					// Importantly, carry over the remainder (which is now stored in `currentSpinRotation`).
					state.currentSpinMaxRotation = mutils.Abs(state.currentSpinRotation())
				}
			}

			state.rotationCountF += delta

			deltaRPM = delta
		}

		state.lastAngle32 = thisAngle

		spinning := mutils.Abs(state.rotationCountF-state.rotationCountFPrev) > 10

		state.rotationCountFPrev = mutils.Lerp(state.rotationCountFPrev, state.rotationCountF, 1-math32.Pow(0.99, float32(player.diff.GetModifiedTime(timeDiff))))

		if len(spinner.players) == 1 {
			if spinning {
				spinner.hitSpinner.StartSpinSample()
			} else {
				spinner.hitSpinner.PauseSpinSample()
			}
		}

		if timeDiff > 0 {
			// We don't use lazer's fancy rpm metre
			decay1 := math.Pow(0.95, timeDiff/FrameTime)
			state.rpm = state.rpm*decay1 + (1.0-decay1)*(math.Abs(float64(deltaRPM)/timeDiff*1000))/360*60
		}

		if len(spinner.players) == 1 {
			spinner.hitSpinner.SetRotation(float64(state.rotationCountFPrev * math32.Pi / 180))
			spinner.hitSpinner.SetRPM(state.rpm)
			spinner.hitSpinner.UpdateCompletion(float64(state.getCompletion()))
		}

		totalSpins := state.maximumBonusSpins + state.requirement + difficulty.LzSpinBonusGap

		for i := state.lastRotationCount; i < state.rotationCount; i++ {
			if i == state.requirement && len(spinner.players) == 1 {
				spinner.hitSpinner.Clear()
			}

			if i < totalSpins {
				if i < state.requirement+difficulty.LzSpinBonusGap {
					spinner.ruleSet.SendResult(player.cursor, createJudgementResult(SpinnerPoints, SpinnerPoints, Hold, time, spinnerPosition, spinner))
				} else {
					if len(spinner.players) == 1 {
						spinner.hitSpinner.Bonus(int(SpinnerBonus.ScoreValueMod(player.diff.Mods)))
					}

					spinner.ruleSet.SendResult(player.cursor, createJudgementResult(SpinnerBonus, SpinnerBonus, Hold, time, spinnerPosition, spinner))
				}
			} else {
				if len(spinner.players) == 1 {
					spinner.hitSpinner.Bonus(0)
				}
			}
		}

		state.lastRotationCount = state.rotationCount
	}
}

func (spinner *Spinner) UpdatePostFor(player *difficultyPlayer, time int64, _ bool) bool {
	state := spinner.state[player]

	if time >= int64(spinner.hitSpinner.GetEndTime()) && !state.finished {
		hit := Miss
		combo := Reset

		if player.diff.CheckModActive(difficulty.Lazer) {
			if spinner.state[player].requirement == 0 || state.getCompletion() >= 1.0 {
				hit = Hit300
			} else if state.getCompletion() >= 0.9 {
				hit = Hit100
			} else if state.getCompletion() >= 0.75 {
				hit = Hit50
			}
		} else {
			if (!player.cursor.OldSpinnerScoring && spinner.state[player].requirement == 0) || state.scoringRotationCount >= spinner.getRequirementGreat(player) {
				hit = Hit300
			} else if state.scoringRotationCount >= spinner.getRequirementOk(player) {
				hit = Hit100
			} else if state.scoringRotationCount >= spinner.getRequirementMeh(player) {
				hit = Hit50
			}
		}

		if hit != Miss {
			combo = Increase
		}

		if len(spinner.players) == 1 {
			spinner.hitSpinner.StopSpinSample()
			spinner.hitSpinner.Hit(float64(time), hit != Miss)
		}

		spinner.ruleSet.SendResult(player.cursor, createJudgementResult(hit, Hit300, combo, time, spinner.hitSpinner.GetPosition(), spinner))

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

func (spinner *Spinner) MissForcefully(_ *difficultyPlayer, _ int64) {
}

func (spinner *Spinner) IsHit(pl *difficultyPlayer) bool {
	return spinner.state[pl].finished
}

func (spinner *Spinner) GetFadeTime() int64 {
	return int64(spinner.hitSpinner.GetStartTime() - spinner.fadeStartRelative)
}

func (spinner *Spinner) GetObject() objects.IHitObject {
	return spinner.hitSpinner
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
