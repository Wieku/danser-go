package api

type Attributes struct {
	// Total Star rating, visible on osu!'s beatmap page
	Total float64

	// Aim stars, needed for Performance Points (aka PP) calculations
	Aim float64

	// Speed stars, needed for Performance Points (aka PP) calculations
	Speed float64

	SpeedNoteCount float64

	AimDifficultStrainCount   float64
	SpeedDifficultStrainCount float64

	// Flashlight stars, needed for Performance Points (aka PP) calculations
	Flashlight float64

	// SliderFactor is a ratio of Aim calculated without sliders to Aim with them
	SliderFactor float64

	ObjectCount int
	Circles     int
	Sliders     int
	Spinners    int
	MaxCombo    int
}
