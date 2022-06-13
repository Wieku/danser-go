package launcher

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
)

type modPopup struct {
	*popup

	bld *builder
}

func newModPopup(bld *builder) *modPopup {
	mP := &modPopup{
		popup: newPopup("Mods", popDynamic),
		bld:   bld,
	}

	mP.internalDraw = mP.drawModMenu

	return mP
}

func (m *modPopup) drawModMenu() {
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, imgui.Vec2{10, 10})

	if imgui.BeginTable("mfa", 6) {
		m.drawRow("Reduction:", func() {
			m.modCheckbox(difficulty.Easy, difficulty.HardRock)

			m.modCheckbox(difficulty.NoFail, difficulty.SuddenDeath|difficulty.Perfect|difficulty.Relax|difficulty.Relax2)

			m.modCheckboxMulti(difficulty.HalfTime, difficulty.Daycore, difficulty.DoubleTime|difficulty.Nightcore)
		})

		m.drawRow("Increase:", func() {
			m.modCheckbox(difficulty.HardRock, difficulty.Easy)

			m.modCheckboxMulti(difficulty.SuddenDeath, difficulty.Perfect, difficulty.NoFail|difficulty.Relax|difficulty.Relax2)

			m.modCheckboxMulti(difficulty.DoubleTime, difficulty.Nightcore, difficulty.HalfTime|difficulty.Daycore)

			m.modCheckbox(difficulty.Hidden, difficulty.None)

			m.modCheckbox(difficulty.Flashlight, difficulty.None)
		})

		m.drawRow("Special:", func() {
			nfSD := difficulty.NoFail | difficulty.SuddenDeath | difficulty.Perfect

			m.modCheckbox(difficulty.Relax, difficulty.Relax2|nfSD)

			m.modCheckbox(difficulty.Relax2, difficulty.Relax|difficulty.SpunOut|nfSD)

			m.modCheckbox(difficulty.SpunOut, difficulty.Relax2)

			m.modCheckbox(difficulty.ScoreV2, difficulty.None)
		})

		imgui.EndTable()
	}

	centerTable("modresettable", -1, func() {
		if imgui.Button("Reset##Mods") {
			m.bld.mods = difficulty.None
		}
	})

	imgui.PopStyleVar()
}

func (m *modPopup) drawRow(name string, work func()) {
	imgui.TableNextRow()
	imgui.TableNextColumn()

	posLocal := imgui.CursorPos()

	work()

	posLocal1 := imgui.CursorPos()

	imgui.TableSetColumnIndex(0)

	imgui.SetCursorPos(imgui.Vec2{posLocal.X, (posLocal.Y + posLocal1.Y - imgui.FrameHeightWithSpacing()) / 2})

	imgui.AlignTextToFramePadding()

	imgui.Text(name)
}

func (m *modPopup) modCheckbox(mod, incompat difficulty.Modifier) (ret bool) {
	imgui.TableNextColumn()

	s := m.bld.mods.Active(mod)

	if s {
		cColor := imgui.CurrentStyle().Color(imgui.StyleColorCheckMark)

		imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{cColor.X, cColor.Y, cColor.Z, 0.8})
		imgui.PushStyleColor(imgui.StyleColorButtonActive, imgui.Vec4{cColor.X * 1.2, cColor.Y * 1.2, cColor.Z * 1.2, 0.8})
		imgui.PushStyleColor(imgui.StyleColorButtonHovered, imgui.Vec4{cColor.X * 1.4, cColor.Y * 1.4, cColor.Z * 1.4, 0.8})
	}

	fH := imgui.FrameHeight()

	if ret = imgui.ButtonV(mod.String(), imgui.Vec2{X: fH * 2, Y: fH * 2}); ret {
		if s {
			m.bld.mods &= ^mod
		} else {
			m.bld.mods &= ^incompat
			m.bld.mods |= mod
		}
	}

	if s {
		imgui.PopStyleColor()
		imgui.PopStyleColor()
		imgui.PopStyleColor()
	}

	if imgui.IsItemHovered() || imgui.IsItemActive() {
		imgui.BeginTooltip()

		sF := mod.StringFull()[0]
		if sF == "Relax2" {
			sF = "AutoPilot"
		}

		imgui.SetTooltip(sF)
		imgui.EndTooltip()
	}

	return
}

func (m *modPopup) modCheckboxMulti(mod1, mod2, incompat difficulty.Modifier) {
	if !m.bld.mods.Active(mod2) {
		if m.modCheckbox(mod1, incompat) && !m.bld.mods.Active(mod1) {
			m.bld.mods |= mod2
		}
	} else {
		m.modCheckbox(mod2, incompat)
	}
}
