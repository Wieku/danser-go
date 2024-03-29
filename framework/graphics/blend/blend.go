package blend

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	color2 "github.com/wieku/danser-go/framework/math/color"
)

type data struct {
	enabled          bool
	source           Factor
	destination      Factor
	sourceAlpha      Factor
	destinationAlpha Factor
	equation         Equation
	equationAlpha    Equation
	color            color2.Color
}

var stack []data

var current data

func Enable() {
	if current.enabled {
		return
	}

	current.enabled = true

	gl.Enable(gl.BLEND)
}

func Disable() {
	if !current.enabled {
		return
	}

	current.enabled = false

	gl.Disable(gl.BLEND)
}

func SetFunction(src Factor, dst Factor) {
	SetFunctionSeparate(src, dst, src, dst)
}

func SetFunctionSeparate(src Factor, dst Factor, srcAlpha Factor, dstAlpha Factor) {
	if current.source == src && current.destination == dst && current.sourceAlpha == srcAlpha && current.destinationAlpha == dstAlpha {
		return
	}

	current.source = src
	current.sourceAlpha = srcAlpha
	current.destination = dst
	current.destinationAlpha = dstAlpha

	gl.BlendFuncSeparate(uint32(src), uint32(dst), uint32(srcAlpha), uint32(dstAlpha))
}

func SetEquation(equation Equation) {
	SetEquationSeparate(equation, equation)
}

func SetEquationSeparate(equation Equation, equationAlpha Equation) {
	if current.equation == equation && current.equationAlpha == equationAlpha {
		return
	}

	current.equation = equation
	current.equationAlpha = equationAlpha

	gl.BlendEquationSeparate(uint32(equation), uint32(equationAlpha))
}

func SetColor(color color2.Color) {
	if current.color == color {
		return
	}

	current.color = color

	gl.BlendColor(color.R, color.G, color.B, color.A)
}

func Push() {
	stack = append(stack, current)
}

func Pop() {
	data := stack[len(stack)-1]
	stack = stack[:len(stack)-1]

	if data.enabled {
		Enable()
	} else {
		Disable()
	}

	SetEquationSeparate(data.equation, data.equationAlpha)
	SetColor(data.color)
	SetFunctionSeparate(data.source, data.destination, data.sourceAlpha, data.destinationAlpha)
}

func ClearStack() {
	stack = stack[:0]
}
