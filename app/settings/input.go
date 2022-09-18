package settings

var Input = initInput()

func initInput() *input {
	return &input{
		LeftKey:              "Z",
		RightKey:             "X",
		RestartKey:           "`",
		SmokeKey:             "C",
		ScreenshotKey:        "F2",
		MouseButtonsDisabled: true,
		MouseHighPrecision:   false,
		MouseSensitivity:     1,
	}
}

type input struct {
	LeftKey              string  `key:"true"`
	RightKey             string  `key:"true"`
	RestartKey           string  `key:"true"`
	SmokeKey             string  `key:"true"`
	ScreenshotKey        string  `key:"true"`
	MouseButtonsDisabled bool    `label:"Disable mouse buttons"`
	MouseHighPrecision   bool    `label:"Mouse raw input"`
	MouseSensitivity     float64 `label:"Raw input sensitivity" min:"0.4" max:"6"`
}
