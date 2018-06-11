package beatmap

import (
	"sync"
	"os"
	"log"
	"path/filepath"
	"strings"
	"github.com/wieku/danser/settings"
)

func LoadBeatmaps() []*BeatMap {
	searchDir := settings.General.OsuSongsDir

	log.Println("Loading beatmaps...")

	var candidates []string

	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(f.Name(), ".osu") {
			candidates = append(candidates, path)
		}
		return nil
	})

	channel := make(chan string, 20)
	channelB := make(chan *BeatMap, len(candidates))
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				path, ok := <- channel
				if !ok {
					break
				}
				if f, err := os.Open(path); err == nil {
					if bMap := ParseBeatMap(f); bMap != nil {
						channelB <- bMap
					}
					f.Close()
				}
			}
		}()
	}

	for _, path := range candidates {
		channel <- path
	}

	close(channel)
	wg.Wait()

	var beatmaps []*BeatMap

	for len(channelB) > 0 {
		beatmap := <-channelB
		beatmaps = append(beatmaps, beatmap)
	}
	return beatmaps
}
