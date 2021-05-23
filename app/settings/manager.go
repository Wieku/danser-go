package settings

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/karrick/godirwalk"
	"github.com/wieku/danser-go/framework/util"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	err := os.Mkdir("settings", 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	migrateSettings()

	initStorage()

	fileName = "default"
	if version != "" {
		fileName = version
	}

	fileName += ".json"

	filePath := filepath.Join("settings", fileName)

	file, err := os.Open(filePath)

	defer file.Close()

	if os.IsNotExist(err) {
		saveSettings(filePath, fileStorage)
		return true
	} else if err != nil {
		panic(err)
	} else {
		load(file, fileStorage)
		saveSettings(filePath, fileStorage) //this is done to save additions from the current format
	}

	if !RECORD {
		setupWatcher(filePath)
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

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("SettingsManager: Detected", file, "modification, reloading...")

					time.Sleep(time.Millisecond * 200)

					sFile, _ := os.Open(fileName)

					load(sFile, fileStorage)

					sFile.Close()
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
	decoder := json.NewDecoder(util.NewUnicodeReader(file))
	if err := decoder.Decode(target); err != nil {
		panic(fmt.Sprintf("Failed to parse %s! Please re-check the file for mistakes. Error: %s", file.Name(), err))
	}
}

func Save() {
	saveSettings(fileName, fileStorage)
}

func saveSettings(path string, source interface{}) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

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

func migrateSettings() {
	_ = godirwalk.Walk("", &godirwalk.Options {
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if osPathname != "." && de.IsDir() {
				return godirwalk.SkipThis
			}

			if !strings.HasSuffix(osPathname, ".json") || !strings.HasPrefix(osPathname, "settings") {
				return nil
			}

			var destName string

			if osPathname == "settings.json" {
				destName = "default.json"
			} else {
				destName = strings.TrimPrefix(osPathname, "settings-")
			}

			err := os.Rename(osPathname, filepath.Join("settings", destName))
			if err != nil {
				panic(err)
			}

			return nil
		},
		Unsorted: true,
	})
}