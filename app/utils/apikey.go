package utils

import (
	"errors"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var apiKey *string
var keyError error

func GetApiKey() (string, error) {
	if apiKey == nil {
		migrateKey()

		var key string

		file, err := os.Open(filepath.Join(env.ConfigDir(), "api.txt"))
		if err == nil {
			defer file.Close()

			var data []byte

			data, err = io.ReadAll(files.NewUnicodeReader(file))
			if err == nil {
				key = strings.TrimSpace(string(data))
			}
		}

		apiKey = &key
		keyError = err
	}

	if *apiKey == "" && keyError == nil {
		return "", errors.New("api.txt is empty")
	}

	return *apiKey, keyError
}

func migrateKey() {
	_, err := os.Stat(filepath.Join(env.DataDir(), "api.txt"))
	if err == nil {
		err = os.Rename(filepath.Join(env.DataDir(), "api.txt"), filepath.Join(env.ConfigDir(), "api.txt"))
		if err != nil {
			panic(err)
		}
	}
}
