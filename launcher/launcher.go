package launcher

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/fsnotify/fsnotify"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sqweek/dialog"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/database"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/osuapi"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/platform"
	"github.com/wieku/danser-go/framework/qpc"
	"github.com/wieku/danser-go/framework/util"
	"github.com/wieku/rplpa"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type Mode int

const (
	CursorDance Mode = iota
	DanserReplay
	Replay
	Knockout
	Play
)

func (m Mode) String() string {
	switch m {
	case CursorDance:
		return "Cursor dance / mandala / tag"
	case DanserReplay:
		return "Cursor dance with UI"
	case Replay:
		return "Watch a replay"
	case Knockout:
		return "Watch knockout"
	case Play:
		return "Play osu!standard"
	}

	return ""
}

var modes = []Mode{CursorDance, DanserReplay, Replay, Knockout, Play}

type PMode int

const (
	Watch PMode = iota
	Record
	Screenshot
)

func (m PMode) String() string {
	switch m {
	case Watch:
		return "Watch"
	case Record:
		return "Record"
	case Screenshot:
		return "Screenshot"
	}

	return ""
}

var pModes = []PMode{Watch, Record, Screenshot}

type ConfigMode int

const (
	Rename ConfigMode = iota
	Clone
	New
)

type launcher struct {
	win *glfw.Window

	bg *common.Background

	batch *batch.QuadBatch
	coin  *common.DanserCoin

	bld *builder

	beatmaps []*beatmap.BeatMap

	configList    []string
	currentConfig *settings.Config

	newDefault bool

	mapsLoaded bool

	newCloneOpened bool

	configManiMode ConfigMode
	configPrevName string

	newCloneName     string
	refreshRate      int
	configEditOpened bool
	configEditPos    imgui.Vec2

	danserRunning       bool
	recordProgress      float32
	recordStatus        string
	recordStatusSpeed   string
	recordStatusElapsed string
	recordStatusETA     string
	showProgressBar     bool

	triangleSpeed    *animation.Glider
	encodeInProgress bool
	encodeStart      time.Time
	danserCmd        *exec.Cmd
	popupStack       []iPopup

	selectWindow *songSelectPopup
	splashText   string

	prevMap *beatmap.BeatMap

	configSearch string

	lastReplayDir   string
	lastKnockoutDir string

	knockoutManager *knockoutManagerPopup

	currentEditor     *settingsEditor
	beatmapDirUpdated bool
	showBeatmapAlert  float64

	winter        bool
	christmas     bool
	recordSnowPos vector.Vector2f

	snow *drawables.Snow

	timeMenu *timePopup

	cHold           map[string]*bool
	configScrolling bool
}

func StartLauncher() {
	defer func() {
		var err any
		var stackTrace []string

		if err = recover(); err != nil {
			stackTrace = goroutines.GetStackTrace(4)
		}

		closeHandler(err, stackTrace)
	}()

	goroutines.SetCrashHandler(closeHandler)

	cTime := time.Now()

	launcher := &launcher{
		bld:        newBuilder(),
		popupStack: make([]iPopup, 0),
		winter:     (cTime.Month() == 12 && cTime.Day() >= 6) || (cTime.Month() < 2),
		christmas:  cTime.Month() == 12 && cTime.Day() >= 6,
		cHold:      make(map[string]*bool),
	}

	platform.StartLogging("launcher")

	loadLauncherConfig()

	settings.CreateDefault()

	settings.Playfield.Background.Triangles.Enabled = true
	settings.Playfield.Background.Triangles.DrawOverBlur = true
	settings.Playfield.Background.Blur.Enabled = false
	settings.Playfield.Background.Parallax.Enabled = true
	settings.Playfield.Background.Parallax.Amount = 0.02

	assets.Init(build.Stream == "Dev")

	goroutines.RunMain(func() {
		defer func() {
			if err := recover(); err != nil {
				stackTrace := goroutines.GetStackTrace(4)
				closeHandler(err, stackTrace)
			}
		}()

		goroutines.CallMain(launcher.startGLFW)

		for !launcher.win.ShouldClose() {
			goroutines.CallMain(func() {
				if launcher.win.GetAttrib(glfw.Iconified) == glfw.False {
					if launcher.win.GetAttrib(glfw.Focused) == glfw.False {
						glfw.SwapInterval(2)
					} else {
						glfw.SwapInterval(1)
					}
				} else {
					glfw.SwapInterval(launcher.refreshRate / 10)
				}
				launcher.Draw()
				launcher.win.SwapBuffers()
				glfw.PollEvents()
			})
		}
	})

	// Save configs on exit
	closeWatcher()
	saveLauncherConfig()
	if launcher.currentConfig != nil {
		launcher.currentConfig.Save("", false)
	}
}

func closeHandler(err any, stackTrace []string) {
	if err != nil {
		log.Println("panic:", err)

		for _, s := range stackTrace {
			log.Println(s)
		}

		showMessage(mError, "Launcher crashed with message:\n %s", err)

		os.Exit(1)
	}

	log.Println("Exiting normally.")
}

