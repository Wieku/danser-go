package replay

import (
	"io/ioutil"
	"strings"
)

var replaydictionary string = "replays/"

func GetOsrFiles() (files []string, err error) {
	dir, err := ioutil.ReadDir(replaydictionary)
	if err != nil {
		return nil, err
	}

	for _, fi := range dir {
		ok := strings.HasSuffix(fi.Name(), ".osr")
		if ok {
			files = append(files, replaydictionary+fi.Name())
		}
	}

	return files,nil
}
