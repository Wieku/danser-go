package main

import (
	"danser/audio"
	"danser/beatmap"
	"danser/bmath"
	"danser/build"
	"danser/dance"
	"danser/database"
	"danser/render"
	"danser/render/font"
	"danser/settings"
	"danser/states"
	"danser/utils"
	"flag"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/wieku/glhf"
	"image"
	"log"
	"os"
)

var player *states.Player
var pressed = false
var pressedM = false
var pressedP = false

func run() {
	var win *glfw.Window

	mainthread.Call(func() {

		artist := flag.String("artist", "", "")
		//title := flag.String("title", "Snow Drive(01.23)", "")
		//difficulty := flag.String("difficulty", "Arigatou", "")
		//title := flag.String("title", "Road of Resistance", "")
		//difficulty := flag.String("difficulty", "Crimson Rebellion", "")
		creator := flag.String("creator", "", "")
		settingsVersion := flag.Int("settings", 0, "")
		cursors := flag.Int("cursors", 1, "")
		tag := flag.Int("tag", 1, "")
		speed := flag.Float64("speed", 1.0, "")
		pitch := flag.Float64("pitch", 1.0, "")
		mover := flag.String("mover", "linear", "")
		debug := flag.Bool("debug", false, "")
		fps := flag.Bool("fps", false, "")

		flag.Parse()


		settings.DEBUG = *debug
		settings.FPS = *fps
		settings.DIVIDES = *cursors
		settings.TAG = *tag
		settings.SPEED = *speed
		settings.PITCH = *pitch
		_ = mover
		dance.SetMover(*mover)

		newSettings := settings.LoadSettings(*settingsVersion)

		player = nil
		var beatMap *beatmap.BeatMap = nil

		// 从设置重新载入map
		title := flag.String("title", settings.General.Title, "")
		difficulty := flag.String("difficulty", settings.General.Difficulty, "")

		if (*artist + *title + *difficulty + *creator) == "" {
			log.Println("No beatmap specified, closing...")
			os.Exit(0)
		}

		database.Init()
		beatmaps := database.LoadBeatmaps()

		for _, b := range beatmaps {
			if (*artist == "" || *artist == b.Artist) && (*title == "" || *title == b.Name) && (*difficulty == "" || *difficulty == b.Difficulty) && (*creator == "" || *creator == b.Creator) {
				beatMap = b
				beatMap.UpdatePlayStats()
				database.UpdatePlayStats(beatMap)
				break
			}
		}

		if beatMap == nil {
			log.Println("Beatmap not found, closing...")
			os.Exit(0)
		}

		//r := replay.ExtractReplay("replays/22-OskaRRRitoS.osr")
		//for k := 0; k < 40; k++ {
		//	log.Println(*r.ReplayData[k])
		//}
		//hitjudge.ParseHits("Song/336414 Wagakki Band - Tengaku/Wagakki Band - Tengaku (Shiro) [Uncompressed Fury of a Raging Japanese God].osu",
		//	"replays/22-OskaRRRitoS.osr")
		//os.Exit(1)

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
			log.Println(mWidth, mHeight)
			settings.Graphics.Width, settings.Graphics.Height = int64(mWidth), int64(mHeight)
			settings.Save()
			win, err = glfw.CreateWindow(mWidth, mHeight, "osu vs player", monitor, nil)
		} else {
			if settings.Graphics.Fullscreen {
				win, err = glfw.CreateWindow(int(settings.Graphics.Width), int(settings.Graphics.Height), "danser", monitor, nil)
			} else {
				win, err = glfw.CreateWindow(int(settings.Graphics.WindowWidth), int(settings.Graphics.WindowHeight), "danser", nil, nil)
			}
		}

		if err != nil {
			panic(err)
		}

		win.SetTitle("osu vs player " + build.VERSION + " by " + build.OWNER + " on " + beatMap.Artist + " - " + beatMap.Name + " [" + beatMap.Difficulty + "]")
		icon, _ := utils.LoadImage("assets/textures/dansercoin.png")
		icon2, _ := utils.LoadImage("assets/textures/dansercoin48.png")
		icon3, _ := utils.LoadImage("assets/textures/dansercoin24.png")
		icon4, _ := utils.LoadImage("assets/textures/dansercoin16.png")
		win.SetIcon([]image.Image{icon, icon2, icon3, icon4})

		win.MakeContextCurrent()
		log.Println("GLFW initialized!")
		glhf.Init()
		glhf.Clear(0, 0, 0, 1)

		batch := render.NewSpriteBatch()
		batch.Begin()
		batch.SetColor(1, 1, 1, 1)
		camera := bmath.NewCamera()
		camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
		camera.SetOrigin(bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
		camera.Update()
		batch.SetCamera(camera.GetProjectionView())

		file, _ := os.Open("assets/fonts/Roboto-Bold.ttf")
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

		audio.Init()
		audio.LoadSamples()

		beatmap.ParseObjects(beatMap)
		beatMap.LoadCustomSamples()
		player = states.NewPlayer(beatMap)

	})

	for !win.ShouldClose() {
		mainthread.Call(func() {
			gl.Enable(gl.MULTISAMPLE)
			gl.Disable(gl.DITHER)
			gl.Disable(gl.SCISSOR_TEST)
			gl.Viewport(0, 0, int32(settings.Graphics.GetWidth()), int32(settings.Graphics.GetHeight()))
			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

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
			glfw.PollEvents()
		})
	}
}

func main() {
	mainthread.CallQueueCap = 100000
	mainthread.Run(run)

	//r := replay.ExtractReplay("replay-osu_807074_2432526116.osr")
	//for k := 0; k < 40; k++ {
	//	log.Println(*r.ReplayData[k].KeyPressed)
	//}

	//files, _ := replay.GetOsrFiles()
	//log.Println(files)

	//log.Println(bmath.Vector2d{291.6093043265869, 320.52429609268916}.Dst(bmath.Vector2d{233.3333,278.2222}))
}
