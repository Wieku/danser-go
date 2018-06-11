package settings

import (
	"strconv"
	"os"
	"encoding/json"
)

var fileStorage *fileformat
var fileName string
func initDefaults() {
	Version = SETTINGSVERSION
	General = &general{os.Getenv("localappdata") + string(os.PathSeparator) + "osu!" + string(os.PathSeparator) + "Songs" + string(os.PathSeparator)}
	Graphics = &graphics{1920, 1080, 1280, 720, true, false, 1000, 16}
	Audio = &audio{0.5, 0.5, 0.5, false}
	Beat = &beat{1.4}
	Cursor = &cursor{&color{true, 8, &hsv{0, 1.0, 1.0}, false, 0, false, 0, 0}, true, -36.0, false, 18, true, true, false, 0.4}
	Objects = &objects{5, true, &color{true, 8, &hsv{0, 1.0, 1.0}, false, 0, true, 100.0, 0}, -1, true, 30, 50, true, true, 0.0, false, &color{false, 8, &hsv{0, 0.0, 1.0}, false, 0, true, 100.0, 0}}
	Playfield = &playfield{5, 0, 0.95, 0.95, 1,true, 1.1}
	fileStorage = &fileformat{&Version, General, Graphics, Audio, Beat, Cursor, Objects, Playfield}
}

func LoadSettings(version int) bool {
	initDefaults()
	fileName = "settings"

	if version > 0 {
		fileName += "-" + strconv.FormatInt(int64(version), 10)
	}
	fileName += ".json"

	file, err := os.Open(fileName)
	defer file.Close()
	if os.IsNotExist(err) {
		saveSettings(fileName)
		return true
	} else if err != nil {
		panic(err)
	} else {
		load(file)
		saveSettings(fileName) //this is done to save additions from the current format
	}

	return false
}

func load(file *os.File) {
	decoder := json.NewDecoder(file)
	decoder.Decode(fileStorage)
}

func Save() {
	saveSettings(fileName)
}

func saveSettings(path string) {
	file, err := os.Create(path)
	defer file.Close()

	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	Version = SETTINGSVERSION
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	encoder.Encode(fileStorage)
}
