package beatmap

import (
	"sort"
	"github.com/wieku/danser/beatmap/objects"
	"strconv"
	"strings"
	"os"
	"bufio"

	"path/filepath"
)

func parseGeneral(line []string, beatMap *BeatMap) bool {

	switch line[0] {
	case "Mode":
		if line[1] != "0" {
			return true
		}
		break
	case "StackLeniency":
		beatMap.StackLeniency, _ = strconv.ParseFloat(line[1], 64)
		break
	case "AudioFilename":
		beatMap.Audio += line[1]
		break
	case "SampleSet":
		switch line[1] {
		case "Normal":
			beatMap.timings.BaseSet = 1
			break
		case "Soft":
			beatMap.timings.BaseSet = 2
			break
		case "Drum":
			beatMap.timings.BaseSet = 3
			break
		}
		beatMap.timings.LastSet = beatMap.timings.BaseSet
		break
	}

	return false
}

func parseMetadata(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "Title":
		beatMap.Name = line[1]
		break
	case "Artist":
		beatMap.Artist = line[1]
		break
	case "Creator":
		beatMap.Creator = line[1]
		break
	case "Version":
		beatMap.Difficulty = line[1]
		break
	}
}

func parseDifficulty(line []string, beatMap *BeatMap) {
	if line[0] == "SliderMultiplier" {
		beatMap.SliderMultiplier, _ = strconv.ParseFloat(line[1], 64)
		beatMap.timings.SliderMult = float64(beatMap.SliderMultiplier)
	}
	if line[0] == "ApproachRate" {
		beatMap.AR, _ = strconv.ParseFloat(line[1], 64)
	}
	if line[0] == "CircleSize" {
		beatMap.CircleSize, _ = strconv.ParseFloat(line[1], 64)
	}
	if line[0] == "SliderTickRate" {
		beatMap.timings.TickRate, _ = strconv.ParseFloat(line[1], 64)
	}

}

func parseEvents(line []string, beatMap *BeatMap) {
	if line[0] == "0" {
		beatMap.Bg += strings.Replace(line[2], "\"", "", -1)
	}
	if line[0] == "2" {
		beatMap.Pauses = append(beatMap.Pauses, objects.NewPause(line))
	}
}

func parseTimingPoints(line []string, beatMap *BeatMap) {
	time, _ := strconv.ParseInt(line[0], 10, 64)
	bpm, _ := strconv.ParseFloat(line[1], 64)
	if len(line) > 3 {
		sampleset, _ := strconv.ParseInt(line[3], 10, 64)
		beatMap.timings.LastSet = int(sampleset)
		beatMap.timings.AddPoint(time, bpm, int(sampleset))
	} else {
		beatMap.timings.AddPoint(time, bpm, beatMap.timings.LastSet)
	}

}

func parseHitObjects(line []string, beatMap *BeatMap) {
	obj := objects.GetObject(line)

	if obj != nil {
		if o, ok := obj.(*objects.Slider); ok {
			o.SetTiming(beatMap.timings)
		}
		if o, ok := obj.(*objects.Circle); ok {
			o.SetTiming(beatMap.timings)
		}
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

func ParseBeatMap(file *os.File) *BeatMap {
	scanner := bufio.NewScanner(file)

	beatMap := NewBeatMap()
	beatMap.Path = file.Name()
	beatMap.Audio = filepath.Dir(file.Name()) + string(os.PathSeparator)
	beatMap.Bg = beatMap.Audio
	var currentSection string

	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "General":
			if arr := tokenize(line, ":"); arr != nil {
				if err := parseGeneral(arr, beatMap); err {
					//log.Println("File is invalid, aborting...")
					return nil
				}
			}
			break
		case "Metadata":
			if arr := tokenize(line, ":"); arr != nil {
				parseMetadata(arr, beatMap)
			}
			break
		case "Difficulty":
			if arr := tokenize(line, ":"); arr != nil {
				parseDifficulty(arr, beatMap)
			}
			break
		case "Events":
			if arr := tokenize(line, ","); arr != nil {
				parseEvents(arr, beatMap)
			}
			break
		}
	}

	return beatMap
}

func ParseObjects(beatMap *BeatMap) {

	file, err := os.Open(beatMap.Path)
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
		case "TimingPoints":
			if arr := tokenize(line, ","); arr != nil {
				parseTimingPoints(arr, beatMap)
			}
			break
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				parseHitObjects(arr, beatMap)
			}
			break
		}
	}

	sort.Slice(beatMap.HitObjects, func(i, j int) bool {return beatMap.HitObjects[i].GetBasicData().StartTime < beatMap.HitObjects[j].GetBasicData().StartTime})

	num := 0

	for _, o := range beatMap.HitObjects {
		_, ok := o.(*objects.Pause)

		if !ok {
			o.GetBasicData().Number = int64(num)
			num++
		}

	}


	calculateStackLeniency(beatMap)
}