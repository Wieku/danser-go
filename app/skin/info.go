package skin

import (
	"fmt"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/math/color"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

const latestVersion = 2.7

type colorI struct {
	index int
	color color.Color
}

type SkinInfo struct {
	Name   string
	Author string

	Version float64

	AnimationFramerate float64

	SpinnerFadePlayfield     bool
	SpinnerNoBlink           bool
	SpinnerFrequencyModulate bool

	LayeredHitSounds bool

	//skipping combo bursts

	CursorCentre bool
	CursorExpand bool
	CursorRotate bool

	ComboColors []color.Color

	DefaultSkinFollowpointBehavior bool

	//slider style unnecessary

	SliderBallTint      bool
	SliderBallFlip      bool
	SliderBorder        color.Color
	SliderTrackOverride *color.Color
	SliderBall          *color.Color

	SongSelectInactiveText color.Color
	SongSelectActiveText   color.Color
	InputOverlayText       color.Color

	//hit circle font settings
	HitCirclePrefix             string
	HitCircleOverlap            float64
	HitCircleOverlayAboveNumber bool

	//score font settings
	ScorePrefix  string
	ScoreOverlap float64

	//combo font settings
	ComboPrefix  string
	ComboOverlap float64
}

func newDefaultInfo() *SkinInfo {
	return &SkinInfo{
		Name:                     "",
		Author:                   "",
		Version:                  latestVersion,
		AnimationFramerate:       -1,
		SpinnerFadePlayfield:     true,
		SpinnerNoBlink:           false,
		SpinnerFrequencyModulate: true,
		LayeredHitSounds:         true,
		CursorCentre:             true,
		CursorExpand:             true,
		CursorRotate:             true,
		ComboColors: []color.Color{
			color.NewIRGB(255, 192, 0),
			color.NewIRGB(0, 202, 0),
			color.NewIRGB(18, 124, 255),
			color.NewIRGB(242, 24, 57),
		},
		SliderBallTint:              false,
		SliderBallFlip:              false,
		SliderBorder:                color.NewL(1),
		SliderTrackOverride:         nil,
		SongSelectInactiveText:      color.NewL(1),
		SongSelectActiveText:        color.NewL(0),
		InputOverlayText:            color.NewL(0),
		HitCirclePrefix:             "default",
		HitCircleOverlap:            -2,
		HitCircleOverlayAboveNumber: false,
		ScorePrefix:                 "score",
		ScoreOverlap:                0,
		ComboPrefix:                 "score",
		ComboOverlap:                0,
	}
}

func (info *SkinInfo) GetFrameTime(frames int) float64 {
	if info.AnimationFramerate > 0 {
		return 1000.0 / info.AnimationFramerate
	}

	return 1000.0 / float64(frames)
}

func tokenize(line, delimiter string) []string {
	line = strings.TrimSpace(line)

	if index := strings.Index(line, "//"); index > -1 {
		line = line[:index]
	}

	if strings.HasPrefix(line, "//") || !strings.Contains(line, delimiter) {
		return nil
	}

	divided := strings.Split(line, delimiter)
	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}

	return divided
}

func ParseFloat(text, errType string) float64 {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(fmt.Sprintf("error while parsing %s: %s", errType, text))
	}

	return value
}

func ParseBool(text, errType string, target *bool) bool {
	if trueVal := strings.EqualFold(text, "true"); trueVal || strings.EqualFold(text, "false") { // We can't use strconv.ParseBool since it won't match osu!'s/.net behavior
		*target = trueVal
	} else {
		value, err := strconv.Atoi(text)
		if err != nil {
			log.Println(fmt.Sprintf("SkinManager: Unable to parse %s: %s. Using default value", errType, text))
		} else {
			*target = value != 0
		}
	}

	return *target
}

func ParseColor(text, errType string) color.Color {
	clr := color.NewL(1)

	divided := strings.Split(text, ",")
	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}

	for i, v := range divided {
		switch i {
		case 0:
			clr.R = float32(ParseFloat(v, errType+".R")) / 255
		case 1:
			clr.G = float32(ParseFloat(v, errType+".G")) / 255
		case 2:
			clr.B = float32(ParseFloat(v, errType+".B")) / 255
			//case 3:
			//	clr.A = float32(ParseFloat(v, errType+".A")) / 255
		}
	}

	return clr
}

