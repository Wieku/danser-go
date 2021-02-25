package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		HitErrorMeter: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		Score: &score{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			ProgressBar:     "Pie",
			ShowGradeAlways: true,
		},
		HpBar: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		ComboCounter: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		PPCounter: &ppCounter{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			XPosition: 5,
			YPosition: 150,
			Align:     "CentreLeft",
		},
		KeyOverlay: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		ScoreBoard: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
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
	Score         *score
	HpBar         *hudElement
	ComboCounter  *hudElement
	PPCounter     *ppCounter
	KeyOverlay    *hudElement
	ScoreBoard    *hudElement
	Boundaries    *boundaries
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

type score struct {
	*hudElement
	ProgressBar     string
	ShowGradeAlways bool
}

type ppCounter struct {
	*hudElement
	XPosition float64
	YPosition float64
	Align     string
}
