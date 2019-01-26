package score

import (
	"danser/settings"
	"github.com/flesnuk/oppai5"
	"math"
	"os"
)

// 部分载入map
func LoadMapbyNum(filename string, objnum int) *oppai.Map {
	f, _ := os.Open(filename)
	return oppai.ParsebyNum(f, objnum)
}

// 计算每帧实时pp
func CalculateRealtimePP(firstpp float64, secondpp float64, firsttime int64, secondtime int64, nowtime float64) (realpp float64) {
	deltapp := secondpp - firstpp
	deltatime := math.Min(float64(secondtime - firsttime), settings.VSplayer.PlayerInfoUI.RealTimePPGap)
	realpp = firstpp + deltapp * math.Max(math.Min(math.Min(nowtime - float64(firsttime + settings.VSplayer.PlayerFieldUI.HitFadeTime), settings.VSplayer.PlayerInfoUI.RealTimePPGap) / deltatime, 1), 0)
	if math.IsNaN(realpp) {
		realpp = 0.0
	}
	return realpp
}