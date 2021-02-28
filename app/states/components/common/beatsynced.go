package common

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"math"
)

type BeatSynced struct {
	*sprite.Sprite

	bMap           *beatmap.BeatMap
	music          *bass.Track
	lastBeatStart  float64
	lastBeatLength float64
	lastBeatProg   float64
	beatProgress   float64
	lastTime       float64
	volAverage     float64
	Progress       float64
	lastProgress   float64
	Kiai           bool
}

func NewBeatSynced() *BeatSynced {
	return &BeatSynced{Sprite: &sprite.Sprite{}}
}

func (bs *BeatSynced) SetMap(bMap *beatmap.BeatMap, track *bass.Track) {
	bs.bMap = bMap
	bs.music = track
}

func (bs *BeatSynced) Update(time float64) {
	if bs.lastTime == 0 {
		bs.lastTime = time
	}

	bs.Sprite.Update(time)

	mTime := bs.music.GetPosition() * 1000

	point := bs.bMap.Timings.GetPoint(mTime)
	bTime := point.BaseBpm

	bs.Kiai = point.Kiai

	if bTime != bs.lastBeatLength {
		bs.lastBeatLength = bTime
		bs.lastBeatStart = (point.Time - mTime) + time
		bs.lastBeatProg = math.Floor((time-bs.lastBeatStart)/bs.lastBeatLength) - 1
	}

	if math.Floor((time-bs.lastBeatStart)/bs.lastBeatLength) > bs.lastBeatProg {
		bs.lastBeatProg++
	}

	bs.beatProgress = (time-bs.lastBeatStart)/bs.lastBeatLength - bs.lastBeatProg

	ratio1 := (time - bs.lastTime) / 16.6666666666667

	bs.lastTime = time

	volume := 0.5

	if bs.music != nil && bs.music.GetState() == bass.MUSIC_PLAYING {
		volume = bs.music.GetLevelCombined()
	}

	ratioVol := math.Pow(0.9, ratio1)

	bs.volAverage = bs.volAverage*ratioVol + volume*(1.0-ratioVol)

	vprog := 1 - ((volume - bs.volAverage) / 0.5)
	pV := math.Min(1.0, math.Max(0.0, 1.0-(vprog*0.5+bs.beatProgress*0.5)))

	ratio := math.Pow(0.5, ratio1)

	bs.Progress = bs.lastProgress*ratio + (pV)*(1-ratio)
	bs.lastProgress = bs.Progress
}
