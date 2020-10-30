package settings

var Input = initInput()

func initInput() *input {
	return &input{
		LeftKey:              "Z",
		RightKey:             "X",
		MouseButtonsDisabled: true,
	}
}

type input struct {
	LeftKey              string
	RightKey             string
	MouseButtonsDisabled bool
}
