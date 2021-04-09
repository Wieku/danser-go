package settings

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"time"
)

var fileStorage *fileformat
var fileName string
var watcher *fsnotify.Watcher

func initStorage() {
	fileStorage = &fileformat{
		General:   General,
		Graphics:  Graphics,
		Audio:     Audio,
		Input:     Input,
		Gameplay:  Gameplay,
		Skin:      Skin,
		Cursor:    Cursor,
		Objects:   Objects,
		Playfield: Playfield,
		Dance:     Dance,
		Knockout:  Knockout,
		Recording: Recording,
	}
}

func LoadSettings(version string) bool {
	initStorage()
	fileName = "settings"

	if version != "" {
		fileName += "-" + version
	}
	fileName += ".json"

	file, err := os.Open(fileName)
	defer file.Close()
	if os.IsNotExist(err) {
		saveSettings(fileName, fileStorage)
		return true
	} else if err != nil {
		panic(err)
	} else {
		load(file, fileStorage)
		saveSettings(fileName, fileStorage) //this is done to save additions from the current format
	}

	if !RECORD {
		setupWatcher(fileName)
	}

	return false
}

func setupWatcher(file string) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)

					time.Sleep(time.Millisecond * 200)

					file, _ := os.Open(fileName)

					load(file, fileStorage)

					file.Close()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	abs, _ := filepath.Abs(file)

	err = watcher.Add(abs)
	if err != nil {
		log.Fatal(err)
	}
}

func CloseWatcher() {
	if watcher != nil {
		err := watcher.Close()
		if err != nil {
			log.Println(err)
		}

		watcher = nil
	}
}

func load(file *os.File, target interface{}) {
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		panic(fmt.Sprintf("Failed to parse %s! Please re-check the file for mistakes. Error: %s", file.Name(), err))
	}
}

func Save() {
	saveSettings(fileName, fileStorage)
}

func saveSettings(path string, source interface{}) {
	file, err := os.Create(path)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	if err := encoder.Encode(source); err != nil {
		panic(err)
	}

	if err := file.Close(); err != nil {
		panic(err)
	}
}

func GetFormat() *fileformat {
	return fileStorage
}
