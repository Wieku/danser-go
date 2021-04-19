package settings

var Input = initInput()

func initInput() *input {
	return &input{
		LeftKey:              "Z",
		RightKey:             "X",
		MouseButtonsDisabled: true,
		MouseHighPrecision:   false,
		MouseSensitivity:     1,
	}
}

type input struct {
	LeftKey              string
	RightKey             string
	MouseButtonsDisabled bool
	MouseHighPrecision   bool
	MouseSensitivity     float64
}
