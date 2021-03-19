package main

import "C"
import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/faiface/mainthread"
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
	"github.com/wieku/danser-go/framework/frame"
	batch2 "github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/statistic"
	"github.com/wieku/rplpa"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"
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
var pressed = false
var pressedM = false
var pressedP = false

var batch *batch2.QuadBatch

var win *glfw.Window
var limiter *frame.Limiter
var screenFBO *buffer.Framebuffer
var lastSamples int
var lastVSync bool

var output string

func run() {

	mainthread.Call(func() {

		md5 := flag.String("md5", "", "Specify the beatmap md5 hash. Overrides other beatmap search flags")

		artist := flag.String("artist", "", artistDesc)
		flag.StringVar(artist, "a", "", artistDesc+shorthand)

		title := flag.String("title", "", titleDesc)
		flag.StringVar(title, "t", "", titleDesc+shorthand)

		difficulty := flag.String("difficulty", "", difficultyDesc)
		flag.StringVar(difficulty, "d", "", difficultyDesc+shorthand)

		creator := flag.String("creator", "", creatorDesc)
		flag.StringVar(creator, "c", "", creatorDesc+shorthand)

		settingsVersion := flag.String("settings", "", "Specify settings version, -settings=a means that settings-a.json will be loaded")
		cursors := flag.Int("cursors", 1, "How many repeated cursors should be visible, recommended 2 for mirror, 8 for mandala")
		tag := flag.Int("tag", 1, "How many cursors should be \"playing\" specific map. 2 means that 1st cursor clicks the 1st object, 2nd clicks 2nd object, 1st clicks 3rd and so on")
		knockout := flag.Bool("knockout", false, "Use knockout feature")
		speed := flag.Float64("speed", 1.0, "Specify music's speed, set to 1.5 to have DoubleTime mod experience")
		pitch := flag.Float64("pitch", 1.0, "Specify music's pitch, set to 1.5 with -speed=1.5 to have Nightcore mod experience")
		debug := flag.Bool("debug", false, "Show info about map and rendering engine, overrides Graphics.ShowFPS setting")

		gldebug := flag.Bool("gldebug", false, "Turns on OpenGL debug logging, may reduce performance heavily")

		play := flag.Bool("play", false, "Practice playing osu!standard maps")
		start := flag.Float64("start", 0, "Start at the given time in seconds")
		end := flag.Float64("end", math.Inf(1), "End at the given time in seconds")

		skip := flag.Bool("skip", false, "Skip straight to map's drain time")
		record := flag.Bool("record", false, "Records a video")

		mods := flag.String("mods", "", "Specify beatmap/play mods. If NC/DT/HT is selected, overrides -speed and -pitch flags")

		replay := flag.String("replay", "", replayDesc)
		flag.StringVar(replay, "r", "", replayDesc+shorthand)

		skin := flag.String("skin", "", "Replace Skin.CurrentSkin setting temporarily")

		out := flag.String("out", "", "Overrides -record flag. Specify the name of recorded video file, extension is managed by settings")

		flag.Parse()

		if *out != "" {
			*record = true
			output = *out
		}

		modsParsed := difficulty2.ParseMods(*mods)

		if *replay != "" {
			bytes, err := ioutil.ReadFile(*replay)
			if err != nil {
				panic(err)
			}

			rp, err := rplpa.ParseReplay(bytes)
			if err != nil {
				panic(err)
			}

			*md5 = rp.BeatmapMD5
			modsParsed = difficulty2.Modifier(rp.Mods)
			*knockout = true
			settings.REPLAY = *replay
		}

		if modsParsed.Active(difficulty2.Target) {
			panic("Target practice mod is not supported!")
		}

		if !modsParsed.Compatible() {
			panic("Incompatible mods selected!")
		}

		closeAfterSettingsLoad := false

		if (*md5 + *artist + *title + *difficulty + *creator) == "" {
			log.Println("No beatmap specified, closing...")
			closeAfterSettingsLoad = true
		}

		settings.DEBUG = *debug
		settings.KNOCKOUT = *knockout
		settings.PLAY = *play
		settings.DIVIDES = *cursors
		settings.TAG = *tag
		settings.SPEED = *speed
		settings.PITCH = *pitch
		settings.SKIP = *skip
		settings.START = *start
		settings.END = *end
		settings.RECORD = *record

		if settings.RECORD {
			bass.Offscreen = true
		}

		newSettings := settings.LoadSettings(*settingsVersion)

		if !newSettings && len(os.Args) == 1 {
			utils.OpenURL("https://youtu.be/dQw4w9WgXcQ")
			closeAfterSettingsLoad = true
		}

		player = nil
		var beatMap *beatmap.BeatMap = nil

		if !closeAfterSettingsLoad {
			database.Init()
			beatmaps := database.LoadBeatmaps()

			if *md5 != "" {
				for _, b := range beatmaps {
					if strings.EqualFold(b.MD5, *md5) {
						beatMap = b
						beatMap.UpdatePlayStats()
						database.UpdatePlayStats(beatMap)
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
						beatMap.UpdatePlayStats()
						database.UpdatePlayStats(beatMap)
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
							beatMap.UpdatePlayStats()
							database.UpdatePlayStats(beatMap)
							break
						}
					}
				}
			}

			if beatMap == nil {
				log.Println("Beatmap not found, closing...")
				closeAfterSettingsLoad = true
			}
		}

		assets.Init(build.Stream == "Dev")

		glfw.Init()
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.Samples, 0)

		var err error

		monitor := glfw.GetPrimaryMonitor()
		mWidth, mHeight := monitor.GetVideoMode().Width, monitor.GetVideoMode().Height

		if newSettings {
			settings.Graphics.SetDefaults(int64(mWidth), int64(mHeight))
			settings.Save()
		}

		if closeAfterSettingsLoad {
			os.Exit(0)
		}

		lastSamples = int(settings.Graphics.MSAA)

		if strings.TrimSpace(*skin) != "" {
			settings.Skin.CurrentSkin = *skin
		}

		if settings.RECORD {
			glfw.WindowHint(glfw.Visible, glfw.False)

			//HACK: some in-app variables depend on these settings so we force them here
			settings.Graphics.VSync = false
			settings.Graphics.ShowFPS = false
			settings.DEBUG = false
			settings.Graphics.Fullscreen = false
			settings.Graphics.WindowWidth = int64(settings.Recording.FrameWidth)
			settings.Graphics.WindowHeight = int64(settings.Recording.FrameHeight)
			settings.Playfield.LeadInTime = 0
		} else {
			discord.Connect()
		}

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

		win.SetTitle("danser " + build.VERSION + " - " + beatMap.Artist + " - " + beatMap.Name + " [" + beatMap.Difficulty + "]")
		input.Win = win

		icon, eee := assets.GetPixmap("assets/textures/dansercoin.png")
		if eee != nil {
			log.Println(eee)
		}
		icon2, _ := assets.GetPixmap("assets/textures/dansercoin48.png")
		icon3, _ := assets.GetPixmap("assets/textures/dansercoin24.png")
		icon4, _ := assets.GetPixmap("assets/textures/dansercoin16.png")

		win.SetIcon([]image.Image{icon.NRGBA(), icon2.NRGBA(), icon3.NRGBA(), icon4.NRGBA()})

		icon.Dispose()
		icon2.Dispose()
		icon3.Dispose()
		icon4.Dispose()

		win.MakeContextCurrent()

		log.Println("GLFW initialized!")

		gl.Init()

		extensionCheck()

		C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.RENDERER))))

		glVendor := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.VENDOR))))
		glRenderer := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.RENDERER))))
		glVersion := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.VERSION))))
		glslVersion := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.SHADING_LANGUAGE_VERSION))))

		var extensions string

		var numExtensions int32
		gl.GetIntegerv(gl.NUM_EXTENSIONS, &numExtensions)

		for i := int32(0); i < numExtensions; i++ {
			extensions += C.GoString((*C.char)(unsafe.Pointer(gl.GetStringi(gl.EXTENSIONS, uint32(i)))))
			extensions += " "
		}

		log.Println("GL Vendor:    ", glVendor)
		log.Println("GL Renderer:  ", glRenderer)
		log.Println("GL Version:   ", glVersion)
		log.Println("GLSL Version: ", glslVersion)
		log.Println("GL Extensions:", extensions)
		log.Println("OpenGL initialized!")

		if *gldebug {
			gl.Enable(gl.DEBUG_OUTPUT)
			gl.DebugMessageCallback(func(
				source uint32,
				gltype uint32,
				id uint32,
				severity uint32,
				length int32,
				message string,
				userParam unsafe.Pointer) {
				log.Println("GL:", message)
			}, gl.Ptr(nil))

			gl.DebugMessageControl(gl.DONT_CARE, gl.DONT_CARE, gl.DONT_CARE, 0, nil, true)
		}

		gl.Enable(gl.BLEND)
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		file, _ := assets.Open("assets/fonts/Exo2-Bold.ttf")
		font.LoadFont(file)
		file.Close()

		batch = batch2.NewQuadBatch()
		batch.Begin()
		batch.SetColor(1, 1, 1, 1)
		camera := camera2.NewCamera()
		camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
		camera.SetOrigin(vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
		camera.Update()
		batch.SetCamera(camera.GetProjectionView())

		font.GetFont("Exo 2 Bold").Draw(batch, 0, 10, 32, "Loading...")

		batch.End()
		win.SwapBuffers()

		glfw.SwapInterval(1)
		lastVSync = true

		bass.Init()
		audio.LoadSamples()

		if modsParsed.Active(difficulty2.Nightcore) {
			settings.SPEED *= 1.5
			settings.PITCH *= 1.5
		} else if modsParsed.Active(difficulty2.DoubleTime) {
			settings.SPEED *= 1.5
		} else if modsParsed.Active(difficulty2.Daycore) {
			settings.PITCH *= 0.75
			settings.SPEED *= 0.75
		} else if modsParsed.Active(difficulty2.HalfTime) {
			settings.SPEED *= 0.75
		}

		beatMap.Diff.SetMods(modsParsed)
		beatmap.ParseTimingPointsAndPauses(beatMap)
		beatmap.ParseObjects(beatMap)
		beatMap.LoadCustomSamples()
		player = states.NewPlayer(beatMap)

		limiter = frame.NewLimiter(int(settings.Graphics.FPSCap))
	})

	count := 0

	if settings.RECORD {
		fps := float64(settings.Recording.FPS)

		if settings.Recording.MotionBlur.Enabled {
			fps *= float64(settings.Recording.MotionBlur.OversampleMultiplier)
		}

		w, h := int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight())

		var fbo *buffer.Framebuffer

		mainthread.Call(func() {
			fbo = buffer.NewFrameMultisampleScreen(w, h, false, 0)
		})

		ffmpeg.StartFFmpeg(int(fps), w, h)

		updateFPS := math.Max(fps, 1000)
		updateDelta := 1000 / updateFPS
		fpsDelta := 1000 / fps

		deltaSumF := fpsDelta

		p, _ := player.(*states.Player)

		//maxFrames := int(p.RunningTime / settings.SPEED / 1000 * fps)

		var lastProgress, progress int

		for !p.Update(updateDelta) {
			deltaSumF += updateDelta
			if deltaSumF >= fpsDelta {
				mainthread.Call(func() {
					fbo.Bind()

					ffmpeg.PreFrame()

					viewport.Push(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
					pushFrame()
					viewport.Pop()

					ffmpeg.MakeFrame()

					fbo.Unbind()

					count++

					progress = int(math.Round(p.GetTimeOffset() / p.RunningTime /*float64(count) / float64(maxFrames)*/ * 100))

					if progress%5 == 0 && lastProgress != progress {
						fmt.Println()
						log.Println(fmt.Sprintf("Progress: %d%%", progress))
						lastProgress = progress
					}
				})

				mainthread.Call(func() {
					ffmpeg.CheckData()
				})

				deltaSumF -= fpsDelta
			}
		}

		mainthread.Call(func() {
			ffmpeg.StopFFmpeg()
		})

		bass.SaveToFile(filepath.Join(settings.Recording.OutputDir, ffmpeg.GetFileName()+".wav"))

		ffmpeg.Combine(output)
	} else {
		for !win.ShouldClose() {
			mainthread.Call(func() {
				if lastVSync != settings.Graphics.VSync {
					if settings.Graphics.VSync {
						glfw.SwapInterval(1)
					} else {
						glfw.SwapInterval(0)
					}

					lastVSync = settings.Graphics.VSync
				}

				pushFrame()

				if win.GetKey(glfw.KeyF2) == glfw.Press {

					if !pressed {
						utils.MakeScreenshot(*win)
					}

					pressed = true
				}

				if win.GetKey(glfw.KeyF2) == glfw.Release {
					pressed = false
				}

				if win.GetKey(glfw.KeyEscape) == glfw.Press {
					win.SetShouldClose(true)
				}

				if win.GetKey(glfw.KeyMinus) == glfw.Press {

					if !pressedM {
						if settings.DIVIDES > 1 {
							settings.DIVIDES -= 1
						}
					}

					pressedM = true
				}

				if win.GetKey(glfw.KeyMinus) == glfw.Release {
					pressedM = false
				}

				if win.GetKey(glfw.KeyEqual) == glfw.Press {

					if !pressedP {
						settings.DIVIDES += 1
					}

					pressedP = true
				}

				if win.GetKey(glfw.KeyEqual) == glfw.Release {
					pressedP = false
				}

				win.SwapBuffers()

				if !settings.Graphics.VSync {
					limiter.Sync()
				}

			})
		}
	}
}

