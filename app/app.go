package app

import "C"
import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	difficulty2 "github.com/wieku/danser-go/app/beatmap/difficulty"
	camera2 "github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/database"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/ffmpeg"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/goroutines"
	batch2 "github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/platform"
	"github.com/wieku/danser-go/framework/profiler"
	"github.com/wieku/danser-go/framework/qpc"
	"github.com/wieku/danser-go/framework/util"
	"github.com/wieku/rplpa"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"
)

const (
	base           = "Specify the"
	artistDesc     = base + " artist of a song"
	titleDesc      = base + " title of a song"
	creatorDesc    = base + " creator of a map"
	difficultyDesc = base + " difficulty(version) of a map"
	replayDesc     = "Play a map from specific replay file. Overrides -knockout, -mods and all beatmap arguments."
	shorthand      = " (shorthand)"
)

var player states.State

var scheduleScreenshot = false

var batch *batch2.QuadBatch

var win *glfw.Window
var limiter *frame.Limiter
var screenFBO *buffer.Framebuffer
var lastSamples int
var lastVSync bool

var output string

var recordMode bool
var screenshotMode bool
var screenshotTime float64

var preciseProgress bool

var monitorHz int

func run() {
	defer func() {
		if err := recover(); err != nil {
			stackTrace := goroutines.GetStackTrace(4)
			closeHandler(err, stackTrace)
		}
	}()

	goroutines.CallMain(func() {
		id := flag.Int64("id", -1, "Specify the beatmap id. Overrides other beatmap search flags")

		md5 := flag.String("md5", "", "Specify the beatmap md5 hash. Overrides other beatmap search flags")

		artist := flag.String("artist", "", artistDesc)
		flag.StringVar(artist, "a", "", artistDesc+shorthand)

		title := flag.String("title", "", titleDesc)
		flag.StringVar(title, "t", "", titleDesc+shorthand)

		difficulty := flag.String("difficulty", "", difficultyDesc)
		flag.StringVar(difficulty, "d", "", difficultyDesc+shorthand)

		creator := flag.String("creator", "", creatorDesc)
		flag.StringVar(creator, "c", "", creatorDesc+shorthand)

		settingsVersion := flag.String("settings", "", "Specify settings version, -settings=b/abc means that settings/b/abc.json will be loaded. \"Credentials\"")
		cursors := flag.Int("cursors", 1, "How many repeated cursors should be visible, recommended 2 for mirror, 8 for mandala")
		tag := flag.Int("tag", 1, "How many cursors should be \"playing\" specific map. 2 means that 1st cursor clicks the 1st object, 2nd clicks 2nd object, 1st clicks 3rd and so on")

		knockout := flag.Bool("knockout", false, "Use (classic) knockout feature. Replays are sourced from \"replays/{a}\" where {a} is an md5 hash of .osu file. Danser automatically organizes replay files put directly in \"replays\", using maps' md5s provided by the replay files.")
		knockout2 := flag.String("knockout2", "", "Use (new) knockout feature, JSON list of paths to compatible replay files has to be provided. \"Knockout.ExcludeMods\" and \"Knockout.MaxPlayers\" options are ignored, they have to be filtered beforehand.")

		speed := flag.Float64("speed", 1.0, "Specify music's speed, set to 1.5 to have DoubleTime mod experience")
		pitch := flag.Float64("pitch", 1.0, "Specify music's pitch, set to 1.5 with -speed=1.5 to have Nightcore mod experience")
		debug := flag.Bool("debug", false, "Show info about map and rendering engine, overrides Graphics.ShowFPS setting. Ignored in record/screenshot modes.")

		gldebug := flag.Bool("gldebug", false, "Turns on OpenGL debug logging, may reduce performance heavily")

		play := flag.Bool("play", false, "Practice playing osu!standard maps")
		start := flag.Float64("start", 0, "Start at the given time in seconds")
		end := flag.Float64("end", math.Inf(1), "End at the given time in seconds")

		skip := flag.Bool("skip", false, "Skip straight to map's drain time")

		quickstart := flag.Bool("quickstart", false, "Sets -skip flag, sets LeadInTime and LeadInHold settings temporarily to 0")

		record := flag.Bool("record", false, "Records a video")
		out := flag.String("out", "", "If -ss flag is used, sets the name of screenshot, extension is PNG. If not, it overrides -record flag, specifies the name of recorded video file, extension is managed by settings")
		ss := flag.Float64("ss", math.NaN(), "Screenshot mode. Snap single frame from danser at given time in seconds. Specify the name of file by -out, resolution is managed by Recording settings")

		mods := flag.String("mods", "", "Specify beatmap/play mods")
		mods2 := flag.String("mods2", "", "Specify beatmap/play mods, lazer style")

		replay := flag.String("replay", "", replayDesc)
		flag.StringVar(replay, "r", "", replayDesc+shorthand)

		skin := flag.String("skin", "", "Replace Skin.CurrentSkin setting temporarily")

		noDbCheck := flag.Bool("nodbcheck", false, "Don't validate the database and only import new beatmap sets if there are any. Useful for slow drives.")
		noUpdCheck := flag.Bool("noupdatecheck", strings.HasPrefix(env.LibDir(), "/usr/lib/"), "Don't check for updates. Speeds up startup if older version of danser is needed for various reasons. Has no effect if danser is running as a linux package")

		ar := flag.Float64("ar", math.NaN(), "Modify map's AR, only in cursordance/play modes")
		od := flag.Float64("od", math.NaN(), "Modify map's OD, only in cursordance/play modes")
		cs := flag.Float64("cs", math.NaN(), "Modify map's CS, only in cursordance/play modes")
		hp := flag.Float64("hp", math.NaN(), "Modify map's HP, only in cursordance/play modes")

		offset := flag.Int("offset", 0, "Specify local audio offset in ms. Applies to recordings, unlike 'Audio.Offset'. Inverted compared to stable's local offset.")

		flag.BoolVar(&preciseProgress, "preciseprogress", false, "Show rendering progress in 1% increments")

		flag.Parse()

		if *mods != "" && *mods2 != "" {
			panic("You can't specify classic and lazer mods at the same time")
		}

		var knockoutReplays []string

		if *knockout2 != "" {
			if err := json.Unmarshal([]byte(*knockout2), &knockoutReplays); err != nil {
				panic(fmt.Sprintf("Failed to parse replay list: %s", err))
			}

			*knockout = true
		}

		if !*noUpdCheck {
			checkForUpdates()
		}

		if *out != "" {
			output = *out
			if math.IsNaN(*ss) {
				*record = true
			}
		}

		recordMode = *record
		screenshotMode = !math.IsNaN(*ss)
		screenshotTime = *ss

		if *record && *play {
			panic("Incompatible flags selected: -record, -play")
		} else if *replay != "" && *play {
			panic("Incompatible flags selected: -replay, -play")
		} else if *knockout && *play {
			panic("Incompatible flags selected: -knockout, -play")
		} else if *replay != "" && *knockout {
			panic("Incompatible flags selected: -replay, -knockout")
		} else if screenshotMode && *play {
			panic("Incompatible flags selected: -ss, -play")
		} else if screenshotMode && recordMode {
			panic("Incompatible flags selected: -ss, -record")
		}

		modsParsed := difficulty2.ParseMods(*mods)
		var modsNew []rplpa.ModInfo = nil

		if *replay != "" {
			bytes, err := ioutil.ReadFile(*replay)
			if err != nil {
				panic(err)
			}

			rp, err := rplpa.ParseReplay(bytes)
			if err != nil {
				panic(err)
			}

			if rp.PlayMode != 0 {
				panic("Modes other than osu!standard are not supported")
			}

			if rp.ReplayData == nil || len(rp.ReplayData) < 2 {
				panic("Replay is missing input data")
			}

			*md5 = rp.BeatmapMD5
			*id = -1
			modsParsed = difficulty2.Modifier(rp.Mods)

			if rp.ScoreInfo != nil && rp.ScoreInfo.Mods != nil && len(rp.ScoreInfo.Mods) > 0 {
				modsNew = make([]rplpa.ModInfo, 0, len(rp.ScoreInfo.Mods))

				for _, mod := range rp.ScoreInfo.Mods {
					modsNew = append(modsNew, *mod)
				}
			}

			if rp.OsuVersion >= 30000000 { // Lazer is 1000 years in the future
				modsParsed |= difficulty2.Lazer

				if modsNew != nil {
					modsNew = append(modsNew, rplpa.ModInfo{Acronym: "LZ"})
				}
			}

			*knockout = true
			settings.REPLAY = *replay
		}

		if *mods2 != "" {
			var mods2I []rplpa.ModInfo

			if err := json.Unmarshal([]byte(*mods2), &mods2I); err != nil {
				panic(fmt.Sprintf("Failed to parse replay list: %s", err))
			}

			modsNew = mods2I
		}

		if modsNew != nil {
			tempDiff := difficulty2.NewDifficulty(1, 1, 1, 1)
			tempDiff.SetMods2(modsNew)
			modsParsed = tempDiff.Mods
		}

		if !modsParsed.Compatible() {
			panic("Incompatible mods selected!")
		}

		closeAfterSettingsLoad := false

		if (*md5+*artist+*title+*difficulty+*creator) == "" && *id < 0 {
			log.Println("No beatmap specified, closing...")
			closeAfterSettingsLoad = true
		}

		settings.DEBUG = *debug
		settings.KNOCKOUT = *knockout
		settings.KNOCKOUTREPLAYS = knockoutReplays
		settings.PLAY = *play
		settings.DIVIDES = *cursors
		settings.TAG = *tag
		settings.SPEED = *speed
		settings.PITCH = *pitch
		settings.SKIP = *skip
		settings.START = *start
		settings.END = *end
		settings.RECORD = recordMode || screenshotMode
		settings.LOCALOFFSET = *offset

		if *settingsVersion == "credentials" || *settingsVersion == "launcher" {
			panic(fmt.Sprintf("flag -settings: name \"%s\" is forbidden", *settingsVersion))
		}

		newSettings := settings.LoadSettings(*settingsVersion)

		log.Println("Current config:", settings.GetCompressedString())

		if !newSettings && len(os.Args) == 1 {
			platform.OpenURL("https://youtu.be/dQw4w9WgXcQ")
			closeAfterSettingsLoad = true
		}

		player = nil
		var beatMap *beatmap.BeatMap = nil

		if !closeAfterSettingsLoad {
			err := database.Init()
			if err != nil {
				log.Println("Failed to initialize database:", err)
			} else {
				beatmaps := database.LoadBeatmaps(*noDbCheck, nil)

				if *id > -1 {
					for _, b := range beatmaps {
						if b.ID == *id {
							beatMap = b

							break
						}
					}
				} else if *md5 != "" {
					for _, b := range beatmaps {
						if strings.EqualFold(b.MD5, *md5) {
							beatMap = b

							break
						}
					}
				} else {
					for _, b := range beatmaps {
						if (*artist == "" || strings.EqualFold(*artist, b.Artist)) &&
							(*title == "" || strings.EqualFold(*title, b.Name)) &&
							(*difficulty == "" || strings.EqualFold(*difficulty, b.Difficulty)) &&
							(*creator == "" || strings.EqualFold(*creator, b.Creator)) {
							beatMap = b

							break
						}
					}

					if beatMap == nil {
						log.Println("Beatmap with exact parameters not found, searching partially...")
						for _, b := range beatmaps {
							if (*artist == "" || strings.Contains(strings.ToLower(b.Artist), strings.ToLower(*artist))) &&
								(*title == "" || strings.Contains(strings.ToLower(b.Name), strings.ToLower(*title))) &&
								(*difficulty == "" || strings.Contains(strings.ToLower(b.Difficulty), strings.ToLower(*difficulty))) &&
								(*creator == "" || strings.Contains(strings.ToLower(b.Creator), strings.ToLower(*creator))) {
								beatMap = b

								break
							}
						}
					}
				}
			}

			if beatMap == nil {
				log.Println("Beatmap not found, closing...")
				closeAfterSettingsLoad = true
			} else {
				beatMap.UpdatePlayStats()
				database.UpdatePlayStats(beatMap)
			}

			database.Close()
		}

		assets.Init(build.Stream == "Dev")

		if !closeAfterSettingsLoad {
			log.Println("Initializing GLFW...")
		}

		err := glfw.Init()
		if err != nil {
			panic("Failed to initialize GLFW: " + err.Error())
		}

		if !closeAfterSettingsLoad {
			log.Println("GLFW Initialized!")
		}

		platform.SetupContext()

		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.Samples, 0)
		glfw.WindowHint(glfw.Visible, glfw.False)

		monitor := glfw.GetPrimaryMonitor()
		mWidth, mHeight := monitor.GetVideoMode().Width, monitor.GetVideoMode().Height

		monitorHz = monitor.GetVideoMode().RefreshRate

		if newSettings {
			settings.Graphics.SetDefaults(int64(mWidth), int64(mHeight))
			settings.Save()
		}

		if closeAfterSettingsLoad {
			os.Exit(0)
		}

		allowDA := false

		// if map was launched not in knockout or play mode but AT mod is present, use replay mode for danser, allowing custom ar,od,cs,hp
		if !settings.KNOCKOUT && modsParsed.Active(difficulty2.Autoplay) {
			settings.PLAY = false
			settings.KNOCKOUT = true
			settings.Knockout.MaxPlayers = 0
			allowDA = true
		}

		lastSamples = int(settings.Graphics.MSAA)

		if strings.TrimSpace(*skin) != "" {
			settings.Skin.CurrentSkin = *skin
		}

		if *quickstart {
			settings.SKIP = true
			settings.Playfield.LeadInTime = 0
			settings.Playfield.LeadInHold = 0
		}

		if settings.RECORD {
			//HACK: some in-app variables depend on these settings so we force them here
			settings.Graphics.VSync = false
			settings.Graphics.ShowFPS = false
			settings.DEBUG = false
			settings.Graphics.Fullscreen = false
			settings.Graphics.WindowWidth = int64(settings.Recording.FrameWidth)
			settings.Graphics.WindowHeight = int64(settings.Recording.FrameHeight)
			settings.Playfield.LeadInTime = 0
		}

		if screenshotMode {
			settings.Playfield.LeadInHold = 0
			settings.START = screenshotTime - 5
			settings.SKIP = false
		}

		log.Println("Creating window...")

		if settings.Graphics.Fullscreen {
			glfw.WindowHint(glfw.RedBits, monitor.GetVideoMode().RedBits)
			glfw.WindowHint(glfw.GreenBits, monitor.GetVideoMode().GreenBits)
			glfw.WindowHint(glfw.BlueBits, monitor.GetVideoMode().BlueBits)
			glfw.WindowHint(glfw.RefreshRate, monitor.GetVideoMode().RefreshRate)
			//glfw.WindowHint(glfw.Decorated, glfw.False)
			win, err = glfw.CreateWindow(int(settings.Graphics.Width), int(settings.Graphics.Height), "danser", monitor, nil)
		} else {
			win, err = glfw.CreateWindow(int(settings.Graphics.WindowWidth), int(settings.Graphics.WindowHeight), "danser", nil, nil)
		}

		if err != nil {
			panic(err)
		}

		if !*record {
			win.SetFocusCallback(func(w *glfw.Window, focused bool) {
				log.Println("Focus changed: ", focused)
				input.Focused = focused
			})
		}

		win.SetTitle("danser " + build.VERSION + " - " + beatMap.Artist + " - " + beatMap.Name + " [" + beatMap.Difficulty + "]")
		input.Win = win

		if cTime := time.Now(); cTime.Month() == 12 && cTime.Day() >= 6 {
			platform.LoadIcons(win, "dansercoin", "-s")
		} else {
			platform.LoadIcons(win, "dansercoin", "")
		}

		win.MakeContextCurrent()

		log.Println("Window created!")

		err = platform.GLInit(*gldebug)
		if err != nil {
			panic("Failed to initialize OpenGL: " + err.Error())
		}

		if !settings.RECORD {
			discord.Connect()
			win.Show()
		}

		gl.Enable(gl.BLEND)
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		file, _ := assets.Open("assets/fonts/Quicksand-Bold.ttf")
		font.LoadFont(file)
		file.Close()

		batch = batch2.NewQuadBatch()
		batch.Begin()
		batch.SetColor(1, 1, 1, 1)
		camera := camera2.NewCamera()
		camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true)
		camera.SetOrigin(vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
		camera.Update()
		batch.SetCamera(camera.GetProjectionView())

		font.GetFont("Quicksand Bold").Draw(batch, 0, settings.Graphics.GetHeightF()-10, 32, "Loading...")

		batch.End()
		win.SwapBuffers()

		glfw.SwapInterval(1)
		lastVSync = true

		bass.Init(settings.RECORD)
		audio.LoadSamples()

		if settings.PLAY || !settings.KNOCKOUT || allowDA {
			if modsNew == nil {
				modsNew = modsParsed.ConvertToModInfoList()
			}

			daMap := make(map[string]any)

			if !math.IsNaN(*ar) {
				daMap["approach_rate"] = *ar
			}

			if !math.IsNaN(*od) {
				daMap["overall_difficulty"] = *od
			}

			if !math.IsNaN(*cs) {
				daMap["circle_size"] = *cs
			}

			if !math.IsNaN(*hp) {
				daMap["drain_rate"] = *hp
			}

			// Add DA only if DA hasn't been added already
			if len(daMap) > 0 && !slices.ContainsFunc(modsNew, func(info rplpa.ModInfo) bool { return info.Acronym == "DA" }) {
				modsNew = append(modsNew, rplpa.ModInfo{
					Acronym:  "DA",
					Settings: daMap,
				})
			}

			if math.Abs(settings.SPEED-1) > 0.001 {
				skipMods := []string{"HT", "DC", "DT", "NC"}

				found := slices.ContainsFunc(modsNew, func(info rplpa.ModInfo) bool { return slices.Contains(skipMods, info.Acronym) })

				// Don't modify current mods
				//if settings.SPEED >= 1 {
				//	if i := slices.IndexFunc(modsNew, func(info rplpa.ModInfo) bool {
				//		return info.Acronym == "DT" || info.Acronym == "NC"
				//	}); i != -1 {
				//		found = true
				//		modsNew[i].Settings["speed_change"] = settings.SPEED
				//	}
				//} else {
				//	if i := slices.IndexFunc(modsNew, func(info rplpa.ModInfo) bool {
				//		return info.Acronym == "HT" || info.Acronym == "DC"
				//	}); i != -1 {
				//		found = true
				//		modsNew[i].Settings["speed_change"] = settings.SPEED
				//	}
				//}

				if !found {
					modsNew = slices.DeleteFunc(modsNew, func(info rplpa.ModInfo) bool {
						return info.Acronym == "DT" || info.Acronym == "NC" || info.Acronym == "HT" || info.Acronym == "DC"
					})

					acr := "HT"
					if settings.SPEED >= 1 {
						acr = "DT"
					}

					modsNew = append(modsNew, rplpa.ModInfo{
						Acronym: acr,
						Settings: map[string]any{
							"speed_change": settings.SPEED,
						},
					})
				}

				settings.SPEED = 1
			}
		}

		if modsNew != nil {
			beatMap.Diff.SetMods2(modsNew)
		} else {
			beatMap.Diff.SetMods(modsParsed)
		}

		beatmap.ParseTimingPointsAndPauses(beatMap)
		beatmap.ParseObjects(beatMap, false, true)
		beatMap.LoadCustomSamples()
		player = states.NewPlayer(beatMap)

		limiter = frame.NewLimiter(int(settings.Graphics.FPSCap))
	})

	if recordMode {
		mainLoopRecord()
	} else if screenshotMode {
		mainLoopSS()
	} else {
		mainLoopNormal()
	}
}

