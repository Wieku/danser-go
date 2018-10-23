package utils

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lucasb-eyer/go-colorful"
)

func GetColors(baseHue, hueShift float64, times int, alpha float64) []mgl32.Vec4 {
	return GetColorsSV(baseHue, hueShift, times, 1, 1, alpha)
}

func GetColor(H, S, V, alpha float64) mgl32.Vec4 {
	color := colorful.Hsv(H, S, V)
	return mgl32.Vec4{float32(color.R), float32(color.G), float32(color.B), float32(alpha)}
}

func GetColorsSV(baseHue, hueShift float64, times int, S, V, alpha float64) []mgl32.Vec4 {
	colors := make([]mgl32.Vec4, times)

	for baseHue < 0.0 {
		baseHue += 360.0
	}

	for baseHue >= 360.0 {
		baseHue -= 360.0
	}

	for i := 0; i < times; i++ {
		hue := baseHue + float64(i)*hueShift

		for hue < 0.0 {
			hue += 360.0
		}

		for hue >= 360.0 {
			hue -= 360.0
		}

		colors[i] = GetColor(hue, S, V, alpha)
	}

	return colors
}

func GetColorsSVT(baseHue, hueShift, tagShift float64, times, tag int, S, V, alpha float64) []mgl32.Vec4 {
	colors := make([]mgl32.Vec4, 0)

	for baseHue < 0.0 {
		baseHue += 360.0
	}

	for baseHue >= 360.0 {
		baseHue -= 360.0
	}

	for i := 0; i < times; i++ {
		hue := baseHue + float64(i)*hueShift

		for hue < 0.0 {
			hue += 360.0
		}

		for hue >= 360.0 {
			hue -= 360.0
		}

		colors = append(colors, GetColorsSV(hue, tagShift, tag, S, V, alpha)...)
	}

	return colors
}

func GetColorShifted(color mgl32.Vec4, hueOffset float64) mgl32.Vec4 {
	tohsv := colorful.Color{float64(color[0]), float64(color[1]), float64(color[2])}
	h, s, v := tohsv.Hsv()
	h += hueOffset

	for h < 0 {
		h += 360.0
	}

	for h > 360.0 {
		h -= 360.0
	}

	col2 := colorful.Hsv(h, s, v)
	return mgl32.Vec4{float32(col2.R), float32(col2.G), float32(col2.B), color.W()}
}
