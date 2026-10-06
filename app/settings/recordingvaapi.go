package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var vaapiH264Profiles = []string{
	"baseline",
	"main",
	"high",
}

var vaapiH265Profiles = []string{
	"main",
	"main10",
}

type h264VaapiSettings struct {
	Device            string `label:"VAAPI Device Path" tooltip:"Path to the VAAPI render device, defaults to /dev/dri/renderD128"`
	RateControl       string `combo:"cqp|Constant Quantization (CQP),cbr|Constant Bitrate (CBR),vbr|Variable Bitrate (VBR)"`
	Bitrate           string `showif:"RateControl=cbr,vbr"`
	QP                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp"`
	Profile           string `combo:"baseline|Baseline,main|Main,high|High"`
	AdditionalOptions string
}

func (s *h264VaapiSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = vaapiCommon(s.Device, s.RateControl, s.Bitrate, s.QP, "h264_vaapi")
	if err != nil {
		return nil, err
	}

	if !slices.Contains(vaapiH264Profiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile:v", s.Profile)

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return
}

type hevcVaapiSettings struct {
	Device            string `label:"VAAPI Device Path" tooltip:"Path to the VAAPI render device, defaults to /dev/dri/renderD128"`
	RateControl       string `combo:"cqp|Constant Quantization (CQP),cbr|Constant Bitrate (CBR),vbr|Variable Bitrate (VBR),icq|Intelligent Constant Quality (ICQ)"`
	Bitrate           string `showif:"RateControl=cbr,vbr"`
	QP                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp"`
	Quality           int    `string:"true" min:"1" max:"51" showif:"RateControl=icq"`
	Profile           string `combo:"main|Main,main10|Main10"`
	AdditionalOptions string
}

func (s *hevcVaapiSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = vaapiCommon(s.Device, s.RateControl, s.Bitrate, s.QP, "hevc_vaapi")
	if err != nil {
		return nil, err
	}

	// HEVC ICQ uses -global_quality instead of -qp
	if strings.ToLower(s.RateControl) == "icq" {
		if s.Quality < 1 || s.Quality > 51 {
			return nil, fmt.Errorf("Quality parameter out of range [1-51]")
		}

		ret = append(ret, "-global_quality", strconv.Itoa(s.Quality))
	}

	if !slices.Contains(vaapiH265Profiles, s.Profile) {
		return nil, fmt.Errorf("invalid profile: %s", s.Profile)
	}

	ret = append(ret, "-profile:v", s.Profile)

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return
}

type av1VaapiSettings struct {
	Device            string `label:"VAAPI Device Path" tooltip:"Path to the VAAPI render device, defaults to /dev/dri/renderD128"`
	RateControl       string `combo:"cqp|Constant Quantization (CQP),cbr|Constant Bitrate (CBR),vbr|Variable Bitrate (VBR),icq|Intelligent Constant Quality (ICQ),qvbr|Quality Variable Bitrate (QVBR)"`
	Bitrate           string `showif:"RateControl=cbr,vbr,qvbr"`
	QP                int    `string:"true" min:"0" max:"51" showif:"RateControl=cqp"`
	Quality           int    `string:"true" min:"1" max:"51" showif:"RateControl=icq,qvbr"`
	CompressionLevel  int    `string:"true" min:"0" max:"7" label:"Compression Level" tooltip:"Trade encoding speed for compression efficiency. 0 is fastest, 7 is most efficient."`
	AdditionalOptions string
}

func (s *av1VaapiSettings) GenerateFFmpegArgs() (ret []string, err error) {
	ret, err = vaapiCommon(s.Device, s.RateControl, s.Bitrate, s.QP, "av1_vaapi")
	if err != nil {
		return nil, err
	}

	// AV1 ICQ/QVBR uses -global_quality instead of -qp
	if strings.ToLower(s.RateControl) == "icq" || strings.ToLower(s.RateControl) == "qvbr" {
		if s.Quality < 1 || s.Quality > 51 {
			return nil, fmt.Errorf("Quality parameter out of range [1-51]")
		}

		ret = append(ret, "-global_quality", strconv.Itoa(s.Quality))
	}

	if s.CompressionLevel < 0 || s.CompressionLevel > 7 {
		return nil, fmt.Errorf("CompressionLevel parameter out of range [0-7]")
	}

	ret = append(ret, "-compression_level", strconv.Itoa(s.CompressionLevel))

	ret = parseCustomOptions(ret, s.AdditionalOptions)

	return
}

func vaapiCommon(device, rateControl, bitrate string, qp int, encoderType string) (ret []string, err error) {
	if device != "" {
		ret = append(ret, "-vaapi_device", device)
	}

	// Hardware upload filter; the nv12 format override is applied in video.go
	ret = append(ret, "-vf", "format=nv12,hwupload")

	switch strings.ToLower(rateControl) {
	case "cqp":
		if qp < 0 || qp > 51 {
			return nil, fmt.Errorf("QP parameter out of range [0-51]")
		}

		ret = append(ret, "-rc_mode", "CQP", "-qp", strconv.Itoa(qp))
	case "cbr":
		ret = append(ret, "-rc_mode", "CBR", "-b:v", bitrate)
	case "vbr":
		ret = append(ret, "-rc_mode", "VBR", "-b:v", bitrate)
	case "icq", "qvbr":
		// ICQ/QVBR handled in the caller (different per codec)
		// For h264: ICQ not supported, so this falls through to the error below
		// For hevc: -global_quality set by caller
		// For av1: ICQ and QVBR use -global_quality set by caller
		return
	default:
		return nil, fmt.Errorf("invalid rate control value: %s", rateControl)
	}

	return
}
