package hitjudge

import (
	"danser/settings"
	"encoding/json"
	"io/ioutil"
)

func SaveError(errors []Error) {
	oerr := ioutil.WriteFile(settings.General.ErrorFixFile, []byte(getErrorCache(errors)), 0666)
	if oerr != nil {
		panic(oerr)
	}
}

func ReadError() ([]Error) {
	oread, _ := ioutil.ReadFile(settings.General.ErrorFixFile)
	return setErrorCache(oread)
}

func getErrorCache(errors []Error) string {
	data, err := json.MarshalIndent(errors, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func setErrorCache(r []byte) []Error {
	var errors []Error
	if err := json.Unmarshal(r, &errors); err != nil {
		panic(err)
	}
	return errors
}

// 过滤Error
func FilterError(replayindex int, errors []Error) []Error {
	reerror := []Error{}
	for _, err := range errors {
		if err.ReplayIndex == replayindex {
			reerror = append(reerror, err)
		}
	}
	return reerror
}