func (l *launcher) startGLFW() {
	err := glfw.Init()
	if err != nil {
		panic("Failed to initialize GLFW: " + err.Error())
	}

	l.refreshRate = glfw.GetPrimaryMonitor().GetVideoMode().RefreshRate

	l.tryCreateDefaultConfig()
	l.createConfigList()
	settings.LoadCredentials()

	l.bld.config = *launcherConfig.Profile

	c, err := l.loadConfig(l.bld.config)
	if err != nil {
		showMessage(mError, "Failed to read \"%s\" profile.\nReverting to \"default\".\nError: %s", l.bld.config, err)

		l.bld.config = "default"
		*launcherConfig.Profile = l.bld.config
		saveLauncherConfig()

		c, err = l.loadConfig(l.bld.config)
		if err != nil {
			panic(err)
		}
	}

	settings.General.OsuSongsDir = c.General.OsuSongsDir

	l.currentConfig = c

	platform.SetupContext()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ScaleToMonitor, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	settings.Graphics.Fullscreen = false
	settings.Graphics.WindowWidth = 800
	settings.Graphics.WindowHeight = 534

	l.win, err = glfw.CreateWindow(800, 534, "danser-go "+build.VERSION+" launcher", nil, nil)

	if err != nil {
		panic(err)
	}

	input.Win = l.win

	if l.christmas {
		platform.LoadIcons(l.win, "dansercoin", "-s")
	} else {
		platform.LoadIcons(l.win, "dansercoin", "")
	}

	l.win.MakeContextCurrent()

	log.Println("GLFW initialized!")

	err = platform.GLInit(false)
	if err != nil {
		panic("Failed to initialize OpenGL: " + err.Error())
	}

	glfw.SwapInterval(1)

	SetupImgui(l.win)

	graphics.LoadTextures()

	if l.winter {
		graphics.LoadWinterTextures()
	}

	bass.Init(false)

	l.triangleSpeed = animation.NewGlider(1)
	l.triangleSpeed.SetEasing(easing.OutQuad)

	l.batch = batch.NewQuadBatch()

	l.bg = common.NewBackground(false)

	settings.Playfield.Background.Triangles.Enabled = true

	if l.winter {
		imgui.PushStyleColorVec4(imgui.ColBorder, vec4(0.76, 0.9, 1, 1))

		settings.Playfield.Background.Triangles.Enabled = false

		l.snow = drawables.NewSnow()
	}

	if l.christmas {
		l.coin = common.NewDanserCoinSanta()
	} else {
		l.coin = common.NewDanserCoin()
	}

	l.coin.DrawVisualiser(true)

	goroutines.RunOS(func() {
		l.splashText = "Initializing..."

		settings.DefaultsFactory.EncoderOptions() // preload to avoid pauses

		l.loadBeatmaps()

		l.win.SetDropCallback(func(w *glfw.Window, names []string) {
			if l.danserRunning {
				return
			}

			if strings.HasSuffix(names[0], ".osz") {
				l.loadOSZs(names)
			} else if len(names) > 1 {
				l.trySelectReplaysFromPaths(names)
			} else {
				l.trySelectReplayFromPath(names[0])
			}
		})

		l.win.SetCloseCallback(func(w *glfw.Window) {
			if l.danserCmd != nil {
				l.win.SetShouldClose(false)

				goroutines.Run(func() {
					if showMessage(mQuestion, "Recording is in progress, do you want to exit?") {
						if l.danserCmd != nil {
							l.danserCmd.Process.Kill()
							l.danserCleanup(false)
						}

						l.win.SetShouldClose(true)
					}
				})
			}
		})

		if len(os.Args) > 2 { //won't work in combined mode
			l.trySelectReplaysFromPaths(os.Args[1:])
		} else if len(os.Args) > 1 {
			l.trySelectReplayFromPath(os.Args[1])
		} else if launcherConfig.LoadLatestReplay {
			l.loadLatestReplay()
		}

		if launcherConfig.CheckForUpdates {
			checkForUpdates(false)
		}

		refreshErr := osuapi.TryRefreshToken()
		if refreshErr != nil {
			showMessage(mError, "Failed to refresh token!\nPlease go to Settings->Credentials and click Authorize.\nError: %s", refreshErr.Error())
		}
	})

	input.RegisterListener(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if l.currentEditor != nil {
			l.currentEditor.updateKey(w, key, scancode, action, mods)
		}
	})
}

func (l *launcher) loadBeatmaps() {
	closeWatcher()
	database.Close()

	l.splashText = "Loading maps...\nThis may take a while..."

	l.beatmaps = make([]*beatmap.BeatMap, 0)

	err := database.Init()
	if err != nil {
		showMessage(mError, "Failed to initialize database! Error: %s\nMake sure Song's folder does exist or change it to the correct directory in settings.", err)
		l.beatmaps = make([]*beatmap.BeatMap, 0)
	} else {
		bSplash := "Loading maps...\nThis may take a while...\n\n"

		beatmaps := database.LoadBeatmaps(launcherConfig.SkipMapUpdate, func(stage database.ImportStage, processed, target int) {
			switch stage {
			case database.Discovery:
				l.splashText = bSplash + "Searching for .osu files...\n\n"
			case database.Comparison:
				l.splashText = bSplash + "Comparing files with database...\n\n"
			case database.Cleanup:
				l.splashText = bSplash + "Removing leftover maps from database...\n\n"
			case database.Import:
				percent := float64(processed) / float64(target) * 100
				l.splashText = bSplash + fmt.Sprintf("Importing maps...\n%d / %d\n%.0f%%", processed, target, percent)
			case database.Finished:
				l.splashText = bSplash + "Finished!\n\n"
			}
		})

		slices.SortFunc(beatmaps, func(a, b *beatmap.BeatMap) int {
			return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		})

		bSplash = "Calculating Star Rating...\nThis may take a while...\n\n\n"

		l.splashText = bSplash + "\n"

		database.UpdateStarRating(beatmaps, func(processed, target int) {
			percent := float64(processed) / float64(target) * 100
			l.splashText = bSplash + fmt.Sprintf("%d / %d\n%.0f%%", processed, target, percent)
		})

		for _, bMap := range beatmaps {
			l.beatmaps = append(l.beatmaps, bMap)
		}

		//database.Close()
	}

	l.setupWatcher()

	l.mapsLoaded = true
}

func (l *launcher) loadLatestReplay() {
	replaysDir := l.currentConfig.General.GetReplaysDir()

	type lastModPath struct {
		tStamp time.Time
		name   string
	}

	var list []*lastModPath

	entries, err := os.ReadDir(replaysDir)
	if err != nil {
		return
	}

	for _, d := range entries {
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".osr") {
			if info, err1 := d.Info(); err1 == nil {
				list = append(list, &lastModPath{
					tStamp: info.ModTime(),
					name:   d.Name(),
				})
			}
		}
	}

	if list == nil {
		return
	}

	slices.SortFunc(list, func(a, b *lastModPath) int {
		return -a.tStamp.Compare(b.tStamp)
	})

	// Load the newest that can be used
	for _, lMP := range list {
		r, err := l.loadReplay(filepath.Join(replaysDir, lMP.name))
		if err == nil {
			l.trySelectReplay(r)
			break
		}
	}
}

