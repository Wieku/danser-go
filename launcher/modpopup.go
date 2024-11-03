package launcher

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"strings"
)

type modPopup struct {
	*popup

	bld *builder

	firstCalc     bool
	settingsDrawn bool
}

func newModPopup(bld *builder) *modPopup {
	mP := &modPopup{
		popup:     newPopup("Mods", popCustom),
		bld:       bld,
		firstCalc: true,
	}

	mP.addFlags = imgui.WindowFlagsAlwaysVerticalScrollbar

	mP.internalDraw = mP.drawModMenu

	return mP
}

func (m *modPopup) drawModMenu() {
	handleDragScroll()
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, vec2(10, 10))

	if imgui.BeginTable("mfa", 6) {
		m.drawRow("Reduction:", func() {
			m.modCheckbox(difficulty.Easy, difficulty.HardRock, difficulty.None)

			m.modCheckbox(difficulty.NoFail, difficulty.SuddenDeath|difficulty.Perfect|difficulty.Relax|difficulty.Relax2, difficulty.None)

			m.modCheckboxMulti(difficulty.HalfTime, difficulty.Daycore, difficulty.DoubleTime|difficulty.Nightcore, difficulty.None)
		})

		m.drawRow("Increase:", func() {
			m.modCheckbox(difficulty.HardRock, difficulty.Easy, difficulty.None)

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

			m.modCheckbox(difficulty.DifficultyAdjust, difficulty.None, difficulty.None)
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
			m.bld.diff = m.bld.sourceDiff.Clone()
			m.bld.baseDiff.SetMods(m.bld.sourceDiff.Mods)
		}
	})

	if m.firstCalc {
		m.height = contentRegionMin().Y + imgui.CurrentStyle().WindowBorderSize()

		imgui.CurrentContext().CurrentWindow().SetSize(vec2(0, m.height))

		m.firstCalc = false
	}

	imgui.PopStyleVar()

	m.drawModSettings()
}

func (m *modPopup) drawModSettings() {
	m.settingsDrawn = false
	m.tryDrawSpeedSettings()
	m.tryDrawEasySettings()
	m.tryDrawClassicSettings()
	m.tryDrawFlashlightSettings()
	m.tryDrawDASettings()
}

func (m *modPopup) tryDrawSpeedSettings() {
	speedAdjust := difficulty.DoubleTime | difficulty.Nightcore | difficulty.HalfTime | difficulty.Daycore

	m.drawSettingsBase(speedAdjust, func() {
		var minV, maxV = 1.01, 2.0
		if m.bld.diff.CheckModActive(difficulty.HalfTime | difficulty.Daycore) {
			minV, maxV = 0.5, 0.99
		}

		conf, _ := difficulty.GetModConfig[difficulty.SpeedSettings](m.bld.diff)

		sliderFloatReset2("Speed", m.bld.baseDiff.GetSpeed(), &conf.SpeedChange, minV, maxV, "%.2f")

		if !m.bld.diff.CheckModActive(difficulty.Daycore | difficulty.Nightcore) {
			checkboxOption("Adjust pitch", &conf.AdjustPitch)
		} else {
			conf.AdjustPitch = true
		}

		difficulty.SetModConfig(m.bld.diff, conf)
	})
}

func (m *modPopup) tryDrawEasySettings() {
	m.drawSettingsBase(difficulty.Easy, func() {
		conf, _ := difficulty.GetModConfig[difficulty.EasySettings](m.bld.diff)

		sliderIntReset2("Extra lives", 2, &conf.Retries, 0, 10, "%d")

		difficulty.SetModConfig(m.bld.diff, conf)
	})
}

func (m *modPopup) tryDrawClassicSettings() {
	m.drawSettingsBase(difficulty.Classic, func() {
		conf, _ := difficulty.GetModConfig[difficulty.ClassicSettings](m.bld.diff)

		checkboxOption("No slider head accuracy requirement", &conf.NoSliderHeadAccuracy)
		checkboxOption("Apply classic note lock", &conf.ClassicNoteLock)
		checkboxOption("Always play a slider's tail sample", &conf.AlwaysPlayTailSample)
		checkboxOption("Classic health", &conf.ClassicHealth)

		difficulty.SetModConfig(m.bld.diff, conf)
	})
}

