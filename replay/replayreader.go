package replay

import (
	"danser/settings"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func GetOsrFiles() (files []string, err error) {
	dir, err := ioutil.ReadDir(settings.VSplayer.ReplayandCache.ReplayDir)
	if err != nil {
		return nil, err
	}

	if settings.VSplayer.PlayerInfo.SpecifiedPlayers {
		specifiedplayers := strings.Split(settings.VSplayer.PlayerInfo.SpecifiedLine, ",")
		specifiedplayerindex := []int{}
		for _, player := range specifiedplayers {
			pl, _ := strconv.Atoi(player)
			if pl <= 0 {
				log.Panic("指定player的字符串有误，请重新检查设定")
			}else {
				specifiedplayerindex = append(specifiedplayerindex, pl)
			}
		}
		for i, fi := range dir {
			ok := strings.HasSuffix(fi.Name(), ".osr")
			if ok {
				for _, pl := range specifiedplayerindex {
					if i+1 == pl {
						files = append(files, settings.VSplayer.ReplayandCache.ReplayDir+fi.Name())
					}
				}
			}
		}
	}else {
		for _, fi := range dir {
			ok := strings.HasSuffix(fi.Name(), ".osr")
			if ok {
				files = append(files, settings.VSplayer.ReplayandCache.ReplayDir+fi.Name())
			}
		}
	}
	return files,nil
}
