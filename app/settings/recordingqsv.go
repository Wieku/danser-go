package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var qsvProfiles = []string{
	"baseline",
	"main",
	"high",
}

var qsvPresets = []string{
	"veryfast",
	"faster",
	"fast",
	"medium",
	"slow",
	"slower",
	"veryslow",
}

type h264QSVSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,icq|Intelligent Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	Quality           int    `string:"true" min:"1" max:"51" showif:"RateControl=icq"`
	Profile           string `combo:"baseline,main,high"`
	Preset            string `combo:"veryfast,faster,fast,medium,slow,slower,veryslow"`
	AdditionalOptions string
}

func (s *h264QSVSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = qsvCommon(s.RateControl, s.Bitrate, s.Quality)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(qsvProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile", s.Profile)

	ret2, err := qsvCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type hevcQSVSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,icq|Intelligent Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	Quality           int    `string:"true" min:"1" max:"51" showif:"RateControl=icq"`
	Preset            string `combo:"veryfast,faster,fast,medium,slow,slower,veryslow"`
	AdditionalOptions string
}

func (s *hevcQSVSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = qsvCommon(s.RateControl, s.Bitrate, s.Quality)
	if err != nil {
		return nil, err
	}

	ret2, err := qsvCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func qsvCommon(rateControl, bitrate string, quality int) (ret []string, err error) {
	switch strings.ToLower(rateControl) {
	case "vbr":
		ret = append(ret, "-b:v", bitrate)
	case "cbr":
		ret = append(ret, "-b:v", bitrate, "-maxrate", bitrate)
	case "icq":
		if quality < 1 || quality > 51 {
			return nil, fmt.Errorf("Quality parameter out of range [0-51]")
		}

		ret = append(ret, "-global_quality", strconv.Itoa(quality))
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}

func qsvCommon2(preset string, additional string) (ret []string, err error) {
	if !slices.Contains(qsvPresets, preset) {
		return nil, fmt.Errorf("invalid preset: %s", preset)
	}

	ret = append(ret, "-preset", preset)

	ret = parseCustomOptions(ret, additional)

	return
}