func (m *modPopup) tryDrawFlashlightSettings() {
	m.drawSettingsBase(difficulty.Flashlight, func() {
		conf, _ := difficulty.GetModConfig[difficulty.FlashlightSettings](m.bld.diff)

		sliderFloatResetStep2("Follow delay", 120, &conf.FollowDelay, 120, 1200, 120, "%.f")
		sliderFloatReset2("Size multiplier", 1, &conf.SizeMultiplier, 0.5, 2, "%.1f")
		checkboxOption("Combo based size", &conf.ComboBasedSize)

		difficulty.SetModConfig(m.bld.diff, conf)
	})
}

func (m *modPopup) tryDrawDASettings() {
	m.drawSettingsBase(difficulty.DifficultyAdjust, func() {
		conf, _ := difficulty.GetModConfig[difficulty.DiffAdjustSettings](m.bld.diff)

		arCSMin, vMax := 0.0, 10.0
		if conf.ExtendedValues {
			arCSMin, vMax = -10, 12
		}

		sliderFloatReset2("Approach Rate (AR)", m.bld.currentMap.Diff.GetBaseAR(), &conf.ApproachRate, arCSMin, vMax, "%.1f")
		sliderFloatReset2("Overall Difficulty (OD)", m.bld.currentMap.Diff.GetBaseOD(), &conf.OverallDifficulty, 0, vMax, "%.1f")
		sliderFloatReset2("Circle Size (CS)", m.bld.currentMap.Diff.GetBaseCS(), &conf.CircleSize, arCSMin, vMax, "%.1f")
		sliderFloatReset2("Health Drain (HP)", m.bld.currentMap.Diff.GetBaseHP(), &conf.DrainRate, 0, vMax, "%.1f")

		checkboxOption("Extended values", &conf.ExtendedValues)

		difficulty.SetModConfig(m.bld.diff, conf)
	})
}

func (m *modPopup) drawSettingsBase(mask difficulty.Modifier, draw func()) {
	if m.bld.diff.CheckModActive(mask) {
		if m.settingsDrawn {
			imgui.Dummy(vec2(0, 10))
		}

		imgui.PushFont(Font32)
		imgui.TextUnformatted((m.bld.diff.Mods & mask).StringFull()[0] + ":")
		imgui.PopFont()
		imgui.WindowDrawList().AddLine(imgui.CursorScreenPos(), imgui.CursorScreenPos().Add(vec2(contentRegionMax().X, 0)), packColor(*imgui.StyleColorVec4(imgui.ColSeparator)))

		imgui.Spacing()

		draw()

		m.settingsDrawn = true
	}
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
			m.bld.baseDiff.RemoveMod(mod)
			m.bld.diff.RemoveMod(mod)
		} else {
			m.bld.baseDiff.RemoveMod(incompat)
			m.bld.diff.RemoveMod(incompat)

			m.bld.baseDiff.AddMod(mod)
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

		modTip := mod.StringFull()[0]
		if modTip == "Relax2" {
			modTip = "AutoPilot"
		}

		imgui.TextUnformatted(modTip)

		imgui.PushFont(Font20)
		imgui.PushTextWrapPosV(300)

		if !req {
			imgui.TextUnformatted("Mods required: " + strings.Join(required.StringFull(), ", "))
		}

		if incompat != difficulty.None {
			imgui.TextUnformatted("Incompatible with: " + strings.Join(incompat.StringFull2(), ", "))
		}

		imgui.PopTextWrapPos()
		imgui.PopFont()

		imgui.EndTooltip()
	}

	return
}

func (m *modPopup) modCheckboxMulti(mod1, mod2, incompat, required difficulty.Modifier) {
	if !m.bld.diff.CheckModActive(mod2) {
		if m.modCheckbox(mod1, incompat, required) && !m.bld.diff.CheckModActive(mod1) {
			m.bld.baseDiff.AddMod(mod2)
			m.bld.diff.AddMod(mod2)
		}
	} else {
		m.modCheckbox(mod2, incompat, required)
	}
}
