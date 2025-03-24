package cstats

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
)

type mod struct {
	Acronym string
	Name    string
}

type StatHolder struct {
	bMap       *beatmap.BeatMap
	diff       *difficulty.Difficulty
	stats      map[string]any
	clickTimes []float64
}

func NewStatHolder(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) *StatHolder {
	holder := &StatHolder{
		stats: make(map[string]any),
		bMap:  bMap,
		diff:  diff,
	}

	holder.SetMapStats(bMap, diff)
	holder.SetScoreStats(osu.Score{})
	holder.SetFCPP(api.PPv2Results{})
	holder.SetSSPP(api.PPv2Results{})
	holder.SetUsername("")
	holder.SetHP(1)

	return holder
}

func (h *StatHolder) SetMapStats(bMap *beatmap.BeatMap, pDiff *difficulty.Difficulty) {
	h.stats["artist"] = bMap.Artist
	h.stats["title"] = bMap.Name
	h.stats["creator"] = bMap.Creator
	h.stats["version"] = bMap.Difficulty

	h.stats["baseAR"] = pDiff.GetBaseAR()
	h.stats["realAR"] = pDiff.ARReal

	h.stats["baseOD"] = pDiff.GetBaseAR()
	h.stats["realOD"] = pDiff.ODReal

	h.stats["baseCS"] = pDiff.GetBaseCS()
	h.stats["realCS"] = pDiff.GetCS()

	h.stats["baseHP"] = pDiff.GetBaseHP()
	h.stats["realHP"] = pDiff.GetHP()

	h.stats["avgBPM"] = (bMap.MinBPM + bMap.MaxBPM) / 2
	h.stats["bpm"] = bMap.Timings.Current.GetBaseBPM() * h.diff.GetSpeed()

	h.stats["speed"] = h.diff.GetSpeed()

	h.stats["timeFromStart"] = int64(0)
	h.stats["timeToEnd"] = int64(0)

	mods := make([]mod, 0)

	for _, m := range pDiff.ExportMods2() {
		mods = append(mods, mod{
			Acronym: m.Acronym,
			Name:    difficulty.ParseFromAcronym(m.Acronym).StringFull()[0],
		})
	}

	h.stats["mods"] = mods
	h.stats["modsA"] = pDiff.Mods.String()
}

func (h *StatHolder) UpdateBPM() {
	h.stats["bpm"] = h.bMap.Timings.Current.GetBaseBPM() * h.diff.GetSpeed()
}

func (h *StatHolder) SetUsername(name string) {
	h.stats["name"] = name
}

func (h *StatHolder) SetHP(hp float64) {
	h.stats["currentHP"] = hp
}

func (h *StatHolder) SetStars(attribs api.Attributes) {
	h.stats["starsAim"] = attribs.Aim
	h.stats["starsSpeed"] = attribs.Speed
	h.stats["starsFL"] = attribs.Flashlight
	h.stats["stars"] = attribs.Total
}

func (h *StatHolder) SetCurrentStars(attribs api.Attributes) {
	h.stats["cStarsAim"] = attribs.Aim
	h.stats["cStarsSpeed"] = attribs.Speed
	h.stats["cStarsFL"] = attribs.Flashlight
	h.stats["cStars"] = attribs.Total
}

func (h *StatHolder) SetScoreStats(score osu.Score) {
	h.stats["score"] = score.Score

	h.stats["countMiss"] = int64(score.CountMiss)
	h.stats["count50"] = int64(score.Count50)
	h.stats["count100"] = int64(score.Count100)
	h.stats["count300"] = int64(score.Count300)

	h.stats["countTicks"] = int64(score.MaxTicks) - int64(score.CountSB)
	h.stats["countSB"] = int64(score.CountSB)
	h.stats["maxTicks"] = int64(score.MaxTicks)

	h.stats["countSliderEnds"] = int64(score.SliderEnd)
	h.stats["maxSliderEnds"] = int64(score.MaxSliderEnd)

	h.stats["combo"] = int64(score.CurrentCombo)
	h.stats["maxCombo"] = int64(score.Combo)

	h.stats["acc"] = score.Accuracy

	h.stats["grade"] = score.Grade.String()

	h.stats["pp"] = score.PP.Total
	h.stats["ppAim"] = score.PP.Aim
	h.stats["ppSpeed"] = score.PP.Speed
	h.stats["ppAcc"] = score.PP.Acc
	h.stats["ppFL"] = score.PP.Flashlight
}

func (h *StatHolder) SetFCPP(pp api.PPv2Results) {
	h.stats["fcPP"] = pp.Total
	h.stats["fcPPAim"] = pp.Aim
	h.stats["fcPPSpeed"] = pp.Speed
	h.stats["fcPPAcc"] = pp.Acc
	h.stats["fcPPFL"] = pp.Flashlight
}

func (h *StatHolder) SetSSPP(pp api.PPv2Results) {
	h.stats["ssPP"] = pp.Total
	h.stats["ssPPAim"] = pp.Aim
	h.stats["ssPPSpeed"] = pp.Speed
	h.stats["ssPPAcc"] = pp.Acc
	h.stats["ssPPFL"] = pp.Flashlight
}

func (h *StatHolder) AddClick(time float64) {
	h.clickTimes = append(h.clickTimes, time)
}

func (h *StatHolder) UpdateTime(time float64) {
	var count int64

	earliestTimeValid := time - 1000*h.diff.GetSpeed()

	for i := len(h.clickTimes) - 1; i >= 0; i-- {
		if earliestTimeValid > h.clickTimes[i] {
			break
		}

		count++
	}

	h.stats["kps"] = count
}
