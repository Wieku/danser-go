package settings

var Input = initInput()

func initInput() *input {
	return &input{
		LeftKey:              "Z",
		RightKey:             "X",
		RestartKey:           "`",
		SmokeKey:             "C",
		MouseButtonsDisabled: true,
		MouseHighPrecision:   false,
		MouseSensitivity:     1,
	}
}

type input struct {
	LeftKey              string
	RightKey             string
	RestartKey           string
	SmokeKey             string
	MouseButtonsDisabled bool
	MouseHighPrecision   bool
	MouseSensitivity     float64 `min:"0.4" max:"6"`
}
