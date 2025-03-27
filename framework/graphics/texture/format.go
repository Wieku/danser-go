package texture

import "github.com/go-gl/gl/v3.3-core/gl"

type Format int

const (
	Red = Format(iota)
	Depth
	RG
	RGB
	BGR
	RGB32F
	RGBA
	BGRA
	RGBA32F
)

func (f Format) InternalFormat() uint32 {
	switch f {
	case Red:
		return gl.R8
	case Depth:
		return gl.DEPTH_COMPONENT32F
	case RG:
		return gl.RG8
	case RGB:
		return gl.RGB8
	case BGR:
		return gl.BGR
	case RGB32F:
		return gl.RGB32F
	case RGBA:
		return gl.RGBA8
	case BGRA:
		return gl.BGRA
	case RGBA32F:
		return gl.RGBA32F
	}

	panic("Wrong texture format!")
}

func (f Format) Format() uint32 {
	switch f {
	case Red:
		return gl.RED
	case Depth:
		return gl.DEPTH_COMPONENT
	case RG:
		return gl.RG
	case RGB, RGB32F:
		return gl.RGB
	case BGR:
		return gl.BGR
	case RGBA, RGBA32F:
		return gl.RGBA
	case BGRA:
		return gl.BGRA
	}

	panic("Wrong texture format!")
}

func (f Format) Size() int {
	switch f {
	case Red, Depth:
		return 1
	case RG:
		return 2
	case RGB, BGR, RGB32F:
		return 3
	case RGBA, BGRA, RGBA32F:
		return 4
	}

	panic("Wrong texture format!")
}

func (f Format) Type() uint32 {
	switch f {
	case Red, RG, RGB, BGR, RGBA, BGRA:
		return gl.UNSIGNED_BYTE
	case Depth, RGB32F, RGBA32F:
		return gl.FLOAT
	}

	panic("Wrong texture format!")
}
