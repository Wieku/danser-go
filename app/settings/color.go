package settings

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/utils"
)

type hsv struct {
	Hue, Saturation, Value float64
}

type color struct {
	EnableRainbow         bool    //true
	RainbowSpeed          float64 //8, degrees per second
	BaseColor             *hsv    //0..360, if EnableRainbow is disabled then this value will be used to calculate base color
	EnableCustomHueOffset bool    //false, false means that every iteration has an offset of i*360/n
	HueOffset             float64 //0, custom hue offset for mirror collages
	FlashToTheBeat        bool    //true, objects size is changing with music peak amplitude
	FlashAmplitude        float64 //50, hue offset for flashes
	currentHue            float64
}

func (cl *color) Update(delta float64) {
	if cl.EnableRainbow {
		cl.currentHue += cl.RainbowSpeed / 1000.0 * delta
		for cl.currentHue >= 360.0 {
			cl.currentHue -= 360.0
		}

		for cl.currentHue < 0.0 {
			cl.currentHue += 360.0
		}
	} else {
		cl.currentHue = 0
	}
}

func (cl *color) GetColors(divides int, beatScale, alpha float64) []mgl32.Vec4 {
	flashOffset := 0.0
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

	return utils.GetColorsSV(hue, offset, divides, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}

func (cl *color) GetColorsH(divides int, beatScale, alpha float64) ([]mgl32.Vec4, []float64) {
	flashOffset := 0.0
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

	return utils.GetColorsSVH(hue, offset, divides, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}
