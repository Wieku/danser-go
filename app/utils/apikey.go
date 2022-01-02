package utils

import (
	"github.com/wieku/danser-go/framework/files"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func GetApiKey() (string, error) {
	file, err := os.Open(filepath.Join(env.DataDir(), "api.txt"))
	if err != nil {
		return "", err
	}

	defer file.Close()

	data, err := ioutil.ReadAll(files.NewUnicodeReader(file))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}