func LoadInfo(path string, local bool) (*SkinInfo, error) {
	var file io.ReadCloser
	var err error

	if local {
		file, err = assets.Open(path)
	} else {
		file, err = os.Open(path)
	}

	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := files.NewScanner(file)

	info := newDefaultInfo()

	versionPresent := false

	colorsI := make([]colorI, 0)

	for scanner.Scan() {
		line := scanner.Text()

		tokenized := tokenize(line, ":")

		if tokenized == nil {
			continue
		}

		switch tokenized[0] {
		case "Name":
			info.Name = tokenized[1]
		case "Author":
			info.Author = tokenized[1]
		case "Version":
			if tokenized[1] == "latest" {
				info.Version = latestVersion
			} else {
				info.Version = ParseFloat(tokenized[1], tokenized[0])
			}

			versionPresent = true
		case "AnimationFramerate":
			info.AnimationFramerate = ParseFloat(tokenized[1], tokenized[0])
		case "SpinnerFadePlayfield":
			ParseBool(tokenized[1], tokenized[0], &info.SpinnerFadePlayfield)
		case "SpinnerNoBlink":
			ParseBool(tokenized[1], tokenized[0], &info.SpinnerNoBlink)
		case "SpinnerFrequencyModulate":
			ParseBool(tokenized[1], tokenized[0], &info.SpinnerFrequencyModulate)
		case "LayeredHitSounds":
			ParseBool(tokenized[1], tokenized[0], &info.LayeredHitSounds)
		case "CursorCentre":
			ParseBool(tokenized[1], tokenized[0], &info.CursorCentre)
		case "CursorExpand":
			ParseBool(tokenized[1], tokenized[0], &info.CursorExpand)
		case "CursorRotate":
			ParseBool(tokenized[1], tokenized[0], &info.CursorRotate)
		case "Combo1", "Combo2", "Combo3", "Combo4", "Combo5", "Combo6", "Combo7", "Combo8":
			index, _ := strconv.ParseInt(strings.TrimPrefix(tokenized[0], "Combo"), 10, 64)
			colorsI = append(colorsI, colorI{
				index: int(index),
				color: ParseColor(tokenized[1], tokenized[0]),
			})
		case "DefaultSkinFollowpointBehavior":
			ParseBool(tokenized[1], tokenized[0], &info.DefaultSkinFollowpointBehavior)
		case "AllowSliderBallTint":
			ParseBool(tokenized[1], tokenized[0], &info.SliderBallTint)
		case "SliderBallFlip":
			ParseBool(tokenized[1], tokenized[0], &info.SliderBallFlip)
		case "SliderBorder":
			info.SliderBorder = ParseColor(tokenized[1], tokenized[0])
		case "SliderTrackOverride":
			col := ParseColor(tokenized[1], tokenized[0])
			info.SliderTrackOverride = &col
		case "SliderBall":
			col := ParseColor(tokenized[1], tokenized[0])
			info.SliderBall = &col
		case "SongSelectInactiveText":
			info.SongSelectInactiveText = ParseColor(tokenized[1], tokenized[0])
		case "SongSelectActiveText":
			info.SongSelectActiveText = ParseColor(tokenized[1], tokenized[0])
		case "InputOverlayText":
			info.InputOverlayText = ParseColor(tokenized[1], tokenized[0])
		case "HitCirclePrefix":
			info.HitCirclePrefix = tokenized[1]
		case "HitCircleOverlap":
			info.HitCircleOverlap = ParseFloat(tokenized[1], tokenized[0])
		case "HitCircleOverlayAboveNumber", "HitCircleOverlayAboveNumer":
			ParseBool(tokenized[1], tokenized[0], &info.HitCircleOverlayAboveNumber)
		case "ScorePrefix":
			info.ScorePrefix = tokenized[1]
		case "ScoreOverlap":
			info.ScoreOverlap = ParseFloat(tokenized[1], tokenized[0])
		case "ComboPrefix":
			info.ComboPrefix = tokenized[1]
		case "ComboOverlap":
			info.ComboOverlap = ParseFloat(tokenized[1], tokenized[0])
		}
	}

	if !versionPresent {
		info.Version = 1.0
	}

	if len(colorsI) > 0 {
		sort.SliceStable(colorsI, func(i, j int) bool {
			return colorsI[i].index <= colorsI[j].index
		})

		info.ComboColors = make([]color.Color, 0)

		for _, c := range colorsI {
			info.ComboColors = append(info.ComboColors, c.color)
		}
	}

	return info, nil
}