func (l *launcher) Draw() {
	w, h := l.win.GetFramebufferSize()
	viewport.Push(w, h)

	if l.bg.HasBackground() {
		gl.ClearColor(0, 0, 0, 1.0)
	} else {
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	}

	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Enable(gl.SCISSOR_TEST)

	w, h = int(settings.Graphics.WindowWidth), int(settings.Graphics.WindowHeight)

	settings.Graphics.Fullscreen = false
	settings.Graphics.WindowWidth = int64(w)
	settings.Graphics.WindowHeight = int64(h)

	if l.currentConfig != nil {
		settings.Audio.GeneralVolume = l.currentConfig.Audio.GeneralVolume
		settings.Audio.MusicVolume = l.currentConfig.Audio.MusicVolume
	}

	t := qpc.GetMilliTimeF()

	l.triangleSpeed.Update(t)

	settings.Playfield.Background.Triangles.Speed = l.triangleSpeed.GetValue()

	pX := (float64(imgui.MousePos().X) * 2 / settings.Graphics.GetWidthF()) - 1
	pY := (float64(imgui.MousePos().Y) * 2 / settings.Graphics.GetHeightF()) - 1

	l.bg.Update(t, -pX, pY)

	l.batch.SetCamera(mgl32.Ortho(-float32(w)/2, float32(w)/2, float32(h)/2, -float32(h)/2, -1, 1))

	bgA := 1.0
	if l.bg.HasBackground() {
		bgA = 0.33
	}

	l.bg.Draw(t, l.batch, 0, bgA, l.batch.Projection)

	l.batch.SetColor(1, 1, 1, 1)
	l.batch.ResetTransform()
	l.batch.SetCamera(mgl32.Ortho(0, float32(w), float32(h), 0, -1, 1))

	l.batch.Begin()

	if l.winter {
		l.snow.Update(t)
		l.snow.Draw(t, l.batch)
	}

	if l.mapsLoaded {
		if l.winter {
			bSnow := *graphics.Snow[0]

			if l.showProgressBar {
				bSnow = *graphics.Snow[5]
			}

			l.batch.DrawStObject(vector.NewVec2d(0, settings.Graphics.GetHeightF()), vector.BottomLeft, vector.NewVec2d(1, 1), false, false, 0, color2.NewL(1), false, bSnow)

			//record button
			if l.bld.currentMode != Play {
				l.batch.DrawStObject(l.recordSnowPos.Copy64().AddS(0, 2), vector.BottomCentre, vector.NewVec2d(1, 1), false, false, 0, color2.NewL(1), false, *graphics.Snow[2])
			}

			//danse button
			l.batch.DrawStObject(vector.NewVec2d(624, 448), vector.BottomCentre, vector.NewVec2d(1, 1), false, false, 0, color2.NewL(1), false, *graphics.Snow[1])

			l.batch.DrawStObject(vector.NewVec2d(115, 240), vector.BottomCentre, vector.NewVec2d(1, 1), false, false, 0, color2.NewL(1), false, *graphics.Snow[4])

			if l.bld.currentMode != Replay {
				l.batch.DrawStObject(vector.NewVec2d(314, 240), vector.BottomCentre, vector.NewVec2d(1, 1), false, false, 0, color2.NewL(1), false, *graphics.Snow[3])
			}

		}

		if l.bld.currentMap != nil && l.prevMap != l.bld.currentMap {
			l.bg.SetBeatmap(l.bld.currentMap, false, false)

			l.prevMap = l.bld.currentMap
		}

		if l.selectWindow != nil {
			if l.selectWindow.PreviewedSong != nil {
				l.selectWindow.PreviewedSong.Update()
				l.coin.SetMap(l.selectWindow.prevMap, l.selectWindow.PreviewedSong)
				l.bg.SetTrack(l.selectWindow.PreviewedSong)
			} else {
				l.coin.SetMap(nil, nil)
				l.bg.SetTrack(nil)
			}
		}

		l.coin.SetPosition(vector.NewVec2d(468+155.5, 180+85))

		if l.christmas {
			l.coin.SetScale(float64(h) / 5)
		} else {
			l.coin.SetScale(float64(h) / 4)
		}

		l.coin.SetRotation(0.1)

		l.coin.Update(t)
		l.coin.Draw(t, l.batch)
	}

	l.batch.End()

	l.drawImgui()

	viewport.Pop()
}

func (l *launcher) drawImgui() {
	Begin()

	resetPopupHierarchyInfo()

	lock := l.danserRunning

	if lock {
		imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), true)
	}

	wW, wH := int(settings.Graphics.WindowWidth), int(settings.Graphics.WindowHeight)

	imgui.SetNextWindowSize(vec2(float32(wW), float32(wH)))

	imgui.SetNextWindowPos(vzero())

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, vec2(20, 20))

	imgui.BeginV("main", nil, imgui.WindowFlagsNoDecoration /*|imgui.WindowFlagsNoMove*/ |imgui.WindowFlagsNoBackground|imgui.WindowFlagsNoScrollWithMouse|imgui.WindowFlagsNoBringToFrontOnFocus)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, vec2(5, 5))

	if l.mapsLoaded {
		l.drawMain()
	} else {
		l.drawSplash()
	}

	imgui.PopStyleVar()

	imgui.End()

	imgui.PopStyleVar()

	if lock {
		imgui.PopItemFlag()
	}

	DrawImgui()
}

