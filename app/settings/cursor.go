package settings

import (
	color2 "github.com/wieku/danser-go/framework/math/color"
)

var Cursor = initCursor()

func initCursor() *cursor {
	return &cursor{
		TrailStyle:   1,
		Style23Speed: 0.18,
		Style4Shift:  0.5,
		Colors: &color{
			EnableRainbow:         true,
			RainbowSpeed:          8,
			BaseColor:             DefaultsFactory.InitHSV(),
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
		CursorSize:                  12,
		CursorExpand:                false,
		ScaleToTheBeat:              false,
		ShowCursorsOnBreaks:         true,
		BounceOnEdges:               false,
		TrailScale:                  1.0,
		TrailEndScale:               0.4,
		TrailDensity:                1,
		TrailMaxLength:              2000,
		TrailRemoveSpeed:            1,
		GlowEndScale:                0.4,
		InnerLengthMult:             0.9,
		AdditiveBlending:            true,
		CursorRipples:               true,
		SmokeEnabled:                true,
	}
}

type cursor struct {
	TrailStyle                  int     `combo:"1|1. Unified color,2|2. Distance-based rainbow,3|3. Time-based rainbow,4|4. Gradient"`
	Style23Speed                float64 `label:"Style 2/3 Speed" scale:"1000" min:"-1" max:"1" format:"%.0f°/(s or 1000px)"`
	Style4Shift                 float64 `label:"Style 4 Hue Shift" scale:"360" min:"-1" max:"1"`
	Colors                      *color  `label:"Color"`
	EnableCustomTagColorOffset  bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used
	TagColorOffset              float64 `min:"-360" max:"360" format:"%.0f°"` //-36, offset of the next tag cursor
	EnableTrailGlow             bool    //true
	EnableCustomTrailGlowOffset bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used (or offset of 180° for single cursor)
	TrailGlowOffset             float64 `min:"-360" max:"360" format:"%.0f°"` //-36, offset of the cursor trail glow
	ScaleToCS                   bool    //false, if enabled, cursor will scale to beatmap CS value
	CursorSize                  float64 `label:"Cursor Size" min:"0.1" max:"50"` //18, cursor radius in osu!pixels
	CursorExpand                bool    //Should cursor be scaled to 1.3 when clicked
	ScaleToTheBeat              bool    //true, cursor size is changing with music peak amplitude
	ShowCursorsOnBreaks         bool    //true
	BounceOnEdges               bool    //false
	TrailScale                  float64 //0.4
	TrailEndScale               float64 //0.4
	TrailDensity                float64 `min:"0.001" max:"3"` //0.5 - 1/TrailDensity = distance between trail points
	TrailMaxLength              int64   `max:"10000"`         //2000 - maximum width (in osu!pixels) of cursortrail
	TrailRemoveSpeed            float64 `max:"5"`             //1.0 - trail removal multiplier, 0.5 means half the speed
	GlowEndScale                float64 //0.4
	InnerLengthMult             float64 //0.9 - if glow is enabled, inner trail will be shortened to 0.9 * length
	AdditiveBlending            bool
	CursorRipples               bool
	SmokeEnabled                bool `label:"Cursor Smoke"`
}

func (cr *cursor) GetColors(divides, cursors int, beatScale, alpha float64) []color2.Color {
	if !cr.EnableCustomTagColorOffset {
		return cr.Colors.GetColors(divides*cursors, beatScale, alpha)
	}

	colors := cr.Colors.GetColors(divides, beatScale, alpha)
	colors1 := make([]color2.Color, divides*cursors)

	for i := 0; i < divides; i++ {
		for j := 0; j < cursors; j++ {
			colors1[i*cursors+j] = colors[i].Shift(float32(j)*float32(cr.TagColorOffset), 0, 0)
		}
	}

	return colors1
}
