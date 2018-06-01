package main

import (
	"danser/audio"
	"sync"
	"danser/beatmap"
	"time"
	"flag"
	"log"
)

func main() {

	audio.Init()
	audio.LoadSamples()

	wg := sync.WaitGroup{}

	beatmaps := beatmap.LoadBeatmaps()

	artist := flag.String("artist", "", "")
	title := flag.String("title", "", "")
	difficulty := flag.String("difficulty", "", "")

	flag.Parse()

	for _, bMap := range beatmaps {
		if (*artist == "" || *artist == bMap.Artist) && (*title == "" || *title == bMap.Name) && (*difficulty == "" || *difficulty == bMap.Difficulty) {
			wg.Add(1)
			beatmap.ParseObjects(bMap)
			bMap.Reset()

			log.Println(bMap.Audio)
			player := audio.NewMusic(bMap.Audio)
			player.RegisterCallback(func() {
				wg.Done()
			})
			player.Play()

			go func() {
				for {
					timMs := player.GetPosition()*1000
					bMap.Update(int64(timMs))
					time.Sleep(time.Millisecond)
				}
			}()

			break
		}
	}

	wg.Wait()
}
