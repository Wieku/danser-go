package main

import "C"
import (
	"flag"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	camera2 "github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/database"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics/font"
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
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/statistic"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"
)

const (
	base           = "Specify the"
	artistDesc     = base + " artist of a song"
	titleDesc      = base + " title of a song"
	creatorDesc    = base + " creator of a map"
	difficultyDesc = base + " difficulty(version) of a map"
	shorthand      = " (shorthand)"
)

var player *states.Player
var pressed = false
var pressedM = false
var pressedP = false

func run() {
	var win *glfw.Window
	var limiter *frame.Limiter

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

		settingsVersion := flag.String("settings", "", "Specify settings version")
		cursors := flag.Int("cursors", 1, "How many repeated cursors should be visible, recommended 2 for mirror, 8 for mandala")
		tag := flag.Int("tag", 1, "How many cursors should be \"playing\" specific map. 2 means that 1st cursor clicks the 1st object, 2nd clicks 2nd object, 1st clicks 3rd and so on")
		knockout := flag.Bool("knockout", false, "Use knockout feature")
		speed := flag.Float64("speed", 1.0, "Specify music's speed, set to 1.5 to have DoubleTime mod experience")
		pitch := flag.Float64("pitch", 1.0, "Specify music's pitch, set to 1.5 with -speed=1.5 to have Nightcore mod experience")
		debug := flag.Bool("debug", false, "Show info about map and rendering engine, overrides Graphics.ShowFPS setting")

		gldebug := flag.Bool("gldebug", false, "Turns on OpenGL debug logging, may reduce performance heavily")

		play := flag.Bool("play", false, "Practice playing osu!standard maps")
		scrub := flag.Float64("scrub", 0, "Start at the given time in seconds")

		skip := flag.Bool("skip", false, "Skip straight to map's drain time")

		flag.Parse()

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
		settings.SCRUB = *scrub

		newSettings := settings.LoadSettings(*settingsVersion)

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
			} else {
				discord.Connect()
			}
		}

		assets.Init(build.Stream == "Dev")

		glfw.Init()
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.Samples, int(settings.Graphics.MSAA))

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
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		batch := batch2.NewQuadBatch()
		batch.Begin()
		batch.SetColor(1, 1, 1, 1)
		camera := camera2.NewCamera()
		camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
		camera.SetOrigin(vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
		camera.Update()
		batch.SetCamera(camera.GetProjectionView())

		file, _ := assets.Open("assets/fonts/Exo2-Bold.ttf")
		font := font.LoadFont(file)
		file.Close()

		font.Draw(batch, 0, 10, 32, "Loading...")

		batch.End()
		win.SwapBuffers()
		glfw.PollEvents()

		glfw.SwapInterval(0)
		if settings.Graphics.VSync {
			glfw.SwapInterval(1)
		}

		bass.Init()
		audio.LoadSamples()

		beatmap.ParseTimingPointsAndPauses(beatMap)
		beatmap.ParseObjects(beatMap)
		beatMap.LoadCustomSamples()
		player = states.NewPlayer(beatMap)
		limiter = frame.NewLimiter(int(settings.Graphics.FPSCap))
	})

	for !win.ShouldClose() {
		mainthread.Call(func() {
			statistic.Reset()
			glfw.PollEvents()

			if settings.Graphics.MSAA > 0 {
				gl.Enable(gl.MULTISAMPLE)
			}

			gl.Enable(gl.SCISSOR_TEST)
			gl.Disable(gl.DITHER)

			viewport.Push(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT)

			if player != nil {
				player.Draw(0)
			}

			if win.GetKey(glfw.KeyEscape) == glfw.Press {
				win.SetShouldClose(true)
			}

			if win.GetKey(glfw.KeyF2) == glfw.Press {

				if !pressed {
					utils.MakeScreenshot(*win)
				}

				pressed = true
			}

			if win.GetKey(glfw.KeyF2) == glfw.Release {
				pressed = false
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

			blend.ClearStack()
			viewport.ClearStack()
		})
	}
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

func main() {
	file, err := os.Create("danser.log")
	if err != nil {
		panic(err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))

	defer func() {
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

	setWorkingDirectory()
	runtime.GOMAXPROCS(runtime.NumCPU())
	mainthread.CallQueueCap = 100000
	mainthread.Run(run)
}
