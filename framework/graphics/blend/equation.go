package blend

import "github.com/go-gl/gl/v3.3-core/gl"

type Equation uint32

const (
	Add             = Equation(gl.FUNC_ADD)
	Subtract        = Equation(gl.FUNC_SUBTRACT)
	ReverseSubtract = Equation(gl.FUNC_REVERSE_SUBTRACT)
	Min             = Equation(gl.MIN)
	Max             = Equation(gl.MAX)
)
