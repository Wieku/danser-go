package settings

import (
	"github.com/fsnotify/fsnotify"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/goroutines"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var currentConfig *Config

var filePath string
var watcher *fsnotify.Watcher

func initSettings() {
	if err := os.MkdirAll(env.ConfigDir(), 0755); err != nil {
		panic(err)
	}

	migrateSettings()
}

func LoadSettings(version string) bool {
	initSettings()

	fileName := "default"
	if version != "" {
		fileName = version
	}

	fileName += ".json"

	filePath = filepath.Join(env.ConfigDir(), fileName)

	file, err := os.Open(filePath)

	newFile := false

	if os.IsNotExist(err) {
		currentConfig = NewConfigFile()
		currentConfig.Save(filePath, true)

		newFile = true
	} else if err != nil {
		panic(err)
	} else {
		currentConfig, err = LoadConfig(file)
		if err != nil {
			panic(err)
		}

		file.Close()

		currentConfig.Save("", false) // this is done to save additions from the current format
	}

	LoadCredentials()

	currentConfig.attachToGlobals()

	if !RECORD {
		setupWatcher(filePath)
	}

	return newFile
}

func setupWatcher(file string) {
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

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("SettingsManager: Detected", file, "modification, reloading...")

					time.Sleep(time.Millisecond * 200)

					sFile, _ := os.Open(event.Name)

					currentConfig, err = LoadConfig(sFile)
					if err != nil {
						panic(err)
					}

					sFile.Close()

					currentConfig.Save("", false)
					currentConfig.attachToGlobals()
				}
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

func CloseWatcher() {
	if watcher != nil {
		err := watcher.Close()
		if err != nil {
			log.Println(err)
		}

		watcher = nil
	}
}

func GetCompressedString() string {
	return currentConfig.GetCompressedString()
}

func Save() {
	currentConfig.Save(filePath, false)
}

func GetFormat() *Config {
	return currentConfig
}

func CreateDefault() {
	initSettings()

	currentConfig = NewConfigFile()
	currentConfig.attachToGlobals()
}

func migrateSettings() {
	currentSettings, _ := files.SearchFiles(env.DataDir(), "settings*.json", 0)

	for _, osPathname := range currentSettings {
		destName := filepath.Base(osPathname)

		if destName == "settings.json" {
			destName = "default.json"
		} else {
			destName = strings.TrimPrefix(destName, "settings-")
		}

		err := files.MoveFile(osPathname, filepath.Join(env.ConfigDir(), destName))
		if err != nil {
			panic(err)
		}
	}
}
