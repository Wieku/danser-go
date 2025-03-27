package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var av1Profiles = []string{
	"main",
}

var av1Presets = []string{
	"0",
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"8",
	"9",
	"10",
	"11",
	"12",
	"13",
}

type av1Settings struct {
	RateControl       string `combo:"crf|Constant Rate Factor (CRF),vbr|VBR,cbr|CBR"`
	Bitrate           string `showif:"RateControl=vbr,cbr"`
	CRF               int    `string:"true" min:"0" max:"63" showif:"RateControl=crf"`
	Profile           string `combo:"main|Main"`
	Preset            string `combo:"0|0 (slowest),1,2,3,4,5,6,7,8,9,10,11,12,13|13 (fastest)"`
	AdditionalOptions string
}

func (s *av1Settings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = av1Common(s.RateControl, s.Bitrate, s.Profile, s.CRF)
	if err != nil {
		return nil, err
	}

	ret2, err := av1Common2(s.Preset, s.AdditionalOptions)
	if err != nil {
		return nil, err
	}

	return append(ret, ret2...), nil
}

func av1Common(rateControl, bitrate, profile string, crf int) (ret []string, err error) {
	if !slices.Contains(av1Profiles, profile) {
		return nil, fmt.Errorf("invalid profile: %s", profile)
	}

	switch strings.ToLower(rateControl) {
	case "vbr":
		// 'enable-overlays' improves quality of keyframes.
		ret = append(ret, "-svtav1-params", fmt.Sprintf("enable-overlays=1:rc=1:tbr=%s:profile=%s", bitrate, profile))
	case "cbr":
		// CBR is not suppoted by libsvtav1's random access prediction structure (SVT_AV1_PRED_RANDOM_ACCESS).
		// Low delay B (SVT_AV1_PRED_LOW_DELAY_B) will be used instead which is a bit worse in encoding speed
		// and does not support overlays but supports CBR.
		ret = append(ret, "-svtav1-params", fmt.Sprintf("pred-struct=1:rc=2:tbr=%s:profile=%s", bitrate, profile))
	case "crf":
		if crf < 0 || crf > 63 {
			return nil, fmt.Errorf("CRF parameter out of range [0-63]")
		}

		ret = append(ret, "-svtav1-params", "enable-overlays=1", "-crf", strconv.Itoa(crf))
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}

func av1Common2(preset string, additional string) (ret []string, err error) {
	if !slices.Contains(av1Presets, preset) {
		return nil, fmt.Errorf("invalid preset: %s", preset)
	}

	ret = append(ret, "-preset", preset)

	ret = parseCustomOptions(ret, additional)

	return
}
