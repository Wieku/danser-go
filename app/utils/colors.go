package utils

import (
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
)

func GetColorsSV(baseHue, hueShift float64, times int, S, V, alpha float64) []color2.Color {
	colors := make([]color2.Color, times)

	baseHue = mutils.Sanitize(baseHue, 360)

	for i := 0; i < times; i++ {
		hue := baseHue + float64(i)*hueShift
		hue = mutils.Sanitize(hue, 360)

		colors[i] = color2.NewHSVA(float32(hue), float32(S), float32(V), float32(alpha))
	}

	return colors
}