func extensionCheck() {
	extensions := []string{
		"GL_ARB_clear_texture",
		"GL_ARB_direct_state_access",
		"GL_ARB_texture_storage",
		"GL_ARB_vertex_attrib_binding",
	}

	if settings.RECORD {
		extensions = append(extensions, "GL_ARB_buffer_storage")
	}

	var notSupported []string

	for _, ext := range extensions {
		if !glfw.ExtensionSupported(ext) {
			notSupported = append(notSupported, ext)
		}
	}

	if len(notSupported) > 0 {
		panic(fmt.Sprintf("Your GPU does not support one or more required OpenGL extensions: %s. Please update your graphics drivers or upgrade your GPU", notSupported))
	}

	_ = extensions
}

func pushFrame() {
	statistic.Reset()
	glfw.PollEvents()

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
}

func setWorkingDirectory() {
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if exec, err = filepath.EvalSymlinks(exec); err != nil {
		panic(err)
	}

	if err = os.Chdir(filepath.Dir(exec)); err != nil {
		panic(err)
	}
}

func checkForUpdates() {
	if build.Stream != "Release" || strings.Contains(build.VERSION, "dev") { //false positive, those are changed during compile
		return
	}

	log.Println("Checking GitHub for a new version of danser...")

	request, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/Wieku/danser-go/releases/latest", nil)
	if err != nil  {
		log.Println("Can't create request")
		return
	}

	client := new(http.Client)
	response, err := client.Do(request)

	if err != nil || response.StatusCode != 200 {
		log.Println("Can't get release info from GitHub")
		return
	}

	var data struct {
		URL string `json:"html_url"`
		Tag string `json:"tag_name"`
	}

	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		log.Println("Failed to decode the response from GitHub")
	}

	githubVersion, _ := strconv.Atoi(strings.ReplaceAll(strings.TrimSuffix(data.Tag, "b"), ".", "")+"9999")

	currentSplit := strings.Split(build.VERSION, "-")

	currentSub := "9999"
	if len(currentSplit) > 1 {
		currentSub = fmt.Sprintf("%04s", strings.TrimPrefix(currentSplit[1], "snapshot"))
	}

	exeVersion, _ := strconv.Atoi(strings.ReplaceAll(currentSplit[0], ".", "")+currentSub)

	if exeVersion >= githubVersion {
		log.Println("You're using the newest version of danser.")

		if strings.Contains(build.VERSION, "snapshot") {
			log.Println("For newer version of snapshots please visit an official danser discord server at: https://discord.gg/UTPvbe8")
		}
	} else {
		log.Println("You're using an older version of danser.")
		log.Println("You can download a newer version here:", data.URL)
		time.Sleep(2*time.Second)
	}
}

func main() {
	file, err := os.Create("danser.log")
	if err != nil {
		panic(err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))

	defer func() {
		settings.CloseWatcher()
		discord.Disconnect()

		if err := recover(); err != nil {
			log.Println("panic:", err)

			for _, s := range utils.GetPanicStackTrace() {
				log.Println(s)
			}

			os.Exit(1)
		}
	}()

	log.Println("Ran using:", os.Args)
	log.Println("Starting danser version", build.VERSION)

	setWorkingDirectory()

	checkForUpdates()

	runtime.GOMAXPROCS(runtime.NumCPU())
	mainthread.CallQueueCap = 100000
	mainthread.Run(run)
}
