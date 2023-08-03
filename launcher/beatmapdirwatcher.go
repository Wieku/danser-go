package launcher

import (
	"github.com/fsnotify/fsnotify"
	"github.com/wieku/danser-go/framework/goroutines"
	"log"
	"os"
	"path/filepath"
)

var watcher *fsnotify.Watcher

func setupWatcher(file string, callback func(event fsnotify.Event)) {
	var err error

	abs, _ := filepath.Abs(file)

	_, err1 := os.Lstat(abs)
	if err1 != nil { // Beatmap dir not found, abort
		return
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	goroutines.Run(func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Println("DirWatcher: New Event:", event.String())

				if callback != nil {
					callback(event)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("DirWatcher: Error:", err)
			}
		}
	})

	err = watcher.Add(abs)
	if err != nil {
		log.Fatal(err)
	}
}

func closeWatcher() {
	if watcher != nil {
		err := watcher.Close()
		if err != nil {
			log.Println(err)
		}

		watcher = nil
	}
}
