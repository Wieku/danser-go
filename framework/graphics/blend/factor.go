package blend

import "github.com/go-gl/gl/v3.3-core/gl"

type Factor uint32

const (
	Zero                  = Factor(gl.ZERO)
	One                   = Factor(gl.ONE)
	SrcColor              = Factor(gl.SRC_COLOR)
	OneMinusSrcColor      = Factor(gl.ONE_MINUS_SRC_COLOR)
	DstColor              = Factor(gl.DST_COLOR)
	OneMinusDstColor      = Factor(gl.ONE_MINUS_DST_COLOR)
	SrcAlpha              = Factor(gl.SRC_ALPHA)
	OneMinusSrcAlpha      = Factor(gl.ONE_MINUS_SRC_ALPHA)
	DstAlpha              = Factor(gl.DST_ALPHA)
	OneMinusDstAlpha      = Factor(gl.ONE_MINUS_DST_ALPHA)
	ConstantColor         = Factor(gl.CONSTANT_COLOR)
	OneMinusConstantColor = Factor(gl.ONE_MINUS_CONSTANT_COLOR)
	ConstantAlpha         = Factor(gl.CONSTANT_ALPHA)
	OneMinusConstantAlpha = Factor(gl.ONE_MINUS_CONSTANT_ALPHA)
	SrcAlphaSaturate      = Factor(gl.SRC_ALPHA_SATURATE)
	Src1Color             = Factor(gl.SRC1_COLOR)
	OneMinusSrc1Color     = Factor(gl.ONE_MINUS_SRC1_COLOR)
	Src1Alpha             = Factor(gl.SRC1_ALPHA)
	OneMinusSrc1Alpha     = Factor(gl.ONE_MINUS_SRC1_ALPHA)
)
