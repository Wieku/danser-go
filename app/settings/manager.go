package settings

import (
	"encoding/json"
	"os"
)

var fileStorage *fileformat
var fileName string

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

	return false
}

func load(file *os.File, target interface{}) {
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		panic(err)
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
