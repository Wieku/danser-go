package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		ShowHitErrorMeter:    true,
		HitErrorMeterOpacity: 1,
		HitErrorMeterScale:   1,
		ShowScore:            true,
		ScoreScale:           1,
		ScoreOpacity:         1,
		ShowCombo:            true,
		ComboScale:           1,
		ComboOpacity:         1,
		ShowKeyOverlay:       true,
		KeyOverlayScale:      1,
		KeyOverlayOpacity:    1,
		ProgressBar:          "Pie",
		Boundaries: &boundaries{
			Enabled:         true,
			BorderThickness: 1,
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
	ShowHitErrorMeter    bool
	HitErrorMeterOpacity float64
	HitErrorMeterScale   float64

	ShowScore    bool
	ScoreScale   float64
	ScoreOpacity float64

	ShowCombo    bool
	ComboScale   float64
	ComboOpacity float64

	ShowKeyOverlay    bool
	KeyOverlayScale   float64
	KeyOverlayOpacity float64

	ProgressBar string

	Boundaries *boundaries
}

type boundaries struct {
	Enabled bool

	BorderThickness float64

	BorderColor   *hsv
	BorderOpacity float64

	BackgroundColor   *hsv
	BackgroundOpacity float64
}
