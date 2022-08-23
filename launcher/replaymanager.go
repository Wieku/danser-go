package launcher

import (
	"fmt"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/utils"
	"strconv"
)

type knockoutManagerPopup struct {
	*popup

	bld *builder

	includeSwitch bool
}

func newKnockoutManagerPopup(bld *builder) *knockoutManagerPopup {
	rm := &knockoutManagerPopup{
		popup:         newPopup("Replay manager", popBig),
		bld:           bld,
		includeSwitch: true,
	}

	rm.internalDraw = rm.drawManager

	return rm
}

func (km *knockoutManagerPopup) drawManager() {
	imgui.PushFont(Font20)
	countEnabled := 0
	for _, replay := range km.bld.knockoutReplays {
		if replay.included {
			countEnabled++
		}
	}

	numText := "No replays"
	if countEnabled == 1 {
		numText = "1 replay"
	} else if countEnabled > 1 {
		numText = fmt.Sprintf("%d replays", countEnabled)
	}

	imgui.Text(numText + " selected")

	imgui.PopFont()

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

		imgui.TableSetColumnIndex(0)

		imgui.PushFont(Font20)

		if imgui.Checkbox("##mass replay disable", &km.includeSwitch) {
			for _, replay := range km.bld.knockoutReplays {
				replay.included = km.includeSwitch
			}
		}

		imgui.TableNextRow()

		changed := false

		for i, replay := range km.bld.knockoutReplays {
			pReplay := replay.parsedReplay

			imgui.TableNextColumn()

			if imgui.Checkbox("##Use"+strconv.Itoa(i), &replay.included) {
				changed = true
			}

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

		if changed {
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
		}

		imgui.PopFont()

		imgui.EndTable()
	}
}
