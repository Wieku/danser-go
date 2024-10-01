package audio

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/math/mutils"
	"strconv"
	"strings"
	"unicode"
)

var Samples [3][7]*bass.Sample
var MapSamples [3][7]map[int]*bass.Sample

var sets = map[string]int{
	"normal": 1,
	"soft":   2,
	"drum":   3,
}

var hitsounds = map[string]int{
	"hitnormal":     1,
	"hitwhistle":    2,
	"hitfinish":     3,
	"hitclap":       4,
	"slidertick":    5,
	"sliderslide":   6,
	"sliderwhistle": 7,
}

var listeners = make([]func(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64), 0)

func AddListener(function func(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64)) {
	listeners = append(listeners, function)
}

func LoadSamples() {
	Samples[0][0] = LoadSample("normal-hitnormal")
	Samples[0][1] = LoadSample("normal-hitwhistle")
	Samples[0][2] = LoadSample("normal-hitfinish")
	Samples[0][3] = LoadSample("normal-hitclap")
	Samples[0][4] = LoadSample("normal-slidertick")
	Samples[0][5] = LoadSample("normal-sliderslide")
	Samples[0][6] = LoadSample("normal-sliderwhistle")

	Samples[1][0] = LoadSample("soft-hitnormal")
	Samples[1][1] = LoadSample("soft-hitwhistle")
	Samples[1][2] = LoadSample("soft-hitfinish")
	Samples[1][3] = LoadSample("soft-hitclap")
	Samples[1][4] = LoadSample("soft-slidertick")
	Samples[1][5] = LoadSample("soft-sliderslide")
	Samples[1][6] = LoadSample("soft-sliderwhistle")

	Samples[2][0] = LoadSample("drum-hitnormal")
	Samples[2][1] = LoadSample("drum-hitwhistle")
	Samples[2][2] = LoadSample("drum-hitfinish")
	Samples[2][3] = LoadSample("drum-hitclap")
	Samples[2][4] = LoadSample("drum-slidertick")
	Samples[2][5] = LoadSample("drum-sliderslide")
	Samples[2][6] = LoadSample("drum-sliderwhistle")
}

func PlaySample(sampleSet, additionSet, hitsound, index int, volume float64, objNum int64, xPos float64) {
	if additionSet == 0 {
		additionSet = sampleSet
	}

	volume = max(volume, 0.08)

	// Play normal
	if skin.GetInfo().LayeredHitSounds || hitsound&1 > 0 || hitsound == 0 {
		playSample(sampleSet, 0, index, volume*0.8, objNum, xPos)
	}

	// Play whistle
	if hitsound&2 > 0 {
		playSample(additionSet, 1, index, volume*0.85, objNum, xPos)
	}

	// Play finish
	if hitsound&4 > 0 {
		playSample(additionSet, 2, index, volume, objNum, xPos)
	}

	// Play clap
	if hitsound&8 > 0 {
		playSample(additionSet, 3, index, volume*0.85, objNum, xPos)
	}
}

func playSample(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64, xPos float64) {
	balance := 0.0
	if settings.DIVIDES == 1 {
		balance = mutils.Clamp((xPos-256)/512*settings.Audio.HitsoundPositionMultiplier, -1, 1)
	}

	if settings.Audio.IgnoreBeatmapSampleVolume {
		volume = 1.0
	}

	if sampleSet == 0 {
		sampleSet = 2
	} else if sampleSet < 0 || sampleSet > 3 {
		sampleSet = 1
	}

	for _, f := range listeners {
		f(sampleSet, hitsoundIndex, index, volume, objNum)
	}

	if sample := MapSamples[sampleSet-1][hitsoundIndex][index]; sample != nil && !settings.Audio.IgnoreBeatmapSamples {
		sample.PlayRVPos(volume, balance)
	} else if Samples[sampleSet-1][hitsoundIndex] != nil {
		Samples[sampleSet-1][hitsoundIndex].PlayRVPos(volume, balance)
	}
}

var whistleChannel *bass.SampleChannel = nil
var slideChannel *bass.SampleChannel = nil
var lastSampleSet = 0
var lastAdditionSet = 0
var lastIndex = 0

