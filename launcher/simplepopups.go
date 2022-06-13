package launcher

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/framework/math/mutils"
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

	imgui.Spacing()

	sliderIntReset("End time", end, 1, end.ogValue, util.FormatSeconds(int(end.value)))

	if start.value >= end.value {
		start.value = end.value - 1
	}

	imgui.Spacing()
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
