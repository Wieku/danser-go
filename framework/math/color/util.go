package color

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/math32"
)

func HSVToRGB(h, s, v float32) (r, g, b float32) {
	h = math32.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	s = bmath.ClampF32(s, 0, 1)
	v = bmath.ClampF32(v, 0, 1)

	hp := h / 60
	c := v * s
	x := c * (1.0 - math32.Abs(math32.Mod(hp, 2.0)-1.0))

	m := v - c

	switch {
	case 0.0 <= hp && hp < 1.0:
		r = c
		g = x
	case 1.0 <= hp && hp < 2.0:
		r = x
		g = c
	case 2.0 <= hp && hp < 3.0:
		g = c
		b = x
	case 3.0 <= hp && hp < 4.0:
		g = x
		b = c
	case 4.0 <= hp && hp < 5.0:
		r = x
		b = c
	case 5.0 <= hp && hp < 6.0:
		r = c
		b = x
	}

	r += m
	g += m
	b += m

	return
}

func RGBToHSV(r, g, b float32) (h, s, v float32) {
	r = bmath.ClampF32(r, 0, 1)
	g = bmath.ClampF32(g, 0, 1)
	b = bmath.ClampF32(b, 0, 1)

	min := math32.Min(math32.Min(r, g), b)
	v = math32.Max(math32.Max(r, g), b)
	c := v - min

	s = 0.0
	if v != 0.0 {
		s = c / v
	}

	h = 0.0

	if min != v {
		if v == r {
			h = math32.Mod((g-b)/c, 6.0)
		}

		if v == g {
			h = (b-r)/c + 2.0
		}

		if v == b {
			h = (r-g)/c + 4.0
		}

		h *= 60.0
		if h < 0.0 {
			h += 360.0
		}
	}

	return
}