func (l *launcher) drawMain() {
	w := contentRegionMax().X

	imgui.PushFont(Font24)

	if imgui.BeginTableV("ltpanel", 2, imgui.TableFlagsSizingStretchProp, vec2(float32(w)/2, 0), -1) {
		imgui.TableSetupColumnV("ltpanel1", imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(0))
		imgui.TableSetupColumnV("ltpanel2", imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.TextUnformatted("Mode:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		if imgui.BeginCombo("##mode", l.bld.currentMode.String()) {
			for _, m := range modes {
				if imgui.SelectableBoolV(m.String(), l.bld.currentMode == m, 0, vzero()) {
					if m == Play {
						l.bld.currentPMode = Watch
					}

					if m != Replay {
						l.bld.replayPath = ""
						l.bld.removeReplay()
					}

					if m != Knockout {
						l.bld.knockoutReplays = nil
					}

					l.bld.currentMode = m
				}
			}

			imgui.EndCombo()
		}

		imgui.EndTable()
	}

	l.drawConfigPanel()

	l.drawControls()

	l.drawLowerPanel()

	if l.selectWindow != nil {
		l.selectWindow.update()
	}

	for i := 0; i < len(l.popupStack); i++ {
		p := l.popupStack[i]
		p.draw()
		if p.shouldClose() {
			l.popupStack = append(l.popupStack[:i], l.popupStack[i+1:]...)
			i--
		}
	}

	imgui.PopFont()

	if imgui.IsMouseClickedBool(0) && !l.danserRunning {
		platform.StopProgress(l.win)
		l.showProgressBar = false
		l.recordStatus = ""
		l.recordProgress = 0
	}

	if !l.danserRunning && l.beatmapDirUpdated && qpc.GetMilliTimeF() >= l.showBeatmapAlert {
		reload := launcherConfig.AutoRefreshDB

		if !reload {
			mapText := "Do you want to refresh the database?"
			if launcherConfig.SkipMapUpdate {
				mapText = "Do you want to load new beatmap sets?"
			}

			reload = showMessage(mQuestion, "Changes in osu!'s Song directory have been detected.\n\n"+mapText)
		}

		l.beatmapDirUpdated = false

		if reload {
			l.reloadMaps(nil)
		}
	}
}

func (l *launcher) drawSplash() {
	w, h := contentRegionMax().X, contentRegionMax().Y

	imgui.PushFont(Font48)

	splText := strings.Split(l.splashText, "\n")

	var height float32

	for _, sText := range splText {
		height += imgui.CalcTextSizeV(sText, false, 0).Y
	}

	var dHeight float32

	for _, sText := range splText {
		tSize := imgui.CalcTextSizeV(sText, false, 0)

		imgui.SetCursorPos(vec2(20+(w-tSize.X)/2, 20+(h-height)/2+dHeight))

		dHeight += tSize.Y

		imgui.TextUnformatted(sText)
	}

	imgui.PopFont()
}

func (l *launcher) drawControls() {
	imgui.SetCursorPos(vec2(20, 88))
	switch l.bld.currentMode {
	case Replay:
		l.selectReplay()
	case Knockout:
		l.newKnockout()
	default:
		l.showSelect()
	}

	imgui.SetCursorPos(vec2(20, 204+34))

	w := contentRegionMax().X

	if imgui.BeginTableV("abtn", 2, imgui.TableFlagsSizingStretchSame, vec2(float32(w)/2, -1), -1) {
		imgui.TableNextColumn()

		if imgui.ButtonV("Speed/Pitch", vec2(-1, imgui.TextLineHeight()*2)) {
			l.openPopup(newPopupF("Speed adjust", popMedium, func() {
				drawSpeedMenu(l.bld)
			}))
		}

		imgui.TableNextColumn()

		nilMap := l.bld.currentMap == nil

		if nilMap {
			imgui.BeginDisabled()
		}

		if imgui.ButtonV("Mods", vec2(-1, imgui.TextLineHeight()*2)) {
			l.openPopup(newModPopup(l.bld))
		}

		if nilMap && imgui.IsItemHoveredV(imgui.HoveredFlagsAllowWhenDisabled) {
			imgui.SetTooltip("Select map/replay first")
		}

		imgui.TableNextColumn()

		if imgui.ButtonV("Time/Offset", vec2(-1, imgui.TextLineHeight()*2)) {
			if l.timeMenu == nil {
				l.timeMenu = newTimePopup(l.bld)

				l.timeMenu.setCloseListener(func() {
					if l.bld.currentMap != nil && l.bld.currentMap.LocalOffset != int(l.bld.offset.value) {
						l.bld.currentMap.LocalOffset = int(l.bld.offset.value)
						database.UpdateLocalOffset(l.bld.currentMap)
					}
				})
			}

			l.openPopup(l.timeMenu)
		}

		if nilMap {
			imgui.EndDisabled()
			if imgui.IsItemHoveredV(imgui.HoveredFlagsAllowWhenDisabled) {
				imgui.SetTooltip("Select map/replay first")
			}
		}

		imgui.TableNextColumn()

		if l.bld.currentMode == CursorDance {
			if imgui.ButtonV("Mirrors/Tags", vec2(-1, imgui.TextLineHeight()*2)) {
				l.openPopup(newPopupF("Difficulty adjust", popDynamic, func() {
					drawCDMenu(l.bld)
				}))
			}
		}

		imgui.EndTable()
	}

}

func (l *launcher) selectReplay() {
	bSize := vec2((imgui.WindowWidth()-40)/4, imgui.TextLineHeight()*2)

	imgui.PushFont(Font32)

	if imgui.ButtonV("Select replay", bSize) {
		dir := l.currentConfig.General.GetReplaysDir()
		if _, err := os.Lstat(dir); err != nil {
			dir = env.DataDir()
		}

		p, err := dialog.File().Filter("osu! replay file (*.osr)", "osr").Title("Select replay file").SetStartDir(dir).Load()
		if err == nil {
			l.trySelectReplayFromPath(p)
		}
	}

	imgui.PopFont()

	imgui.PushFont(Font20)
	imgui.IndentV(5)

	if l.bld.currentReplay != nil {
		b := l.bld.currentMap

		mString := fmt.Sprintf("%s - %s [%s]\nPlayed by: %s", b.Artist, b.Name, b.Difficulty, l.bld.currentReplay.Username)

		imgui.PushTextWrapPosV(contentRegionMax().X / 2)
		imgui.TextUnformatted(mString)
		imgui.PopTextWrapPos()
	} else {
		imgui.TextUnformatted("No replay selected")
	}

	imgui.UnindentV(5)
	imgui.PopFont()
}

func (l *launcher) trySelectReplayFromPath(p string) {
	replay, err := l.loadReplay(p)

	if err != nil {
		e := []rune(err.Error())
		showMessage(mError, string(unicode.ToUpper(e[0]))+string(e[1:]))
		return
	}

	l.trySelectReplay(replay)
}

func (l *launcher) trySelectReplaysFromPaths(p []string) {
	var errorCollection string
	var replays []*knockoutReplay

	for _, rPath := range p {
		replay, err := l.loadReplay(rPath)

		if err != nil {
			if errorCollection != "" {
				errorCollection += "\n"
			}

			errorCollection += fmt.Sprintf("%s:\n\t%s", filepath.Base(rPath), err)
		} else {
			replays = append(replays, replay)
		}
	}

	if errorCollection != "" {
		showMessage(mError, "There were errors opening replays:\n%s", errorCollection)
	}

	if replays != nil && len(replays) > 0 {
		found := false

		for _, replay := range replays {
			for _, bMap := range l.beatmaps {
				if strings.ToLower(bMap.MD5) == strings.ToLower(replay.parsedReplay.BeatmapMD5) {
					l.bld.currentMode = Knockout
					l.bld.setMap(bMap)

					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			showMessage(mError, "Replays use an unknown map. Please download the map beforehand.")
		} else {
			var finalReplays []*knockoutReplay

			for _, replay := range replays {
				if strings.ToLower(l.bld.currentMap.MD5) == strings.ToLower(replay.parsedReplay.BeatmapMD5) {
					finalReplays = append(finalReplays, replay)
				}
			}

			slices.SortFunc(finalReplays, func(a, b *knockoutReplay) int {
				return -cmp.Compare(a.parsedReplay.Score, b.parsedReplay.Score)
			})

			l.bld.knockoutReplays = finalReplays
			l.knockoutManager = newKnockoutManagerPopup(l.bld)
		}
	}
}

func (l *launcher) trySelectReplay(replay *knockoutReplay) {
	for _, bMap := range l.beatmaps {
		if strings.ToLower(bMap.MD5) == strings.ToLower(replay.parsedReplay.BeatmapMD5) {
			l.bld.currentMode = Replay
			l.bld.replayPath = replay.path
			l.bld.setMap(bMap)
			l.bld.setReplay(replay.parsedReplay)

			return
		}
	}

	showMessage(mError, "Replay uses an unknown map. Please download the map beforehand.")
}

func (l *launcher) newKnockout() {
	bSize := vec2((imgui.WindowWidth()-40)/4, imgui.TextLineHeight()*2)

	imgui.PushFont(Font32)

	if imgui.ButtonV("Select replays", bSize) {
		kPath := getAbsPath(launcherConfig.LastKnockoutPath)

		if _, err := os.Lstat(kPath); err != nil {
			kPath = env.DataDir()
		}

		p, err := dialog.File().Filter("osu! replay file (*.osr)", "osr").Title("Select replay files").SetStartDir(kPath).LoadMultiple()
		if err == nil {
			launcherConfig.LastKnockoutPath = getRelativeOrABSPath(filepath.Dir(p[0]))
			saveLauncherConfig()

			l.trySelectReplaysFromPaths(p)
		}
	}

	imgui.PopFont()

	imgui.PushFont(Font20)

	imgui.IndentV(5)

	if l.bld.knockoutReplays != nil && l.bld.currentMap != nil {
		b := l.bld.currentMap

		imgui.PushTextWrapPosV(contentRegionMax().X / 2)

		imgui.TextUnformatted(fmt.Sprintf("%s - %s [%s]", b.Artist, b.Name, b.Difficulty))

		imgui.AlignTextToFramePadding()

		imgui.TextUnformatted(fmt.Sprintf("%d replays loaded", len(l.bld.knockoutReplays)))

		imgui.PopTextWrapPos()

		imgui.SameLine()

		if imgui.Button("Manage##knockout") && l.knockoutManager != nil {
			l.openPopup(l.knockoutManager)
		}
	} else {
		imgui.TextUnformatted("No replays selected")
	}

	imgui.UnindentV(5)

	imgui.PopFont()
}

func (l *launcher) loadReplay(p string) (*knockoutReplay, error) {
	if !strings.HasSuffix(p, ".osr") {
		return nil, fmt.Errorf("it's not a replay file")
	}

	rData, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}

	replay, err := rplpa.ParseReplay(rData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse replay: %s", err)
	}

	if replay.PlayMode != 0 {
		return nil, errors.New("only osu!standard mode is supported")
	}

	if replay.ReplayData == nil || len(replay.ReplayData) < 2 {
		return nil, errors.New("replay is missing input data")
	}

	// dump unneeded data as it's not needed anymore to save memory
	replay.LifebarGraph = nil
	replay.ReplayData = nil

	return &knockoutReplay{
		path:         p,
		parsedReplay: replay,
		included:     true,
	}, nil
}

func (l *launcher) showSelect() {
	bSize := vec2((imgui.WindowWidth()-40)/4, imgui.TextLineHeight()*2)

	imgui.PushFont(Font32)

	if imgui.ButtonV("Select map", bSize) {
		if l.selectWindow == nil {
			l.selectWindow = newSongSelectPopup(l.bld, l.beatmaps)
		}

		l.selectWindow.open()
		l.openPopup(l.selectWindow)
	}

	imgui.PopFont()

	imgui.PushFont(Font20)

	imgui.IndentV(5)

	if l.bld.currentMap != nil {
		b := l.bld.currentMap

		mString := fmt.Sprintf("%s - %s [%s]", b.Artist, b.Name, b.Difficulty)

		imgui.PushTextWrapPosV(contentRegionMax().X / 2)
		imgui.TextUnformatted(mString)
		imgui.PopTextWrapPos()
	} else {
		imgui.TextUnformatted("No map selected")
	}

	imgui.UnindentV(5)

	imgui.PopFont()
}

func (l *launcher) drawLowerPanel() {
	w, h := contentRegionMax().X, contentRegionMax().Y

	if l.bld.currentMode != Play {
		showProgress := l.bld.currentPMode == Record && l.showProgressBar

		spacing := imgui.FrameHeightWithSpacing()
		if showProgress {
			spacing *= 2
		}

		imgui.SetCursorPos(vec2(20, h-spacing))

		imgui.SetNextItemWidth((imgui.WindowWidth() - 40) / 4)

		l.recordSnowPos = vector.NewVec2f(20+(imgui.WindowWidth()-40)/4/2, h-spacing-2)

		if imgui.BeginCombo("##Watch mode", l.bld.currentPMode.String()) {
			for _, m := range pModes {
				if imgui.SelectableBoolV(m.String(), l.bld.currentPMode == m, 0, vzero()) {
					l.bld.currentPMode = m
				}
			}

			imgui.EndCombo()
		}

		if l.bld.currentPMode != Watch {
			imgui.SameLine()
			if imgui.Button("Configure") {
				l.openPopup(newPopupF("Record settings", popDynamic, func() {
					drawRecordMenu(l.bld)
				}))
			}
		}

		imgui.SetCursorPos(vec2(contentRegionMin().X, h-imgui.FrameHeightWithSpacing()))

		if showProgress {
			if strings.HasPrefix(l.recordStatus, "Done") {
				imgui.PushStyleColorVec4(imgui.ColPlotHistogram, imgui.Vec4{
					X: 0.16,
					Y: 0.75,
					Z: 0.18,
					W: 1,
				})
			} else {
				imgui.PushStyleColorVec4(imgui.ColPlotHistogram, *imgui.StyleColorVec4(imgui.ColCheckMark))
			}

			imgui.ProgressBarV(l.recordProgress, vec2(w/2, imgui.FrameHeight()), l.recordStatus)

			if l.encodeInProgress {
				imgui.PushFont(Font16)

				cPos := imgui.CursorPos()

				imgui.TextUnformatted(l.recordStatusSpeed)

				cPos.X += 95

				imgui.SetCursorPos(cPos)

				eta := int(time.Since(l.encodeStart).Seconds())

				imgui.TextUnformatted("| Elapsed: " + util.FormatSeconds(eta))

				cPos.X += 135

				imgui.SetCursorPos(cPos)

				imgui.TextUnformatted("| " + l.recordStatusETA)

				imgui.PopFont()
			}

			imgui.PopStyleColor()
		}
	}

	fHwS := imgui.FrameHeightWithSpacing()*2 - imgui.CurrentStyle().FramePadding().X

	bW := (w) / 4

	imgui.SetCursorPos(vec2(contentRegionMax().X-w/2.5, h-imgui.FrameHeightWithSpacing()*2))

	centerTable("dansebutton", w/2.5, func() {
		imgui.PushFont(Font48)
		{
			dRun := l.danserRunning && l.bld.currentPMode == Record

			s := (l.bld.currentMode == Replay && l.bld.currentReplay == nil) ||
				(l.bld.currentMode != Replay && l.bld.currentMap == nil) ||
				(l.bld.currentMode == Knockout && l.bld.numKnockoutReplays() == 0)

			if !dRun {
				if s {
					imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), true)
				}
			} else {
				imgui.PopItemFlag()
			}

			name := "danse!"
			if dRun {
				name = "CANCEL"
			}

			if imgui.ButtonV(name, vec2(bW, fHwS)) {
				if dRun {
					if l.danserCmd != nil {
						goroutines.Run(func() {
							res := showMessage(mQuestion, "Do you really want to cancel?")

							if res && l.danserCmd != nil {
								l.danserCmd.Process.Kill()
								l.danserCleanup(false)
							}
						})
					}
				} else {
					if l.selectWindow != nil {
						l.selectWindow.stopPreview()
					}

					log.Println(l.bld.getArguments())

					l.triangleSpeed.AddEventS(l.triangleSpeed.GetTime(), l.triangleSpeed.GetTime()+1000, 50, 1)

					if l.bld.currentPMode != Watch {
						l.startDanser()
					} else {
						goroutines.Run(func() {
							time.Sleep(500 * time.Millisecond)
							l.startDanser()
						})
					}
				}
			}

			if !dRun {
				if s {
					imgui.PopItemFlag()
				}
			} else {
				imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), true)
			}

			imgui.PopFont()
		}
	})
}

