package osu

import (
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
	rotationCountF       float64
	rotationCountFD      float64
	frameVariance        float64
	theoreticalVelocity  float64
	currentVelocity      float64
	finished             bool
	zeroCount            int64
	rpm                  float64
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
	return spinner.hitSpinner.GetBasicData().Number
}

func (spinner *Spinner) Init(ruleSet *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer) {
	spinner.ruleSet = ruleSet
	spinner.hitSpinner = object.(*objects.Spinner)
	spinner.players = players
	spinner.state = make(map[*difficultyPlayer]*spinnerstate)

	rSpinner := object.(*objects.Spinner)

	spinnerTime := rSpinner.GetBasicData().EndTime - rSpinner.GetBasicData().StartTime

	spinner.fadeStartRelative = 100000

	for _, player := range spinner.players {
		spinner.state[player] = new(spinnerstate)
		spinner.fadeStartRelative = math.Min(spinner.fadeStartRelative, player.diff.Preempt)
		spinner.state[player].requirement = int64(float64(spinnerTime) / 1000 * player.diff.SpinnerRatio)
		spinner.state[player].frameVariance = FrameTime
	}

	spinner.maxAcceleration = 0.00008 + math.Max(0, (5000-float64(spinnerTime))/1000/2000)
}

func (spinner *Spinner) UpdateClickFor(*difficultyPlayer, int64) bool {
	return true
}

func (spinner *Spinner) UpdateFor(player *difficultyPlayer, time int64) bool {
	numFinishedTotal := 0

	spinnerPosition := spinner.hitSpinner.GetBasicData().StartPos

	state := spinner.state[player]

	timeDiff := float64(time - player.cursor.LastFrameTime)
	if player.cursor.LastFrameTime == 0 {
		timeDiff = FrameTime
	}

	if !state.finished {
		numFinishedTotal++

		if player.cursor.IsReplayFrame && time > spinner.hitSpinner.GetBasicData().StartTime && time < spinner.hitSpinner.GetBasicData().EndTime {
			decay1 := math.Pow(0.9, timeDiff/FrameTime)
			state.rpm = state.rpm*decay1 + (1.0-decay1)*(math.Abs(state.currentVelocity)*1000)/(math.Pi*2)*60

			mouseAngle := float64(player.cursor.Position.Sub(spinnerPosition).AngleR())

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

				if !player.gameDownState || time < spinner.hitSpinner.GetBasicData().StartTime || time > spinner.hitSpinner.GetBasicData().EndTime {
					angleDiff = 0
				}

				if math.Abs(angleDiff) < math.Pi {
					if player.diff.GetModifiedTime(state.frameVariance) > FrameTime*1.04 {
						state.theoreticalVelocity = angleDiff / player.diff.GetModifiedTime(timeDiff)
					} else {
						state.theoreticalVelocity = angleDiff / FrameTime
					}
				} else {
					state.theoreticalVelocity = 0
				}
			}

			state.lastAngle = mouseAngle

			maxAccelThisFrame := player.diff.GetModifiedTime(spinner.maxAcceleration * timeDiff)

			if state.theoreticalVelocity > state.currentVelocity {
				state.currentVelocity += math.Min(state.theoreticalVelocity-state.currentVelocity, maxAccelThisFrame)
			} else {
				state.currentVelocity += math.Max(state.theoreticalVelocity-state.currentVelocity, -maxAccelThisFrame)
			}

			state.currentVelocity = math.Max(-0.05, math.Min(state.currentVelocity, 0.05))

			if len(spinner.players) == 1 {
				if state.currentVelocity == 0 {
					spinner.hitSpinner.StopSpinSample()
				} else {
					spinner.hitSpinner.StartSpinSample()
				}
			}

			rotationAddition := state.currentVelocity * timeDiff

			state.rotationCountFD += rotationAddition
			state.rotationCountF += math.Abs(rotationAddition / math.Pi)

			if len(spinner.players) == 1 {
				spinner.hitSpinner.SetRotation(player.diff.GetModifiedTime(state.rotationCountFD))
				spinner.hitSpinner.SetRPM(player.diff.GetModifiedTime(state.rpm))
				spinner.hitSpinner.UpdateCompletion(state.rotationCountF / float64(state.requirement))
			}

			state.rotationCount = int64(state.rotationCountF)

			if state.rotationCount != state.lastRotationCount {
				state.scoringRotationCount++

				if state.scoringRotationCount == state.requirement && len(spinner.players) == 1 {
					spinner.hitSpinner.Clear()
				}

				if state.scoringRotationCount > state.requirement+3 && (state.scoringRotationCount-(state.requirement+3))%2 == 0 {
					if len(spinner.players) == 1 {
						spinner.hitSpinner.Bonus()
					}

					spinner.ruleSet.SendResult(time, player.cursor, spinner.hitSpinner.GetBasicData().Number, spinnerPosition.X, spinnerPosition.Y, SpinnerBonus, true, ComboResults.Hold)
				} else if state.scoringRotationCount > 1 && state.scoringRotationCount%2 == 0 {
					spinner.ruleSet.SendResult(time, player.cursor, spinner.hitSpinner.GetBasicData().Number, spinnerPosition.X, spinnerPosition.Y, SpinnerPoints, true, ComboResults.Hold)
				} else if state.scoringRotationCount > 1 {
					spinner.ruleSet.SendResult(time, player.cursor, spinner.hitSpinner.GetBasicData().Number, spinnerPosition.X, spinnerPosition.Y, SpinnerSpin, true, ComboResults.Hold)
				}

				state.lastRotationCount = state.rotationCount
			}
		}
	}

	return numFinishedTotal == 0
}

func (spinner *Spinner) UpdatePostFor(player *difficultyPlayer, time int64) bool {
	state := spinner.state[player]

	if time >= spinner.hitSpinner.GetBasicData().EndTime && !state.finished {
		hit := Miss
		combo := ComboResults.Reset

		if state.scoringRotationCount > state.requirement+1 {
			hit = Hit300
		} else if state.scoringRotationCount > state.requirement {
			hit = Hit100
		} else if state.scoringRotationCount == state.requirement {
			hit = Hit50
		}

		if hit != Miss {
			combo = ComboResults.Increase
		}

		if len(spinner.players) == 1 {
			spinner.hitSpinner.StopSpinSample()
			spinner.hitSpinner.Hit(time, hit != Miss)
		}

		spinner.ruleSet.SendResult(time, player.cursor, spinner.hitSpinner.GetBasicData().Number, spinner.hitSpinner.GetPosition().X, spinner.hitSpinner.GetPosition().Y, hit, false, combo)

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
	return spinner.hitSpinner.GetBasicData().StartTime - int64(spinner.fadeStartRelative)
}
