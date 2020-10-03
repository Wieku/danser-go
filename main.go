package main

import (
	"flag"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/database"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/statistic"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

		settingsVersion := flag.Int("settings", 0, "Specify settings version")
		cursors := flag.Int("cursors", 1, "How many repeated cursors should be visible, recommended 2 for mirror, 8 for mandala")
		tag := flag.Int("tag", 1, "How many cursors should be \"playing\" specific map. 2 means that 1st cursor clicks the 1st object, 2nd clicks 2nd object, 1st clicks 3rd and so on")
		knockout := flag.Bool("knockout", false, "Use knockout feature")
		speed := flag.Float64("speed", 1.0, "Specify music's speed, set to 1.5 to have DoubleTime mod experience")
		pitch := flag.Float64("pitch", 1.0, "Specify music's pitch, set to 1.5 with -speed=1.5 to have Nightcore mod experience")
		mover := flag.String("mover", "flower", "It will be moved to settings")
		debug := flag.Bool("debug", false, "Show info about map and rendering engine, overrides Graphics.ShowFPS setting")

		play := flag.Bool("play", false, "Practice playing osu!standard maps")

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
		_ = mover
		dance.SetMover(*mover)

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
		icon, _ := utils.LoadImageN("assets/textures/dansercoin.png")
		icon2, _ := utils.LoadImageN("assets/textures/dansercoin48.png")
		icon3, _ := utils.LoadImageN("assets/textures/dansercoin24.png")
		icon4, _ := utils.LoadImageN("assets/textures/dansercoin16.png")
		win.SetIcon([]image.Image{icon, icon2, icon3, icon4})

		win.MakeContextCurrent()
		log.Println("GLFW initialized!")
		gl.Init()
		gl.Enable(gl.BLEND)
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		batch := sprite.NewSpriteBatch()
		batch.Begin()
		batch.SetColor(1, 1, 1, 1)
		camera := bmath.NewCamera()
		camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
		camera.SetOrigin(vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
		camera.Update()
		batch.SetCamera(camera.GetProjectionView())

		file, _ := os.Open("assets/fonts/Exo2-Bold.ttf")
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
	defer discord.Disconnect()
	setWorkingDirectory()
	runtime.GOMAXPROCS(runtime.NumCPU())
	mainthread.CallQueueCap = 100000
	mainthread.Run(run)
}
