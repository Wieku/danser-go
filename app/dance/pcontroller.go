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
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
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

	leftClick       bool
	rightClick      bool
	relaxController *input.RelaxInputProcessor
	mouseController schedulers.Scheduler
}

func NewPlayerController() Controller {
	return &PlayerController{}
}

func (controller *PlayerController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *PlayerController) InitCursors() {
	controller.cursors = []*graphics.Cursor{graphics.NewCursor()}
	controller.cursors[0].IsPlayer = true
	controller.cursors[0].ScoreTime = time.Now()
	controller.window = glfw.GetCurrentContext()
	controller.ruleset = osu.NewOsuRuleset(controller.bMap, controller.cursors, []difficulty.Modifier{controller.bMap.Diff.Mods})
	controller.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	if !controller.bMap.Diff.CheckModActive(difficulty.Relax) {
		controller.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if strings.EqualFold(glfw.GetKeyName(key, scancode), settings.Input.LeftKey) {
				if action == glfw.Press {
					controller.cursors[0].LeftKey = true
				} else if action == glfw.Release {
					controller.cursors[0].LeftKey = false
				}
			}

			if strings.EqualFold(glfw.GetKeyName(key, scancode), settings.Input.RightKey) {
				if action == glfw.Press {
					controller.cursors[0].RightKey = true
				} else if action == glfw.Release {
					controller.cursors[0].RightKey = false
				}
			}
		})
	} else {
		controller.relaxController = input.NewRelaxInputProcessor(controller.ruleset, controller.cursors[0])
	}

	if controller.bMap.Diff.CheckModActive(difficulty.Relax2) {
		controller.mouseController = schedulers.NewGenericScheduler(movers.NewLinearMover)
		controller.mouseController.Init(controller.bMap.GetObjectsCopy(), controller.bMap.Diff.Mods, controller.cursors[0], spinners.GetMoverCtorByName("circle"), false)
	}
}

func (controller *PlayerController) Update(time float64, delta float64) {
	controller.bMap.Update(time)

	if controller.window != nil {
		if !controller.bMap.Diff.CheckModActive(difficulty.Relax2) {
			controller.cursors[0].SetScreenPos(vector.NewVec2d(controller.window.GetCursorPos()).Copy32())
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
	}

	controller.counter += time - controller.lastTime

	if controller.counter >= 1000.0/60 {
		controller.cursors[0].IsReplayFrame = true
		controller.counter -= 1000.0 / 60
	} else {
		controller.cursors[0].IsReplayFrame = false
	}

	controller.ruleset.UpdateClickFor(controller.cursors[0], int64(time))
	controller.ruleset.UpdateNormalFor(controller.cursors[0], int64(time))
	controller.ruleset.UpdatePostFor(controller.cursors[0], int64(time))
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
