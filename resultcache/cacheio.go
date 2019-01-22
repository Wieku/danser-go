package resultcache

import (
	"danser/hitjudge"
	"danser/settings"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

func SaveResult(oresult []hitjudge.ObjectResult, tresult []hitjudge.TotalResult, num int) {
	oerr := ioutil.WriteFile(settings.VSplayer.ReplayandCache.CacheDir+getFileNum(num)+".ocache", []byte(getObjectCache(oresult)), 0666)
	if oerr != nil {
		panic(oerr)
	}
	terr := ioutil.WriteFile(settings.VSplayer.ReplayandCache.CacheDir+getFileNum(num)+".tcache", []byte(getTotalCache(tresult)), 0666)
	if terr != nil {
		panic(terr)
	}
}

func ReadResult(num int) ([]hitjudge.ObjectResult, []hitjudge.TotalResult) {
	oread, _ := ioutil.ReadFile(settings.VSplayer.ReplayandCache.CacheDir+getFileNum(num)+".ocache")
	tread, _ := ioutil.ReadFile(settings.VSplayer.ReplayandCache.CacheDir+getFileNum(num)+".tcache")
	return setObjectCache(oread), setTotalCache(tread)
}

func getObjectCache(oresult []hitjudge.ObjectResult) string {
	data, err := json.MarshalIndent(oresult, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func getTotalCache(tresult []hitjudge.TotalResult) string {
	data, err := json.MarshalIndent(tresult, "", "     ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func setObjectCache(r []byte) []hitjudge.ObjectResult {
	var oresult []hitjudge.ObjectResult
	if err := json.Unmarshal(r, &oresult); err != nil {
		panic(err)
	}
	return oresult
}

func setTotalCache(r []byte) []hitjudge.TotalResult {
	var tresult []hitjudge.TotalResult
	if err := json.Unmarshal(r, &tresult); err != nil {
		panic(err)
	}
	return tresult
}

func getFileNum(num int) string {
	if num < 10 {
		return "0"+strconv.Itoa(num)
	}else{
		return strconv.Itoa(num)
	}
}



