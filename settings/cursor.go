package settings

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/utils"
)

var Cursor = initCursor()

func initCursor() *cursor {
	return &cursor{
		TrailStyle:   1,
		Style23Speed: 0.18,
		Style4Shift:  0.5,
		Colors: &color{
			EnableRainbow: true,
			RainbowSpeed:  8,
			BaseColor: &hsv{
				0,
				1.0,
				1.0},
			EnableCustomHueOffset: false,
			HueOffset:             0,
			FlashToTheBeat:        false,
			FlashAmplitude:        0,
			currentHue:            0,
		},
		EnableCustomTagColorOffset:  true,
		TagColorOffset:              -36,
		EnableTrailGlow:             true,
		EnableCustomTrailGlowOffset: true,
		TrailGlowOffset:             -36.0,
		ScaleToCS:                   false,
		CursorSize:                  18,
		ScaleToTheBeat:              true,
		ShowCursorsOnBreaks:         true,
		BounceOnEdges:               false,
		TrailEndScale:               0.4,
		TrailDensity:                0.5,
		TrailMaxLength:              2000,
		TrailRemoveSpeed:            1,
		GlowEndScale:                0.4,
		InnerLengthMult:             0.9,
		AdditiveBlending:            true,
	}
}

type cursor struct {
	TrailStyle                  int
	Style23Speed                float64
	Style4Shift                 float64
	Colors                      *color
	EnableCustomTagColorOffset  bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used
	TagColorOffset              float64 //-36, offset of the next tag cursor
	EnableTrailGlow             bool    //true
	EnableCustomTrailGlowOffset bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used (or offset of 180° for single cursor)
	TrailGlowOffset             float64 //-36, offset of the cursor trail glow
	ScaleToCS                   bool    //false, if enabled, cursor will scale to beatmap CS value
	CursorSize                  float64 //18, cursor radius in osu!pixels
	ScaleToTheBeat              bool    //true, cursor size is changing with music peak amplitude
	ShowCursorsOnBreaks         bool    //true
	BounceOnEdges               bool    //false
	TrailEndScale               float64 //0.4
	TrailDensity                float64 //0.5 - 1/TrailDensity = distance between trail points
	TrailMaxLength              int64   //2000 - maximum width (in osu!pixels) of cursortrail
	TrailRemoveSpeed            float64 //1.0 - trail removal multiplier, 0.5 means half the speed
	GlowEndScale                float64 //0.4
	InnerLengthMult             float64 //0.9 - if glow is enabled, inner trail will be shortened to 0.9 * length
	AdditiveBlending            bool
}

func (cr *cursor) GetColors(divides, tag int, beatScale, alpha float64) ([]mgl32.Vec4, []float64) {
	if !cr.EnableCustomTagColorOffset {
		return cr.Colors.GetColorsH(divides*tag, beatScale, alpha)
	}
	flashOffset := 0.0
	cl := cr.Colors
	if cl.FlashToTheBeat {
		flashOffset = cl.FlashAmplitude * (beatScale - 1.0) / (0.4 * Beat.BeatScale)
	}
	hue := cl.BaseColor.Hue + cl.currentHue + flashOffset

	for hue >= 360.0 {
		hue -= 360.0
	}

	for hue < 0.0 {
		hue += 360.0
	}

	offset := 360.0 / float64(divides)

	if cl.EnableCustomHueOffset {
		offset = cl.HueOffset
	}

	return utils.GetColorsSVT(hue, offset, cr.TagColorOffset, divides, tag, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}

func (cr *cursor) GetColorsA(divides, tag int, beatScale, alpha float64) ([]mgl32.Vec4, []float64) {
	if !cr.EnableCustomTagColorOffset {
		return cr.Colors.GetColorsH(divides*tag, beatScale, alpha)
	}
	flashOffset := 0.0
	cl := cr.Colors
	if cl.FlashToTheBeat {
		flashOffset = cl.FlashAmplitude * (beatScale - 1.0) / (0.4 * Beat.BeatScale)
	}
	hue := cl.BaseColor.Hue + cl.currentHue + flashOffset

	for hue >= 360.0 {
		hue -= 360.0
	}

	for hue < 0.0 {
		hue += 360.0
	}

	offset := 360.0 / float64(divides)

	if cl.EnableCustomHueOffset {
		offset = cl.HueOffset
	}

	return utils.GetColorsSVTA(hue, offset, cr.TagColorOffset, divides, tag, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}