func (l *launcher) drawConfigPanel() {
	if l.currentEditor != nil {
		l.currentEditor.setDanserRunning(l.danserRunning && l.bld.currentPMode == Watch)
	}

	w := contentRegionMax().X

	imgui.SetCursorPos(vec2(contentRegionMax().X-float32(w)/2.5, 20))

	if imgui.BeginTableV("rtpanel", 2, imgui.TableFlagsSizingStretchProp, vec2(float32(w)/2.5, 0), -1) {
		imgui.TableSetupColumnV("rtpanel1", imgui.TableColumnFlagsWidthStretch, 0, 0)
		imgui.TableSetupColumnV("rtpanel2", imgui.TableColumnFlagsWidthFixed, 0, 1)

		imgui.TableNextColumn()

		if imgui.ButtonV("Launcher settings", vec2(-1, 0)) {
			wSize := imgui.WindowSize()

			lEditor := newPopupF("About", popCustom, drawLauncherConfig)
			lEditor.width = wSize.X / 2
			lEditor.height = wSize.Y * 0.9

			lEditor.setCloseListener(func() {
				saveLauncherConfig()
			})

			l.openPopup(lEditor)
		}

		imgui.TableNextColumn()

		if imgui.Button("About") {
			l.openPopup(newPopupF("About", popDynamic, func() {
				drawAbout(l.coin.Texture.Texture)
			}))
		}

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.TextUnformatted("Config:")

		imgui.SameLine()

		imgui.SetNextItemWidth(-1)

		mWidth := imgui.CalcItemWidth() - imgui.CurrentStyle().FramePadding().X*2

		if imgui.BeginComboV("##config", l.bld.config, imgui.ComboFlagsHeightLarge) {
			for _, s := range l.configList {
				mWidth = max(mWidth, imgui.CalcTextSizeV(s, false, 0).X+20)
			}

			imgui.SetNextItemWidth(mWidth)

			focusScroll := searchBox("##configSearch", &l.configSearch)

			if !imgui.IsMouseClickedBool(0) && !imgui.IsMouseClickedBool(1) && !imgui.IsAnyItemActive() && !l.configEditOpened && !l.configScrolling {
				imgui.SetKeyboardFocusHereV(-1)
			}

			if imgui.SelectableBool("Create new...") {
				l.newCloneOpened = true
				l.configManiMode = New
			}

			imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
			imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)
			imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, vzero())
			imgui.PushStyleColorVec4(imgui.ColFrameBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})

			searchResults := make([]string, 0, len(l.configList))

			search := strings.ToLower(l.configSearch)

			for _, s := range l.configList {
				if l.configSearch == "" || strings.Contains(strings.ToLower(s), search) {
					searchResults = append(searchResults, s)
				}
			}

			if len(searchResults) > 0 {
				sHeight := float32(min(8, len(searchResults)))*imgui.FrameHeightWithSpacing() - imgui.CurrentStyle().ItemSpacing().Y/2

				if imgui.BeginListBoxV("##blistbox", vec2(mWidth, sHeight)) {
					l.configScrolling = handleDragScroll()
					focusScroll = focusScroll || imgui.IsWindowAppearing()

					for _, s := range searchResults {
						if selectableFocus(s, s == l.bld.config, focusScroll) {
							if s != l.bld.config {
								l.setConfig(s)
							}
						}

						if _, ok := l.cHold[s]; !ok {
							l.cHold[s] = new(bool)
						}

						if imgui.IsMouseClickedBool(1) && imgui.IsItemHovered() {
							*l.cHold[s] = true
							l.configEditOpened = true

							imgui.SetNextWindowPosV(imgui.MousePos(), imgui.CondAlways, vzero())

							imgui.OpenPopupStr("##context" + s)
						}

						befHold := *l.cHold[s]

						if imgui.BeginPopupModalV("##context"+s, l.cHold[s], imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar) {
							if s != "default" {
								if imgui.SelectableBool("Rename") {
									l.newCloneOpened = true
									l.configPrevName = s
									l.configManiMode = Rename
								}
							}

							if imgui.SelectableBool("Clone") {
								l.newCloneOpened = true
								l.configPrevName = s
								l.configManiMode = Clone
							}

							if s != "default" {
								if imgui.SelectableBool("Remove") {
									if showMessage(mQuestion, "Are you sure you want to remove \"%s\" profile?", s) {
										l.removeConfig(s)
									}
								}
							}

							if (imgui.IsMouseClickedBool(0) || imgui.IsMouseClickedBool(1)) && !imgui.IsWindowHoveredV(imgui.HoveredFlagsRootAndChildWindows|imgui.HoveredFlagsAllowWhenBlockedByActiveItem|imgui.HoveredFlagsAllowWhenBlockedByPopup) {
								*l.cHold[s] = false
								//imgui.CloseCurrentPopup()
							}

							imgui.EndPopup()
						}

						if befHold && !*l.cHold[s] {
							l.configEditOpened = false
						}
					}

					imgui.EndListBox()
				}
			}

			imgui.PopStyleVar()
			imgui.PopStyleVar()
			imgui.PopStyleVar()
			imgui.PopStyleColor()

			imgui.EndCombo()
		}

		imgui.TableNextColumn()

		dRun := l.danserRunning && l.bld.currentPMode == Watch

		if dRun {
			imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), false)
		}

		if imgui.ButtonV("Edit", vec2(-1, 0)) {
			l.openCurrentSettingsEditor()
		}

		if dRun {
			imgui.PopItemFlag()
		}

		imgui.EndTable()
	}

	if l.newCloneOpened {
		popupSmall("Clone/new box", &l.newCloneOpened, true, 0, 0, func() {
			if imgui.BeginTable("rfa", 1) {
				imgui.TableNextColumn()

				imgui.TextUnformatted("Name:")

				imgui.SameLine()

				imgui.SetNextItemWidth(imgui.TextLineHeight() * 10)

				if inputTextV("##nclonename", &l.newCloneName, imgui.InputTextFlagsCallbackCharFilter, imguiPathFilter) {
					l.newCloneName = strings.TrimSpace(l.newCloneName)
				}

				if !imgui.IsAnyItemActive() && !imgui.IsMouseClickedBool(0) {
					imgui.SetKeyboardFocusHereV(-1)
				}

				imgui.TableNextColumn()

				cPos := imgui.CursorPos()

				imgui.SetCursorPos(vec2(cPos.X+(imgui.ContentRegionAvail().X-imgui.CalcTextSizeV("Save", false, 0).X-imgui.CurrentStyle().FramePadding().X*2)/2, cPos.Y))

				e := l.newCloneName == ""

				if e {
					imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), true)
				}

				if imgui.Button("Save##newclone") || (!e && (imgui.IsKeyPressedBool(imgui.KeyEnter) || imgui.IsKeyPressedBool(imgui.KeyKeypadEnter))) {
					_, err := os.Stat(filepath.Join(env.ConfigDir(), l.newCloneName+".json"))
					if err == nil {
						showMessage(mError, "Config with that name already exists!\nPlease pick a different name")
					} else {
						log.Println("ok")
						switch l.configManiMode {
						case Rename:
							l.renameConfig(l.configPrevName, l.newCloneName)
						case Clone:
							l.cloneConfig(l.configPrevName, l.newCloneName)
						case New:
							l.createConfig(l.newCloneName)
						}

						l.newCloneOpened = false
						l.newCloneName = ""
					}
				}

				if e {
					imgui.PopItemFlag()
				}

				imgui.EndTable()
			}
		})
	}
}

