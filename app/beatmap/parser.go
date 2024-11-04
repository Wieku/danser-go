package beatmap

import (
	"cmp"
	"errors"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
)

const bufferSize = 10 * 1024 * 1024

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
		case "Soft", "None":
			beatMap.Timings.BaseSet = 2
		case "Drum":
			beatMap.Timings.BaseSet = 3
		}
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
	case "BeatmapID":
		beatMap.ID, _ = strconv.ParseInt(line[1], 10, 64)
	case "BeatmapSetID":
		beatMap.SetID, _ = strconv.ParseInt(line[1], 10, 64)
	}
}

func parseDifficulty(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "SliderMultiplier":
		beatMap.SliderMultiplier, _ = strconv.ParseFloat(line[1], 64)
		beatMap.Timings.SliderMult = beatMap.SliderMultiplier
	case "ApproachRate":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetAR(mutils.Clamp(parsed, 0, 10))
		beatMap.ARSpecified = true
	case "CircleSize":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetCS(mutils.Clamp(parsed, 0, 10))
	case "SliderTickRate":
		beatMap.Timings.TickRate, _ = strconv.ParseFloat(line[1], 64)
	case "HPDrainRate":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetHP(mutils.Clamp(parsed, 0, 10))
	case "OverallDifficulty":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatMap.Diff.SetOD(mutils.Clamp(parsed, 0, 10))

		if !beatMap.ARSpecified {
			beatMap.Diff.SetAR(beatMap.Diff.GetOD())
		}
	}
}

func parseEvents(line []string, beatMap *BeatMap) {
	switch line[0] {
	case "Background", "0":
		beatMap.Bg = strings.Replace(line[2], "\"", "", -1)
	case "Break", "2":
		beatMap.Pauses = append(beatMap.Pauses, NewPause(line))
	}
}

func parseHitObjects(line []string, beatMap *BeatMap) {
	obj := objects.CreateObject(line)

	if obj != nil {
		beatMap.HitObjects = append(beatMap.HitObjects, obj)
	}
}

func tokenize(line, delimiter string) []string {
	return tokenizeN(line, delimiter, -1)
}

func tokenizeN(line, delimiter string, n int) []string {
	if strings.HasPrefix(line, "//") || !strings.Contains(line, delimiter) {
		return nil
	}

	divided := strings.SplitN(line, delimiter, n)

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

var bufferPool = &sync.Pool{
	New: func() any {
		buf := make([]byte, bufferSize)
		return &buf
	},
}

func ParseBeatMap(beatMap *BeatMap) error {
	file, err := os.Open(filepath.Join(settings.General.GetSongsDir(), beatMap.Dir, beatMap.File))
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := files.NewScanner(file)

	buf := bufferPool.Get().(*[]byte)
	scanner.Buffer(*buf, cap(*buf))

	defer bufferPool.Put(buf)

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
			if arr := tokenizeN(line, ":", 2); len(arr) > 1 {
				if err := parseGeneral(arr, beatMap); err {
					return errors.New("wrong mode")
				}
			}
		case "Metadata":
			if arr := tokenizeN(line, ":", 2); len(arr) > 1 {
				parseMetadata(arr, beatMap)
			}
		case "Difficulty":
			if arr := tokenizeN(line, ":", 2); len(arr) > 1 {
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

				objTypeI, _ := strconv.Atoi(arr[3])
				objType := objects.Type(objTypeI)
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

				beatMap.Length = max(beatMap.Length, timeI)
			}
		}
	}

	beatMap.FinalizePoints()

	file.Seek(0, 0)

	if beatMap.Name+beatMap.Artist+beatMap.Creator == "" || counter == 0 {
		return errors.New("corrupted file")
	}

	return nil
}

func ParseBeatMapFile(file *os.File) *BeatMap {
	beatMap := NewBeatMap()
	beatMap.Dir, _ = filepath.Rel(settings.General.GetSongsDir(), filepath.Dir(file.Name()))
	beatMap.Dir = filepath.ToSlash(beatMap.Dir)

	f, _ := file.Stat()
	beatMap.File = f.Name()

	err := ParseBeatMap(beatMap)

	if err != nil {
		return nil
	}

	return beatMap
}

func ParseTimingPointsAndPauses(beatMap *BeatMap) {
	if beatMap.Timings.HasPoints() {
		return
	}

	file, err := os.Open(filepath.Join(settings.General.GetSongsDir(), beatMap.Dir, beatMap.File))
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := files.NewScanner(file)

	buf := bufferPool.Get().(*[]byte)
	scanner.Buffer(*buf, cap(*buf))

	defer bufferPool.Put(buf)

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
				beatMap.Pauses = append(beatMap.Pauses, NewPause(arr))
			}
		case "TimingPoints":
			if arr := tokenize(line, ","); arr != nil {
				beatMap.ParsePoint(line)
			}
		}
	}

	beatMap.FinalizePoints()
}

func ParseObjects(beatMap *BeatMap, diffCalcOnly, parseColors bool) {
	file, err := os.Open(filepath.Join(settings.General.GetSongsDir(), beatMap.Dir, beatMap.File))
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := files.NewScanner(file)

	buf := bufferPool.Get().(*[]byte)
	scanner.Buffer(*buf, cap(*buf))

	defer bufferPool.Put(buf)

	var currentSection string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "osu file format v") {
			trim := strings.TrimPrefix(line, "osu file format v")
			beatMap.Version, _ = strconv.Atoi(trim)
		}

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "Colours": //nolint:misspell
			if parseColors {
				if arr := tokenize(line, ":"); arr != nil {
					skin.AddBeatmapColor(arr)
				}
			}
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				parseHitObjects(arr, beatMap)
			}
		}
	}

	slices.SortStableFunc(beatMap.HitObjects, func(a, b objects.IHitObject) int {
		return cmp.Compare(a.GetStartTime(), b.GetStartTime())
	})

	if parseColors {
		skin.FinishBeatmapColors()
	}

	num := 0
	comboNumber := 1
	comboSet := 0
	comboSetHax := 0
	forceNewCombo := false

	for i, iO := range beatMap.HitObjects {
		if iO.GetType() == objects.SPINNER {
			forceNewCombo = true
		} else if iO.IsNewCombo() || forceNewCombo {
			iO.SetNewCombo(true)
			comboNumber = 1
			comboSet++
			comboSetHax += int(iO.GetColorOffset()) + 1

			forceNewCombo = false
		}

		if iO.IsNewCombo() && i > 0 {
			beatMap.HitObjects[i-1].SetLastInCombo(true)
		}

		iO.SetID(int64(num))
		iO.SetComboNumber(int64(comboNumber))
		iO.SetComboSet(int64(comboSet))
		iO.SetComboSetHax(int64(comboSetHax))
		iO.SetStackLeniency(beatMap.StackLeniency)

		comboNumber++
		num++
	}

	for _, obj := range beatMap.HitObjects {
		obj.SetTiming(beatMap.Timings, beatMap.Version, diffCalcOnly)
	}

	if settings.Objects.StackEnabled || settings.KNOCKOUT || settings.PLAY || diffCalcOnly {
		beatMap.CalculateStackLeniency(beatMap.Diff)
	}
}
