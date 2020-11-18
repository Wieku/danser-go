package beatmap

import (
	"bufio"
	"errors"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func parseGeneral(line []string, beatMap *BeatMap) bool {
	switch line[0] {
	case "Mode":
		beatMap.Mode, _ = strconv.ParseInt(line[1], 10, 64)
	case "StackLeniency":
		beatMap.StackLeniency, _ = strconv.ParseFloat(line[1], 64)
		if math.IsNaN(beatMap.StackLeniency) {
			beatMap.StackLeniency = 0.0
		}
	case "AudioFilename":
		beatMap.Audio += line[1]
	case "PreviewTime":
		beatMap.PreviewTime, _ = strconv.ParseInt(line[1], 10, 64)
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
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetAR(parsed)
	case "CircleSize":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetCS(parsed)
	case "SliderTickRate":
		beatMap.Timings.TickRate, _ = strconv.ParseFloat(line[1], 64)
	case "HPDrainRate":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetHPDrain(parsed)
	case "OverallDifficulty":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetOD(parsed)
	}
}

func parseEvents(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "Background", "0":
		beatMap.Bg = strings.Replace(line[2], "\"", "", -1)
	case "Break", "2":
		beatMap.Pauses = append(beatMap.Pauses, objects.NewPause(line))
	}
}

func parseHitObjects(line []string, beatMap *BeatMap) {
	obj := objects.GetObject(line)

	if obj != nil {
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
	file, err := os.Open(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.File))
	defer file.Close()

	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))

	var currentSection string
	counter := 0

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
				parseEvents(arr, beatMap)
			}
		case "TimingPoints":
			if arr := tokenize(line, ","); len(arr) > 1 {
				beatMap.ParsePoint(line)
				counter++
			}
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				var time string

				objType, _ := strconv.ParseInt(arr[3], 10, 64)
				if (objType & objects.CIRCLE) > 0 {
					beatMap.Circles++
					time = arr[2]
				} else if (objType & objects.SPINNER) > 0 {
					beatMap.Spinners++
					time = arr[5]
				} else if (objType & objects.SLIDER) > 0 {
					beatMap.Sliders++
					time = arr[2]
				} else if (objType & objects.LONGNOTE) > 0 {
					beatMap.Sliders++
					time = strings.Split(arr[5], ":")[0]
				}
				timeI, _ := strconv.Atoi(time)

				beatMap.Length = bmath.MaxI(beatMap.Length, timeI)
			}
		}
	}

	//beatMap.LoadTimingPoints()

	file.Seek(0, 0)

	if beatMap.Name+beatMap.Artist+beatMap.Creator == "" || counter == 0 {
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

func ParseTimingPointsAndPauses(beatMap *BeatMap) {
	if len(beatMap.Timings.Points) > 0 {
		return
	}

	file, err := os.Open(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.File))
	defer file.Close()

	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))
	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "Events":
			if arr := tokenize(line, ","); len(arr) > 1 && (arr[0] == "2" || arr[0] == "Break") {
				beatMap.Pauses = append(beatMap.Pauses, objects.NewPause(arr))
			}
		case "TimingPoints":
			if arr := tokenize(line, ","); arr != nil {
				beatMap.ParsePoint(line)
			}
		}
	}
}

func ParseObjects(beatMap *BeatMap) {
	file, err := os.Open(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.File))
	defer file.Close()

	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))
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

	sort.Slice(beatMap.HitObjects, func(i, j int) bool {
		return beatMap.HitObjects[i].GetBasicData().StartTime < beatMap.HitObjects[j].GetBasicData().StartTime
	})

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

	for _, obj := range beatMap.HitObjects {
		obj.SetTiming(beatMap.Timings)
	}

	calculateStackLeniency(beatMap)
}
