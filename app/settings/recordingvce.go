package settings

import (
	"fmt"
	"strings"
)

type vce264Settings struct {
	RateControl       string `combo:"qp|Quality Preset,cbr|CBR"`
	Bitrate           string `showif:"RateControl=cbr"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality" showif:"RateControl=qp"`
	AdditionalOptions string
}

func (s *vce264Settings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = vce264Common(s.RateControl, s.Bitrate, s.Preset)
	if err != nil {
		return nil, err
	}

	ret2, err := vceCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type vce265Settings struct {
	RateControl       string `combo:"qp|Quality Preset,cbr|CBR"`
	Bitrate           string `showif:"RateControl=cbr"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality" showif:"RateControl=qp"`
	AdditionalOptions string
}

func (s *vce265Settings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = vce265Common(s.RateControl, s.Bitrate, s.Preset)
	if err != nil {
		return nil, err
	}

	ret2, err := vceCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func vce264Common(rateControl string, bitrate string, qp string) (ret []string, err error) {
	switch strings.ToLower(rateControl) {
	case "qp":
		if qp == "quality" {
			ret = append(ret, "-quality", "2", "-rc", "0")
		} else if qp == "balanced" {
			ret = append(ret, "-quality", "0", "-rc", "0")
		} else if qp == "speed" {
			ret = append(ret, "-quality", "1", "-rc", "0")
		} else {
			return nil, fmt.Errorf("Invalid preset")
		}
	case "cbr":
		ret = append(ret, "-rc", "1", "-b:v", bitrate)
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}
	return
}

func vce265Common(rateControl string, bitrate string, qp string) (ret []string, err error) {
	switch strings.ToLower(rateControl) {
	case "qp":
		if qp == "quality" {
			ret = append(ret, "-quality", "0", "-rc", "0")
		} else if qp == "balanced" {
			ret = append(ret, "-quality", "5", "-rc", "0")
		} else if qp == "speed" {
			ret = append(ret, "-quality", "10", "-rc", "0")
		} else {
			return nil, fmt.Errorf("Invalid preset")
		}
	case "cbr":
		ret = append(ret, "-rc", "1", "-b:v", bitrate)
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}
	return
}

func vceCommon2(preset string, additional string) (ret []string, err error) {

	ret = parseCustomOptions(ret, additional)

	return
}
