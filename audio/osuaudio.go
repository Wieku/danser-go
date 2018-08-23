package audio

import (
	"github.com/wieku/danser/settings"
	"os"
	"log"
	"strconv"
)

var Samples [5][3]*Sample
var MapSamples [5][3]map[int]*Sample

func LoadSamples() {

	Samples[0][0] = NewSample("assets/sounds/normal-hitnormal.wav")
	Samples[1][0] = NewSample("assets/sounds/normal-hitwhistle.wav")
	Samples[2][0] = NewSample("assets/sounds/normal-hitfinish.wav")
	Samples[3][0] = NewSample("assets/sounds/normal-hitclap.wav")
	Samples[4][0] = NewSample("assets/sounds/normal-slidertick.wav")

	Samples[0][1] = NewSample("assets/sounds/soft-hitnormal.wav")
	Samples[1][1] = NewSample("assets/sounds/soft-hitwhistle.wav")
	Samples[2][1] = NewSample("assets/sounds/soft-hitfinish.wav")
	Samples[3][1] = NewSample("assets/sounds/soft-hitclap.wav")
	Samples[4][1] = NewSample("assets/sounds/soft-slidertick.wav")

	Samples[0][2] = NewSample("assets/sounds/drum-hitnormal.wav")
	Samples[1][2] = NewSample("assets/sounds/drum-hitwhistle.wav")
	Samples[2][2] = NewSample("assets/sounds/drum-hitfinish.wav")
	Samples[3][2] = NewSample("assets/sounds/drum-hitclap.wav")
	Samples[4][2] = NewSample("assets/sounds/drum-slidertick.wav")
}

func PlaySample(sampleSet, additionSet, hitsound, index int) {
	playSample(sampleSet, 0, index)

	if additionSet == 0 {
		additionSet = sampleSet
	}

	if hitsound&2 > 0 {
		playSample(additionSet, 1, index)
	}
	if hitsound&4 > 0 {
		playSample(additionSet, 2, index)
	}
	if hitsound&8 > 0 {
		playSample(additionSet, 3, index)
	}
}

func playSample(sampleSet int, hitsoundIndex, index int) {
	if sample := MapSamples[hitsoundIndex][sampleSet-1][index]; sample != nil {
		sample.Play()
	} else {
		Samples[hitsoundIndex][sampleSet-1].Play()
	}
}

func PlaySliderTick(sampleSet, index int) {
	if sample := MapSamples[4][sampleSet-1][index]; sample != nil {
		sample.Play()
	} else {
		Samples[4][sampleSet-1].Play()
	}
}

func RegisterBeatmapSample(dir string, sampleSet, hitsound, index int) {
	if index == 0 {
		return
	}

	sampleSetText := "normal"

	if sampleSet == 2 {
		sampleSetText = "soft"
	} else if sampleSet == 3 {
		sampleSetText = "drum"
	}

	loadSample(dir, sampleSetText, "hitnormal", sampleSet, 0, index)
	loadSample(dir, sampleSetText, "slidertick", sampleSet, 4, index)

	if hitsound&2 > 0 {
		loadSample(dir, sampleSetText, "hitwhistle", sampleSet, 1, index)
	}
	if hitsound&4 > 0 {
		loadSample(dir, sampleSetText, "hitfinish", sampleSet, 2, index)
	}
	if hitsound&8 > 0 {
		loadSample(dir, sampleSetText, "hitclap", sampleSet, 3, index)
	}
}

func loadSample(dir, sampleSet, hitsound string, sampleSetIndex, hitsoundIndex, index int) {
	path := settings.General.OsuSongsDir + string(os.PathSeparator) + dir + string(os.PathSeparator) + sampleSet + "-" + hitsound

	if index > 1 {
		path += strconv.FormatInt(int64(index), 10)
	}

	path += ".wav"

	if sample := NewSample(path); sample != nil {

		if MapSamples[hitsoundIndex][sampleSetIndex-1] == nil {
			MapSamples[hitsoundIndex][sampleSetIndex-1] = make(map[int]*Sample)
		}

		if MapSamples[hitsoundIndex][sampleSetIndex-1][index] != nil {
			return
		}

		log.Println("Loaded:", path)
		MapSamples[hitsoundIndex][sampleSetIndex-1][index] = sample
	}
}