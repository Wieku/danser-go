package dance

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/wieku/danser-go/beatmap"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/difficulty"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/rulesets/osu"
	"github.com/wieku/danser-go/settings"
)

type PlayerController struct {
	bMap     *beatmap.BeatMap
	cursors  []*render.Cursor
	window   *glfw.Window
	ruleset  *osu.OsuRuleSet
	lastTime int64
	counter  int64

	leftClick  bool
	rightClick bool
}

func NewPlayerController() Controller {
	return &PlayerController{}
}

func (controller *PlayerController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *PlayerController) InitCursors() {
	controller.cursors = []*render.Cursor{render.NewCursor()}
	controller.cursors[0].IsPlayer = true
	controller.window = glfw.GetCurrentContext()
	controller.ruleset = osu.NewOsuRuleset(controller.bMap, controller.cursors, []difficulty.Modifier{difficulty.None})
	controller.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	controller.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if glfw.GetKeyName(key, scancode) == settings.Input.LeftKey {
			if action == glfw.Press {
				controller.leftClick = true
			} else if action == glfw.Release {
				controller.leftClick = false
			}
		}

		if glfw.GetKeyName(key, scancode) == settings.Input.RightKey {
			if action == glfw.Press {
				controller.rightClick = true
			} else if action == glfw.Release {
				controller.rightClick = false
			}
		}
	})
}

func (controller *PlayerController) Update(time int64, delta float64) {

	controller.bMap.Update(time)

	if controller.window != nil {
		controller.cursors[0].SetScreenPos(bmath.NewVec2d(controller.window.GetCursorPos()).Copy32())

		mouseEnabled := !settings.Input.MosueButtonsDisabled

		controller.cursors[0].LeftButton = controller.leftClick || (mouseEnabled && controller.window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press)
		controller.cursors[0].RightButton = controller.rightClick || (mouseEnabled && controller.window.GetMouseButton(glfw.MouseButtonRight) == glfw.Press)
	}

	controller.counter += time - controller.lastTime

	if controller.counter >= 12 {
		controller.cursors[0].LastFrameTime = time - 12
		controller.cursors[0].CurrentFrameTime = time
		controller.cursors[0].IsReplayFrame = true
		controller.counter -= 12
	} else {
		controller.cursors[0].IsReplayFrame = false
	}

	controller.lastTime = time

	controller.ruleset.UpdateClickFor(controller.cursors[0], time)
	controller.ruleset.UpdateNormalFor(controller.cursors[0], time)
	controller.ruleset.Update(time)

	controller.cursors[0].Update(delta)
}

func (controller *PlayerController) GetRuleset() *osu.OsuRuleSet {
	return controller.ruleset
}

func (controller *PlayerController) GetCursors() []*render.Cursor {
	return controller.cursors
}