func (l *launcher) openCurrentSettingsEditor() {
	saveFunc := func() {
		settings.SaveCredentials(false)
		l.currentConfig.Save("", false)

		if !compareDirs(l.currentConfig.General.OsuSongsDir, settings.General.OsuSongsDir) {
			showMessage(mInfo, "This config has different osu! Songs directory.\nRestart the launcher to see updated maps")
		}
	}

	if l.currentEditor == nil || l.currentEditor.current != l.currentConfig {
		l.currentEditor = newSettingsEditor(l.currentConfig)
	}

	l.currentEditor.setDanserRunning(l.danserRunning)
	l.currentEditor.setCloseListener(saveFunc)
	l.currentEditor.setSaveListener(saveFunc)

	l.openPopup(l.currentEditor)
}

func (l *launcher) tryCreateDefaultConfig() {
	_, err := os.Stat(filepath.Join(env.ConfigDir(), "default.json"))
	if err != nil {
		l.createConfig("default")
	}
}

func (l *launcher) createConfig(name string) {
	vm := glfw.GetPrimaryMonitor().GetVideoMode()

	conf := settings.NewConfigFile()
	conf.Graphics.SetDefaults(int64(vm.Width), int64(vm.Height))
	conf.Save(filepath.Join(env.ConfigDir(), name+".json"), true)

	l.createConfigList()

	l.setConfig(name)
}

