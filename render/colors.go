package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lucasb-eyer/go-colorful"
)

func GetColors(baseHue, hueShift float64, times int, alpha float64) []mgl32.Vec4 {
	colors := make([]mgl32.Vec4, times)

	for baseHue < 0.0 {
		baseHue += 360.0
	}

	for baseHue >= 360.0 {
		baseHue -= 360.0
	}

	for i:=0; i < times; i++ {
		hue := baseHue + float64(i)*hueShift
		for hue >= 360.0 {
			hue -= 360.0
		}
		color := colorful.Hsv(hue, 1, 1)
		colors[i] = mgl32.Vec4{float32(color.R), float32(color.G), float32(color.B), float32(alpha)}
	}

	return colors
}