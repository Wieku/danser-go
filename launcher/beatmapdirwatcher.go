package launcher

import (
	"github.com/fsnotify/fsnotify"
	"github.com/wieku/danser-go/framework/goroutines"
	"log"
	"path/filepath"
)

var watcher *fsnotify.Watcher

func setupWatcher(file string, callback func(event fsnotify.Event)) {
	var err error

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

				if callback != nil {
					callback(event)
				}
				//log.Println("SettingsManager: Detected", file, "modification, reloading...", event.String())
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("error:", err)
			}
		}
	})

	abs, _ := filepath.Abs(file)

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
