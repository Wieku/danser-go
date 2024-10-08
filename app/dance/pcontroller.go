package dance

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/dance/input"
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/dance/schedulers"
	"github.com/wieku/danser-go/app/dance/spinners"
	"github.com/wieku/danser-go/app/graphics"
	input2 "github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/platform"
	"log"
	"strings"
	"time"
)

type PlayerController struct {
	bMap     *beatmap.BeatMap
	cursors  []*graphics.Cursor
	window   *glfw.Window
	ruleset  *osu.OsuRuleSet
	lastTime float64
	counter  float64

	relaxController *input.RelaxInputProcessor
	mouseController schedulers.Scheduler
	firstTime       bool
	previousPos     vector.Vector2f
	position        vector.Vector2f

	rawInput bool
	inside   bool

	quickRestart     bool
	quickRestartTime float64
}

func NewPlayerController() Controller {
	return &PlayerController{firstTime: true}
}

func (controller *PlayerController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *PlayerController) InitCursors() {
	controller.cursors = []*graphics.Cursor{graphics.NewCursor()}
	controller.cursors[0].IsPlayer = true
	controller.cursors[0].Name = settings.Gameplay.PlayUsername
	controller.cursors[0].ScoreTime = time.Now()
	controller.window = glfw.GetCurrentContext()
	controller.ruleset = osu.NewOsuRuleset(controller.bMap, controller.cursors, []*difficulty.Difficulty{controller.bMap.Diff.Clone()})

	if !controller.bMap.Diff.CheckModActive(difficulty.Relax) {
		input2.RegisterListener(controller.KeyEvent)
	} else {
		controller.relaxController = input.NewRelaxInputProcessor(controller.ruleset, controller.cursors[0])
	}

	controller.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	if controller.bMap.Diff.CheckModActive(difficulty.Relax2) {
		controller.mouseController = schedulers.NewGenericScheduler(movers.NewLinearMoverSimple, 0, 0)
		controller.mouseController.Init(controller.bMap.GetObjectsCopy(), controller.bMap.Diff, controller.cursors[0], spinners.GetMoverCtorByName("circle"), false)
	} else if settings.Input.MouseHighPrecision {
		if glfw.RawMouseMotionSupported() {
			controller.rawInput = true
			controller.window.SetInputMode(glfw.RawMouseMotion, glfw.True)
		} else {
			log.Println("InputManager: Raw input not supported!")
		}
	}
}

func (controller *PlayerController) KeyEvent(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, _ glfw.ModifierKey) {
	kName, ok := platform.GetKeyName(key, scancode)
	if !ok {
		return
	}

	if strings.EqualFold(kName, settings.Input.LeftKey) {
		if action == glfw.Press {
			controller.cursors[0].LeftKey = true
		} else if action == glfw.Release {
			controller.cursors[0].LeftKey = false
		}
	}

	if strings.EqualFold(kName, settings.Input.RightKey) {
		if action == glfw.Press {
			controller.cursors[0].RightKey = true
		} else if action == glfw.Release {
			controller.cursors[0].RightKey = false
		}
	}

	if strings.EqualFold(kName, settings.Input.RestartKey) {
		if action == glfw.Press {
			controller.quickRestartTime = controller.lastTime
			controller.quickRestart = true
		} else if action == glfw.Release {
			controller.quickRestart = false
		}
	}

	if strings.EqualFold(kName, settings.Input.SmokeKey) {
		if action == glfw.Press {
			controller.cursors[0].SmokeKey = true
		} else if action == glfw.Release {
			controller.cursors[0].SmokeKey = false
		}
	}
}

func (controller *PlayerController) Update(time float64, delta float64) {
	controller.bMap.Update(time)

	if controller.window != nil {
		if !controller.bMap.Diff.CheckModActive(difficulty.Relax2) {
			mousePosition := vector.NewVec2d(controller.window.GetCursorPos()).Copy32()

			if controller.rawInput {
				controller.updateRaw(mousePosition)
			} else {
				controller.position = mousePosition
			}

			controller.cursors[0].SetScreenPos(controller.position)
		} else {
			controller.mouseController.Update(time)
		}

		if !controller.bMap.Diff.CheckModActive(difficulty.Relax) {
			mouseEnabled := !settings.Input.MouseButtonsDisabled

			controller.cursors[0].LeftMouse = mouseEnabled && controller.window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press
			controller.cursors[0].RightMouse = mouseEnabled && controller.window.GetMouseButton(glfw.MouseButtonRight) == glfw.Press

			controller.cursors[0].LeftButton = controller.cursors[0].LeftKey || controller.cursors[0].LeftMouse
			controller.cursors[0].RightButton = controller.cursors[0].RightKey || controller.cursors[0].RightMouse
		} else {
			controller.relaxController.Update(time)
		}

		if controller.quickRestart && time-controller.quickRestartTime > 500 {
			controller.quickRestart = false

			utils.QuickRestart()
		}
	}

	controller.counter += time - controller.lastTime

	if controller.counter >= 1000.0/60 {
		controller.cursors[0].IsReplayFrame = true
		controller.counter -= 1000.0 / 60
	} else {
		controller.cursors[0].IsReplayFrame = false
	}

	controller.ruleset.UpdateClickFor(controller.cursors[0], int64(time))
	controller.ruleset.UpdateNormalFor(controller.cursors[0], int64(time), false)
	controller.ruleset.UpdatePostFor(controller.cursors[0], int64(time), false)
	controller.ruleset.Update(int64(time))

	controller.lastTime = time

	controller.cursors[0].Update(delta)
}

func (controller *PlayerController) GetRuleset() *osu.OsuRuleSet {
	return controller.ruleset
}

func (controller *PlayerController) GetCursors() []*graphics.Cursor {
	return controller.cursors
}

func (controller *PlayerController) updateRaw(mousePos vector.Vector2f) {
	hovered := controller.window.GetAttrib(glfw.Hovered) == 1

	if controller.firstTime {
		controller.previousPos = vector.NewVec2d(controller.window.GetCursorPos()).Copy32()
		controller.position = controller.previousPos
		controller.firstTime = false

		if hovered && input2.Focused {
			controller.setRawStatus(true)
		} else {
			controller.setRawStatus(false)
		}
	}

	if controller.inside {
		direction := mousePos.Sub(controller.previousPos).Scl(float32(settings.Input.MouseSensitivity))
		controller.position = controller.position.Add(direction)
		controller.previousPos = controller.position
	} else {
		controller.position = mousePos
	}

	if controller.inside &&
		(controller.position.X < 0 || controller.position.X64() > settings.Graphics.GetWidthF() ||
			controller.position.Y < 0 || controller.position.Y64() > settings.Graphics.GetHeightF() || !hovered) {
		controller.setRawStatus(false)
	} else if input2.Focused && hovered && !controller.inside {
		controller.setRawStatus(true)
	}

	controller.previousPos = mousePos
}

func (controller *PlayerController) setRawStatus(state bool) {
	goroutines.CallMain(func() {
		if state {
			log.Println("InputManager: Switching to raw input mode")

			controller.position = vector.NewVec2d(controller.window.GetCursorPos()).Copy32()

			controller.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

			controller.previousPos = vector.NewVec2d(controller.window.GetCursorPos()).Copy32()
		} else {
			log.Println("InputManager: Switching to normal input mode")

			controller.previousPos = controller.position

			controller.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

			for i := 0; i < 20; i++ {
				controller.window.SetCursorPos(controller.position.X64(), controller.position.Y64())
			}
		}
	})

	controller.inside = state
}
