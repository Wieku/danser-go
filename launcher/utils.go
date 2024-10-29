package launcher

import (
	"fmt"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/sqweek/dialog"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/platform"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type messageType int

const (
	mInfo = messageType(iota)
	mError
	mQuestion
)

func showMessage(typ messageType, format string, args ...any) bool {
	message := fmt.Sprintf(format, args...)

	switch typ {
	case mInfo:
		dialog.Message(message).Info()
	case mError:
		if urlIndex := strings.Index(message, "http"); urlIndex > -1 {
			if dialog.Message(message + "\n\nDo you want to go there?").ErrorYesNo() {
				url := message[urlIndex:]
				platform.OpenURL(url)
				return true
			}
		} else {
			dialog.Message(message).Error()
		}
	case mQuestion:
		return dialog.Message(message).YesNo()
	}

	return false
}

func checkForUpdates(pingUpToDate bool) {
	status, url, err := utils.CheckForUpdate()

	switch status {
	case utils.Ignored, utils.UpToDate:
		if pingUpToDate {
			showMessage(mInfo, "You're using the newest version of danser.")
		}
	case utils.Failed:
		showMessage(mError, "Can't get version from GitHub:", err)
	case utils.Snapshot:
		if showMessage(mQuestion, "You're using a snapshot version of danser.\nFor newer version of snapshots please visit the official danser discord server at: %s\n\nDo you want to go there?", url) {
			platform.OpenURL(url)
		}
	case utils.UpdateAvailable:
		if showMessage(mQuestion, "You're using an older version of danser.\nYou can download a newer version here: %s\n\nDo you want to go there?", url) {
			platform.OpenURL(url)
		}
	}
}

func textColumn(text string) {
	imgui.TableNextColumn()
	imgui.TextUnformatted(text)
}

func imguiPathFilter(data imgui.InputTextCallbackData) int {
	if data.EventFlag() == imgui.InputTextFlagsCallbackCharFilter {
		run := data.EventChar()

		switch run {
		case '\\':
			data.SetEventChar('/')
		case '<', '>', '|', '?', '*', ':', '"':
			data.SetEventChar(0)
		}
	}

	return 0
}

func vec2(x, y float32) imgui.Vec2 {
	return imgui.Vec2{
		X: x,
		Y: y,
	}
}

func vec4(x, y, z, w float32) imgui.Vec4 {
	return imgui.Vec4{
		X: x,
		Y: y,
		Z: z,
		W: w,
	}
}

func vzero() imgui.Vec2 {
	return vec2(0, 0)
}

func dummyExactY(y float32) {
	dummyExact(vec2(0, y))
}

func dummyExact(v imgui.Vec2) {
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, vzero())
	imgui.Dummy(v)
	imgui.PopStyleVar()
}

func packColor(vec imgui.Vec4) uint32 {
	convert := func(f float32) uint32 {
		scaled := (f * math.MaxUint8) + 0.5 // nolint: gomnd
		switch {
		case scaled <= 0:
			return 0
		case scaled >= math.MaxUint8:
			return math.MaxUint8
		default:
			return uint32(scaled)
		}
	}
	return convert(vec.X) |
		convert(vec.Y)<<8 |
		convert(vec.Z)<<16 |
		convert(vec.W)<<24
}

// covers cases:
// one of them is an abs path to data dir
// has path separator suffixed
// one of them has backwards slashes
func compareDirs(str1, str2 string) bool {
	abPath := strings.TrimSuffix(strings.ReplaceAll(env.DataDir(), "\\", "/"), "/") + "/"

	str1D := strings.TrimSuffix(strings.ReplaceAll(str1, "\\", "/"), "/")
	str2D := strings.TrimSuffix(strings.ReplaceAll(str2, "\\", "/"), "/")

	return strings.TrimPrefix(str1D, abPath) == strings.TrimPrefix(str2D, abPath)
}

func getAbsPath(path string) string {
	if strings.TrimSpace(path) != "" && filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(env.DataDir(), path)
}

func getRelativeOrABSPath(path string) string {
	slashPath := strings.TrimSuffix(strings.ReplaceAll(path, "\\", "/"), "/")

	slashBase := strings.TrimSuffix(strings.ReplaceAll(env.DataDir(), "\\", "/"), "/") + "/"

	return strings.ReplaceAll(strings.TrimPrefix(slashPath, slashBase), "/", string(os.PathSeparator))
}

func setTooltip(txt string) {
	imgui.SetTooltip(strings.ReplaceAll(txt, "%", "%%"))
}

func contentRegionMin() imgui.Vec2 {
	return imgui.CursorScreenPos().Sub(imgui.WindowPos())
}

func contentRegionMax() imgui.Vec2 {
	return imgui.ContentRegionAvail().Add(imgui.CursorScreenPos()).Sub(imgui.WindowPos())
}