func mainLoopRecord() {
	count := int64(0)

	fps := float64(settings.Recording.FPS)
	audioFPS := 1000.0

	if settings.Recording.MotionBlur.Enabled {
		fps *= float64(settings.Recording.MotionBlur.OversampleMultiplier)
	}

	w, h := int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight())

	var fbo *buffer.Framebuffer

	goroutines.CallMain(func() {
		fbo = buffer.NewFrameMultisampleScreen(w, h, false, 0)
	})

	ffmpeg.StartFFmpeg(int(fps), w, h, audioFPS, output)

	updateFPS := max(fps, 1000)
	updateDelta := 1000 / updateFPS
	fpsDelta := 1000 / fps
	audioDelta := 1000.0 / audioFPS

	deltaSumF := fpsDelta
	deltaSumA := 0.0

	p, _ := player.(*states.Player)

	lastCount := int64(0)
	lastRealTime := qpc.GetMilliTimeF()

	var lastProgress, progress int

	if preciseProgress {
		lastProgress = -1
	}

	for !p.Update(updateDelta) {
		deltaSumA += updateDelta
		for deltaSumA >= audioDelta {
			ffmpeg.PushAudio()

			deltaSumA -= audioDelta
		}

		deltaSumF += updateDelta
		if deltaSumF >= fpsDelta {
			goroutines.CallMain(func() {
				fbo.Bind()

				ffmpeg.PreFrame()

				viewport.Push(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
				pushFrame()
				viewport.Pop()

				ffmpeg.MakeFrame()

				fbo.Unbind()

				count++

				timeOffset := p.GetTimeOffset()
				progress = int(math.Round(timeOffset / p.RunningTime * 100))

				if (preciseProgress || progress%5 == 0) && lastProgress != progress {
					speed := float64(count-lastCount) * (1000 / fps) / (qpc.GetMilliTimeF() - lastRealTime)

					eta := int((p.RunningTime - timeOffset) / 1000 / speed)

					etaText := util.FormatSeconds(eta)

					if settings.Recording.ShowFFmpegLogs {
						fmt.Println()
					}

					log.Println(fmt.Sprintf("Progress: %d%%, Speed: %.2fx, ETA: %s", progress, speed, etaText))

					lastProgress = progress

					lastCount = count
					lastRealTime = qpc.GetMilliTimeF()
				}
			})

			deltaSumF -= fpsDelta
		}
	}

	goroutines.CallMain(func() {
		ffmpeg.StopFFmpeg()
	})
}

