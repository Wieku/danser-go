package launcher

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"strings"
)

type modPopup struct {
	*popup

	bld *builder

	baseDiff *difficulty.Difficulty

	firstCalc bool
}

func newModPopup(bld *builder) *modPopup {
	mP := &modPopup{
		popup:     newPopup("Mods", popCustom),
		bld:       bld,
		baseDiff:  bld.currentMap.Diff.Clone(),
		firstCalc: true,
	}

	mP.baseDiff.SetMods(bld.diff.Mods)

	mP.internalDraw = mP.drawModMenu

	return mP
}

func (m *modPopup) drawModMenu() {
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, vec2(10, 10))

	if imgui.BeginTable("mfa", 6) {
		m.drawRow("Reduction:", func() {
			m.modCheckbox(difficulty.Easy, difficulty.HardRock|difficulty.DifficultyAdjust, difficulty.None)

			m.modCheckbox(difficulty.NoFail, difficulty.SuddenDeath|difficulty.Perfect|difficulty.Relax|difficulty.Relax2, difficulty.None)

			m.modCheckboxMulti(difficulty.HalfTime, difficulty.Daycore, difficulty.DoubleTime|difficulty.Nightcore, difficulty.None)
		})

		m.drawRow("Increase:", func() {
			m.modCheckbox(difficulty.HardRock, difficulty.Easy|difficulty.DifficultyAdjust, difficulty.None)

			m.modCheckboxMulti(difficulty.SuddenDeath, difficulty.Perfect, difficulty.NoFail|difficulty.Relax|difficulty.Relax2, difficulty.None)

			m.modCheckboxMulti(difficulty.DoubleTime, difficulty.Nightcore, difficulty.HalfTime|difficulty.Daycore, difficulty.None)

			m.modCheckbox(difficulty.Hidden, difficulty.None, difficulty.None)

			m.modCheckbox(difficulty.Flashlight, difficulty.None, difficulty.None)
		})

		m.drawRow("Special:", func() {
			nfSD := difficulty.NoFail | difficulty.SuddenDeath | difficulty.Perfect

			m.modCheckbox(difficulty.Relax, difficulty.Relax2|nfSD, difficulty.None)

			m.modCheckbox(difficulty.Relax2, difficulty.Relax|difficulty.SpunOut|nfSD, difficulty.None)

			m.modCheckbox(difficulty.SpunOut, difficulty.Relax2, difficulty.None)

			m.modCheckbox(difficulty.DifficultyAdjust, difficulty.Easy|difficulty.HardRock, difficulty.None)
		})

		m.drawRow("Conversion:", func() {
			m.modCheckbox(difficulty.ScoreV2, difficulty.Lazer|difficulty.Classic, difficulty.None)
			m.modCheckbox(difficulty.Lazer, difficulty.ScoreV2, difficulty.None)
			m.modCheckbox(difficulty.Classic, difficulty.ScoreV2, difficulty.Lazer)
		})

		imgui.EndTable()
	}

	centerTable("modresettable", -1, func() {
		if imgui.Button("Reset##Mods") {
			m.bld.diff.RemoveMod(^difficulty.None)
			m.baseDiff.RemoveMod(^difficulty.None)
		}
	})

	if m.firstCalc {
		cStyle := imgui.CurrentContext().Style()
		bSize := (&cStyle).WindowBorderSize()

		m.height = contentRegionMin().Y + bSize
		imgui.CurrentContext().CurrentWindow().SetSize(vec2(0, m.height))

		m.firstCalc = false
	}

	imgui.PopStyleVar()
}

func (m *modPopup) drawRow(name string, work func()) {
	imgui.TableNextRow()
	imgui.TableNextColumn()

	posLocal := imgui.CursorPos()

	work()

	posLocal1 := imgui.CursorPos()

	imgui.TableSetColumnIndex(0)

	imgui.SetCursorPos(vec2(posLocal.X, (posLocal.Y+posLocal1.Y-imgui.FrameHeightWithSpacing())/2))

	imgui.AlignTextToFramePadding()

	imgui.TextUnformatted(name)
}

func (m *modPopup) modCheckbox(mod, incompat, required difficulty.Modifier) (ret bool) {
	imgui.TableNextColumn()

	req := required == difficulty.None || m.bld.diff.CheckModActive(required)

	if !req {
		imgui.BeginDisabled()
	}

	s := m.bld.diff.CheckModActive(mod)

	if s {
		cColor := *imgui.StyleColorVec4(imgui.ColCheckMark)

		imgui.PushStyleColorVec4(imgui.ColButton, vec4(cColor.X, cColor.Y, cColor.Z, 0.8))
		imgui.PushStyleColorVec4(imgui.ColButtonActive, vec4(cColor.X*1.2, cColor.Y*1.2, cColor.Z*1.2, 0.8))
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, vec4(cColor.X*1.4, cColor.Y*1.4, cColor.Z*1.4, 0.8))
	}

	fH := imgui.FrameHeight()

	if ret = imgui.ButtonV(mod.String(), imgui.Vec2{X: fH * 2, Y: fH * 2}); ret {
		if s {
			m.baseDiff.RemoveMod(mod)
			m.bld.diff.RemoveMod(mod)
		} else {
			m.baseDiff.RemoveMod(incompat)
			m.bld.diff.RemoveMod(incompat)

			m.baseDiff.AddMod(mod)
			m.bld.diff.AddMod(mod)
		}
	}

	if s {
		imgui.PopStyleColor()
		imgui.PopStyleColor()
		imgui.PopStyleColor()
	}

	if !req {
		imgui.EndDisabled()
	}

	if imgui.IsItemHoveredV(imgui.HoveredFlagsAllowWhenDisabled) || imgui.IsItemActive() {
		imgui.BeginTooltip()

		ttip := mod.StringFull()[0]
		if ttip == "Relax2" {
			ttip = "AutoPilot"
		}

		if !req {
			if ttip != "" {
				ttip += "\n"
			}

			ttip += "Mods required: " + strings.Join(required.StringFull(), ", ")
		}

		imgui.SetTooltip(ttip)
		imgui.EndTooltip()
	}

	return
}

func (m *modPopup) modCheckboxMulti(mod1, mod2, incompat, required difficulty.Modifier) {
	if !m.bld.diff.CheckModActive(mod2) {
		if m.modCheckbox(mod1, incompat, required) && !m.bld.diff.CheckModActive(mod1) {
			m.baseDiff.AddMod(mod2)
			m.bld.diff.AddMod(mod2)
		}
	} else {
		m.modCheckbox(mod2, incompat, required)
	}
}
