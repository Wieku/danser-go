package launcher

import (
	"fmt"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/utils"
	"strconv"
)

type knockoutManagerPopup struct {
	*popup

	bld *builder

	includeSwitch bool

	lastSelected int

	countEnabled int
}

func newKnockoutManagerPopup(bld *builder) *knockoutManagerPopup {
	rm := &knockoutManagerPopup{
		popup:         newPopup("Replay manager", popBig),
		bld:           bld,
		includeSwitch: true,
		lastSelected:  -1,
	}

	rm.internalDraw = rm.drawManager

	rm.refreshCount()

	return rm
}

func (km *knockoutManagerPopup) refreshCount() {
	countIncluded := 0

	for _, replay := range km.bld.knockoutReplays {
		if replay.included {
			countIncluded++
		}
	}

	if countIncluded == 0 {
		km.includeSwitch = false
	} else if countIncluded == len(km.bld.knockoutReplays) {
		km.includeSwitch = true
	}

	km.countEnabled = countIncluded
}

func (km *knockoutManagerPopup) drawManager() {
	imgui.PushFont(Font20)

	numText := "No replays"
	if km.countEnabled == 1 {
		numText = "1 replay"
	} else if km.countEnabled > 1 {
		numText = fmt.Sprintf("%d replays", km.countEnabled)
	}

	imgui.TextUnformatted(numText + " selected")

	imgui.PopFont()

	if imgui.BeginTableV("replay table", 9, imgui.TableFlagsBorders|imgui.TableFlagsScrollY, vec2(-1, imgui.ContentRegionAvail().Y), -1) {
		imgui.TableSetupScrollFreeze(0, 1)

		imgui.TableSetupColumnV("", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(0))
		imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch|imgui.TableColumnFlagsNoSort, 0, imgui.ID(1))
		imgui.TableSetupColumnV("Score", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(2))
		imgui.TableSetupColumnV("Mods", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(3))
		imgui.TableSetupColumnV("300", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(4))
		imgui.TableSetupColumnV("100", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(5))
		imgui.TableSetupColumnV("50", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(6))
		imgui.TableSetupColumnV("Miss", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(7))
		imgui.TableSetupColumnV("Combo", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoSort, 0, imgui.ID(8))

		imgui.TableHeadersRow()

		imgui.TableSetColumnIndex(0)

		imgui.PushFont(Font20)

		if imgui.Checkbox("##mass replay disable", &km.includeSwitch) {
			for _, replay := range km.bld.knockoutReplays {
				replay.included = km.includeSwitch
			}

			km.refreshCount()
		}

		imgui.TableNextRow()

		changed := -1

		for i, replay := range km.bld.knockoutReplays {
			pReplay := replay.parsedReplay

			imgui.TableNextColumn()

			if imgui.Checkbox("##Use"+strconv.Itoa(i), &replay.included) {
				changed = i
			}

			textColumn(pReplay.Username)

			textColumn(utils.Humanize(pReplay.Score))

			textColumn(difficulty.Modifier(pReplay.Mods).String())

			textColumn(utils.Humanize(pReplay.Count300))

			textColumn(utils.Humanize(pReplay.Count100))

			textColumn(utils.Humanize(pReplay.Count50))

			textColumn(utils.Humanize(pReplay.CountMiss))

			textColumn(utils.Humanize(pReplay.MaxCombo))
		}

		if changed > -1 {
			if km.lastSelected > -1 && (imgui.IsKeyDown(imgui.KeyLeftShift) || imgui.IsKeyDown(imgui.KeyRightShift)) {
				value := km.bld.knockoutReplays[changed].included

				lower := min(km.lastSelected, changed)
				higher := max(km.lastSelected, changed)

				for i := lower; i <= higher; i++ {
					km.bld.knockoutReplays[i].included = value
				}
			}

			km.lastSelected = changed

			km.refreshCount()
		}

		imgui.PopFont()

		imgui.EndTable()
	}
}
