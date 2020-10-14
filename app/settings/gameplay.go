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
}