func mainLoopSS() {
	w, h := int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight())

	var fbo *buffer.Framebuffer

	goroutines.CallMain(func() {
		fbo = buffer.NewFrameMultisampleScreen(w, h, false, 0)
	})

	p, _ := player.(*states.Player)

	for !p.Update(1) {
		if p.GetTime() >= screenshotTime*1000 {
			log.Println("Scheduling screenshot")
			goroutines.CallMain(func() {
				fbo.Bind()

				viewport.Push(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
				pushFrame()
				viewport.Pop()

				utils.MakeScreenshot(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), output, false)

				fbo.Unbind()
			})

			break
		}
	}
}

func mainLoopNormal() {
	goroutines.CallMain(func() {
		win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if action == glfw.Press {
				switch key {
				case glfw.KeyF11:
					switch mods {
					case glfw.ModShift:
						settings.PerfGraph = !settings.PerfGraph
					case glfw.ModControl:
						settings.CallGraph = !settings.CallGraph
					default:
						settings.DEBUG = !settings.DEBUG
					}
				case glfw.KeyEscape:
					win.SetShouldClose(true)
				case glfw.KeyMinus:
					settings.DIVIDES = max(1, settings.DIVIDES-1)
				case glfw.KeyEqual:
					settings.DIVIDES += 1
				case glfw.KeyO:
					if mods == glfw.ModControl {
						log.Println("Launcher: Open settings")
					}
				default:
					if kName, ok := platform.GetKeyName(key, scancode); ok && kName == settings.Input.ScreenshotKey {
						scheduleScreenshot = true
					}
				}
			}

			input.CallListeners(w, key, scancode, action, mods)
		})
	})

	goroutines.RunMainLoop(func() bool {
		return !win.ShouldClose()
	}, func() {
		if lastVSync != settings.Graphics.VSync {
			if settings.Graphics.VSync {
				glfw.SwapInterval(1)
			} else {
				glfw.SwapInterval(0)
			}

			lastVSync = settings.Graphics.VSync
		}

		profiler.StartGroup("glfw.PollEvents", profiler.PInput)
		glfw.PollEvents()
		profiler.EndGroup()

		pushFrame()

		if scheduleScreenshot {
			w, h := win.GetFramebufferSize()
			utils.MakeScreenshot(w, h, "", true)
			scheduleScreenshot = false
		}

		profiler.StartGroup("App.mainLoopNormal", profiler.PSwapBuffers)

		win.SwapBuffers()

		profiler.EndGroup()

		profiler.StartGroup("App.mainLoopNormal", profiler.PSleep)
		if !settings.Graphics.VSync {
			fCap := int(settings.Graphics.FPSCap)

			if fCap < 0 {
				fCap = -fCap * monitorHz
			}

			limiter.FPS = fCap
			limiter.Sync()
		}
		profiler.EndGroup()
	})

	settings.CloseWatcher()
}

