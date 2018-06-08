package main

import (
	"github.com/wieku/danser/audio"
	"github.com/wieku/danser/beatmap"
	"flag"
	"log"
	"github.com/faiface/mainthread"
	"github.com/faiface/glhf"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/wieku/danser/states"
	"github.com/go-gl/gl/v3.3-core/gl"
	"image"
	"unsafe"
	"os"
	"image/png"
	"strconv"
	"time"
	"github.com/wieku/danser/settings"
)

var player *states.Player
var pressed = false
func run() {
	var win *glfw.Window

	mainthread.Call(func() {

		audio.Init()
		audio.LoadSamples()
		settings.LoadSettings(0)
		glfw.Init()
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.Samples, 16)

		var err error

		win, err = glfw.CreateWindow(1920, 1080, "osuda", glfw.GetPrimaryMonitor(), nil)
		if err != nil {
			panic(err)
		}

		win.MakeContextCurrent()
		log.Println("GLFW initialized!")
		glhf.Init()
		glhf.Clear(0,0,0,1)
		win.SwapBuffers()
		glfw.PollEvents()

		glfw.SwapInterval(0)
		player = nil

		artist := flag.String("artist", "", "")
		title := flag.String("title", "", "")
		difficulty := flag.String("difficulty", "", "")

		flag.Parse()

		go func() {
			beatmaps := beatmap.LoadBeatmaps()
			for _, b := range beatmaps {
				if (*artist == "" || *artist == b.Artist) && (*title == "" || *title == b.Name) && (*difficulty == "" || *difficulty == b.Difficulty) {//if b.Difficulty == "200BPM t+pazolite_cheatreal GO TO HELL  AR10" {

					mainthread.Call(func(){
						win.SetTitle("osudancer - " + b.Artist + " - " + b.Name + " [" + b.Difficulty + "]")
						beatmap.ParseObjects(b)
						player = states.NewPlayer(b)
					})

					break
				}
			}
		}()

	})

	for !win.ShouldClose() {
		mainthread.Call(func() {
			gl.Enable(gl.MULTISAMPLE)
			//gl.Enable(gl.POLYGON_SMOOTH)
			gl.Disable(gl.DITHER)
			gl.Viewport(0, 0, 1920, 1080)
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
					w, h := win.GetFramebufferSize()
					img := image.NewNRGBA(image.Rectangle{Min: image.Point{0, 0} , Max: image.Point{w, h} })
					buff := make([]uint8, w*h*4)
					buff1 := make([]uint8, w*h*4)
					gl.PixelStorei(gl.PACK_ALIGNMENT, int32(1))
					gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&buff[0]))

					for i := h-1; i >=0; i-- {
						for j:=0; j < w*4; j++ {
							if (j+1)%4 == 0 {
								buff1[(h-1-i)*w*4+j] = 0xFF
							} else {
								buff1[(h-1-i)*w*4+j] = buff[i*w*4+j]
							}
						}
					}

					img.Pix = buff1
					os.Mkdir("screenshots", 0644)
					f, _ := os.OpenFile("screenshots/"+strconv.FormatInt(time.Now().UnixNano(), 10)+".png", os.O_WRONLY | os.O_CREATE, 0644)
					defer f.Close()
					png.Encode(f, img)
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
