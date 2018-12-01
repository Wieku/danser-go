package replay

import (
	"io/ioutil"
	"github.com/Mempler/rplpa"
	"log"
)

func PrintReplay() {
	buf, err := ioutil.ReadFile("replay-osu_807074_2644463955.osr")
	if err != nil {
		panic(err)
	}
	replay, err := rplpa.ParseReplay(buf)
	if err != nil {
		panic(err)
	}
	log.Println("Playmode: ", replay.PlayMode)
	log.Println("OsuVersion: ", replay.OsuVersion)
	log.Println("BeatmapMD5: ", replay.BeatmapMD5)
	log.Println("Username: ", replay.Username)
	log.Println("ReplayMD5: ", replay.ReplayMD5)
	log.Println("Count300: ", replay.Count300)
	log.Println("Count100: ", replay.Count100)
	log.Println("Count50: ", replay.Count50)
	for _, rdata := range replay.ReplayData {
		log.Println(rdata)
	}
}

func ExtractReplay(name string) *rplpa.Replay {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	replay, err := rplpa.ParseReplay(buf)
	if err != nil {
		panic(err)
	}
	return replay
}