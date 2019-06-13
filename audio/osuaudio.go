package audio

import (
	"github.com/wieku/danser-go/settings"
	"os"
	"strconv"
	"path/filepath"
	"strings"
	"unicode"
)

var Samples [3][5]*Sample
var MapSamples [3][5]map[int]*Sample

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

func LoadSamples() {
	Samples[0][0] = NewSample("assets/sounds/normal-hitnormal.wav")
	Samples[0][1] = NewSample("assets/sounds/normal-hitwhistle.wav")
	Samples[0][2] = NewSample("assets/sounds/normal-hitfinish.wav")
	Samples[0][3] = NewSample("assets/sounds/normal-hitclap.wav")
	Samples[0][4] = NewSample("assets/sounds/normal-slidertick.wav")

	Samples[1][0] = NewSample("assets/sounds/soft-hitnormal.wav")
	Samples[1][1] = NewSample("assets/sounds/soft-hitwhistle.wav")
	Samples[1][2] = NewSample("assets/sounds/soft-hitfinish.wav")
	Samples[1][3] = NewSample("assets/sounds/soft-hitclap.wav")
	Samples[1][4] = NewSample("assets/sounds/soft-slidertick.wav")

	Samples[2][0] = NewSample("assets/sounds/drum-hitnormal.wav")
	Samples[2][1] = NewSample("assets/sounds/drum-hitwhistle.wav")
	Samples[2][2] = NewSample("assets/sounds/drum-hitfinish.wav")
	Samples[2][3] = NewSample("assets/sounds/drum-hitclap.wav")
	Samples[2][4] = NewSample("assets/sounds/drum-slidertick.wav")
}

func PlaySample(sampleSet, additionSet, hitsound, index int, volume float64) {
	playSample(sampleSet, 0, index, volume)

	if additionSet == 0 {
		additionSet = sampleSet
	}

	if hitsound&2 > 0 {
		playSample(additionSet, 1, index, volume)
	}
	if hitsound&4 > 0 {
		playSample(additionSet, 2, index, volume)
	}
	if hitsound&8 > 0 {
		playSample(additionSet, 3, index, volume)
	}
}

func playSample(sampleSet int, hitsoundIndex, index int, volume float64) {
	if settings.Audio.IgnoreBeatmapSampleVolume {
		volume = 1.0
	}

	if sample := MapSamples[sampleSet-1][hitsoundIndex][index]; sample != nil {
		sample.PlayRV(volume)
	} else {
		Samples[sampleSet-1][hitsoundIndex].PlayRV(volume)
	}
}

func PlaySliderTick(sampleSet, index int, volume float64) {
	playSample(sampleSet, 4, index, volume)
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
		if !strings.HasSuffix(info.Name(), ".wav") {
			return nil
		}

		rawName := strings.TrimSuffix(info.Name(), ".wav")

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
				MapSamples[setID-1][hitSoundID-1] = make(map[int]*Sample)
			}

			MapSamples[setID-1][hitSoundID-1][hitSoundIndex] = NewSample(path)

		}

		return nil
	})
}
