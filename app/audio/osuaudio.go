package audio

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

var Samples [3][5]*bass.Sample
var MapSamples [3][5]map[int]*bass.Sample

var sets = map[string]int{
	"normal": 1,
	"soft":   2,
	"drum":   3,
}

var hitsounds = map[string]int{
	"hitnormal":  1,
	"hitwhistle": 2,
	"hitfinish":  3,
	"hitclap":    4,
	"slidertick": 5,
}

var listeners = make([]func(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64), 0)

func AddListener(function func(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64)) {
	listeners = append(listeners, function)
}

func LoadSamples() {
	Samples[0][0] = LoadSample("assets/sounds/normal-hitnormal")
	Samples[0][1] = LoadSample("assets/sounds/normal-hitwhistle")
	Samples[0][2] = LoadSample("assets/sounds/normal-hitfinish")
	Samples[0][3] = LoadSample("assets/sounds/normal-hitclap")
	Samples[0][4] = LoadSample("assets/sounds/normal-slidertick")

	Samples[1][0] = LoadSample("assets/sounds/soft-hitnormal")
	Samples[1][1] = LoadSample("assets/sounds/soft-hitwhistle")
	Samples[1][2] = LoadSample("assets/sounds/soft-hitfinish")
	Samples[1][3] = LoadSample("assets/sounds/soft-hitclap")
	Samples[1][4] = LoadSample("assets/sounds/soft-slidertick")

	Samples[2][0] = LoadSample("assets/sounds/drum-hitnormal")
	Samples[2][1] = LoadSample("assets/sounds/drum-hitwhistle")
	Samples[2][2] = LoadSample("assets/sounds/drum-hitfinish")
	Samples[2][3] = LoadSample("assets/sounds/drum-hitclap")
	Samples[2][4] = LoadSample("assets/sounds/drum-slidertick")
}

func PlaySample(sampleSet, additionSet, hitsound, index int, volume float64, objNum int64, xPos float64) {
	playSample(sampleSet, 0, index, volume, objNum, xPos)

	if additionSet == 0 {
		additionSet = sampleSet
	}

	if hitsound&2 > 0 {
		playSample(additionSet, 1, index, volume, objNum, xPos)
	}
	if hitsound&4 > 0 {
		playSample(additionSet, 2, index, volume, objNum, xPos)
	}
	if hitsound&8 > 0 {
		playSample(additionSet, 3, index, volume, objNum, xPos)
	}
}

func playSample(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64, xPos float64) {
	balance := 0.0
	if settings.DIVIDES == 1 {
		balance = (xPos - 256) / 512 * settings.Audio.HitsoundPositionMultiplier
	}

	if settings.Audio.IgnoreBeatmapSampleVolume {
		volume = 1.0
	}

	for _, f := range listeners {
		f(sampleSet, hitsoundIndex, index, volume, objNum)
	}

	if sample := MapSamples[sampleSet-1][hitsoundIndex][index]; sample != nil && !settings.Audio.IgnoreBeatmapSamples {
		sample.PlayRVPos(volume, balance)
	} else {
		Samples[sampleSet-1][hitsoundIndex].PlayRVPos(volume, balance)
	}
}

func PlaySliderTick(sampleSet, index int, volume float64, objNum int64, xPos float64) {
	playSample(sampleSet, 4, index, volume, objNum, xPos)
}

func LoadBeatmapSamples(dir string) {
	splitBeforeDigit := func(name string) []string {
		for i, r := range name {
			if unicode.IsDigit(r) {
				return []string{name[:i], name[i:]}
			}
		}
		return []string{name}
	}

	fullPath := settings.General.OsuSongsDir + string(os.PathSeparator) + dir

	filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(info.Name(), ".wav") && !strings.HasSuffix(info.Name(), ".mp3") && !strings.HasSuffix(info.Name(), ".ogg") {
			return nil
		}

		rawName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(info.Name(), ".wav"), ".ogg"), ".mp3")

		if separated := strings.Split(rawName, "-"); len(separated) == 2 {

			setID := sets[separated[0]]

			if setID == 0 {
				return nil
			}

			subSeparated := splitBeforeDigit(separated[1])

			hitSoundIndex := 1

			if len(subSeparated) > 1 {
				index, err := strconv.ParseInt(subSeparated[1], 10, 32)

				if err != nil {
					return nil
				}

				hitSoundIndex = int(index)
			}

			hitSoundID := hitsounds[subSeparated[0]]

			if hitSoundID == 0 {
				return nil
			}

			if MapSamples[setID-1][hitSoundID-1] == nil {
				MapSamples[setID-1][hitSoundID-1] = make(map[int]*bass.Sample)
			}

			MapSamples[setID-1][hitSoundID-1][hitSoundIndex] = bass.NewSample(path)

		}

		return nil
	})
}

func LoadSample(name string) *bass.Sample {
	if sam := bass.NewSample(name + ".wav"); sam != nil {
		return sam
	}

	if sam := bass.NewSample(name + ".ogg"); sam != nil {
		return sam
	}

	if sam := bass.NewSample(name + ".mp3"); sam != nil {
		return sam
	}

	return nil
}

func LoadSampleLoop(name string) *bass.Sample {
	if sam := bass.NewSampleLoop(name + ".wav"); sam != nil {
		return sam
	}

	if sam := bass.NewSampleLoop(name + ".ogg"); sam != nil {
		return sam
	}

	if sam := bass.NewSampleLoop(name + ".mp3"); sam != nil {
		return sam
	}

	return nil
}
