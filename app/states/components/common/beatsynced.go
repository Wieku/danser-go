package common

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type BeatSynced struct {
	*sprite.Sprite

	bMap  *beatmap.BeatMap
	music bass.ITrack

	timingPoint objects.TimingPoint

	Divisor float64

	rawProgress  float64
	beatProgress float64
	beatIndex    int

	lastTime float64

	averageVolume float64

	Beat     float64
	lastBeat float64

	IsSynced bool
	Kiai     bool
}

func NewBeatSynced() *BeatSynced {
	return &BeatSynced{
		Sprite:   sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(0, 0), vector.Centre),
		lastTime: math.NaN(),
		Divisor:  1,
	}
}

func (bs *BeatSynced) SetMap(bMap *beatmap.BeatMap, track bass.ITrack) {
	bs.bMap = bMap
	bs.music = track
}

func (bs *BeatSynced) Update(time float64) {
	if math.IsNaN(bs.lastTime) {
		bs.lastTime = time
	}

	bs.Sprite.Update(time)

	if bs.music != nil && bs.bMap != nil {
		var mTime float64

		if bs.music.GetState() == bass.MusicPlaying {
			mTime = bs.music.GetPosition() * 1000
			bs.timingPoint = bs.bMap.Timings.GetOriginalPointAt(mTime)
			bs.IsSynced = true
		} else {
			mTime = time
			bs.timingPoint = bs.bMap.Timings.GetDefault()
			bs.IsSynced = false
		}

		beatLength := bs.timingPoint.GetBaseBeatLength() / bs.Divisor

		bs.Kiai = bs.timingPoint.Kiai

		bs.rawProgress = (mTime - bs.timingPoint.Time) / beatLength

		bs.beatProgress = math.Mod(bs.rawProgress, 1)
		if mTime < bs.timingPoint.Time {
			bs.beatProgress += 1
		}

		bs.beatIndex = int(bs.rawProgress)
		if bs.timingPoint.OmitFirstBarLine {
			bs.beatIndex--
		}

		if mTime < bs.timingPoint.Time {
			bs.beatIndex--
		}
	} else {
		bs.beatProgress = 1
	}

	delta := max(0, time-bs.lastTime)

	ratio60 := delta / 16.6666666666667

	volume := 1.0
	if bs.music != nil && bs.music.GetState() == bass.MusicPlaying {
		volume = bs.music.GetLevelCombined()
	}

	volumeRatio := math.Pow(0.9, ratio60)

	bs.averageVolume = bs.averageVolume*volumeRatio + volume*(1.0-volumeRatio)

	volumeProgress := 1 - (volume-bs.averageVolume)*2

	beatRatio := math.Pow(0.5, ratio60)

	beat := mutils.Clamp(1.0-(volumeProgress*0.5+bs.beatProgress*0.5), 0.0, 1.0)

	bs.Beat = bs.lastBeat*beatRatio + beat*(1-beatRatio)
	bs.lastBeat = bs.Beat

	bs.lastTime = time
}
