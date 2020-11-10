package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		HitErrorMeter: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		Score: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		ComboCounter: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		KeyOverlay: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		ProgressBar: "Pie",
		Boundaries: &boundaries{
			Enabled:         true,
			BorderThickness: 1,
			BorderFill:      1,
			BorderColor: &hsv{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			BorderOpacity: 1,
			BackgroundColor: &hsv{
				Hue:        0,
				Saturation: 1,
				Value:      0,
			},
			BackgroundOpacity: 0.5,
		},
	}
}

type gameplay struct {
	HitErrorMeter *hudElement
	Score         *hudElement
	ComboCounter  *hudElement
	KeyOverlay    *hudElement

	ProgressBar string

	Boundaries *boundaries
}

type boundaries struct {
	Enabled bool

	BorderThickness float64
	BorderFill      float64

	BorderColor   *hsv
	BorderOpacity float64

	BackgroundColor   *hsv
	BackgroundOpacity float64
}

type hudElement struct {
	Show    bool
	Scale   float64
	Opacity float64
}
