package settings

import (
	"encoding/json"
	"os"
	"strconv"
)

var fileStorage *fileformat
var fileName string

func initStorage() {
	fileStorage = &fileformat{
		General:   General,
		Graphics:  Graphics,
		Audio:     Audio,
		Cursor:    Cursor,
		Objects:   Objects,
		Playfield: Playfield,
		Dance:     Dance,
	}
}

func LoadSettings(version int) bool {
	initStorage()
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

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	encoder.Encode(fileStorage)
}
