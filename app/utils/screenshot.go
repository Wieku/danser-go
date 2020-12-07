package utils

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
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

//var cmd *exec.Cmd
//var pipe io.WriteCloser
//
//var FFMpegStart int64
//
//func StartFFmpeg(fps, w, h int) {
//	os.Mkdir("videos", 0655)
//	cmd = exec.Command("ffmpeg",
//		"-y", //(optional) overwrite output file if it exists
//		"-f", "rawvideo",
//		"-vcodec","rawvideo",
//		"-s", fmt.Sprintf("%dx%d", w, h), //size of one frame
//		"-pix_fmt", "rgb24",
//		"-r", strconv.Itoa(fps), //frames per second
//		"-i", "-", //The imput comes from a pipe
//		"-vf", "vflip",
//		"-an", //Tells FFMPEG not to expect any audio
//		"-vcodec", "h264_nvenc",
//		"-rc", "constqp",
//		"-qp", "15",
//		"-threads", "4",
//		"videos/video.mp4",
//	)
//
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//
//	var err error
//
//	pipe, err = cmd.StdinPipe()
//	if err != nil {
//		panic(err)
//	}
//
//	err = cmd.Start()
//	if err != nil {
//		panic(err)
//	}
//}
//
//func StopFFmpeg() {
//	pipe.Close()
//	//cmd.Process.Signal(os.Kill)
//	cmd.Wait()
//}
//
//func Combine() {
//	cmd2 := exec.Command("ffmpeg",
//		"-y",
//		"-i", "videos/video.mp4",
//		"-i", "videos/audio.wav",
//		"-c:v", "copy",
//		"-c:a", "aac",
//		"-ab", "320k",
//		"videos/danser_"+time.Now().Format("2006-01-02_15-04-05")+".mp4",
//	)
//	cmd2.Start()
//	cmd2.Wait()
//	os.Remove("videos/video.mp4")
//	os.Remove("videos/audio.wav")
//}
//
//func MakeFrame(win *glfw.Window, num int) {
//	if num == 0 {
//		FFMpegStart = qpc.GetNanoTime()
//	}
//
//	w, h := win.GetFramebufferSize()
//
//	pixmap := texture.NewPixMapC(w, h, 3)
//
//	gl.PixelStorei(gl.PACK_ALIGNMENT, int32(1))
//	gl.ReadPixels(0, 0, int32(w), int32(h), gl.RGB, gl.UNSIGNED_BYTE, pixmap.RawPointer)
//
//	mainthread.CallNonBlock(func() {
//		_, err := pipe.Write(pixmap.Data)
//		if err != nil {
//			panic(err)
//		}
//		//log.Println("sent frame", num)
//
//		pixmap.Dispose()
//	})
//}