func (l *launcher) removeConfig(name string) {
	os.Remove(filepath.Join(env.ConfigDir(), name+".json"))

	l.createConfigList()

	if l.bld.config == name {
		l.setConfig("default")
	}
}

func (l *launcher) cloneConfig(toClone, name string) {
	cConfig, err := l.loadConfig(toClone)

	if err != nil {
		showMessage(mError, err.Error())
		return
	}

	cConfig.Save(filepath.Join(env.ConfigDir(), name+".json"), true)

	l.createConfigList()

	l.setConfig(name)
}

func (l *launcher) renameConfig(toRename, name string) {
	cConfig, err := l.loadConfig(toRename)

	if err != nil {
		showMessage(mError, err.Error())
		return
	}

	cConfig.Save(filepath.Join(env.ConfigDir(), name+".json"), true)

	os.Remove(filepath.Join(env.ConfigDir(), toRename+".json"))

	l.createConfigList()

	l.setConfig(name)
}

func (l *launcher) createConfigList() {
	l.configList = []string{}

	filepath.Walk(env.ConfigDir(), func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			stPath := strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(path, ".json"), env.ConfigDir()+string(os.PathSeparator)), "\\", "/")

			if stPath != "credentials" && stPath != "default" && stPath != "launcher" {
				l.configList = append(l.configList, stPath)
			}
		}

		return nil
	})

	log.Println("Available configs:", strings.Join(l.configList, ", "))

	sort.Strings(l.configList)

	l.configList = append([]string{"default"}, l.configList...)
}

func (l *launcher) loadConfig(name string) (*settings.Config, error) {
	f, err := os.Open(filepath.Join(env.ConfigDir(), name+".json"))
	if err != nil {
		return nil, fmt.Errorf("invalid file state. Please don't modify the folder while launcher is running. Error: %s", err)
	}

	defer f.Close()

	return settings.LoadConfig(f)
}

func (l *launcher) setConfig(s string) {
	eConfig, err := l.loadConfig(s)

	if err != nil {
		showMessage(mError, "Failed to read \"%s\" profile. Error: %s", s, err)
	} else {
		if !compareDirs(eConfig.General.OsuSongsDir, settings.General.OsuSongsDir) {
			showMessage(mInfo, "This config has different osu! Songs directory.\nRestart the launcher to see updated maps")
		}

		l.bld.config = s
		l.currentConfig = eConfig

		*launcherConfig.Profile = l.bld.config
		saveLauncherConfig()
	}
}

