package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var amfProfiles = []string{
	"main",
	"high",
}

var amfPresets = []string{
	"speed",
	"balanced",
	"quality",
	"high_quality",
}

type h264AmfSettings struct {
	RateControl       string `combo:"cqp|Constant Quantization,cbr|Constant Bitrate,vbr_peak|Variable Bitrate,qvbr|Quality Variable Bitrate,hqvbr|High Quality Variable Bitrate,hqcbr|High Quality Constant Bitrate"`
	Bitrate           string `showif:"RateControl=cbr,vbr_peak,qvbr,hqvbr,hqcbr"`
	CQ                int    `string:"true" min:"-1" max:"51" showif:"RateControl=cqp,qvbr"`
	Profile           string `combo:"main|Main,high|High"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality"`
	AdditionalOptions string
}

func (s *h264AmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon("transcoding", s.RateControl, s.Bitrate, s.CQ, "h264_amf")
	if err != nil {
		return nil, err
	}

	if !slices.Contains(amfProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile:v", s.Profile)

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type hevcAmfSettings struct {
	RateControl       string `combo:"cqp|Constant Quantization,cbr|Constant Bitrate,vbr_peak|Variable Bitrate,qvbr|Quality Variable Bitrate,hqvbr|High Quality Variable Bitrate,hqcbr|High Quality Constant Bitrate"`
	Bitrate           string `showif:"RateControl=cbr,vbr_peak,qvbr,hqvbr,hqcbr"`
	CQ                int    `string:"true" min:"-1" max:"51" showif:"RateControl=cqp,qvbr"`
	Profile           string `combo:"main|Main,high|High"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality"`
	AdditionalOptions string
}

func (s *hevcAmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon("transcoding", s.RateControl, s.Bitrate, s.CQ, "hevc_amf")
	if err != nil {
		return nil, err
	}

	if !slices.Contains(amfProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile tier: %s", s.Profile)
	}

	ret = append(ret, "-profile_tier:v", s.Profile)

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type av1AmfSettings struct {
	RateControl       string `combo:"cqp|Constant Quantization,cbr|Constant Bitrate,vbr_peak|Variable Bitrate,qvbr|Quality Variable Bitrate,hqvbr|High Quality Variable Bitrate,hqcbr|High Quality Constant Bitrate"`
	Bitrate           string `showif:"RateControl=cbr,vbr_peak,vbr_latency,qvbr,hqvbr,hqcbr"`
	CQ                int    `string:"true" min:"-1" max:"51" showif:"RateControl=cqp,qvbr"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality,high_quality|High Quality"`
	AdditionalOptions string
}

func (s *av1AmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon("transcoding", s.RateControl, s.Bitrate, s.CQ, "av1_amf")
	if err != nil {
		return nil, err
	}

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func amfCommon(usage, rateControl, bitrate string, cq int, encoderType string) (ret []string, err error) {
	ret = append(ret, "-usage", usage)

	switch strings.ToLower(rateControl) {
	case "cqp":
		// The AV1 encoder allows values from -1 to 255 in the '-qp_i' and 'qp_p' arguments, but the
		// '-qvbr_quality_level' argument (which also uses the 'cq' variable) allows values from -1 to 51.
		// So, '-qp_i' and 'qp_p' are artificially capped here at 51 as well.
		if cq < -1 || cq > 51 {
			return nil, fmt.Errorf("CQ parameter out of range [-1-51]")
		}

		switch encoderType {
		case "h264_amf":
			ret = append(ret, "-rc", "cqp", "-qp_i", strconv.Itoa(cq), "-qp_p", strconv.Itoa(cq), "-qp_b", strconv.Itoa(cq))
		case "hevc_amf", "av1_amf":
			ret = append(ret, "-rc", "cqp", "-qp_i", strconv.Itoa(cq), "-qp_p", strconv.Itoa(cq))
		}
	case "cbr":
		ret = append(ret, "-rc", "cbr", "-b:v", bitrate)
	case "vbr_peak":
		ret = append(ret, "-rc", "vbr_peak", "-b:v", bitrate)
	case "qvbr":
		ret = append(ret, "-rc", "qvbr", "-b:v", bitrate, "-qvbr_quality_level", strconv.Itoa(cq))
	case "hqvbr":
		ret = append(ret, "-rc", "hqvbr", "-b:v", bitrate)
	case "hqcbr":
		ret = append(ret, "-rc", "hqcbr", "-b:v", bitrate)
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}

func amfCommon2(preset string, additional string) (ret []string, err error) {
	if !slices.Contains(amfPresets, preset) {
		return nil, fmt.Errorf("invalid preset: %s", preset)
	}

	ret = append(ret, "-preset:v", preset)
	ret = parseCustomOptions(ret, additional)

	return
}
