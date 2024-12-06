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
	"constrained_baseline",
	"constrained_high",
}

var amfPresets = []string{
	"speed",
	"balanced",
	"quality",
	"high_quality",
}

type h264AmfSettings struct {
	RateControl       string `combo:"cqp|Constant Quantization,cbr|Constant Bitrate,vbr_peak|Constrained Variable Bitrate,vbr_latency|Latency Constrained Variable Bitrate,qvbr|Quality Variable Bitrate,hqvbr|High Quality Variable Bitrate,hqcbr|High Quality Constant Bitrate"`
	Bitrate           string `showif:"RateControl=cbr,vbr_peak,vbr_latency,qvbr,hqvbr,hqcbr"`
	CQ                int    `string:"true" min:"-1" max:"51" showif:"RateControl=cqp,qvbr"`
	Profile           string `combo:"main|Main,high|High,constrained_baseline|Constrained Baseline,constrained_high|Constrained High"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality"`
	Usage             string `combo:"ultralowlatency|Ultra Low Latency,lowlatency|Low Latency,lowlatency_high_quality|Low Latency High Quality,transcoding|Transcoding,high_quality|High Quality" tooltip:"'Transcoding' is preferred over 'High Quality' as it provides good quality without sacrifcing encoding speed. 'Low Latency' variants are almost on par with 'Transcoding'."`
	AdditionalOptions string
}

func (s *h264AmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon(s.Usage, s.RateControl, s.Bitrate, s.CQ, "h264_amf")
	if err != nil {
		return nil, err
	}

	if !slices.Contains(amfProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	switch s.Profile {
	case "main":
		ret = append(ret, "-profile:v", "77")
	case "high":
		ret = append(ret, "-profile:v", "100")
	case "constrained_baseline":
		ret = append(ret, "-profile:v", "256")
	case "constrained_high":
		ret = append(ret, "-profile:v", "257")
	}

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type hevcAmfSettings struct {
	RateControl       string `combo:"cqp|Constant Quantization,cbr|Constant Bitrate,vbr_peak|Constrained Variable Bitrate,vbr_latency|Latency Constrained Variable Bitrate,qvbr|Quality Variable Bitrate,hqvbr|High Quality Variable Bitrate,hqcbr|High Quality Constant Bitrate"`
	Bitrate           string `showif:"RateControl=cbr,vbr_peak,vbr_latency,qvbr,hqvbr,hqcbr"`
	CQ                int    `string:"true" min:"-1" max:"51" showif:"RateControl=cqp,qvbr"`
	Profile           string `combo:"main|Main,high|High"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality"`
	Usage             string `combo:"ultralowlatency|Ultra Low Latency,lowlatency|Low Latency,lowlatency_high_quality|Low Latency High Quality,transcoding|Transcoding,high_quality|High Quality" tooltip:"'Transcoding' is preferred over 'High Quality' as it provides good quality without sacrifcing encoding speed. 'Low Latency' variants are almost on par with 'Transcoding'."`
	AdditionalOptions string
}

func (s *hevcAmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon(s.Usage, s.RateControl, s.Bitrate, s.CQ, "hevc_amf")
	if err != nil {
		return nil, err
	}

	if !slices.Contains(amfProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	switch s.Profile {
	case "main":
		ret = append(ret, "-profile_tier:v", "0")
	case "high":
		ret = append(ret, "-profile_tier:v", "1")
	}

	ret2, err := amfCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type av1AmfSettings struct {
	RateControl       string `combo:"cqp|Constant Quantization,cbr|Constant Bitrate,vbr_peak|Constrained Variable Bitrate,vbr_latency|Latency Constrained Variable Bitrate,qvbr|Quality Variable Bitrate,hqvbr|High Quality Variable Bitrate,hqcbr|High Quality Constant Bitrate"`
	Bitrate           string `showif:"RateControl=cbr,vbr_peak,vbr_latency,qvbr,hqvbr,hqcbr"`
	CQ                int    `string:"true" min:"-1" max:"51" showif:"RateControl=cqp,qvbr"`
	Preset            string `combo:"speed|Speed,balanced|Balanced,quality|Quality,high_quality|High Quality"`
	Usage             string `combo:"ultralowlatency|Ultra Low Latency,lowlatency|Low Latency,lowlatency_high_quality|Low Latency High Quality,transcoding|Transcoding,high_quality|High Quality" tooltip:"'Transcoding' is preferred over 'High Quality' as it provides good quality without sacrifcing encoding speed. 'Low Latency' variants are almost on par with 'Transcoding'."`
	AdditionalOptions string
}

func (s *av1AmfSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = amfCommon(s.Usage, s.RateControl, s.Bitrate, s.CQ, "av1_amf")
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
		// So, '-qp_i' and 'qp_p' are capped at 51 as well.
		if cq < -1 || cq > 51 {
			return nil, fmt.Errorf("CQ parameter out of range [-1-51]")
		}

		switch encoderType {
		case "h264_amf":
			ret = append(ret, "-rc", "0", "-qp_i", strconv.Itoa(cq), "-qp_p", strconv.Itoa(cq), "-qp_b", strconv.Itoa(cq))
		case "hevc_amf", "av1_amf":
			ret = append(ret, "-rc", "0", "-qp_i", strconv.Itoa(cq), "-qp_p", strconv.Itoa(cq))
		}
	case "cbr":
		// The AMF rate control values for CBR and VBR low latency are inversely mapped.
		// For H264: CBR = 1, VBR low latency = 3
		// For HEVC/AV1: CBR = 3, VBR low latency = 1
		switch encoderType {
		case "h264_amf":
			ret = append(ret, "-rc", "1", "-b:v", bitrate)
		case "hevc_amf", "av1_amf":
			ret = append(ret, "-rc", "3", "-b:v", bitrate)
		}
	case "vbr_peak":
		ret = append(ret, "-rc", "2", "-b:v", bitrate)
	case "vbr_latency":
		switch encoderType {
		case "h264_amf":
			ret = append(ret, "-rc", "3", "-b:v", bitrate)
		case "hevc_amf", "av1_amf":
			ret = append(ret, "-rc", "1", "-b:v", bitrate)
		}
	case "qvbr":
		ret = append(ret, "-rc", "4", "-b:v", bitrate, "-qvbr_quality_level", strconv.Itoa(cq))
	case "hqvbr":
		ret = append(ret, "-rc", "5", "-b:v", bitrate)
	case "hqcbr":
		ret = append(ret, "-rc", "6", "-b:v", bitrate)
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
