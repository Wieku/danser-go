package launcher

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/platform"
	"strconv"
)

func drawSpeedMenu(bld *builder) {
	sliderFloatResetStep("Speed", &bld.speed, 0.1, 3, 0.05, "%.2f")
	imgui.Spacing()

	sliderFloatResetStep("Pitch", &bld.pitch, 0.1, 3, 0.05, "%.2f")
	imgui.Spacing()
}

func drawCDMenu(bld *builder) {
	if imgui.BeginTable("dfa", 2) {
		imgui.TableNextColumn()

		imgui.TextUnformatted("Mirrored cursors:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(imgui.TextLineHeight() * 4.5)

		if imgui.InputIntV("##mirrors", &bld.mirrors, 1, 1, 0) {
			if bld.mirrors < 1 {
				bld.mirrors = 1
			}
		}

		imgui.TableNextColumn()

		imgui.TextUnformatted("Tag cursors:")

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
		imgui.TableSetupColumnV("c1rfa", imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(0))
		imgui.TableSetupColumnV("c2rfa", imgui.TableColumnFlagsWidthFixed, imgui.TextLineHeight()*7, imgui.ID(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.TextUnformatted("Output name:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		inputTextV("##oname", &bld.outputName, imgui.InputTextFlagsCallbackCharFilter, imguiPathFilter)

		if bld.currentPMode == Screenshot {
			imgui.TableNextColumn()

			imgui.AlignTextToFramePadding()
			imgui.TextUnformatted("Screenshot at:")

			imgui.TableNextColumn()

			if imgui.BeginTableV("rrfa", 2, 0, vec2(-1, 0), -1) {
				imgui.TableSetupColumnV("c1rrfa", imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(0))
				imgui.TableSetupColumnV("c2rrfa", imgui.TableColumnFlagsWidthFixed, imgui.CalcTextSizeV("s", false, 0).X+imgui.CurrentStyle().CellPadding().X*2, imgui.ID(1))

				imgui.TableNextColumn()

				imgui.SetNextItemWidth(-1)

				valText := strconv.FormatFloat(float64(bld.ssTime), 'f', 3, 64)
				prevText := valText

				if inputText("##sstime", &valText) {
					parsed, err := strconv.ParseFloat(valText, 64)
					if err != nil {
						valText = prevText
					} else {
						parsed = mutils.Clamp(parsed, 0, float64(bld.end.ogValue))
						bld.ssTime = float32(parsed)
					}
				}

				imgui.TableNextColumn()

				imgui.AlignTextToFramePadding()
				imgui.TextUnformatted("s")

				imgui.EndTable()
			}
		}

		imgui.EndTable()
	}
}

func drawAbout(dTex texture.Texture) {
	centerTable("about1", -1, func() {
		imgui.Image(imgui.TextureID{Data: uintptr(dTex.GetID())}, vec2(100, 100))
	})

	centerTable("about2", -1, func() {
		imgui.TextUnformatted("danser-go " + build.VERSION)
	})

	centerTable("about3", -1, func() {
		if imgui.Button("Check for updates") {
			checkForUpdates(true)
		}
	})

	imgui.Dummy(vec2(1, imgui.FrameHeight()))

	centerTable("about4.1", -1, func() {
		imgui.TextUnformatted("Advanced visualisation multi-tool")
	})

	centerTable("about4.2", -1, func() {
		imgui.TextUnformatted("for osu!")
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

	checkboxOption("Check for updates on startup", &launcherConfig.CheckForUpdates)

	checkboxOption("Load latest replay on startup", &launcherConfig.LoadLatestReplay)

	checkboxOption("Speed up startup on slow HDDs.\nWon't detect deleted/updated\nmaps!", &launcherConfig.SkipMapUpdate)

	checkboxOption("Load changes in Songs folder automatically", &launcherConfig.AutoRefreshDB)

	checkboxOption("Show JSON paths in config editor", &launcherConfig.ShowJSONPaths)

	checkboxOption("Show exported videos/images\nin explorer", &launcherConfig.ShowFileAfter)

	checkboxOption("Preview selected maps", &launcherConfig.PreviewSelected)

	imgui.AlignTextToFramePadding()
	imgui.TextUnformatted("Preview volume")

	volume := int32(launcherConfig.PreviewVolume * 100)

	imgui.PushFont(Font16)

	imgui.SetNextItemWidth(-1)

	if sliderIntSlide("##previewvolume", &volume, 0, 100, "%d%%", 0) {
		launcherConfig.PreviewVolume = float64(volume) / 100
	}

	imgui.PopFont()

	imgui.PopStyleVar()
}
