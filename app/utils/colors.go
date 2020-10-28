package utils

import (
	color2 "github.com/wieku/danser-go/framework/math/color"
)

func GetColors(baseHue, hueShift float64, times int, alpha float64) []color2.Color {
	return GetColorsSV(baseHue, hueShift, times, 1, 1, alpha)
}

func GetColorsH(baseHue, hueShift float64, times int, alpha float64) ([]color2.Color, []float64) {
	return GetColorsSVH(baseHue, hueShift, times, 1, 1, alpha)
}

func GetColor(H, S, V, alpha float64) color2.Color {
	return color2.NewHSVA(float32(H), float32(S), float32(V), float32(alpha))
}

func GetColorsSV(baseHue, hueShift float64, times int, S, V, alpha float64) []color2.Color {
	colors := make([]color2.Color, times)

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

func GetColorsSVH(baseHue, hueShift float64, times int, S, V, alpha float64) ([]color2.Color, []float64) {
	colors := make([]color2.Color, times)
	shifts := make([]float64, times)

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
		shifts[i] = hue
	}

	return colors, shifts
}

func GetColorsSVHA(baseHue, hueShift float64, times int, S, V, alpha float64) ([]color2.Color, []float64) {
	colors := make([]color2.Color, times)
	shifts := make([]float64, times)

	for baseHue < 0.0 {
		baseHue += 360.0
	}

	for baseHue >= 360.0 {
		baseHue -= 360.0
	}

	for i := 0; i < times; i++ {
		hue := baseHue

		if i == 1 {
			hue += 180
		} else {
			hue += float64(i) * hueShift
		}

		for hue < 0.0 {
			hue += 360.0
		}

		for hue >= 360.0 {
			hue -= 360.0
		}

		colors[i] = GetColor(hue, S, V, alpha)
		shifts[i] = hue
	}

	return colors, shifts
}

func GetColorsSVT(baseHue, hueShift, tagShift float64, times, tag int, S, V, alpha float64) ([]color2.Color, []float64) {
	colors := make([]color2.Color, 0)
	shifts := make([]float64, 0)

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

		c, h := GetColorsSVH(hue, tagShift, tag, S, V, alpha)

		colors = append(colors, c...)
		shifts = append(shifts, h...)
	}

	return colors, shifts
}

func GetColorsSVTA(baseHue, hueShift, tagShift float64, times, tag int, S, V, alpha float64) ([]color2.Color, []float64) {
	colors := make([]color2.Color, 0)
	shifts := make([]float64, 0)

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

		c, h := GetColorsSVHA(hue, tagShift, tag, S, V, alpha)

		colors = append(colors, c...)
		shifts = append(shifts, h...)
	}

	return colors, shifts
}
