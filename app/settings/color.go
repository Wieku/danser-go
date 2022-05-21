package settings

import (
	"github.com/wieku/danser-go/app/utils"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"math"
)

type HSV struct {
	Hue, Saturation, Value float64
}

func (d *defaultsFactory) InitHSV() *HSV {
	return &HSV{
		Hue:        0,
		Saturation: 1,
		Value:      1,
	}
}

type color struct {
	EnableRainbow         bool    `label:"Enable Rainbow"`                                                //true
	RainbowSpeed          float64 `label:"Rainbow Speed" min:"-360" max:"360" format:"%.0f°/s"`           //8, degrees per second
	BaseColor             *HSV    `label:"Basic Color" short:"true"`                                      //0..360, if EnableRainbow is disabled then this value will be used to calculate base color
	EnableCustomHueOffset bool    `label:"Enable Custom Hue Offset"`                                      //false, false means that every iteration has an offset of i*360/n
	HueOffset             float64 `min:"-360" max:"360" format:"%.0f°" label:"Mirror Collage Hue Offset"` //0, custom hue offset for mirror collages
	FlashToTheBeat        bool    //true, objects size is changing with music peak amplitude
	FlashAmplitude        float64 `min:"-360" max:"360" format:"%.0f°"` //50, hue offset for flashes
	currentHue            float64
}

func (cl *color) Update(delta float64) {
	if cl.EnableRainbow {
		cl.currentHue += cl.RainbowSpeed / 1000.0 * delta

		cl.currentHue = math.Mod(cl.currentHue, 360)
		if cl.currentHue < 0.0 {
			cl.currentHue += 360.0
		}
	} else {
		cl.currentHue = 0
	}
}

func (cl *color) GetColors(divides int, beatScale, alpha float64) []color2.Color {
	flashOffset := 0.0
	if cl.FlashToTheBeat {
		flashOffset = cl.FlashAmplitude * (beatScale - 1.0) / (Audio.BeatScale - 1)
	}

	hue := math.Mod(cl.BaseColor.Hue+cl.currentHue+flashOffset, 360)
	if hue < 0.0 {
		hue += 360.0
	}

	offset := 360.0 / float64(divides)
	if cl.EnableCustomHueOffset {
		offset = cl.HueOffset
	}

	return utils.GetColorsSV(hue, offset, divides, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}
