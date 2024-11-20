package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var amfProfiles = []string{
	"main",
}

var nvencPresets = []string{
  "speed",
  "balanced",
  "quality",
}

type h264AmfSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,cqp|Constant Frame Compression (CQP),cq|Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CQ                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp,cq"`
	Profile           string `combo:"main"`
	Preset            string `combo:"speed,balanced,quality"`
	AdditionalOptions string
}

func (s *h264AmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon(s.RateControl, s.Bitrate, s.CQ)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(amfProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile", s.Profile)

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type hevcAmfSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,cqp|Constant Frame Compression (CQP),cq|Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CQ                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp,cq"`
	Preset            string `combo:"speed, balanced, quality"`
	AdditionalOptions string
}

func (s *hevcAmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = nvencCommon(s.RateControl, s.Bitrate, s.CQ)
	if err != nil {
		return nil, err
	}

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type av1AmfSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,cqp|Constant Frame Compression (CQP),cq|Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CQ                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp,cq"`
	Preset            string `combo:"speed, balanced, quality"`
	AdditionalOptions string
}

func (s *av1AmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon(s.RateControl, s.Bitrate, s.CQ)
	if err != nil {
		return nil, err
	}

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func amfCommon(rateControl, bitrate string, cq int) (ret []string, err error) {
	switch strings.ToLower(rateControl) {
	case "vbr":
		ret = append(ret, "-rc", "vbr", "-b:v", bitrate)
	case "cbr":
		ret = append(ret, "-rc", "cbr", "-b:v", bitrate)
	case "cqp":
		if cq < 0 || cq > 51 {
			return nil, fmt.Errorf("CQ parameter out of range [0-51]")
		}

		ret = append(ret, "-rc", "cqp", "-qp_i", strconv.Itoa(cq), "-qp_p", strconv.Itoa(cq), "-qp_b", strconv.Itoa(cq))
	case "cq":
		if cq < 0 || cq > 51 {
			return nil, fmt.Errorf("CQ parameter out of range [0-51]")
		}

		ret = append(ret, "-rc", "cq", "-cq", strconv.Itoa(cq))
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}

func amfCommon2(preset string, additional string) (ret []string, err error) {
	if !slices.Contains(amfPresets, preset) {
		return nil, fmt.Errorf("invalid preset: %s", preset)
	}

	ret = append(ret, "-preset", preset)

	ret = parseCustomOptions(ret, additional)

	return
}
