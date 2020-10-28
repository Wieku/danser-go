package utils

import (
	color2 "github.com/wieku/danser-go/framework/math/color"
	"math"
)

func GetColorsSV(baseHue, hueShift float64, times int, S, V, alpha float64) []color2.Color {
	colors := make([]color2.Color, times)

	baseHue = math.Mod(baseHue, 360)
	if baseHue < 0.0 {
		baseHue += 360.0
	}

	for i := 0; i < times; i++ {
		hue := baseHue + float64(i)*hueShift

		hue = math.Mod(hue, 360)
		if hue < 0.0 {
			hue += 360.0
		}

		colors[i] = color2.NewHSVA(float32(hue), float32(S), float32(V), float32(alpha))
	}

	return colors
}
