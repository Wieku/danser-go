package settings

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

var libx264Profiles = []string{
	"baseline",
	"main",
	"high",
	"high444",
}

var libx265Profiles = []string{
	"main",
	"main444-8",
}

var libxPresets = []string{
	"ultrafast",
	"superfast",
	"veryfast",
	"faster",
	"fast",
	"medium",
	"slow",
	"slower",
	"veryslow",
	"placebo",
}

type x264Settings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,crf|Constant Rate Factor (CRF)"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CRF               int    `string:"true" min:"0" max:"51" showif:"RateControl=crf"`
	Profile           string `combo:"baseline,main,high,high444"`
	Preset            string `combo:"ultrafast,superfast,veryfast,faster,fast,medium,slow,slower,veryslow,placebo"`
	AdditionalOptions string
}

func (s *x264Settings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = libxCommon(s.RateControl, s.Bitrate, s.CRF)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(libx264Profiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile:v", s.Profile)

	ret2, err := libxCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type x265Settings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,crf|Constant Rate Factor (CRF)"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CRF               int    `string:"true" min:"0" max:"51" showif:"RateControl=crf"`
	Profile           string `combo:"main,main444-8"`
	Preset            string `combo:"ultrafast,superfast,veryfast,faster,fast,medium,slow,slower,veryslow,placebo"`
	AdditionalOptions string
}

func (s *x265Settings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = libxCommon(s.RateControl, s.Bitrate, s.CRF)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(libx265Profiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile:v", s.Profile)

	ret2, err := libxCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func libxCommon(rateControl, bitrate string, crf int) (ret []string, err error) {
	switch strings.ToLower(rateControl) {
	case "vbr":
		ret = append(ret, "-b:v", bitrate)
	case "cbr":
		ret = append(ret, "-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-bufsize", bitrate)
	case "crf":
		if crf < 0 {
			return nil, fmt.Errorf("CRF parameter out of range [0-%d]", math.MaxInt)
		}

		ret = append(ret, "-crf", strconv.Itoa(crf))
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}

func libxCommon2(preset string, additional string) (ret []string, err error) {
	if !slices.Contains(libxPresets, preset) {
		return nil, fmt.Errorf("invalid preset: %s", preset)
	}

	ret = append(ret, "-preset", preset)

	ret = parseCustomOptions(ret, additional)

	return
}
