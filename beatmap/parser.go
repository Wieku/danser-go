package beatmap

import (
	"sort"
	"github.com/wieku/danser-go/beatmap/objects"
	"strconv"
	"strings"
	"os"
	"bufio"
	"github.com/wieku/danser-go/settings"
	"errors"
	"path/filepath"
)

func parseGeneral(line []string, beatMap *BeatMap) bool {
	switch line[0] {
	case "Mode":
		if line[1] != "0" {
			return true
		}
	case "StackLeniency":
		beatMap.StackLeniency, _ = strconv.ParseFloat(line[1], 64)
	case "AudioFilename":
		beatMap.Audio += line[1]
	case "SampleSet":
		switch line[1] {
		case "Normal", "All":
			beatMap.Timings.BaseSet = 1
		case "Soft":
			beatMap.Timings.BaseSet = 2
		case "Drum":
			beatMap.Timings.BaseSet = 3
		}
		beatMap.Timings.LastSet = beatMap.Timings.BaseSet
	}

	return false
}

func parseMetadata(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "Title":
		beatMap.Name = line[1]
	case "TitleUnicode":
		beatMap.NameUnicode = line[1]
	case "Artist":
		beatMap.Artist = line[1]
	case "ArtistUnicode":
		beatMap.ArtistUnicode = line[1]
	case "Creator":
		beatMap.Creator = line[1]
	case "Version":
		beatMap.Difficulty = line[1]
	case "Source":
		beatMap.Source = line[1]
	case "Tags":
		beatMap.Tags = line[1]
	}
}

func parseDifficulty(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "SliderMultiplier":
		beatMap.SliderMultiplier, _ = strconv.ParseFloat(line[1], 64)
		beatMap.Timings.SliderMult = float64(beatMap.SliderMultiplier)
	case "ApproachRate":
		beatMap.AR, _ = strconv.ParseFloat(line[1], 64)
	case "CircleSize":
		beatMap.CircleSize, _ = strconv.ParseFloat(line[1], 64)
	case "SliderTickRate":
		beatMap.Timings.TickRate, _ = strconv.ParseFloat(line[1], 64)
	case "HPDrainRate":
		beatMap.HPDrainRate, _ = strconv.ParseFloat(line[1], 64)
	case "OverallDifficulty":
		beatMap.OverallDifficulty, _ = strconv.ParseFloat(line[1], 64)
	}
}

func parseEvents(line []string, beatMap *BeatMap) {
	if line[0] == "0" {
		beatMap.Bg += strings.Replace(line[2], "\"", "", -1)
	}
	if line[0] == "2" {
		beatMap.PausesText += line[1] + "," + line[2]
		beatMap.Pauses = append(beatMap.Pauses, objects.NewPause(line))
	}
}

func parseHitObjects(line []string, beatMap *BeatMap) {
	obj := objects.GetObject(line)

	if obj != nil {
		obj.SetTiming(beatMap.Timings)
		/*if o, ok := obj.(*objects.Slider); ok {
			o.SetTiming(beatMap.Timings)
		}
		if o, ok := obj.(*objects.Circle); ok {
			o.SetTiming(beatMap.Timings)
		}*/
		beatMap.HitObjects = append(beatMap.HitObjects, obj)
	}
}

func tokenize(line, delimiter string) []string {
	if strings.HasPrefix(line, "//") || !strings.Contains(line, delimiter) {
		return nil
	}
	divided := strings.Split(line, delimiter)
	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}
	return divided
}

func getSection(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") {
		return strings.TrimRight(strings.TrimLeft(line, "["), "]")
	}
	return ""
}

func ParseBeatMap(beatMap *BeatMap) error {
	file, err := os.Open(settings.General.OsuSongsDir + string(os.PathSeparator) + beatMap.Dir + string(os.PathSeparator) + beatMap.File)
	defer file.Close()

	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)

	var currentSection string
	counter := 0
	counter1 := 0

	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "General":
			if arr := tokenize(line, ":"); len(arr) > 1 {
				if err := parseGeneral(arr, beatMap); err {
					return errors.New("wrong mode")
				}
			}
		case "Metadata":
			if arr := tokenize(line, ":"); len(arr) > 1 {
				parseMetadata(arr, beatMap)
			}
		case "Difficulty":
			if arr := tokenize(line, ":"); len(arr) > 1 {
				parseDifficulty(arr, beatMap)
			}
		case "Events":
			if arr := tokenize(line, ","); len(arr) > 1 {
				if arr[0] == "2" {
					if counter1 > 0 {
						beatMap.PausesText += ","
					}
					counter1++
				}

				parseEvents(arr, beatMap)
			}
		case "TimingPoints":
			if arr := tokenize(line, ","); len(arr) > 1 {
				if counter > 0 {
					beatMap.TimingPoints += "|"
				}
				counter++

				beatMap.TimingPoints += line
			}
		}
	}

	beatMap.LoadTimingPoints()

	file.Seek(0, 0)

	if beatMap.Name+beatMap.Artist+beatMap.Creator == "" || beatMap.TimingPoints == "" {
		return errors.New("corrupted file")
	}

	return nil
}

func ParseBeatMapFile(file *os.File) *BeatMap {
	beatMap := NewBeatMap()
	beatMap.Dir = filepath.Base(filepath.Dir(file.Name()))
	f, _ := file.Stat()
	beatMap.File = f.Name()

	err := ParseBeatMap(beatMap)

	if err != nil {
		return nil
	}

	return beatMap
}

func ParseObjects(beatMap *BeatMap) {

	file, err := os.Open(settings.General.OsuSongsDir + string(os.PathSeparator) + beatMap.Dir + string(os.PathSeparator) + beatMap.File)
	defer file.Close()

	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)

	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				parseHitObjects(arr, beatMap)
			}
			break
		}
	}

	sort.Slice(beatMap.HitObjects, func(i, j int) bool { return beatMap.HitObjects[i].GetBasicData().StartTime < beatMap.HitObjects[j].GetBasicData().StartTime })

	num := 0
	comboNumber := 1
	comboSet := 0
	for _, o := range beatMap.HitObjects {
		_, ok := o.(*objects.Pause)

		if !ok {
			o.GetBasicData().Number = int64(num)
			if o.GetBasicData().NewCombo {
				comboNumber = 1
				comboSet++
			}

			o.GetBasicData().ComboNumber = int64(comboNumber)
			o.GetBasicData().ComboSet = int64(comboSet)

			comboNumber++
			num++
		}

	}

	calculateStackLeniency(beatMap)
}