func pushFrame() {
	profiler.StartGroup("App.pushFrame", profiler.PDraw)
	profiler.ResetStats()

	gl.Enable(gl.SCISSOR_TEST)
	gl.Disable(gl.DITHER)

	blend.Enable()
	blend.SetFunction(blend.One, blend.OneMinusSrcAlpha)

	viewport.Push(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	if screenFBO == nil ||
		lastSamples != int(settings.Graphics.MSAA) ||
		screenFBO.GetWidth() != int(settings.Graphics.GetWidth()) ||
		screenFBO.GetHeight() != int(settings.Graphics.GetHeight()) {
		if screenFBO != nil {
			screenFBO.Dispose()
		}

		screenFBO = buffer.NewFrameMultisampleScreen(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false, int(settings.Graphics.MSAA))

		lastSamples = int(settings.Graphics.MSAA)
	}

	if lastSamples > 0 {
		screenFBO.Bind()
	}

	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	if player != nil {
		player.Draw(0)
	}

	if lastSamples > 0 {
		screenFBO.Unbind()
	}

	blend.ClearStack()
	viewport.Pop()

	profiler.EndGroup()
}

func checkForUpdates() {
	status, url, err := utils.CheckForUpdate()

	switch status {
	case utils.Failed:
		log.Println("Can't get version from GitHub:", err)
	case utils.UpToDate:
		log.Println("You're using the newest version of danser.")
	case utils.Snapshot:
		log.Println("You're using a snapshot version of danser.")
		log.Println("For newer version of snapshots please visit the official danser discord server at:", url)
	case utils.UpdateAvailable:
		log.Println("You're using an older version of danser.")
		log.Println("You can download a newer version here:", url)
		time.Sleep(2 * time.Second)
	}
}

func Run() {
	defer func() {
		var err any
		var stackTrace []string

		if err = recover(); err != nil {
			stackTrace = goroutines.GetStackTrace(4)
		}

		closeHandler(err, stackTrace)
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	goroutines.SetCrashHandler(closeHandler)

	platform.StartLogging("danser")

	platform.DisableQuickEdit()

	goroutines.RunMain(run)
}

func closeHandler(err any, stackTrace []string) {
	settings.CloseWatcher()
	discord.Disconnect()
	platform.EnableQuickEdit()

	if err != nil {
		log.Println("panic:", err)

		for _, s := range stackTrace {
			log.Println(s)
		}

		os.Exit(1)
	}

	log.Println("Exiting normally.")
}
