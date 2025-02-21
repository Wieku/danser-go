package settings

import (
	"fmt"
	"strconv"
	"strings"
)

type aacSettings struct {
	Bitrate           string `combo:"8k,16k,32k,64k,96k,128k,160k,192k,224k,256k,288k,320k"`
	AdditionalOptions string
}

func (s *aacSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret = append(ret, "-b:a", s.Bitrate)

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return ret, nil
}

type mp3Settings struct {
	RateControl       string `combo:"cbr|CBR,abr|ABR"`
	TargetBitrate     string `combo:"8k,16k,32k,64k,96k,128k,160k,192k,224k,256k,288k,320k"`
	AdditionalOptions string
}

func (s *mp3Settings) GenerateFFmpegArgs() (ret []string, err error) {
	switch strings.ToLower(s.RateControl) {
	case "cbr":
		break
	case "abr":
		ret = append(ret, "-abr", "true")
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", s.RateControl)
	}

	ret = append(ret, "-b:a", s.TargetBitrate)

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return ret, nil
}

type opusSettings struct {
	RateControl       string `combo:"cbr|CBR,vbr|VBR"`
	TargetBitrate     string `combo:"8k,16k,32k,64k,96k,128k,160k,192k,224k,256k,288k,320k"`
	AdditionalOptions string
}

func (s *opusSettings) GenerateFFmpegArgs() (ret []string, err error) {
	switch strings.ToLower(s.RateControl) {
	case "vbr":
		break
	case "cbr":
		ret = append(ret, "-vbr", "off")
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", s.RateControl)
	}

	ret = append(ret, "-b:a", s.TargetBitrate)

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return ret, nil
}

type flacSettings struct {
	CompressionLevel  int `combo:"0|0 (Biggest size),1,2,3,4,5,6,7,8,9,10,11,12|12 (Smallest size)"`
	AdditionalOptions string
}

func (s *flacSettings) GenerateFFmpegArgs() (ret []string, err error) {
	if s.CompressionLevel < 0 || s.CompressionLevel > 12 {
		return nil, fmt.Errorf("CompressionLevel out of range [0-12]")
	}

	ret = append(ret, "-compression_level", strconv.Itoa(s.CompressionLevel))
	ret = append(ret, "-sample_fmt", "s32")
	ret = append(ret, "-bits_per_raw_sample", "24")

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return ret, nil
}