func PlaySliderLoops(sampleSet, additionSet, hitsound, index int, volume float64, objNum int64, xPos float64) {
	if additionSet == 0 {
		additionSet = sampleSet
	}

	whistleUpdate := lastAdditionSet != additionSet || index != lastIndex || whistleChannel == nil
	slideUpdate := lastSampleSet != sampleSet || index != lastIndex || slideChannel == nil

	if hitsound&2 > 0 && whistleUpdate {
		if whistleChannel != nil {
			bass.StopSample(whistleChannel)
		}

		whistleChannel = playSampleLoop(additionSet, 6, index, volume, objNum, xPos)
	}

	if (hitsound&2 == 0 || skin.GetInfo().LayeredHitSounds) && slideUpdate {
		if slideChannel != nil {
			bass.StopSample(slideChannel)
		}

		slideChannel = playSampleLoop(sampleSet, 5, index, volume, objNum, xPos)
	}

	lastSampleSet = sampleSet
	lastAdditionSet = additionSet
	lastIndex = index
}

func StopSliderLoops() {
	lastSampleSet = 0
	lastAdditionSet = 0
	lastIndex = 0

	if whistleChannel != nil {
		bass.StopSample(whistleChannel)
	}

	if slideChannel != nil {
		bass.StopSample(slideChannel)
	}

	whistleChannel = nil
	slideChannel = nil
}

func playSampleLoop(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64, xPos float64) *bass.SampleChannel {
	balance := 0.0
	if settings.DIVIDES == 1 {
		balance = (xPos - 256) / 512 * settings.Audio.HitsoundPositionMultiplier
	}

	if settings.Audio.IgnoreBeatmapSampleVolume {
		volume = 1.0
	}

	if sampleSet == 0 {
		sampleSet = 2
	} else if sampleSet < 0 || sampleSet > 3 {
		sampleSet = 1
	}

	for _, f := range listeners {
		f(sampleSet, hitsoundIndex, index, volume, objNum)
	}

	if sample := MapSamples[sampleSet-1][hitsoundIndex][index]; sample != nil && !settings.Audio.IgnoreBeatmapSamples {
		return sample.PlayRVPosLoop(volume, balance)
	} else if Samples[sampleSet-1][hitsoundIndex] != nil {
		return Samples[sampleSet-1][hitsoundIndex].PlayRVPosLoop(volume, balance)
	}

	return nil
}

func PlaySliderTick(sampleSet, index int, volume float64, objNum int64, xPos float64) {
	playSample(sampleSet, 4, index, volume, objNum, xPos)
}

func LoadBeatmapSamples(fMap map[string]string) {
	splitBeforeDigit := func(name string) []string {
		for i, r := range name {
			if unicode.IsDigit(r) {
				return []string{name[:i], name[i:]}
			}
		}

		return []string{name}
	}

	for lName, fName := range fMap {
		if strings.Contains(lName, "/") || (!strings.HasSuffix(lName, ".wav") && !strings.HasSuffix(lName, ".mp3") && !strings.HasSuffix(lName, ".ogg")) {
			continue
		}

		rawName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(lName, ".wav"), ".ogg"), ".mp3")

		if separated := strings.Split(rawName, "-"); len(separated) == 2 {
			setID := sets[separated[0]]

			if setID == 0 {
				continue
			}

			subSeparated := splitBeforeDigit(separated[1])

			hitSoundIndex := 1

			if len(subSeparated) > 1 {
				index, err := strconv.ParseInt(subSeparated[1], 10, 32)

				if err != nil {
					continue
				}

				hitSoundIndex = int(index)
			}

			hitSoundID := hitsounds[subSeparated[0]]

			if hitSoundID == 0 {
				continue
			}

			if MapSamples[setID-1][hitSoundID-1] == nil {
				MapSamples[setID-1][hitSoundID-1] = make(map[int]*bass.Sample)
			}

			MapSamples[setID-1][hitSoundID-1][hitSoundIndex] = bass.NewSample(fName)
		}
	}
}

func LoadSample(name string) *bass.Sample {
	return skin.GetSample(name)
}

func PlayFailSound() {
	sample := LoadSample("failsound")
	if sample != nil {
		sample.Play()
	}
}
