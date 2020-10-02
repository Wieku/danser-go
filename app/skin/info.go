package skin

import (
	"bufio"
	"fmt"
	"github.com/wieku/danser-go/framework/math/color"
	"os"
	"strconv"
	"strings"
)

const latestVersion = 2.7

type SkinInfo struct {
	Name   string
	Author string

	Version float64

	AnimationFramerate float64

	SpinnerFadePlayfield bool
	SpinnerNoBlink       bool

	//skipping combo bursts

	//skipping cursor settings for now
	//renderer for old cursor trails may come in the future

	ComboColors []color.Color

	//slider style unnecessary

	SliderBallTint      bool
	SliderBallFlip      bool
	SliderBorder        color.Color
	SliderTrackOverride *color.Color

	InputOverlayText color.Color

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
		Name:                        "",
		Author:                      "",
		Version:                     2.7,
		AnimationFramerate:          -1,
		SpinnerFadePlayfield:        true,
		SpinnerNoBlink:              false,
		ComboColors:                 []color.Color{},
		SliderBallTint:              false,
		SliderBallFlip:              false,
		SliderBorder:                color.NewL(1),
		SliderTrackOverride:         nil,
		InputOverlayText:            color.NewL(1),
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
		panic(fmt.Sprintf("Error while parsing %s: %s", errType, text))
	}

	return value
}

func ParseColor(text, errType string) color.Color {
	color := color.NewL(1)

	divided := strings.Split(text, ",")
	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}

	for i, v := range divided {
		switch i {
		case 0:
			color.R = float32(ParseFloat(v, errType+".R")) / 255
		case 1:
			color.G = float32(ParseFloat(v, errType+".G")) / 255
		case 2:
			color.B = float32(ParseFloat(v, errType+".B")) / 255
		case 3:
			color.A = float32(ParseFloat(v, errType+".A")) / 255
		}
	}

	return color
}

func LoadInfo(path string) *SkinInfo {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	info := newDefaultInfo()

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
		case "AnimationFramerate":
			info.AnimationFramerate = ParseFloat(tokenized[1], tokenized[0])
		case "SpinnerFadePlayfield":
			info.SpinnerFadePlayfield = tokenized[1] == "1"
		case "SpinnerNoBlink":
			info.SpinnerNoBlink = tokenized[1] == "1"
		case "Combo1", "Combo2", "Combo3", "Combo4", "Combo5", "Combo6", "Combo7", "Combo8":
			info.ComboColors = append(info.ComboColors, ParseColor(tokenized[1], tokenized[0]))
		case "SliderBallTint":
			info.SliderBallTint = tokenized[1] == "1"
		case "SliderBallFlip":
			info.SliderBallFlip = tokenized[1] == "1"
		case "SliderBorder":
			info.SliderBorder = ParseColor(tokenized[1], tokenized[0])
		case "SliderTrackOverride":
			col := ParseColor(tokenized[1], tokenized[0])
			info.SliderTrackOverride = &col
		case "InputOverlayText":
			info.InputOverlayText = ParseColor(tokenized[1], tokenized[0])
		case "HitCirclePrefix":
			info.HitCirclePrefix = tokenized[1]
		case "HitCircleOverlap":
			info.HitCircleOverlap = ParseFloat(tokenized[1], tokenized[0])
		case "HitCircleOverlayAboveNumber", "HitCircleOverlayAboveNumer":
			info.HitCircleOverlayAboveNumber = tokenized[1] == "1"
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

	return info
}
