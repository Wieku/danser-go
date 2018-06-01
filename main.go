package main



import (
	/*"log"
	"os"
	"unsafe"
	"sync"*/
	"danser/audio"
	"time"
	"log"
	"os"
	"sync"
)

func main() {

	audio.Init()

	file := os.Getenv("localappdata") + "\\osu!\\Songs\\342773 TheFatRat - Windfall\\TheFatRat - Windfall.mp3"
	log.Println(file)
	player := audio.NewMusic(file)

	wg := sync.WaitGroup{}
	wg.Add(1)

	player.RegisterCallback(func() {
		log.Println("end!")
		wg.Done()
	})

	player.Play()

	log.Println(player.GetLength())

	wg.Wait()
}
