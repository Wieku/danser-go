package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var nvencProfiles = []string{
	"baseline",
	"main",
	"high",
}

var nvencPresets = []string{
	"slow",
	"medium",
	"fast",
	"p1",
	"p2",
	"p3",
	"p4",
	"p5",
	"p6",
	"p7",
}

type h264NvencSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,cqp|Constant Frame Compression (CQP),cq|Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CQ                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp,cq"`
	Profile           string `combo:"baseline,main,high"`
	Preset            string `combo:"fast|fast (legacy),medium|medium (legacy),slow|slow (legacy),p1|fastest (p1),p2|faster (p2),p3|fast (p3),p4|medium (p4),p5|slow (p5),p6|slower (p6),p7|slowest (p7)"`
	AdditionalOptions string
}

func (s *h264NvencSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = nvencCommon(s.RateControl, s.Bitrate, s.CQ)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(nvencProfiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile", s.Profile)

	ret2, err := nvencCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type hevcNvencSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,cqp|Constant Frame Compression (CQP),cq|Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CQ                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp,cq"`
	Preset            string `combo:"fast|fast (legacy),medium|medium (legacy),slow|slow (legacy),p1|fastest (p1),p2|faster (p2),p3|fast (p3),p4|medium (p4),p5|slow (p5),p6|slower (p6),p7|slowest (p7)"`
	AdditionalOptions string
}

func (s *hevcNvencSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = nvencCommon(s.RateControl, s.Bitrate, s.CQ)
	if err != nil {
		return nil, err
	}

	ret2, err := nvencCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

type av1NvencSettings struct {
	RateControl       string `combo:"vbr|VBR,cbr|CBR,cqp|Constant Frame Compression (CQP),cq|Constant Quality"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CQ                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp,cq"`
	Preset            string `combo:"fast|fast (legacy),medium|medium (legacy),slow|slow (legacy),p1|fastest (p1),p2|faster (p2),p3|fast (p3),p4|medium (p4),p5|slow (p5),p6|slower (p6),p7|slowest (p7)"`
	AdditionalOptions string
}

func (s *av1NvencSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = nvencCommon(s.RateControl, s.Bitrate, s.CQ)
	if err != nil {
		return nil, err
	}

	ret2, err := nvencCommon2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func nvencCommon(rateControl, bitrate string, cq int) (ret []string, err error) {
	switch strings.ToLower(rateControl) {
	case "vbr":
		ret = append(ret, "-rc", "vbr", "-b:v", bitrate)
	case "cbr":
		ret = append(ret, "-rc", "cbr", "-b:v", bitrate)
	case "cqp":
		if cq < 0 || cq > 51 {
			return nil, fmt.Errorf("CQ parameter out of range [0-51]")
		}

		ret = append(ret, "-rc", "constqp", "-qp", strconv.Itoa(cq))
	case "cq":
		if cq < 0 || cq > 51 {
			return nil, fmt.Errorf("CQ parameter out of range [0-51]")
		}

		ret = append(ret, "-rc", "vbr", "-b:v", "400M", "-cq", strconv.Itoa(cq))
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}

func nvencCommon2(preset string, additional string) (ret []string, err error) {
	if !slices.Contains(nvencPresets, preset) {
		return nil, fmt.Errorf("invalid preset: %s", preset)
	}

	ret = append(ret, "-preset", preset)

	ret = parseCustomOptions(ret, additional)

	return
}
