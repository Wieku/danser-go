package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		ShowHitErrorMeter:    true,
		HitErrorMeterOpacity: 1,
		HitErrorMeterScale:   1,
		ScoreScale:           1,
		ScoreOpacity:         1,
		ComboScale:           1,
		ComboOpacity:         1,
		KeyOverlayScale:      1,
		KeyOverlayOpacity:    1,
	}
}

type gameplay struct {
	ShowHitErrorMeter    bool
	HitErrorMeterOpacity float64
	HitErrorMeterScale   float64

	ScoreScale   float64
	ScoreOpacity float64

	ComboScale   float64
	ComboOpacity float64

	KeyOverlayScale   float64
	KeyOverlayOpacity float64
}
