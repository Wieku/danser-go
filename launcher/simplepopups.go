package launcher

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/platform"
	"github.com/wieku/danser-go/framework/util"
	"strconv"
)

func drawTimeMenu(bld *builder) {
	imgui.Checkbox("Skip map's beginning", &bld.skip)

	start := &bld.start
	end := &bld.end

	sliderIntReset("Start time", start, 0, end.ogValue-1, util.FormatSeconds(int(start.value)))

	if start.value >= end.value {
		end.value = start.value + 1
	}

	iPos1 := imgui.CursorPosY()

	sliderIntReset("End time", end, 1, end.ogValue, util.FormatSeconds(int(end.value)))

	iPos2 := imgui.CursorPosY()

	if start.value >= end.value {
		start.value = end.value - 1
	}

	imgui.Dummy(imgui.Vec2{0, iPos2 - iPos1})

	sliderIntReset("Audio offset", &bld.offset, -300, 300, "%dms")
}

func drawSpeedMenu(bld *builder) {
	sliderFloatResetStep("Speed", &bld.speed, 0.1, 3, 0.05, "%.2f")
	imgui.Spacing()

	sliderFloatResetStep("Pitch", &bld.pitch, 0.1, 3, 0.05, "%.2f")
	imgui.Spacing()
}

func drawParamMenu(bld *builder) {
	sliderFloatReset("Approach Rate (AR)", &bld.ar, 0, 10, "%.1f")
	imgui.Spacing()

	if bld.currentMode == Play || bld.currentMode == DanserReplay {
		sliderFloatReset("Overall Difficulty (OD)", &bld.od, 0, 10, "%.1f")
		imgui.Spacing()
	}

	sliderFloatReset("Circle Size (CS)", &bld.cs, 0, 10, "%.1f")
	imgui.Spacing()

	if bld.currentMode == Play || bld.currentMode == DanserReplay {
		sliderFloatReset("Health Drain (HP)", &bld.hp, 0, 10, "%.1f")
		imgui.Spacing()
	}
}

func drawCDMenu(bld *builder) {
	if imgui.BeginTable("dfa", 2) {
		imgui.TableNextColumn()

		imgui.Text("Mirrored cursors:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(imgui.TextLineHeight() * 4.5)

		if imgui.InputIntV("##mirrors", &bld.mirrors, 1, 1, 0) {
			if bld.mirrors < 1 {
				bld.mirrors = 1
			}
		}

		imgui.TableNextColumn()

		imgui.Text("Tag cursors:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(imgui.TextLineHeight() * 4.5)

		if imgui.InputIntV("##tags", &bld.tags, 1, 1, 0) {
			if bld.tags < 1 {
				bld.tags = 1
			}
		}

		imgui.EndTable()
	}
}

func drawRecordMenu(bld *builder) {
	if imgui.BeginTable("rfa", 2) {
		imgui.TableNextColumn()

		imgui.Text("Output name:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(imgui.TextLineHeight() * 10)

		imgui.InputText("##oname", &bld.outputName)

		if bld.currentPMode == Screenshot {
			imgui.TableNextColumn()

			imgui.Text("Screenshot at:")

			imgui.TableNextColumn()

			imgui.SetNextItemWidth(imgui.TextLineHeight() * 10)

			valText := strconv.FormatFloat(float64(bld.ssTime), 'f', 3, 64)
			prevText := valText

			if imgui.InputText("##sstime", &valText) {
				parsed, err := strconv.ParseFloat(valText, 64)
				if err != nil {
					valText = prevText
				} else {
					parsed = mutils.ClampF(parsed, 0, float64(bld.end.ogValue))
					bld.ssTime = float32(parsed)
				}
			}
		}

		imgui.EndTable()
	}
}

func drawAbout(dTex texture.Texture) {
	centerTable("about1", -1, func() {
		imgui.Image(imgui.TextureID(dTex.GetID()), imgui.Vec2{100, 100})
	})

	centerTable("about2", -1, func() {
		imgui.Text("danser-go " + build.VERSION)
	})

	centerTable("about3", -1, func() {
		if imgui.Button("Check for updates") {
			checkForUpdates(true)
		}
	})

	imgui.Dummy(imgui.Vec2{1, imgui.FrameHeight()})

	centerTable("about4.1", -1, func() {
		imgui.Text("Advanced visualisation multi-tool")
	})

	centerTable("about4.2", -1, func() {
		imgui.Text("for osu!")
	})

	imgui.Dummy(imgui.Vec2{1, imgui.FrameHeight()})

	if imgui.BeginTableV("about5", 3, imgui.TableFlagsSizingStretchSame, imgui.Vec2{-1, 0}, -1) {
		imgui.TableNextColumn()

		centerTable("aboutgithub", -1, func() {
			if imgui.Button("GitHub") {
				platform.OpenURL("https://wieku.me/danser")
			}
		})

		imgui.TableNextColumn()

		centerTable("aboutdonate", -1, func() {
			if imgui.Button("Donate") {
				platform.OpenURL("https://wieku.me/donate")
			}
		})

		imgui.TableNextColumn()

		centerTable("aboutdiscord", -1, func() {
			if imgui.Button("Discord") {
				platform.OpenURL("https://wieku.me/lair")
			}
		})

		imgui.EndTable()
	}
}

func drawLauncherConfig() {
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, imgui.Vec2{imgui.CurrentStyle().CellPadding().X, 10})

	if imgui.BeginTableV("lconfigtable", 2, 0, imgui.Vec2{-1, 0}, -1) {
		imgui.TableSetupColumnV("1lconfigtable", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
		imgui.TableSetupColumnV("2lconfigtable", imgui.TableColumnFlagsWidthFixed, 0, uint(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Check for updates on startup")

		imgui.TableNextColumn()

		imgui.Checkbox("##CheckForUpdates", &launcherConfig.CheckForUpdates)

		imgui.TableNextColumn()

		posLocal := imgui.CursorPos()

		imgui.AlignTextToFramePadding()
		imgui.Text("Show exported videos/images\nin explorer")

		posLocal1 := imgui.CursorPos()

		imgui.TableNextColumn()

		imgui.SetCursorPos(imgui.Vec2{imgui.CursorPosX(), (posLocal.Y + posLocal1.Y - imgui.FrameHeightWithSpacing()) / 2})

		imgui.Checkbox("##ShowFileAfter", &launcherConfig.ShowFileAfter)

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Preview selected maps")

		imgui.TableNextColumn()

		imgui.Checkbox("##PreviewSelected", &launcherConfig.PreviewSelected)

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Load latest replay on startup")

		imgui.TableNextColumn()

		imgui.Checkbox("##LoadLatestReplay", &launcherConfig.LoadLatestReplay)

		imgui.EndTable()
	}

	imgui.PopStyleVar()
}