func (l *launcher) startDanser() {
	l.recordProgress = 0
	l.recordStatus = ""
	l.recordStatusSpeed = ""
	l.recordStatusETA = ""
	l.encodeInProgress = false

	dExec := os.Args[0]

	if build.Stream == "Release" {
		dExec = filepath.Join(env.LibDir(), build.DanserExec)
	}

	l.danserCmd = exec.Command(dExec, l.bld.getArguments()...)

	rFile, oFile, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	l.danserCmd.Stderr = os.Stderr
	l.danserCmd.Stdin = os.Stdin
	l.danserCmd.Stdout = io.MultiWriter(os.Stdout, oFile)

	err = l.danserCmd.Start()
	if err != nil {
		showMessage(mError, "danser failed to start! %s", err.Error())
		return
	}

	if l.bld.currentPMode == Watch {
		l.win.Iconify()
	} else if l.bld.currentPMode == Record {
		l.showProgressBar = true
	}

	l.danserRunning = true
	l.recordStatus = "Preparing..."

	panicMessage := ""
	panicWait := &sync.WaitGroup{}
	panicWait.Add(1)

	resultFile := ""

	goroutines.Run(func() {
		sc := bufio.NewScanner(rFile)

		l.encodeInProgress = false

		for sc.Scan() {
			line := sc.Text()

			if strings.Contains(line, "Launcher: Open settings") {
				if l.currentEditor == nil || !l.currentEditor.opened {
					l.openCurrentSettingsEditor()
				}

				l.win.Restore()
				l.win.Focus()
			}

			if strings.Contains(line, "panic:") {
				panicMessage = line[strings.Index(line, "panic:"):]
				panicWait.Done()
			}

			if strings.Contains(line, "Starting encoding!") {
				l.encodeInProgress = true
				l.encodeStart = time.Now()

				platform.StartProgress(l.win)
			}

			if strings.Contains(line, "Finishing rendering") {
				l.encodeInProgress = false

				l.recordProgress = 1
				platform.SetProgress(l.win, 100, 100)
				l.recordStatus = "Finalizing..."
				l.recordStatusSpeed = ""
				l.recordStatusETA = ""
			}

			if idx := strings.Index(line, "Video is available at: "); idx > -1 {
				resultFile = strings.TrimPrefix(line[idx:], "Video is available at: ")
			}

			if idx := strings.Index(line, "Screenshot "); idx > -1 && strings.Contains(line, " saved!") {
				resultFile = strings.TrimSuffix(strings.TrimPrefix(line[idx:], "Screenshot "), " saved!")
				resultFile = filepath.Join(env.DataDir(), "screenshots", resultFile)
			}

			if strings.Contains(line, "Progress") && l.encodeInProgress {
				line = line[strings.Index(line, "Progress"):]

				rStats := strings.Split(line, ",")

				spl := strings.TrimSpace(strings.Split(rStats[0], ":")[1])

				l.recordStatus = spl

				l.recordStatusSpeed = strings.TrimSpace(rStats[1])
				l.recordStatusETA = strings.TrimSpace(rStats[2])

				speed := strings.TrimSpace(strings.Split(rStats[1], ":")[1])

				speedP, _ := strconv.ParseFloat(speed[:len(speed)-1], 32)

				l.triangleSpeed.AddEvent(l.triangleSpeed.GetTime(), l.triangleSpeed.GetTime()+500, speedP)

				at, _ := strconv.Atoi(spl[:len(spl)-1])

				l.recordProgress = float32(at) / 100
				platform.SetProgress(l.win, at, 100)
			}
		}

		l.recordProgress = 1
		l.recordStatus = "Done in " + util.FormatSeconds(int(time.Since(l.encodeStart).Seconds()))
		l.recordStatusSpeed = ""
		l.recordStatusETA = ""
	})

	goroutines.Run(func() {
		err = l.danserCmd.Wait()

		l.danserCleanup(err == nil)

		if err != nil {
			panicWait.Wait()

			goroutines.CallMain(func() {
				pMsg := panicMessage
				if idx := strings.Index(pMsg, "Error:"); idx > -1 {
					pMsg = pMsg[:idx-1] + "\n\n" + pMsg[idx+7:]
				}

				showMessage(mError, "danser crashed! %s\n\n%s", err.Error(), pMsg)
			})
		} else if l.bld.currentPMode != Watch && l.bld.currentMode != Play {
			if launcherConfig.ShowFileAfter && resultFile != "" {
				platform.ShowFileInManager(resultFile)
			}

			platform.Beep(platform.Ok)
		}

		rFile.Close()
		oFile.Close()

		l.win.Restore()
	})
}

func (l *launcher) danserCleanup(success bool) {
	l.recordStatusSpeed = ""
	l.recordStatusETA = ""
	l.danserRunning = false
	l.triangleSpeed.AddEvent(l.triangleSpeed.GetTime(), l.triangleSpeed.GetTime()+500, 1)
	l.danserCmd = nil

	if !success {
		platform.StopProgress(l.win)
		l.recordStatus = ""
		l.showProgressBar = false
	}
}

func (l *launcher) openPopup(p iPopup) {
	p.open()
	l.popupStack = append(l.popupStack, p)
}

func (l *launcher) loadOSZs(names []string) {
	closeWatcher()
	l.beatmapDirUpdated = false

	reload := false

	for _, name := range names {
		if strings.HasSuffix(name, ".osz") {
			fileName := filepath.Base(name)

			err := files.MoveFile(name, filepath.Join(settings.General.GetSongsDir(), fileName))
			if err != nil {
				showMessage(mError, "Failed to move \"%s\" to Songs folder: %s", name, err)
			} else {
				reload = true
			}
		}
	}

	if reload {
		l.reloadMaps(func() {
			if l.selectWindow == nil {
				l.selectWindow = newSongSelectPopup(l.bld, l.beatmaps)
			}

			if l.bld.knockoutReplays == nil && l.bld.currentReplay == nil {
				l.selectWindow.selectNewest()
			}
		})
	} else {
		l.setupWatcher()
	}
}

func (l *launcher) reloadMaps(after func()) {
	goroutines.RunOS(func() {
		l.mapsLoaded = false
		l.loadBeatmaps()

		// Add to main thread scheduler to avoid race conditions
		goroutines.CallNonBlockMain(func() {
			if l.bld.currentMap != nil {
				for _, b := range l.beatmaps {
					if b.MD5 == l.bld.currentMap.MD5 {
						l.bld.currentMap = b
						break
					}
				}
			}

			if l.selectWindow != nil {
				l.selectWindow.setBeatmaps(l.beatmaps)
			}

			if after != nil {
				after()
			}
		})
	})
}

func (l *launcher) setupWatcher() {
	setupWatcher(settings.General.GetSongsDir(), func(event fsnotify.Event) {
		delay := 3000.0                   //Wait for the last map to load on osu side
		if launcherConfig.AutoRefreshDB { // Wait a bit longer if we're about to refresh the DB automatically
			delay = 6000
		}

		l.showBeatmapAlert = qpc.GetMilliTimeF() + delay
		l.beatmapDirUpdated = true
	})
}
