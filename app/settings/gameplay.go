package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		ShowHitErrorMeter:  true,
		HitErrorMeterScale: 1,
		ScoreScale:         1,
		ScoreOpacity:       1,
		ComboScale:         1,
		ComboOpacity:       1,
	}
}

type gameplay struct {
	ShowHitErrorMeter  bool
	HitErrorMeterScale float64

	ScoreScale   float64
	ScoreOpacity float64

	ComboScale   float64
	ComboOpacity float64
}
