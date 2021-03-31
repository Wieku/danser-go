package utils

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"log"
	"os"
	"time"
)

func MakeScreenshot(win glfw.Window) {
	w, h := win.GetFramebufferSize()

	pixmap := texture.NewPixMapC(w, h, 3)

	gl.PixelStorei(gl.PACK_ALIGNMENT, int32(1))
	gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGB, gl.UNSIGNED_BYTE, pixmap.RawPointer)

	go func() {
		defer pixmap.Dispose()

		err := os.Mkdir("screenshots", 0644)
		if !os.IsExist(err) {
			log.Println("Failed to save the screenshot! Error:", err)
			return
		}

		fileName := "danser_" + time.Now().Format("2006-01-02_15-04-05") + ".png"

		err = pixmap.WritePng("screenshots/"+fileName, true)
		if err != nil {
			log.Println("Failed to save the screenshot! Error:", err)
			return
		}

		log.Println("Screenshot", fileName, "saved!")
	}()
}
