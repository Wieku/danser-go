package color

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

func HSVToRGB(h, s, v float32) (r, g, b float32) {
	h = math32.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	s = mutils.Clamp(s, 0, 1)
	v = mutils.Clamp(v, 0, 1)

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
	r = mutils.Clamp(r, 0, 1)
	g = mutils.Clamp(g, 0, 1)
	b = mutils.Clamp(b, 0, 1)

	minV := min(min(r, g), b)
	v = max(max(r, g), b)
	c := v - minV

	s = 0.0
	if v != 0.0 {
		s = c / v
	}

	h = 0.0

	if minV != v {
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

func PackInt(r, g, b, a float32) uint32 {
	rI := uint32(r * 255)
	gI := uint32(g * 255)
	bI := uint32(b * 255)
	aI := uint32(a * 255)

	return aI<<24 | bI<<16 | gI<<8 | rI
}

func PackFloat(r, g, b, a float32) float32 {
	return math.Float32frombits(PackInt(r, g, b, a))
}
