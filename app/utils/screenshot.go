package utils

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"image"
	"image/png"
	"os"
	"runtime"
	"time"
	"unsafe"
)

func MakeScreenshot(win glfw.Window) {
	w, h := win.GetFramebufferSize()
	buff := make([]uint8, w*h*4)
	gl.PixelStorei(gl.PACK_ALIGNMENT, int32(1))
	gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&buff[0]))

	go func() {
		img := image.NewNRGBA(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{w, h}})
		buff1 := make([]uint8, w*h*4)
		for i := h - 1; i >= 0; i-- {
			for j := 0; j < w*4; j++ {
				if (j+1)%4 == 0 {
					buff1[(h-1-i)*w*4+j] = 0xFF
				} else {
					buff1[(h-1-i)*w*4+j] = buff[i*w*4+j]
				}
			}
		}
		runtime.KeepAlive(buff)
		img.Pix = buff1
		os.Mkdir("screenshots", 0644)

		dateString := time.Now().Format("2006-01-02_15-04-05")
		f, _ := os.OpenFile("screenshots/danser_"+dateString+".png", os.O_WRONLY|os.O_CREATE, 0644)

		defer f.Close()

		png.Encode(f, img)
	}()
}
