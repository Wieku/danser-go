package main

import (
	"github.com/wieku/danser/audio"
	"github.com/wieku/danser/beatmap"
	"flag"
	"log"
	"github.com/faiface/mainthread"
	"github.com/wieku/glhf"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/wieku/danser/states"
	"github.com/go-gl/gl/v3.3-core/gl"
	"os"
	"github.com/wieku/danser/settings"
	"github.com/wieku/danser/utils"
)

var player *states.Player
var pressed = false
func run() {
	var win *glfw.Window

	mainthread.Call(func() {

		artist := flag.String("artist", "", "")
		title := flag.String("title", "", "")
		difficulty := flag.String("difficulty", "", "")
		creator := flag.String("creator", "", "")
		settingsVersion := flag.Int("settings", 0, "")
		cursors := flag.Int("cursors", 2, "")
		tag := flag.Int("tag", 1, "")
		speed := flag.Float64("speed", 1.0, "")
		mover := flag.String("mover", "flower", "")

		flag.Parse()

		if (*artist+*title+*difficulty+*creator) == "" {
			log.Println("Any beatmap specified, closing...")
			os.Exit(0)
		}

		settings.DIVIDES = *cursors
		settings.TAG = *tag
		settings.SPEED = *speed
		beatmap.SetMover(*mover)

		newSettings := settings.LoadSettings(*settingsVersion)

		player = nil
		var beatMap *beatmap.BeatMap = nil

		beatmaps := beatmap.LoadBeatmaps()
		for _, b := range beatmaps {
			if (*artist == "" || *artist == b.Artist) && (*title == "" || *title == b.Name) && (*difficulty == "" || *difficulty == b.Difficulty) && (*creator == "" || *creator == b.Creator) {//if b.Difficulty == "200BPM t+pazolite_cheatreal GO TO HELL  AR10" {
				beatMap = b
				break
			}
		}

		if beatMap == nil {
			log.Println("No beatmaps found, closing...")
			os.Exit(0)
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
			log.Println(mWidth, mHeight)
			settings.Graphics.Width, settings.Graphics.Height = int64(mWidth), int64(mHeight)
			settings.Save()
			win, err = glfw.CreateWindow(mWidth, mHeight, "danser", monitor, nil)
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

		win.SetTitle("danser - " + beatMap.Artist + " - " + beatMap.Name + " [" + beatMap.Difficulty + "]")

		win.MakeContextCurrent()
		log.Println("GLFW initialized!")
		glhf.Init()
		glhf.Clear(0,0,0,1)
		win.SwapBuffers()
		glfw.PollEvents()

		glfw.SwapInterval(0)
		if settings.Graphics.VSync {
			glfw.SwapInterval(1)
		}

		audio.Init()
		audio.LoadSamples()

		beatmap.ParseObjects(beatMap)
		player = states.NewPlayer(beatMap)

	})

	for !win.ShouldClose() {
		mainthread.Call(func() {
			gl.Enable(gl.MULTISAMPLE)
			gl.Disable(gl.DITHER)
			gl.Viewport(0, 0, int32(settings.Graphics.GetWidth()), int32(settings.Graphics.GetHeight()))
			gl.ClearColor(0,0,0,1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			if player != nil {
				player.Update()
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

			win.SwapBuffers()
			glfw.PollEvents()
		})
	}
}

func main() {
	mainthread.Run(run)
}
