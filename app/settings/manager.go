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
var watcher *fsnotify.Watcher
var allSettingsFile string
var baseFileName string

func initSettings() *fileformat {
	return &fileformat{
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

func InitStorage() {
	fileStorage = initSettings()
}

func LoadOverride(name string) (newOverride bool) {
	newOverride = false

	fileName := "overrides/local/" + name + ".json"
	file, err := os.Open(fileName);

	if os.IsNotExist(err) {
		fileName = "overrides/" + name + ".json"
		file, err = os.Open(fileName)

		if os.IsNotExist(err) {
			fileName = "overrides/local/" + name + ".json"
			createEmpty(fileName)
			newOverride = true
		}
	}

	defer file.Close()

	if err != nil {
		panic(err)
	} else {
		load(file, fileStorage)
	}

	if !RECORD {
		setupWatcher(fileName)
	}

	return
}

func LoadSettings(version string) (newSettings bool) {
	newSettings = false

	baseFileName = "settings"

	if version != "" {
		baseFileName += "-" + version
	}
	baseFileName += ".json"


	file, err := os.Open(baseFileName)
	defer file.Close()
	if os.IsNotExist(err) {
		saveBase()
		newSettings = true
	} else if err != nil {
		panic(err)
	} else {
		load(file, fileStorage)
		saveBase()
	}

	if !RECORD {
		setupWatcher(baseFileName)
	}

	return
}

func setupWatcher(fileName string) {
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

					SaveAll()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	abs, _ := filepath.Abs(fileName)

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

func saveBase() {
	saveSettings(baseFileName, fileStorage)
}

func UpdateBase(updateJSON json.RawMessage) {
	newBase := initSettings()
	file, err := os.Open(baseFileName)
	if err != nil{
		panic(err)
	}
	defer file.Close()
	load(file, newBase)
	if err := json.Unmarshal(updateJSON, &newBase); err != nil {
		panic(err)
	}
	saveSettings(baseFileName, newBase)
	SaveAll()
}

func SetAllSettingsFile(fileName string) {
	allSettingsFile = fileName
}

func SaveAll() {
	if allSettingsFile != "" {
		saveSettings(allSettingsFile, fileStorage)
	}
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

func createEmpty(path string) {
	file, err := os.Create(path)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	file.Write([]byte("{}\n"))
}

func GetFormat() *fileformat {
	return fileStorage
}
