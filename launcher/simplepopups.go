package launcher

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/utils"
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

	imgui.Dummy(vec2(0, iPos2-iPos1))

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
		imgui.TableSetupColumnV("c1rfa", imgui.TableColumnFlagsWidthFixed, 0, uint(0))
		imgui.TableSetupColumnV("c2rfa", imgui.TableColumnFlagsWidthFixed, imgui.TextLineHeight()*7, uint(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Output name:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		imgui.InputTextV("##oname", &bld.outputName, imgui.InputTextFlagsCallbackCharFilter, imguiPathFilter)

		if bld.currentPMode == Screenshot {
			imgui.TableNextColumn()

			imgui.AlignTextToFramePadding()
			imgui.Text("Screenshot at:")

			imgui.TableNextColumn()

			if imgui.BeginTableV("rrfa", 2, 0, vec2(-1, 0), -1) {
				imgui.TableSetupColumnV("c1rrfa", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
				imgui.TableSetupColumnV("c2rrfa", imgui.TableColumnFlagsWidthFixed, imgui.CalcTextSize("s", false, 0).X+imgui.CurrentStyle().CellPadding().X*2, uint(1))

				imgui.TableNextColumn()

				imgui.SetNextItemWidth(-1)

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

				imgui.TableNextColumn()

				imgui.AlignTextToFramePadding()
				imgui.Text("s")

				imgui.EndTable()
			}
		}

		imgui.EndTable()
	}
}

func drawReplayManager(bld *builder) {
	if imgui.BeginTableV("replay table", 9, imgui.TableFlagsBorders|imgui.TableFlagsScrollY, vec2(-1, imgui.ContentRegionAvail().Y), -1) {
		imgui.TableSetupScrollFreeze(0, 1)

		imgui.TableSetupColumnV("", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(0))
		imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch|imgui.TableColumnFlagsNoSort, 0, uint(1))
		imgui.TableSetupColumnV("Score", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(2))
		imgui.TableSetupColumnV("Mods", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(3))
		imgui.TableSetupColumnV("300", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(4))
		imgui.TableSetupColumnV("100", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(5))
		imgui.TableSetupColumnV("50", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(6))
		imgui.TableSetupColumnV("Miss", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(7))
		imgui.TableSetupColumnV("Combo", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, uint(8))

		imgui.TableHeadersRow()

		imgui.PushFont(Font20)

		for i, replay := range bld.knockoutReplays {
			pReplay := replay.parsedReplay

			imgui.TableNextColumn()

			imgui.Checkbox("##Use"+strconv.Itoa(i), &replay.included)

			imgui.TableNextColumn()

			imgui.Text(pReplay.Username)

			imgui.TableNextColumn()

			imgui.Text(utils.Humanize(pReplay.Score))

			imgui.TableNextColumn()

			imgui.Text(difficulty.Modifier(pReplay.Mods).String())

			imgui.TableNextColumn()

			imgui.Text(utils.Humanize(pReplay.Count300))

			imgui.TableNextColumn()

			imgui.Text(utils.Humanize(pReplay.Count100))

			imgui.TableNextColumn()

			imgui.Text(utils.Humanize(pReplay.Count50))

			imgui.TableNextColumn()

			imgui.Text(utils.Humanize(pReplay.CountMiss))

			imgui.TableNextColumn()

			imgui.Text(utils.Humanize(pReplay.MaxCombo))
		}

		imgui.PopFont()

		imgui.EndTable()
	}
}

func drawAbout(dTex texture.Texture) {
	centerTable("about1", -1, func() {
		imgui.Image(imgui.TextureID(dTex.GetID()), vec2(100, 100))
	})

	centerTable("about2", -1, func() {
		imgui.Text("danser-go " + build.VERSION)
	})

	centerTable("about3", -1, func() {
		if imgui.Button("Check for updates") {
			checkForUpdates(true)
		}
	})

	imgui.Dummy(vec2(1, imgui.FrameHeight()))

	centerTable("about4.1", -1, func() {
		imgui.Text("Advanced visualisation multi-tool")
	})

	centerTable("about4.2", -1, func() {
		imgui.Text("for osu!")
	})

	imgui.Dummy(vec2(1, imgui.FrameHeight()))

	if imgui.BeginTableV("about5", 3, imgui.TableFlagsSizingStretchSame, vec2(-1, 0), -1) {
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
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, vec2(imgui.CurrentStyle().CellPadding().X, 10))

	if imgui.BeginTableV("lconfigtable", 2, 0, vec2(-1, 0), -1) {
		imgui.TableSetupColumnV("1lconfigtable", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
		imgui.TableSetupColumnV("2lconfigtable", imgui.TableColumnFlagsWidthFixed, 0, uint(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Check for updates on startup")

		imgui.TableNextColumn()

		imgui.Checkbox("##CheckForUpdates", &launcherConfig.CheckForUpdates)

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Load latest replay on startup")

		imgui.TableNextColumn()

		imgui.Checkbox("##LoadLatestReplay", &launcherConfig.LoadLatestReplay)

		imgui.TableNextColumn()

		posLocalSMU := imgui.CursorPos()

		imgui.AlignTextToFramePadding()
		imgui.Text("Speed up startup on slow HDDs.\nWon't detect deleted/updated\nmaps!")

		posLocalSMU1 := imgui.CursorPos()

		imgui.TableNextColumn()

		imgui.SetCursorPos(vec2(imgui.CursorPosX(), (posLocalSMU.Y+posLocalSMU1.Y-imgui.FrameHeightWithSpacing())/2))
		imgui.Checkbox("##SkipMapUpdate", &launcherConfig.SkipMapUpdate)

		imgui.TableNextColumn()

		posLocalSFA := imgui.CursorPos()

		imgui.AlignTextToFramePadding()
		imgui.Text("Show exported videos/images\nin explorer")

		posLocalSFA1 := imgui.CursorPos()

		imgui.TableNextColumn()

		imgui.SetCursorPos(vec2(imgui.CursorPosX(), (posLocalSFA.Y+posLocalSFA1.Y-imgui.FrameHeightWithSpacing())/2))
		imgui.Checkbox("##ShowFileAfter", &launcherConfig.ShowFileAfter)

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Preview selected maps")

		imgui.TableNextColumn()

		imgui.Checkbox("##PreviewSelected", &launcherConfig.PreviewSelected)

		imgui.EndTable()
	}

	imgui.AlignTextToFramePadding()
	imgui.Text("Preview volume")

	volume := int32(launcherConfig.PreviewVolume * 100)

	imgui.PushFont(Font16)

	imgui.SetNextItemWidth(-1)

	if sliderIntSlide("##previewvolume", &volume, 0, 100, "%d%%", 0) {
		launcherConfig.PreviewVolume = float64(volume) / 100
	}

	imgui.PopFont()

	imgui.PopStyleVar()
}
