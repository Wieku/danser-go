package osu

import (
	"github.com/wieku/danser/beatmap/objects"
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
	frameVariance        float64
	theoreticalVelocity  float64
	currentVelocity      float64
	finished             bool
	zeroCount            int64
}

type Spinner struct {
	ruleSet           *OsuRuleSet
	hitSpinner        *objects.Spinner
	players           []*difficultyPlayer
	state             []*spinnerstate
	fadeStartRelative float64
	maxAcceleration   float64
}

var spinners = 0

func (spinner *Spinner) Init(ruleSet *OsuRuleSet, object objects.BaseObject, players []*difficultyPlayer) {
	spinner.ruleSet = ruleSet
	spinner.hitSpinner = object.(*objects.Spinner)
	spinner.players = players
	spinner.state = make([]*spinnerstate, len(players))

	rSpinner := object.(*objects.Spinner)

	spinnerTime := rSpinner.GetBasicData().EndTime - rSpinner.GetBasicData().StartTime

	spinner.fadeStartRelative = 100000

	for i, player := range spinner.players {
		spinner.state[i] = new(spinnerstate)
		spinner.fadeStartRelative = math.Min(spinner.fadeStartRelative, player.diff.Preempt)
		spinners++
		spinner.state[i].requirement = int64(float64(spinnerTime) / 1000 * player.diff.SpinnerRatio)
		spinner.state[i].frameVariance = FrameTime
	}

	spinner.maxAcceleration = 0.00008 + math.Max(0, (5000-float64(spinnerTime))/1000/2000)
}

func (spinner *Spinner) Update(time int64) bool {
	numFinishedTotal := 0

	spinnerPosition := spinner.hitSpinner.GetBasicData().StartPos

	for i, player := range spinner.players {
		state := spinner.state[i]
		timeDiff := time - player.cursor.LastFrameTime

		if !state.finished {
			numFinishedTotal++
			if player.cursor.IsReplayFrame {

				if (player.cursorLock == -1 || player.cursorLock == spinner.hitSpinner.GetBasicData().Number) && timeDiff > 0 {

					mouseAngle := player.cursor.Position.Sub(spinnerPosition).AngleR()

					angleDiff := mouseAngle - state.lastAngle

					if mouseAngle-state.lastAngle < -math.Pi {
						angleDiff = (2 * math.Pi) + mouseAngle - state.lastAngle
					} else if state.lastAngle-mouseAngle < -math.Pi {
						angleDiff = (-2 * math.Pi) - state.lastAngle + mouseAngle
					}

					decay := math.Pow(0.999, float64(timeDiff))
					state.frameVariance = decay*state.frameVariance + (1-decay)*float64(timeDiff)

					if angleDiff == 0 {
						state.zeroCount += 1

						if state.zeroCount < 2 {
							state.theoreticalVelocity /= 3
						} else {
							state.theoreticalVelocity = 0
						}
					} else {
						state.zeroCount = 0

						if (!player.cursor.LeftButton && !player.cursor.RightButton) || time < spinner.hitSpinner.GetBasicData().StartTime || time > spinner.hitSpinner.GetBasicData().EndTime {
							angleDiff = 0
						}

						if math.Abs(angleDiff) < math.Pi {
							if player.diff.GetModifiedTime(state.frameVariance) > FrameTime*1.04 {
								state.theoreticalVelocity = angleDiff / player.diff.GetModifiedTime(float64(timeDiff))
							} else {
								state.theoreticalVelocity = angleDiff / FrameTime
							}
						} else {
							state.theoreticalVelocity = 0
						}
					}

					state.lastAngle = mouseAngle

					maxAccelThisFrame := player.diff.GetModifiedTime(spinner.maxAcceleration * float64(timeDiff))

					if state.theoreticalVelocity > state.currentVelocity {
						state.currentVelocity += math.Min(state.theoreticalVelocity-state.currentVelocity, maxAccelThisFrame)
					} else {
						state.currentVelocity += math.Max(state.theoreticalVelocity-state.currentVelocity, -maxAccelThisFrame)
					}

					state.currentVelocity = math.Max(-0.05, math.Min(state.currentVelocity, 0.05))

					rotationAddition := state.currentVelocity * float64(timeDiff)

					state.rotationCountF += math.Abs(rotationAddition / math.Pi)
					state.rotationCount = int64(state.rotationCountF)

					if state.rotationCount != state.lastRotationCount {

						state.scoringRotationCount++

						if state.scoringRotationCount > state.requirement+3 && (state.scoringRotationCount-(state.requirement+3))%2 == 0 {
							spinner.ruleSet.SendResult(time, player.cursor, spinnerPosition.X, spinnerPosition.Y, HitResults.SpinnerBonus, true, ComboResults.Hold)
						} else if state.scoringRotationCount > 1 && state.scoringRotationCount%2 == 0 {
							spinner.ruleSet.SendResult(time, player.cursor, spinnerPosition.X, spinnerPosition.Y, HitResults.SpinnerScore, true, ComboResults.Hold)
						} else if state.scoringRotationCount > 1 {
							//hp inpact in the future
						}

						state.lastRotationCount = state.rotationCount
					}

					player.cursorLock = spinner.hitSpinner.GetBasicData().Number
				}
			}

			if time >= spinner.hitSpinner.GetBasicData().EndTime {

				hit := HitResults.Miss
				combo := ComboResults.Reset

				if state.scoringRotationCount > state.requirement+1 {
					hit = HitResults.Hit300
				} else if state.scoringRotationCount > state.requirement {
					hit = HitResults.Hit100
				} else if state.scoringRotationCount == state.requirement {
					hit = HitResults.Hit50
				}

				if hit != HitResults.Miss {
					combo = ComboResults.Increase
				}

				spinner.ruleSet.SendResult(time, player.cursor, spinner.hitSpinner.GetPosition().X, spinner.hitSpinner.GetPosition().Y, hit, false, combo)

				player.cursorLock = -1
				state.finished = true
				continue
			}
		}
	}

	return numFinishedTotal == 0
}

func (spinner *Spinner) GetFadeTime() int64 {
	return spinner.hitSpinner.GetBasicData().StartTime - int64(spinner.fadeStartRelative)
}
